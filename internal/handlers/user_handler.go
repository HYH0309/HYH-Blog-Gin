package handlers

import (
	"strings"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc services.UserService
	jwt *auth.JWTService
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // email or username
	Password   string `json:"password" binding:"required"`
}

type RegisterResponse struct {
	ID       uint   `json:"id" example:"1"`
	Email    string `json:"email" example:"user@example.com"`
	Username string `json:"username" example:"alice"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// NewUserHandler 创建并返回 UserHandler 实例（使用 service 层和 JWT 服务）。
func NewUserHandler(svc services.UserService, jwt *auth.JWTService) *UserHandler {
	return &UserHandler{svc: svc, jwt: jwt}
}

// Register 注册新用户
// @Summary 用户注册
// @Description 使用邮箱、用户名和密码注册新用户
// @Tags 用户
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "注册信息"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 输入规范化
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	user, err := h.svc.Register(req.Email, req.Username, req.Password)
	if err != nil {
		// 这里可以根据错误类型返回不同的状态码，例如冲突等
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, RegisterResponse{ID: user.ID, Email: user.Email, Username: user.Username})
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名或邮箱与密码登录，返回 JWT Token
// @Tags 用户
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	ident := strings.TrimSpace(req.Identifier)
	// 如果是邮箱，转换为小写
	if strings.Contains(ident, "@") {
		ident = strings.ToLower(ident)
	}

	user, err := h.svc.Authenticate(ident, req.Password)
	if err != nil {
		utils.Unauthorized(c, "invalid credentials")
		return
	}

	token, err := h.jwt.GenerateToken(user.ID)
	if err != nil {
		utils.InternalError(c, "failed to generate token")
		return
	}
	utils.OK(c, LoginResponse{Token: token})
}

// GetProfile 获取当前用户信息
// @Summary 获取个人信息
// @Description 获取当前登录用户的个人资料（需要鉴权）
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserSwagger
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	id, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.Unauthorized(c, "unauthorized")
		return
	}
	user, err := h.svc.GetByID(id)
	if err != nil {
		utils.NotFound(c, "user not found")
		return
	}
	// service 已返回去除密码的副本，直接返回
	utils.OK(c, user)
}
