// filepath: d:\Dev\golang\api_golang_estudos\application\services\test\loanService_test.go
package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
	"github.com/henrygoeszanin/api_golang_estudos/domain/entities"
)

// MockLoanRepository é uma implementação mock do LoanRepository
// Usada para simular o comportamento do repositório de empréstimos durante os testes
type MockLoanRepository struct {
	mock.Mock
}

// Create implementa o método Create da interface LoanRepository
func (m *MockLoanRepository) Create(loan *entities.Loan) error {
	args := m.Called(loan)
	// Simular atribuição de ID ao criar (comportamento do banco)
	if loan.ID == 0 && args.Error(0) == nil {
		loan.ID = 1
	}
	return args.Error(0)
}

// FindByID implementa o método FindByID da interface LoanRepository
func (m *MockLoanRepository) FindByID(id uint) (*entities.Loan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Loan), args.Error(1)
}

// FindByUserID implementa o método FindByUserID da interface LoanRepository
func (m *MockLoanRepository) FindByUserID(userID uint) ([]*entities.Loan, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Loan), args.Error(1)
}

// Update implementa o método Update da interface LoanRepository
func (m *MockLoanRepository) Update(loan *entities.Loan) error {
	args := m.Called(loan)
	return args.Error(0)
}

// ReturnLoan implementa o método ReturnLoan da interface LoanRepository
func (m *MockLoanRepository) ReturnLoan(id uint, returnDate time.Time) error {
	args := m.Called(id, returnDate)
	return args.Error(0)
}

