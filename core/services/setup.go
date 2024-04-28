package services

import (
	"log"
	"os"
	"strings"

	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
)

// 初始化一些数据库数据
func InitData() {
	initRoles()
	initSuperUser()
	initApiKey()
	initPassport()
}

func initRoles() {
	opRole := operation.NewRoleOperation()
	roles, _ := opRole.GetAllRoles()
	if len(roles) == 0 {

		for _, roleConfig := range global.CONFIG.Roles {
			role := model.Role{
				Name: roleConfig.Name,
				Desc: roleConfig.Desc,
			}
			opRole.Create(&role)
		}
	}
}

func initSuperUser() {
	op := operation.NewUserOperationWithDebug()
	users, _ := op.GetAllUsers()
	log.Println(users)
	if len(users) == 0 {
		//交互式添加初始的超级管理员
		user := &model.User{
			Username:  global.CONFIG.User1st.Username,
			Name:      global.CONFIG.User1st.Name,
			Nickname:  global.CONFIG.User1st.Nickname,
			Password:  global.CONFIG.User1st.Password,
			PublicKey: global.CONFIG.User1st.PublicKey,
			Email:     global.CONFIG.User1st.Email,
		}
		user, err := op.CreateUser(user)
		if err != nil {
			log.Println(err)
		}
		//添加超级管理员角色
		opRole := operation.NewRoleOperation()
		//获取角色的ID
		roles := strings.Split(global.CONFIG.User1st.Roles, ",")
		for _, roleName := range roles {
			role, err := opRole.GetRoleByName(roleName)
			if err != nil {
				log.Printf("初始化用户角色表没发现%s", roleName)
				os.Exit(0)
			}
			err = op.AddRoleToUser(user.ID, role.ID)
			if err != nil {
				log.Println(err)
			}
		}

	}
}

func initApiKey() {
	op := operation.NewApikeyOperation()
	keys, _ := op.GetAllApiKeys()
	if len(keys) == 0 {
		key := &model.Apikey{
			Apikey:      global.CONFIG.ApiKey.Prefix + global.CONFIG.ApiKey.Key,
			Description: "default apikey",
		}
		op.Create(key)
	}
}

func initPassport() {
	op := operation.NewPassportOperation()
	passports, _ := op.GetPassports()
	if len(passports) == 0 {
		passport := &model.Passport{
			ServiceUser:  global.CONFIG.ControlPassport.ServiceUser,
			ResourceType: global.CONFIG.ControlPassport.ResourceType,
			PassportPub:  global.CONFIG.ControlPassport.PassportPub,
			Passport:     global.CONFIG.ControlPassport.Passport,
			Description:  global.CONFIG.ControlPassport.Description,
		}
		op.CreatePublicPassport(passport)
	}
}
