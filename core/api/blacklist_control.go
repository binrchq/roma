package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"binrc.com/roma/core/middleware"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"binrc.com/roma/core/utils/logger"
	"github.com/gin-gonic/gin"
)

type BlacklistController struct{}

func NewBlacklistController() *BlacklistController {
	return &BlacklistController{}
}

// GetAllBlacklists 获取所有黑名单记录
// @Summary 获取黑名单列表
// @Description 获取所有黑名单记录，支持分页
// @Tags blacklist
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} utils.Response{data=[]model.Blacklist}
// @Failure 500 {object} utils.Response{data=""}
// @Router /api/v1/blacklist [get]
func (bc *BlacklistController) GetAllBlacklists(c *gin.Context) {
	utilG := utils.Gin{C: c}
	
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	
	op := operation.NewBlacklistOperation()
	blacklists, err := op.GetAll(pageSize, offset)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取黑名单列表失败: "+err.Error())
		return
	}
	
	// 解析IP信息JSON
	for i := range blacklists {
		if blacklists[i].IPInfo != "" {
			var ipInfo middleware.IPInfo
			if err := json.Unmarshal([]byte(blacklists[i].IPInfo), &ipInfo); err == nil {
				// IP信息已解析，可以用于前端显示
				_ = ipInfo
			}
		}
	}
	
	utilG.Response(http.StatusOK, utils.SUCCESS, gin.H{
		"list":      blacklists,
		"total":     len(blacklists),
		"page":      page,
		"page_size": pageSize,
	})
}

// GetBlacklistByIP 根据IP获取黑名单记录
// @Summary 获取IP的黑名单信息
// @Description 根据IP地址获取黑名单记录和IP信息
// @Tags blacklist
// @Produce json
// @Param ip path string true "IP地址"
// @Success 200 {object} utils.Response{data=model.Blacklist}
// @Failure 404 {object} utils.Response{data=""}
// @Router /api/v1/blacklist/:ip [get]
func (bc *BlacklistController) GetBlacklistByIP(c *gin.Context) {
	utilG := utils.Gin{C: c}
	ip := c.Param("ip")
	
	if ip == "" {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "IP地址不能为空")
		return
	}
	
	op := operation.NewBlacklistOperation()
	blacklist, err := op.GetByIP(ip)
	if err != nil {
		// 如果不存在，尝试获取IP信息
		ipInfo, infoErr := middleware.GetIPInfoParsed(ip)
		if infoErr == nil {
			utilG.Response(http.StatusOK, utils.SUCCESS, gin.H{
				"ip":        ip,
				"blacklisted": false,
				"ip_info":    ipInfo,
			})
			return
		}
		utilG.Response(http.StatusNotFound, utils.ERROR, "IP不在黑名单中")
		return
	}
	
	// 解析IP信息
	var ipInfo *middleware.IPInfo
	if blacklist.IPInfo != "" {
		var info middleware.IPInfo
		if err := json.Unmarshal([]byte(blacklist.IPInfo), &info); err == nil {
			ipInfo = &info
		}
	} else {
		// 如果数据库中没有IP信息，尝试获取
		info, err := middleware.GetIPInfoParsed(ip)
		if err == nil {
			ipInfo = info
			// 更新数据库中的IP信息
			ipInfoJSON, _ := json.Marshal(ipInfo)
			blacklist.IPInfo = string(ipInfoJSON)
			op.CreateOrUpdate(blacklist)
		}
	}
	
	utilG.Response(http.StatusOK, utils.SUCCESS, gin.H{
		"blacklist": blacklist,
		"ip_info":   ipInfo,
	})
}

// AddToBlacklist 手动添加IP到黑名单
// @Summary 添加IP到黑名单
// @Description 手动添加IP到黑名单
// @Tags blacklist
// @Accept json
// @Produce json
// @Param request body AddBlacklistRequest true "黑名单信息"
// @Success 200 {object} utils.Response{data=model.Blacklist}
// @Failure 400 {object} utils.Response{data=""}
// @Router /api/v1/blacklist [post]
func (bc *BlacklistController) AddToBlacklist(c *gin.Context) {
	utilG := utils.Gin{C: c}
	
	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}
	
	if req.IP == "" {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "IP地址不能为空")
		return
	}
	
	// 检查是否为内网IP，不允许封禁内网IP
	if utils.IsPrivateIP(req.IP) {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "不允许封禁内网IP地址")
		return
	}
	
	// 获取IP信息
	ipInfo, err := middleware.GetIPInfoParsed(req.IP)
	if err != nil {
		// IP信息获取失败不影响添加黑名单
		logger.Logger.Warning(fmt.Sprintf("Failed to get IP info for %s: %v", req.IP, err))
	}
	
	blacklist := &model.Blacklist{
		IP:     req.IP,
		Reason: req.Reason,
		Source: "manual",
	}
	
	if req.Duration > 0 {
		banUntil := time.Now().Add(time.Duration(req.Duration) * time.Second)
		blacklist.BanUntil = &banUntil
	}
	
	if ipInfo != nil {
		ipInfoJSON, _ := json.Marshal(ipInfo)
		blacklist.IPInfo = string(ipInfoJSON)
	}
	
	op := operation.NewBlacklistOperation()
	result, err := op.CreateOrUpdate(blacklist)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "添加黑名单失败: "+err.Error())
		return
	}
	
	utilG.Response(http.StatusOK, utils.SUCCESS, result)
}

// RemoveFromBlacklist 从黑名单移除IP（解禁）
// @Summary 解禁IP
// @Description 从黑名单移除IP，解禁该IP
// @Tags blacklist
// @Produce json
// @Param ip path string true "IP地址"
// @Success 200 {object} utils.Response{data="解禁成功"}
// @Failure 404 {object} utils.Response{data=""}
// @Router /api/v1/blacklist/:ip [delete]
func (bc *BlacklistController) RemoveFromBlacklist(c *gin.Context) {
	utilG := utils.Gin{C: c}
	ip := c.Param("ip")
	
	if ip == "" {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "IP地址不能为空")
		return
	}
	
	op := operation.NewBlacklistOperation()
	err := op.Delete(ip)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "解禁失败: "+err.Error())
		return
	}
	
	// 同时从内存黑名单移除
	middleware.RemoveFromBlacklist(ip)
	
	utilG.Response(http.StatusOK, utils.SUCCESS, "IP已解禁")
}

// GetIPInfo 获取IP信息（不检查黑名单）
// @Summary 获取IP信息
// @Description 从 ipseek.cc API 获取IP的地理位置和ISP信息
// @Tags blacklist
// @Produce json
// @Param ip path string true "IP地址"
// @Success 200 {object} utils.Response{data=middleware.IPInfo}
// @Failure 400 {object} utils.Response{data=""}
// @Router /api/v1/blacklist/ip-info/:ip [get]
func (bc *BlacklistController) GetIPInfo(c *gin.Context) {
	utilG := utils.Gin{C: c}
	ip := c.Param("ip")
	
	if ip == "" {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "IP地址不能为空")
		return
	}
	
	ipInfo, err := middleware.GetIPInfoParsed(ip)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取IP信息失败: "+err.Error())
		return
	}
	
	utilG.Response(http.StatusOK, utils.SUCCESS, ipInfo)
}

// AddBlacklistRequest 添加黑名单请求
type AddBlacklistRequest struct {
	IP       string `json:"ip" binding:"required"`        // IP地址
	Reason   string `json:"reason"`                       // 封禁原因
	Duration int64  `json:"duration"`                     // 封禁时长（秒），0表示永久封禁
}

