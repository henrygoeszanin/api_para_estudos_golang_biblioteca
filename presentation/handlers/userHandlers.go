package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"

	"github.com/gin-gonic/gin"
	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
)

// UserHandler manipula as requisições relacionadas a usuários
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler cria uma nova instância de UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register registra um novo usuário
func (userHandler *UserHandler) Register(c *gin.Context) {
	var userDTO dtos.UserCreateDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdUser, err := userHandler.userService.Create(userDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// GetByID busca um usuário pelo ID
func (userHandler *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	user, err := userHandler.userService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update atualiza os dados de um usuário
func (userHandler *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var userDTO dtos.UserUpdateDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := userHandler.userService.Update(uint(id), userDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete remove um usuário
func (userHandler *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := userHandler.userService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário removido com sucesso"})
}

// List lista todos os usuários
func (userHandler *UserHandler) List(c *gin.Context) {
	users, err := userHandler.userService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// PromoteToAdmin promove um usuário para administrador
func (userHandler *UserHandler) PromoteToAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	user, err := userHandler.userService.PromoteToAdmin(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usuário promovido a administrador",
		"user":    user,
	})
}

func (userHandler *UserHandler) GetMe(c *gin.Context) {
	// Extrair claims JWT para obter o ID do usuário autenticado
	claims := jwt.ExtractClaims(c)
	// Verificar se o claim id existe
	idVal, exists := claims["id"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "Token inválido: ID do usuário não encontrado",
			"details":          "O token não contém a identificação do usuário",
			"available_claims": claims,
		})
		return
	}

	// Conversão segura de tipos
	var userID uint
	if idFloat, ok := idVal.(float64); ok {
		userID = uint(idFloat)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Token inválido: formato de ID incorreto",
			"details": fmt.Sprintf("Esperava um número, recebeu %T", idVal),
		})
		return
	}

	// Buscar o usuário no serviço
	user, err := userHandler.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateMe atualiza o perfil do usuário logado
func (userHandler *UserHandler) UpdateMe(c *gin.Context) {
	// Extrair claims JWT para obter o ID do usuário autenticado
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))

	// Fazer o binding dos dados de atualização
	var userDTO dtos.UserUpdateDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Atualizar o usuário
	updatedUser, err := userHandler.userService.Update(userID, userDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}
