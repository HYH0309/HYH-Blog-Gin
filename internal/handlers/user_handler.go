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
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// NewUserHandler 创建并返回 UserHandler 实例（使用 service 层和 JWT 服务）。
func NewUserHandler(svc services.UserService, jwt *auth.JWTService) *UserHandler {
	return &UserHandler{svc: svc, jwt: jwt}
}

// Register 处理用户注册请求。
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

// Login 处理用户登录请求。
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

// GetProfile 获取当前用户的个人资料。
func (h *UserHandler) GetProfile(c *gin.Context) {
	id, ok := getUserIDFromContext(c)
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
