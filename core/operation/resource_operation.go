package operation

import (
	"errors"
	"fmt"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"github.com/rs/zerolog/log"
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
func (r *ResourceOperation) UpdateLinuxResource(resource *model.LinuxConfig) (*model.LinuxConfig, error) {
	// Find the existing resource by its Hostname
	existingResource := &model.LinuxConfig{}
	if err := r.DB.Where("hostname = ?", resource.Hostname).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("hostname not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}

func (r *ResourceOperation) UpdateWindowsResource(resource *model.WindowsConfig) (*model.WindowsConfig, error) {
	// Find the existing resource by its Hostname
	existingResource := &model.WindowsConfig{}
	if err := r.DB.Where("hostname = ?", resource.Hostname).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("hostname not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}

func (r *ResourceOperation) UpdateDatabaseResource(resource *model.DatabaseConfig) (*model.DatabaseConfig, error) {
	// Find the existing resource by its DatabaseNick
	existingResource := &model.DatabaseConfig{}
	if err := r.DB.Where("database_nick = ?", resource.DatabaseNick).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("database nick name not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}

func (r *ResourceOperation) UpdateRouterResource(resource *model.RouterConfig) (*model.RouterConfig, error) {
	// Find the existing resource by its RouterName
	existingResource := &model.RouterConfig{}
	if err := r.DB.Where("router_name = ?", resource.RouterName).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("router name not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}
func (r *ResourceOperation) UpdateSwitchResource(resource *model.SwitchConfig) (*model.SwitchConfig, error) {
	// Find the existing resource by its SwitchName
	existingResource := &model.SwitchConfig{}
	if err := r.DB.Where("switch_name = ?", resource.SwitchName).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("switch name not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}

func (r *ResourceOperation) UpdateDockerResource(resource *model.DockerConfig) (*model.DockerConfig, error) {
	// Find the existing resource by its ContainerName
	existingResource := &model.DockerConfig{}
	if err := r.DB.Where("container_name = ?", resource.ContainerName).First(existingResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("docker container name not found: %w", err)
		}
		return nil, err
	}

	// Update the resource with the new data
	if err := r.DB.Model(existingResource).Updates(resource).Error; err != nil {
		return nil, err
	}

	return existingResource, nil
}

func (r *ResourceOperation) UpdateResource(resource model.Resource, resourceType string) (model.Resource, error) {
	switch resourceType {
	case constants.ResourceTypeLinux:
		return r.UpdateLinuxResource(resource.(*model.LinuxConfig))
	case constants.ResourceTypeRouter:
		return r.UpdateRouterResource(resource.(*model.RouterConfig))
	case constants.ResourceTypeWindows:
		return r.UpdateWindowsResource(resource.(*model.WindowsConfig))
	case constants.ResourceTypeDocker:
		return r.UpdateDockerResource(resource.(*model.DockerConfig))
	case constants.ResourceTypeDatabase:
		return r.UpdateDatabaseResource(resource.(*model.DatabaseConfig))
	case constants.ResourceTypeSwitch:
		return r.UpdateSwitchResource(resource.(*model.SwitchConfig))
	default:
		return nil, errors.New("unknown resource type: " + resourceType)
	}
}

func (r *ResourceOperation) DeleteLinuxResource(identifier string) error {
	if err := r.DB.Where("hostname = ?", identifier).Or("id = ?", identifier).Delete(&model.LinuxConfig{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *ResourceOperation) DeleteRouterResource(identifier string) error {
	if err := r.DB.Where("router_name = ?", identifier).Or("id = ?", identifier).Delete(&model.RouterConfig{}).Error; err != nil {
		return err
	}
	return nil
}
func (r *ResourceOperation) DeleteWindowsResource(identifier string) error {
	if err := r.DB.Where("hostname = ?", identifier).Or("id = ?", identifier).Delete(&model.WindowsConfig{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *ResourceOperation) DeleteDockerResource(identifier string) error {
	if err := r.DB.Where("container_name = ?", identifier).Or("id = ?", identifier).Delete(&model.DockerConfig{}).Error; err != nil {
		return err
	}
	return nil
}
func (r *ResourceOperation) DeleteDatabaseResource(identifier string) error {
	if err := r.DB.Where("database_nick = ?", identifier).Or("id = ?", identifier).Delete(&model.DatabaseConfig{}).Error; err != nil {
		return err
	}
	return nil
}
func (r *ResourceOperation) DeleteSwitchResource(identifier string) error {
	if err := r.DB.Where("switch_name = ?", identifier).Or("id = ?", identifier).Delete(&model.SwitchConfig{}).Error; err != nil {
		return err
	}
	return nil
}
func (r *ResourceOperation) DeleteResource(identifier string, resourceType string) error {
	switch resourceType {
	case constants.ResourceTypeLinux:
		return r.DeleteLinuxResource(identifier)
	case constants.ResourceTypeRouter:
		return r.DeleteRouterResource(identifier)
	case constants.ResourceTypeWindows:
		return r.DeleteWindowsResource(identifier)
	case constants.ResourceTypeDocker:
		return r.DeleteDockerResource(identifier)
	case constants.ResourceTypeDatabase:
		return r.DeleteDatabaseResource(identifier)
	case constants.ResourceTypeSwitch:
		return r.DeleteSwitchResource(identifier)
	default:
		return errors.New("unknown resource type: " + resourceType)
	}
}

// GetResourceListByRoleId 根据角色ID和资源类型获取资源列表
func (r *ResourceOperation) GetResourceListByRoleId(roleId uint, resourceType string) ([]model.Resource, error) {
	var resourceList []model.Resource

	// 根据资源类型查询对应的资源配置
	log.Info().Msgf("roleId: %d, resourceType: %s", roleId, resourceType)
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
			err := r.DB.Model(&model.LinuxConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&linuxConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if linuxConfig.ID == 0 {
				continue
			}
			linuxConfigs = append(linuxConfigs, &linuxConfig)
		}
		log.Info().Msgf("linuxConfigs: %v", linuxConfigs)
		// 将具体类型的资源配置转换为 model.Resource 接口类型，并添加到 resourceList 中
		for _, cfg := range linuxConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			log.Info().Msgf("cfg: %v", cfg)
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeDatabase:
		var databaseConfigs []*model.DatabaseConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var databaseConfig model.DatabaseConfig
			err := r.DB.Model(&model.DatabaseConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&databaseConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if databaseConfig.ID == 0 {
				continue
			}
			databaseConfigs = append(databaseConfigs, &databaseConfig)
		}
		log.Info().Msgf("databaseConfigs: %v", databaseConfigs)
		// 将具体类型的资源配置转换为 model.Resource 接口类型，并添加到 resourceList 中
		for _, cfg := range databaseConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			resourceList = append(resourceList, cfg)
		}

	case constants.ResourceTypeWindows:
		var windowsConfigs []*model.WindowsConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var windowsConfig model.WindowsConfig
			err := r.DB.Model(&model.WindowsConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&windowsConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if windowsConfig.ID == 0 {
				continue
			}
			windowsConfigs = append(windowsConfigs, &windowsConfig)
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range windowsConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeRouter:
		var routerConfigs []*model.RouterConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var routerConfig model.RouterConfig
			err := r.DB.Model(&model.RouterConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&routerConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if routerConfig.ID == 0 {
				continue
			}
			routerConfigs = append(routerConfigs, &routerConfig)
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range routerConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeDocker:
		var dockerConfigs []*model.DockerConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var dockerConfig model.DockerConfig
			err := r.DB.Model(&model.DockerConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&dockerConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if dockerConfig.ID == 0 {
				continue
			}
			dockerConfigs = append(dockerConfigs, &dockerConfig)
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range dockerConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			resourceList = append(resourceList, cfg)
		}
	case constants.ResourceTypeSwitch:
		var switchConfigs []*model.SwitchConfig
		var resArole []*model.ResourceRole
		err := r.DB.Model(&model.ResourceRole{}).
			Where("role_id = ? and resource_type = ?", roleId, resourceType).
			Find(&resArole).Error
		if err != nil {
			return nil, err
		}
		for _, res := range resArole {
			var switchConfig model.SwitchConfig
			err := r.DB.Model(&model.SwitchConfig{}).Where("id = ? AND deleted_at IS NULL", res.ResourceID).First(&switchConfig).Error
			if err != nil {
				// 如果资源不存在或已删除，跳过
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// 再次检查 ID 是否为 0（防止零值记录）
			if switchConfig.ID == 0 {
				continue
			}
			switchConfigs = append(switchConfigs, &switchConfig)
		}
		// 将配置转换为 model.Resource 接口类型的指针，并添加到 resourceList 中
		for _, cfg := range switchConfigs {
			// 过滤已删除的资源
			if cfg.DeletedAt.Valid {
				continue
			}
			resourceList = append(resourceList, cfg)
		}
	default:
		return nil, errors.New("unknown resource type")
	}
	return resourceList, nil
}
