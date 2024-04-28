package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils"
	"github.com/gin-gonic/gin"
)

func AddResource(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var resourceData struct {
		Type string            `json:"type"`
		Data []json.RawMessage `json:"data"` // 使用 json.RawMessage 保存未解码的 JSON 字符串
	}
	if err := c.ShouldBindJSON(&resourceData); err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, err.Error())
		return
	}
	// 检查 role 参数是否为空，如果为空，则设置默认值为 "ops"
	roleName := "ops"
	// if resourceData.Role == "" {
	// 	resourceData.Role = "ops"
	// }
	// 创建资源
	// var resourceMap = map[string]interface{}{
	// 	constants.ResourceTypeLinux:    model.LinuxConfig{},
	// 	constants.ResourceTypeRouter:   model.RouterConfig{},
	// 	constants.ResourceTypeWindows:  model.WindowsConfig{},
	// 	constants.ResourceTypeDocker:   model.DockerConfig{},
	// 	constants.ResourceTypeDatabase: model.DatabaseConfig{},
	// 	constants.ResourceTypeSwitch:   model.SwitchConfig{},
	// }

	// 开启事务
	tx := global.GetDB().Begin()
	if tx.Error != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "服务器错误,数据库异常Q4A")
		return
	}
	// resourceSlice, ok := resourceMap[resourceData.Type]
	// if !ok {
	// 	utilG.Response(utils.ERROR, utils.ERROR, fmt.Sprintf("%s 资源类型不存在", resourceData.Type))
	// 	return
	// }
	var failedCount int // 记录失败的条目数
	var failedMsgs []string

	opRes := operation.NewResourceOperation()
	opRole := operation.NewRoleOperation()
	for id, r := range resourceData.Data {
		//
		// linuxConfog := model.LinuxConfig{}
		// if err := json.Unmarshal(r, &linuxConfog); err != nil {
		// 	fmt.Println("JSON解析失败s:原因.%s; 数据:%s", err.Error(), r)
		// }
		// // 解码 JSON 数据到对应的切片
		// if err := json.Unmarshal(r, &resourceSlice); err != nil {
		// 	errMsg := fmt.Sprintf("JSON解析失败:原因.%s; 数据:%s", err.Error(), r)
		// 	failedMsgs = append(failedMsgs, errMsg)
		// 	log.Println(errMsg) // 记录错误到日志
		// 	failedCount++
		// 	continue // 继续处理下一个数据
		// }

		// 将 r 转换为相应的资源类型并创建资源
		var resModel model.Resource
		// 将 r 转换为相应的资源类型并创建资源
		switch resourceData.Type {
		case constants.ResourceTypeLinux:
			resModel = new(model.LinuxConfig)
		case constants.ResourceTypeRouter:
			resModel = new(model.RouterConfig)
		case constants.ResourceTypeWindows:
			resModel = new(model.WindowsConfig)
		case constants.ResourceTypeDocker:
			resModel = new(model.DockerConfig)
		case constants.ResourceTypeDatabase:
			resModel = new(model.DatabaseConfig)
		case constants.ResourceTypeSwitch:
			resModel = new(model.SwitchConfig)
		default:
			utilG.Response(utils.ERROR, utils.ERROR, "未知的资源类型")
			return
		}
		// 解码 JSON 数据到资源模型
		if err := json.Unmarshal(r, resModel); err != nil {
			errMsg := fmt.Sprintf("JSON解析失败s2:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			continue // 继续处理下一个数据
		}
		// 创建资源
		resModel, err := opRes.CreateResource(resModel, resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("写入数据库失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
		// 绑定资源角色
		// role, err := opRole.GetRoleByName(resourceData.Role)
		role, err := opRole.GetRoleByName(roleName)
		if err != nil {
			errMsg := fmt.Sprintf("资源赋值失败1:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
		err = opRes.CreateResourceAndAssociate(int64(role.ID), resModel.GetID(), resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("资源赋值失败2:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
	}

	if failedCount > 0 {
		utilG.Response(utils.ERROR, utils.ERROR, fmt.Sprintf("%d 个资源创建失败(%s)", failedCount, strings.Join(failedMsgs, ";")))
		tx.Rollback() // 回滚事务
		return
	}
	// 提交事务
	tx.Commit()

	utilG.Response(utils.SUCCESS, utils.SUCCESS, "资源创建成功")
}