// TestLoanService_Create testa todos os cenários do método Create do LoanService
func TestLoanService_Create(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockLoanRepo := new(MockLoanRepository)
	mockBookRepo := new(MockBookRepository)
	loanService := services.NewLoanService(mockLoanRepo, mockBookRepo)

	t.Run("Deve_Criar_Emprestimo_Com_Sucesso_Quando_Livro_Disponivel", func(t *testing.T) {
		// Arrange - Configuração específica para este caso de teste
		// Simular livro encontrado com disponibilidade
		mockBook := &entities.Book{
			Model:     gorm.Model{ID: 1},
			Title:     "Livro Disponível",
			Author:    "Autor Teste",
			Quantity:  5,
			Available: 3,
		}
		mockBookRepo.On("FindByID", uint(1)).Return(mockBook, nil).Once()

		// Simular criação do empréstimo bem-sucedida
		mockLoanRepo.On("Create", mock.AnythingOfType("*entities.Loan")).Return(nil).Once()

		// Simular carregamento completo do empréstimo após criação
		returnDate := time.Now().AddDate(0, 0, 14) // Daqui a 14 dias
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		mockLoan := &entities.Loan{
			Model:      gorm.Model{ID: 1},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now(),
			ReturnDate: returnDate,
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(1)).Return(mockLoan, nil).Once()

		// Act - Execução da funcionalidade que está sendo testada
		loanDTO := dtos.LoanCreateDTO{
			BookID:     1,
			ReturnDate: returnDate,
		}
		result, err := loanService.Create(2, loanDTO)

		// Assert - Verificação dos resultados esperados
		assert.NoError(t, err, "Não deve retornar erro ao criar empréstimo com livro disponível")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, uint(1), result.BookID, "O ID do livro deve corresponder ao fornecido")
		assert.Equal(t, uint(2), result.UserID, "O ID do usuário deve corresponder ao fornecido")
		assert.Equal(t, "Livro Disponível", result.BookTitle, "O título do livro deve corresponder")
		assert.Equal(t, "Usuário Teste", result.UserName, "O nome do usuário deve corresponder")
		assert.False(t, result.IsReturned, "O empréstimo não deve estar marcado como devolvido")
		mockBookRepo.AssertExpectations(t)
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Livro_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Simular livro não encontrado
		mockBookRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Tenta criar empréstimo com livro inexistente
		loanDTO := dtos.LoanCreateDTO{
			BookID:     999,
			ReturnDate: time.Now().AddDate(0, 0, 7),
		}
		result, err := loanService.Create(2, loanDTO)

		// Assert - Verifica erro apropriado
		assert.Error(t, err, "Deve retornar erro quando livro não é encontrado")
		assert.Nil(t, result, "Não deve retornar resposta quando há erro")
		assert.Contains(t, err.Error(), "livro não encontrado", "Mensagem deve indicar livro não encontrado")
		mockBookRepo.AssertExpectations(t)
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Livro_Indisponivel", func(t *testing.T) {
		// Arrange - Simular livro sem disponibilidade
		mockBook := &entities.Book{
			Model:     gorm.Model{ID: 2},
			Title:     "Livro Indisponível",
			Author:    "Autor Teste",
			Quantity:  3,
			Available: 0, // Sem exemplares disponíveis
		}
		mockBookRepo.On("FindByID", uint(2)).Return(mockBook, nil).Once()

		// Act - Tenta criar empréstimo com livro indisponível
		loanDTO := dtos.LoanCreateDTO{
			BookID:     2,
			ReturnDate: time.Now().AddDate(0, 0, 7),
		}
		result, err := loanService.Create(2, loanDTO)

		// Assert - Verifica erro apropriado
		assert.Error(t, err, "Deve retornar erro quando livro não está disponível")
		assert.Nil(t, result, "Não deve retornar resposta quando há erro")
		assert.Contains(t, err.Error(), "livro não disponível", "Mensagem deve indicar que livro não está disponível")
		mockBookRepo.AssertExpectations(t)
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca_Do_Livro", func(t *testing.T) {
		// Arrange - Simular erro no banco ao buscar livro
		dbError := errors.New("database error: connection failed")
		mockBookRepo.On("FindByID", uint(3)).Return(nil, dbError).Once()

		// Act - Tenta criar empréstimo com erro na busca
		loanDTO := dtos.LoanCreateDTO{
			BookID:     3,
			ReturnDate: time.Now().AddDate(0, 0, 7),
		}
		result, err := loanService.Create(2, loanDTO)

		// Assert - Verifica propagação do erro
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar resposta quando há erro")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do banco")
		mockBookRepo.AssertExpectations(t)
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Criacao_Do_Emprestimo", func(t *testing.T) {
		// Arrange - Simular livro encontrado mas falha ao criar empréstimo
		mockBook := &entities.Book{
			Model:     gorm.Model{ID: 4},
			Title:     "Livro Erro",
			Author:    "Autor Teste",
			Quantity:  5,
			Available: 3,
		}
		mockBookRepo.On("FindByID", uint(4)).Return(mockBook, nil).Once()

		// Simular erro na criação
		createError := errors.New("database error: constraint violation")
		mockLoanRepo.On("Create", mock.AnythingOfType("*entities.Loan")).Return(createError).Once()

		// Act - Tenta criar empréstimo com erro na criação
		loanDTO := dtos.LoanCreateDTO{
			BookID:     4,
			ReturnDate: time.Now().AddDate(0, 0, 7),
		}
		result, err := loanService.Create(2, loanDTO)

		// Assert - Verifica propagação do erro
		assert.Error(t, err, "Deve retornar erro quando falha na criação")
		assert.Nil(t, result, "Não deve retornar resposta quando há erro")
		assert.Equal(t, createError, err, "Deve propagar o erro original do banco")
		mockBookRepo.AssertExpectations(t)
		mockLoanRepo.AssertExpectations(t)
	})
}

