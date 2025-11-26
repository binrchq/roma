package services

import (
	"log"
	"os"
	"strings"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/permissions"
)

// 初始化一些数据库数据
func InitData() {
	initRoles()
	initDefaultSpace()
	initSpaces()
	initSuperUser()
	initApiKey()
	initPassport()
}

func initRoles() {
	opRole := operation.NewRoleOperation()
	if global.CONFIG == nil || len(global.CONFIG.Roles) == 0 {
		return
	}

	for _, roleConfig := range global.CONFIG.Roles {
		if roleConfig == nil {
			continue
		}
		name := strings.TrimSpace(roleConfig.Name)
		if name == "" {
			continue
		}

		desc, err := permissions.BuildRoleDescriptor(roleConfig)
		if err != nil || strings.TrimSpace(desc) == "" {
			if strings.TrimSpace(roleConfig.Desc) == "" {
				log.Printf("skip role %s: %v", name, err)
				continue
			}
			desc = strings.TrimSpace(roleConfig.Desc)
		}

		role := &model.Role{
			Name: name,
			Desc: desc,
		}
		if _, err := opRole.CreateOrUpdate(role); err != nil {
			log.Printf("sync role %s failed: %v", name, err)
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

// initDefaultSpace 初始化默认空间
func initDefaultSpace() {
	opSpace := operation.NewSpaceOperation()

	// 检查默认空间是否已存在
	existing, err := opSpace.GetSpaceByName("default")
	if err == nil && existing != nil {
		log.Printf("default space already exists, skipping")
		return
	}

	// 创建默认空间
	defaultSpace := &model.Space{
		Name:        "default",
		Description: "默认空间，所有资源默认归属此空间",
		IsActive:    true,
		CreatedBy:   0, // 系统创建
	}

	space, err := opSpace.CreateSpace(defaultSpace)
	if err != nil {
		log.Printf("create default space failed: %v", err)
		return
	}

	log.Printf("default space created successfully (ID: %d)", space.ID)
}

func initSpaces() {
	if global.CONFIG == nil || len(global.CONFIG.Spaces) == 0 {
		return
	}

	opSpace := operation.NewSpaceOperation()
	opUser := operation.NewUserOperation()
	opRole := operation.NewRoleOperation()

	for _, spaceConfig := range global.CONFIG.Spaces {
		if spaceConfig == nil || strings.TrimSpace(spaceConfig.Name) == "" {
			continue
		}

		// 跳过 default 空间（已由 initDefaultSpace 创建）
		if strings.ToLower(strings.TrimSpace(spaceConfig.Name)) == "default" {
			continue
		}

		// 检查空间是否已存在
		existing, err := opSpace.GetSpaceByName(spaceConfig.Name)
		if err == nil && existing != nil {
			log.Printf("space %s already exists, skipping", spaceConfig.Name)
			continue
		}

		// 创建空间（需要找到创建者，默认使用第一个 super 用户，通过权限描述符判断）
		createdBy := uint(0)
		users, err := opUser.GetAllUsers()
		if err == nil && len(users) > 0 {
			for _, u := range users {
				roles, err := opUser.GetUserRoles(u.ID)
				if err == nil {
					for _, r := range roles {
						if permissions.IsSuperRole(r) {
							createdBy = u.ID
							break
						}
					}
					if createdBy > 0 {
						break
					}
				}
			}
			if createdBy == 0 {
				createdBy = users[0].ID
			}
		}

		space := &model.Space{
			Name:        spaceConfig.Name,
			Description: spaceConfig.Description,
			IsActive:    true,
			CreatedBy:   createdBy,
		}
		space, err = opSpace.CreateSpace(space)
		if err != nil {
			log.Printf("create space %s failed: %v", spaceConfig.Name, err)
			continue
		}

		// 获取默认角色
		defaultRoleID := uint(0)
		if spaceConfig.DefaultRole != "" {
			role, err := opRole.GetRoleByName(spaceConfig.DefaultRole)
			if err == nil && role != nil {
				defaultRoleID = role.ID
			}
		}

		// 添加空间成员
		for _, username := range spaceConfig.Members {
			user, err := opUser.GetUserByUsername(strings.TrimSpace(username))
			if err != nil {
				log.Printf("user %s not found for space %s", username, spaceConfig.Name)
				continue
			}

			roleID := defaultRoleID
			if roleID == 0 {
				// 如果没有默认角色，尝试使用用户的第一个角色
				userRoles, err := opUser.GetUserRoles(user.ID)
				if err == nil && len(userRoles) > 0 {
					roleID = userRoles[0].ID
				}
			}

			_, err = opSpace.AddSpaceMember(space.ID, user.ID, roleID)
			if err != nil {
				log.Printf("add member %s to space %s failed: %v", username, spaceConfig.Name, err)
			}
		}

		log.Printf("space %s initialized successfully", spaceConfig.Name)
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
