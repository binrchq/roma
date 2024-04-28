package operation

import (
	"errors"
	"fmt"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/gorm"
)

type ResourceOperation struct {
	DB *gorm.DB
}

func NewResourceOperation() *ResourceOperation {
	return &ResourceOperation{DB: global.GetDB()}
}

func NewResourceOperationWithDebug() *ResourceOperation {
	return &ResourceOperation{DB: global.GetDB().Debug()}
}

func NewResourceOperationWithDB(db *gorm.DB) *ResourceOperation {
	return &ResourceOperation{DB: db}
}

// 创建资源并与角色关联
func (r *ResourceOperation) CreateResourceAndAssociate(roleId int64, resourceId int64, resourceType string) error {
	// 将资源与角色关联
	resourceRole := &model.ResourceRole{
		ResourceID:   resourceId,
		ResourceType: resourceType,
		RoleID:       roleId,
	}
	// 保存关联到数据库
	_, err := r.CreateResourceRole(resourceRole)
	if err != nil {
		return err
	}
	return nil
}

func (r *ResourceOperation) CreateResourceRole(resource_role *model.ResourceRole) (*model.ResourceRole, error) {
	if err := r.DB.Create(resource_role).Error; err != nil {
		return nil, err
	}
	return resource_role, nil
}

func (r *ResourceOperation) CreateLinuxResource(resource *model.LinuxConfig) (*model.LinuxConfig, error) {
	if err := r.DB.Where(model.LinuxConfig{Hostname: resource.Hostname}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("hostname already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateWindowsResource(resource *model.WindowsConfig) (*model.WindowsConfig, error) {
	if err := r.DB.Where(model.WindowsConfig{Hostname: resource.Hostname}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("hostname already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateDatabaseResource(resource *model.DatabaseConfig) (*model.DatabaseConfig, error) {
	if err := r.DB.Where(model.DatabaseConfig{DatabaseNick: resource.DatabaseNick}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("database nick name already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateRouterResource(resource *model.RouterConfig) (*model.RouterConfig, error) {
	if err := r.DB.Where(model.RouterConfig{RouterName: resource.RouterName}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("router name already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateSwitchResource(resource *model.SwitchConfig) (*model.SwitchConfig, error) {
	if err := r.DB.Where(model.SwitchConfig{SwitchName: resource.SwitchName}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("switch name already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateDockerResource(resource *model.DockerConfig) (*model.DockerConfig, error) {
	if err := r.DB.Where(model.DockerConfig{ContainerName: resource.ContainerName}).FirstOrCreate(resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("docker container name already exists: %w", err)
		}
		return nil, err
	}
	return resource, nil
}

func (r *ResourceOperation) CreateResource(resource model.Resource, resourceType string) (model.Resource, error) {
	switch resourceType {
	case constants.ResourceTypeLinux:
		return r.CreateLinuxResource(resource.(*model.LinuxConfig))
	case constants.ResourceTypeRouter:
		return r.CreateRouterResource(resource.(*model.RouterConfig))
	case constants.ResourceTypeWindows:
		return r.CreateWindowsResource(resource.(*model.WindowsConfig))
	case constants.ResourceTypeDocker:
		return r.CreateDockerResource(resource.(*model.DockerConfig))
	case constants.ResourceTypeDatabase:
		return r.CreateDatabaseResource(resource.(*model.DatabaseConfig))
	case constants.ResourceTypeSwitch:
		return r.CreateSwitchResource(resource.(*model.SwitchConfig))
	default:
		return nil, errors.New("unknown resource type:" + resourceType)
	}
}

// GetResourceListByRoleId 根据角色ID和资源类型获取资源列表
func (r *ResourceOperation) GetResourceListByRoleId(roleId uint, resourceType string) ([]model.Resource, error) {
	var resourceList []model.Resource

	// 根据资源类型查询对应的资源配置
	fmt.Println("roleId:", roleId, "resourceType:", resourceType)
	switch resourceType {
	case constants.ResourceTypeLinux:
		var linuxConfigs []*model.LinuxConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var linuxConfig model.LinuxConfig
			err := r.DB.Model(&model.LinuxConfig{}).Where("id = ?", res.ResourceID).Find(&linuxConfig).Error
			if err != nil {
				return nil, err
			}
			linuxConfigs = append(linuxConfigs, &linuxConfig)
		}
		fmt.Println(linuxConfigs)
		// 将具体类型的资源配置转换为 model.Resource 接口类型，并添加到 resourceList 中
		for _, cfg := range linuxConfigs {
			fmt.Println(cfg)
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeDatabase:
		var databaseConfigs []*model.DatabaseConfig
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&databaseConfigs).Error
		if err != nil {
			return nil, err
		}
		fmt.Println(databaseConfigs)
		// 将具体类型的资源配置转换为 model.Resource 接口类型，并添加到 resourceList 中
		for _, cfg := range databaseConfigs {
			resourceList = append(resourceList, cfg)
		}

	case constants.ResourceTypeWindows:
		var windowsConfigs []*model.WindowsConfig
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&windowsConfigs).Error
		if err != nil {
			return nil, err
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range windowsConfigs {
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeRouter:
		var routerConfigs []*model.RouterConfig
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&routerConfigs).Error
		if err != nil {
			return nil, err
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range routerConfigs {
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeDocker:
		var dockerConfigs []*model.DockerConfig
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&dockerConfigs).Error
		if err != nil {
			return nil, err
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range dockerConfigs {
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeSwitch:
		var switchConfigs []*model.SwitchConfig
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&switchConfigs).Error
		if err != nil {
			return nil, err
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range switchConfigs {
			resourceList = append(resourceList, cfg)
		}
	default:
		return nil, errors.New("unknown resource type")
	}
	return resourceList, nil
}