// TestLoanService_GetByID testa todos os cenários do método GetByID do LoanService
func TestLoanService_GetByID(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockLoanRepo := new(MockLoanRepository)
	mockBookRepo := new(MockBookRepository)
	loanService := services.NewLoanService(mockLoanRepo, mockBookRepo)

	t.Run("Deve_Retornar_Emprestimo_Quando_ID_E_Usuario_Validos", func(t *testing.T) {
		// Arrange - Configura mock para retornar um empréstimo específico
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		mockLoan := &entities.Loan{
			Model:      gorm.Model{ID: 1},
			UserID:     2, // Mesmo ID do usuário logado
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour), // 7 dias atrás
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),  // 7 dias à frente
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(1)).Return(mockLoan, nil).Once()

		// Act - Busca empréstimo com ID existente e usuário autorizado
		result, err := loanService.GetByID(1, 2)

		// Assert - Verifica se o empréstimo correto foi retornado
		assert.NoError(t, err, "Não deve retornar erro ao encontrar empréstimo")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, uint(1), result.ID, "O ID deve corresponder ao empréstimo encontrado")
		assert.Equal(t, uint(2), result.UserID, "O ID do usuário deve corresponder")
		assert.Equal(t, uint(1), result.BookID, "O ID do livro deve corresponder")
		assert.Equal(t, "Livro Teste", result.BookTitle, "O título do livro deve corresponder")
		assert.Equal(t, "Usuário Teste", result.UserName, "O nome do usuário deve corresponder")
		assert.False(t, result.IsReturned, "Status de devolução deve corresponder")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Emprestimo_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Configura mock para simular empréstimo não encontrado
		mockLoanRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Busca empréstimo com ID inexistente
		result, err := loanService.GetByID(999, 2)

		// Assert - Verifica se o erro apropriado foi retornado
		assert.Error(t, err, "Deve retornar erro quando empréstimo não é encontrado")
		assert.Nil(t, result, "Não deve retornar empréstimo quando não encontrado")
		assert.Contains(t, err.Error(), "empréstimo não encontrado", "A mensagem deve indicar empréstimo não encontrado")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Autorizado", func(t *testing.T) {
		// Arrange - Configura mock para retornar empréstimo de outro usuário
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 3}, Name: "Outro Usuário"}
		mockLoan := &entities.Loan{
			Model:      gorm.Model{ID: 2},
			UserID:     3, // Diferente do usuário logado (ID 2)
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(2)).Return(mockLoan, nil).Once()

		// Act - Busca empréstimo de outro usuário
		result, err := loanService.GetByID(2, 2) // Usuário 2 tentando acessar empréstimo do usuário 3

		// Assert - Verifica se o erro de autorização foi retornado
		assert.Error(t, err, "Deve retornar erro quando usuário não está autorizado")
		assert.Nil(t, result, "Não deve retornar empréstimo quando não autorizado")
		assert.Contains(t, err.Error(), "acesso negado", "A mensagem deve indicar acesso negado")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_No_Acesso_Ao_Banco", func(t *testing.T) {
		// Arrange - Configura mock para simular erro de banco de dados
		dbError := errors.New("database error: connection timeout")
		mockLoanRepo.On("FindByID", uint(3)).Return(nil, dbError).Once()

		// Act - Executa com cenário de falha no banco
		result, err := loanService.GetByID(3, 2)

		// Assert - Verifica se o erro do banco é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando há falha no banco de dados")
		assert.Nil(t, result, "Não deve retornar empréstimo quando há erro de banco")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do banco")
		mockLoanRepo.AssertExpectations(t)
	})
}

