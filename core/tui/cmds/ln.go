package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"bitrec.ai/roma/core/connect"
	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/tui/cmds/itface"
	"bitrec.ai/roma/core/utils"
	"github.com/brckubo/ssh"
	"github.com/rs/zerolog/log"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewLn(nil, ""), Weight: 50})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewLn(nil, ""), Weight: 50})
}

func (cmd *Ln) Name() string {
	return "ln"
}

type Ln struct {
	baseLen int
	flags   *Flags
	target  string
	sess    ssh.Session
}

func NewLn(sess ssh.Session, typo string) *Ln {
	flags := &Flags{}
	flags.AddOption("t", "type", "Resource type", StringOption, typo)
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Ln{baseLen: 2, flags: flags, target: "", sess: sess}
}

func (cmd *Ln) Execute(commands string) (interface{}, error) {
	//看看cmd是否是ln
	if !strings.HasPrefix(commands, "ln") {
		cmd.target = strings.TrimSpace(commands)
	} else {
		argParts := commands[cmd.baseLen:]
		args := strings.Fields(strings.TrimSpace(argParts))
		// Use Parse to handle the arguments and set the options in cmd.flags
		cmd.target = cmd.flags.Parse(args)
	}

	resourceTypes := constants.GetResourceType()
	if cmd.flags.GetOptionValue("type").(string) == "~" {
		cmd.flags.SetOptionValue("type", resourceTypes[0])
	}
	if !sliceContains(resourceTypes, cmd.flags.GetOptionValue("type").(string)) {
		return nil, errors.New("invalid resource type,please ln -h to get itfacece")
	}

	return cmd.handle()
}

// 处理 Linux 资源类型的逻辑
func (cmd *Ln) handle() (interface{}, error) {
	roles, err := operation.NewUserOperation().GetUserRolesByUsername(cmd.sess.User())
	if err != nil {
		log.Error().Err(err).Msg("unable to get user roles")
		return nil, err
	}
	searchType, resA := utils.DetermineSearchType(cmd.target)
	var resListA []model.Resource
	op := operation.NewResourceOperation()
	log.Info().Msg("roles:")
	for _, role := range roles {
		resList, _ := op.GetResourceListByRoleId(role.ID, cmd.flags.GetOptionValue("type").(string))
		for _, res := range resList {
			log.Info().Msgf("-------------------------%v", res.GetName())
			log.Info().Msgf("searchType: %v", searchType)
			log.Info().Msgf("resA: %v", resA)
			if matchResource(res, searchType, resA) {
				resListA = append(resListA, res)
			}
		}
	}
	if len(resListA) == 0 {
		return nil, errors.New("resource not found")
	}
	if len(resListA) > 1 {
		return NewLs(cmd.sess, "").Resources(resListA), nil
	}
	Res := resListA[0]
	log.Info().Msgf("connecting to %v", Res.GetName())
	err = connect.NewConnectionLoop(&cmd.sess, Res, cmd.flags.GetOptionValue("type").(string))
	if err != nil {
		return nil, err
	}
	return "连接成功", nil
}

// 获取字段值的通用函数
func getFieldValue(res model.Resource, fieldName string) (string, bool) {
	val := reflect.ValueOf(res).Elem()
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return "", false
	}
	switch field.Kind() {
	case reflect.String:
		return field.String(), true
	case reflect.Int:
		return strconv.Itoa(int(field.Int())), true
	default:
		return "", false
	}
}

// 匹配资源
func matchResource(res model.Resource, searchType, resA string) bool {
	fieldMappings := map[string][]string{
		utils.DETERMINE_IP:          {"IPv4Priv", "IPv4Pub", "IPv6"},
		utils.DETERMINE_DOMAIN:      {"IPv4Pub", "IPv6"},
		utils.DETERMINE_DOMAIN_PORT: {"IPv4Pub", "IPv6"},
		utils.DETERMINE_IP_PORT:     {"IPv4Pub", "IPv4Priv", "IPv6"},
		utils.DETERMINE_HOSTNAME:    {"GetName"},
	}

	fields, exists := fieldMappings[searchType]
	log.Info().Msgf("fields: %v", fields)
	if !exists {
		return false
	}

	if searchType == utils.DETERMINE_HOSTNAME && strings.Contains(res.GetName(), resA) {
		return true
	}
	for _, field := range fields {
		fieldValue, ok := getFieldValue(res, field)
		if !ok {
			continue
		}
		if searchType == utils.DETERMINE_DOMAIN_PORT || searchType == utils.DETERMINE_IP_PORT {
			parts := strings.Split(resA, ":")
			if len(parts) != 2 {
				return false
			}
			port, _ := getFieldValue(res, "Port")
			portActual, _ := getFieldValue(res, "PortActual")
			portIPv6, _ := getFieldValue(res, "PortIPv6")

			if fieldValue == parts[0] && (port == parts[1] || portActual == parts[1] || portIPv6 == parts[1]) {
				return true
			}
		} else if fieldValue == resA {
			return true
		}
	}

	return false
}

// Help 返回 ln 命令的帮助信息
func (cmd *Ln) Usage() string {
	resourceTypes := constants.GetResourceType()
	usageMsg := cmd.flags.FormatUsagef("🍂 %s", green(cmd.Name()+" [-t TYPE] RESOURCE or RESOURCE"))
	usageMsg += cmd.flags.FormatUsagef("Login the specified TYPE of resource,TYPE is %s;RESOURCE for ls Query, etc.", cyan(strings.Join(resourceTypes, ", ")))
	usageMsg += cmd.flags.FormatUsagef("Usage:")
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	// 写入Options
	log.Info().Msgf("flags: %v", cmd.flags.Options)
	tw = cmd.flags.ColorUsage(tw)
	fmt.Fprint(tw) // 换行
	tw.Flush()
	return usageMsg + buffer.String()
}
