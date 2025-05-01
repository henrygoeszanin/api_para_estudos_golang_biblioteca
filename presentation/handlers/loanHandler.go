package handlers

import (
	"net/http"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
)

// LoanHandler manipula as requisições relacionadas a empréstimos
type LoanHandler struct {
	loanService *services.LoanService
}

// NewLoanHandler cria uma nova instância de LoanHandler
func NewLoanHandler(loanService *services.LoanService) *LoanHandler {
	return &LoanHandler{
		loanService: loanService,
	}
}

// List lista todos os empréstimos do usuário atual
func (loalHandler *LoanHandler) List(c *gin.Context) {
	// Obter ID do usuário das claims do JWT
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))

	loans, err := loalHandler.loanService.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// Create manipula a criação de um novo empréstimo
func (loalHandler *LoanHandler) Create(c *gin.Context) {
	// Obter ID do usuário das claims do JWT
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))

	// Processar body da requisição
	var loanDTO dtos.LoanCreateDTO
	if err := c.ShouldBindJSON(&loanDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan, err := loalHandler.loanService.Create(userID, loanDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, loan)
}

// GetByID obtém detalhes de um empréstimo específico
func (loalHandler *LoanHandler) GetByID(c *gin.Context) {
	// Obter ID do usuário das claims do JWT
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))

	// Obter ID do empréstimo da URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	loan, err := loalHandler.loanService.GetByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// ReturnLoan marca um empréstimo como devolvido
func (loalHandler *LoanHandler) ReturnLoan(c *gin.Context) {
	// Obter ID do usuário das claims do JWT
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))

	// Obter ID do empréstimo da URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	loan, err := loalHandler.loanService.ReturnLoan(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Livro devolvido com sucesso",
		"loan":    loan,
	})
}