// TestLoanService_ListByUser testa todos os cenários do método ListByUser do LoanService
func TestLoanService_ListByUser(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockLoanRepo := new(MockLoanRepository)
	mockBookRepo := new(MockBookRepository)
	loanService := services.NewLoanService(mockLoanRepo, mockBookRepo)

	t.Run("Deve_Listar_Emprestimos_Do_Usuario_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista de empréstimos
		mockBook1 := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro 1"}
		mockBook2 := &entities.Book{Model: gorm.Model{ID: 2}, Title: "Livro 2"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}

		mockLoans := []*entities.Loan{
			{
				Model:      gorm.Model{ID: 1},
				UserID:     2,
				User:       *mockUser,
				BookID:     1,
				Book:       *mockBook1,
				LoanDate:   time.Now().Add(-14 * 24 * time.Hour),
				ReturnDate: time.Now().Add(-7 * 24 * time.Hour),
				IsReturned: true,
				ReturnedAt: func() *time.Time { t := time.Now().Add(-8 * 24 * time.Hour); return &t }(),
			},
			{
				Model:      gorm.Model{ID: 2},
				UserID:     2,
				User:       *mockUser,
				BookID:     2,
				Book:       *mockBook2,
				LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
				ReturnDate: time.Now().Add(7 * 24 * time.Hour),
				IsReturned: false,
			},
		}
		mockLoanRepo.On("FindByUserID", uint(2)).Return(mockLoans, nil).Once()

		// Act - Executa listagem
		result, err := loanService.ListByUser(2)

		// Assert - Verifica resultado da listagem
		assert.NoError(t, err, "Não deve retornar erro ao listar empréstimos do usuário")
		assert.Len(t, result, 2, "Deve retornar 2 empréstimos")
		assert.Equal(t, uint(1), result[0].ID, "O ID do primeiro empréstimo deve corresponder")
		assert.Equal(t, "Livro 1", result[0].BookTitle, "O título do livro deve corresponder")
		assert.True(t, result[0].IsReturned, "O status de devolução deve corresponder")
		assert.Equal(t, uint(2), result[1].ID, "O ID do segundo empréstimo deve corresponder")
		assert.Equal(t, "Livro 2", result[1].BookTitle, "O título do livro deve corresponder")
		assert.False(t, result[1].IsReturned, "O status de devolução deve corresponder")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Lista_Vazia_Quando_Usuario_Sem_Emprestimos", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista vazia
		emptyList := []*entities.Loan{}
		mockLoanRepo.On("FindByUserID", uint(3)).Return(emptyList, nil).Once()

		// Act - Executa listagem
		result, err := loanService.ListByUser(3)

		// Assert - Verifica resultado da listagem
		assert.NoError(t, err, "Não deve retornar erro com lista vazia")
		assert.Empty(t, result, "Deve retornar lista vazia")
		assert.Len(t, result, 0, "Lista deve ter comprimento 0")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Listagem", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: connection timeout")
		mockLoanRepo.On("FindByUserID", uint(4)).Return(nil, dbError).Once()

		// Act - Executa listagem com erro de banco
		result, err := loanService.ListByUser(4)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar lista de empréstimos")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do banco")
		mockLoanRepo.AssertExpectations(t)
	})
}

