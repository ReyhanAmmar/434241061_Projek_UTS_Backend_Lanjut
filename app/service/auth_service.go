package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
	"api-buku-kas/app/repository"
	"api-buku-kas/helper"
)

const refreshTokenBytes = 32

type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	jwt        *helper.JWTManager
	perms      *helper.PermissionSet
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		jwt:        jwtManager,
		perms:      perms,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	created, err := s.users.Create(ctx, model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
		Role:     model.RoleUser,
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("email sudah dipakai")
		}
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusCreated,
		"pendaftaran berhasil",
		created,
	)
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.VerifyDummyPassword(req.Password)
			return helper.Unauthorized("email atau password salah")
		}
		return helper.Internal(err)
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("email atau password salah")
	}

	if !user.IsActive {
		return helper.Forbidden("akun dinonaktifkan")
	}

	pair, stored, err := s.newTokenPair(user)
	if err != nil {
		return helper.Internal(err)
	}

	if err := s.tokens.Save(ctx, stored); err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c, fiber.StatusOK, "login berhasil", pair,
	)
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	oldHash := helper.SHA256Hex(req.RefreshToken)

	stored, err := s.tokens.FindActive(ctx, oldHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized(
				"refresh token tidak valid atau sudah kedaluwarsa",
			)
		}
		return helper.Internal(err)
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("akun tidak dapat dipakai")
		}
		return helper.Internal(err)
	}

	if !user.IsActive {
		return helper.Forbidden("akun dinonaktifkan")
	}

	pair, next, err := s.newTokenPair(user)
	if err != nil {
		return helper.Internal(err)
	}

	if err := s.tokens.Rotate(ctx, oldHash, next); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized(
				"refresh token tidak valid atau sudah dipakai",
			)
		}
		return helper.Internal(err)
	}

	return helper.Success(
		c, fiber.StatusOK, "token berhasil diperbarui", pair,
	)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.tokens.Revoke(
		ctx, helper.SHA256Hex(req.RefreshToken),
	); err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c, fiber.StatusOK, "logout berhasil", nil,
	)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByID(ctx, current.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("akun tidak dapat dipakai")
		}

		return helper.Internal(err)
	}

	if !user.IsActive {
		return helper.Forbidden("akun dinonaktifkan")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"profil berhasil diambil",
		fiber.Map{
			"user":        user,
			"permissions": s.perms.PermissionsOf(user.Role),
		},
	)
}

func (s *AuthService) newTokenPair(
	user model.User,
) (model.TokenPair, model.RefreshToken, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, model.RefreshToken{}, err
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, model.RefreshToken{}, err
	}

	pair := model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}

	stored := model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.SHA256Hex(refreshToken),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}

	return pair, stored, nil
}