// TestLoanService_ReturnLoan testa todos os cenários do método ReturnLoan do LoanService
func TestLoanService_ReturnLoan(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockLoanRepo := new(MockLoanRepository)
	mockBookRepo := new(MockBookRepository)
	loanService := services.NewLoanService(mockLoanRepo, mockBookRepo)

	t.Run("Deve_Devolver_Emprestimo_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mocks para cenário de devolução bem-sucedida
		// 1. Simular busca pelo empréstimo ativo
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		activeLoan := &entities.Loan{
			Model:      gorm.Model{ID: 1},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(1)).Return(activeLoan, nil).Once()

		// 2. Simular processo de devolução
		mockLoanRepo.On("ReturnLoan", uint(1), mock.AnythingOfType("time.Time")).Return(nil).Once()

		// 3. Simular busca pelo empréstimo atualizado
		returnTime := time.Now()
		returnedLoan := &entities.Loan{
			Model:      gorm.Model{ID: 1},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: true,
			ReturnedAt: &returnTime,
		}
		mockLoanRepo.On("FindByID", uint(1)).Return(returnedLoan, nil).Once()

		// Act - Executa devolução
		result, err := loanService.ReturnLoan(1, 2)

		// Assert - Verifica resultado da devolução
		assert.NoError(t, err, "Não deve retornar erro ao devolver empréstimo")
		assert.NotNil(t, result, "Deve retornar resposta não nula")
		assert.Equal(t, uint(1), result.ID, "O ID do empréstimo deve corresponder")
		assert.True(t, result.IsReturned, "Empréstimo deve estar marcado como devolvido")
		assert.NotNil(t, result.ReturnedAt, "Data de devolução deve estar preenchida")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Emprestimo_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Configura mock para simular empréstimo não encontrado
		mockLoanRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Tenta devolver empréstimo inexistente
		result, err := loanService.ReturnLoan(999, 2)

		// Assert - Verifica erro apropriado
		assert.Error(t, err, "Deve retornar erro quando empréstimo não existe")
		assert.Nil(t, result, "Não deve retornar resultado")
		assert.Contains(t, err.Error(), "empréstimo não encontrado", "A mensagem deve indicar empréstimo não encontrado")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Autorizado", func(t *testing.T) {
		// Arrange - Configura mock para simular empréstimo de outro usuário
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 3}, Name: "Outro Usuário"}
		mockLoan := &entities.Loan{
			Model:      gorm.Model{ID: 2},
			UserID:     3, // Diferente do usuário solicitante (ID 2)
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(2)).Return(mockLoan, nil).Once()

		// Act - Tenta devolver empréstimo de outro usuário
		result, err := loanService.ReturnLoan(2, 2)

		// Assert - Verifica erro de autorização
		assert.Error(t, err, "Deve retornar erro quando usuário não está autorizado")
		assert.Nil(t, result, "Não deve retornar resultado")
		assert.Contains(t, err.Error(), "acesso negado", "A mensagem deve indicar acesso negado")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Emprestimo_Ja_Devolvido", func(t *testing.T) {
		// Arrange - Configura mock para simular empréstimo já devolvido
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		returnTime := time.Now().Add(-1 * 24 * time.Hour)
		mockLoan := &entities.Loan{
			Model:      gorm.Model{ID: 3},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(-3 * 24 * time.Hour),
			IsReturned: true,
			ReturnedAt: &returnTime,
		}
		mockLoanRepo.On("FindByID", uint(3)).Return(mockLoan, nil).Once()

		// Act - Tenta devolver empréstimo que já foi devolvido
		result, err := loanService.ReturnLoan(3, 2)

		// Assert - Verifica erro apropriado
		assert.Error(t, err, "Deve retornar erro quando empréstimo já foi devolvido")
		assert.Nil(t, result, "Não deve retornar resultado")
		assert.Contains(t, err.Error(), "já foi devolvido", "A mensagem deve indicar que já foi devolvido")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Devolucao", func(t *testing.T) {
		// Arrange - Configura mocks para cenário de falha na devolução
		// 1. Simular busca pelo empréstimo
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		activeLoan := &entities.Loan{
			Model:      gorm.Model{ID: 4},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(4)).Return(activeLoan, nil).Once()

		// 2. Simular erro na devolução
		returnError := errors.New("database error: update failed")
		mockLoanRepo.On("ReturnLoan", uint(4), mock.AnythingOfType("time.Time")).Return(returnError).Once()

		// Act - Tenta devolver com erro na operação
		result, err := loanService.ReturnLoan(4, 2)

		// Assert - Verifica propagação do erro
		assert.Error(t, err, "Deve retornar erro quando falha na devolução")
		assert.Nil(t, result, "Não deve retornar resultado")
		assert.Equal(t, returnError, err, "Deve propagar o erro original")
		mockLoanRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca_Apos_Devolucao", func(t *testing.T) {
		// Arrange - Configura mocks para cenário de falha após devolução
		// 1. Simular busca pelo empréstimo
		mockBook := &entities.Book{Model: gorm.Model{ID: 1}, Title: "Livro Teste"}
		mockUser := &entities.User{Model: gorm.Model{ID: 2}, Name: "Usuário Teste"}
		activeLoan := &entities.Loan{
			Model:      gorm.Model{ID: 5},
			UserID:     2,
			User:       *mockUser,
			BookID:     1,
			Book:       *mockBook,
			LoanDate:   time.Now().Add(-7 * 24 * time.Hour),
			ReturnDate: time.Now().Add(7 * 24 * time.Hour),
			IsReturned: false,
		}
		mockLoanRepo.On("FindByID", uint(5)).Return(activeLoan, nil).Once()

		// 2. Simular sucesso na devolução
		mockLoanRepo.On("ReturnLoan", uint(5), mock.AnythingOfType("time.Time")).Return(nil).Once()

		// 3. Simular erro na busca pelo empréstimo atualizado
		dbError := errors.New("database error: connection lost")
		mockLoanRepo.On("FindByID", uint(5)).Return(nil, dbError).Once()

		// Act - Executa devolução com erro após atualização
		result, err := loanService.ReturnLoan(5, 2)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha ao buscar empréstimo atualizado")
		assert.Nil(t, result, "Não deve retornar resultado")
		assert.Equal(t, dbError, err, "Deve propagar o erro original")
		mockLoanRepo.AssertExpectations(t)
	})
}
