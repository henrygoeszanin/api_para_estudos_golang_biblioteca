// filepath: d:\Dev\golang\api_golang_estudos\application\services\test\bookService_test.go
package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
	"github.com/henrygoeszanin/api_golang_estudos/domain/entities"
)

// MockBookRepository é uma implementação mock do BookRepository
// Usada para simular o comportamento do repositório real durante os testes
// sem depender de acesso ao banco de dados
type MockBookRepository struct {
	mock.Mock
}

// Create implementa o método Create da interface BookRepository
func (m *MockBookRepository) Create(book *entities.Book) error {
	args := m.Called(book)
	return args.Error(0)
}

// FindByID implementa o método FindByID da interface BookRepository
func (m *MockBookRepository) FindByID(id uint) (*entities.Book, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Book), args.Error(1)
}

// List implementa o método List da interface BookRepository
func (m *MockBookRepository) List() ([]*entities.Book, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Book), args.Error(1)
}

// Update implementa o método Update da interface BookRepository
func (m *MockBookRepository) Update(book *entities.Book) error {
	args := m.Called(book)
	return args.Error(0)
}

// Delete implementa o método Delete da interface BookRepository
func (m *MockBookRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// TestBookService_Create testa todos os cenários do método Create do BookService
func TestBookService_Create(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockBookRepository)
	bookService := services.NewBookService(mockRepo)

	t.Run("Deve_Criar_Livro_Com_Sucesso_Quando_Dados_Validos", func(t *testing.T) {
		// Arrange - Configuração específica para este caso de teste
		mockRepo.On("Create", mock.AnythingOfType("*entities.Book")).Return(nil).Once()

		// Act - Execução da funcionalidade que está sendo testada
		bookDTO := dtos.BookCreateDTO{
			Title:       "O Senhor dos Anéis",
			Author:      "J.R.R. Tolkien",
			Description: "Uma história épica de fantasia",
			Quantity:    10,
		}
		result, err := bookService.Create(bookDTO)

		// Assert - Verificação dos resultados esperados
		assert.NoError(t, err, "Não deve retornar erro ao criar livro com dados válidos")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, "O Senhor dos Anéis", result.Title, "O título no DTO de resposta deve corresponder ao fornecido")
		assert.Equal(t, "J.R.R. Tolkien", result.Author, "O autor no DTO de resposta deve corresponder ao fornecido")
		assert.Equal(t, 10, result.Quantity, "A quantidade no DTO de resposta deve corresponder ao fornecido")
		assert.Equal(t, 10, result.Available, "A disponibilidade inicial deve ser igual à quantidade")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_No_Repositorio", func(t *testing.T) {
		// Arrange - Configuração para simular erro no repositório
		dbError := errors.New("database error: connection failed")
		mockRepo.On("Create", mock.AnythingOfType("*entities.Book")).Return(dbError).Once()

		// Act - Tenta criar um livro com erro de repositório
		bookDTO := dtos.BookCreateDTO{
			Title:       "Livro com Erro",
			Author:      "Autor Teste",
			Description: "Descrição de teste",
			Quantity:    5,
		}
		result, err := bookService.Create(bookDTO)

		// Assert - Verifica se o erro é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando o repositório falha")
		assert.Nil(t, result, "Não deve retornar livro quando há erro no repositório")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do repositório")
		mockRepo.AssertExpectations(t)
	})
}

// TestBookService_GetByID testa todos os cenários do método GetByID do BookService
func TestBookService_GetByID(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockBookRepository)
	bookService := services.NewBookService(mockRepo)

	t.Run("Deve_Retornar_Livro_Quando_ID_Existente", func(t *testing.T) {
		// Arrange - Configura mock para retornar um livro específico
		mockBook := &entities.Book{
			Model:       gorm.Model{ID: 1},
			Title:       "Livro Encontrado",
			Author:      "Autor Encontrado",
			Description: "Descrição do livro encontrado",
			Quantity:    5,
			Available:   3,
		}
		mockRepo.On("FindByID", uint(1)).Return(mockBook, nil).Once()

		// Act - Busca livro com ID existente
		result, err := bookService.GetByID(1)

		// Assert - Verifica se o livro correto foi retornado
		assert.NoError(t, err, "Não deve retornar erro ao encontrar livro existente")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, "Livro Encontrado", result.Title, "O título deve corresponder ao livro encontrado")
		assert.Equal(t, "Autor Encontrado", result.Author, "O autor deve corresponder ao livro encontrado")
		assert.Equal(t, 5, result.Quantity, "A quantidade deve corresponder ao livro encontrado")
		assert.Equal(t, 3, result.Available, "A disponibilidade deve corresponder ao livro encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Livro_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Configura mock para simular ID inexistente
		mockRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Busca livro com ID inexistente
		result, err := bookService.GetByID(999)

		// Assert - Verifica se o erro apropriado foi retornado
		assert.Error(t, err, "Deve retornar erro quando livro não é encontrado")
		assert.Nil(t, result, "Não deve retornar livro quando não encontrado")
		assert.Contains(t, err.Error(), "livro não encontrado", "A mensagem deve indicar livro não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_No_Acesso_Ao_Banco", func(t *testing.T) {
		// Arrange - Configura mock para simular erro de banco de dados
		dbError := errors.New("database error: connection timeout")
		mockRepo.On("FindByID", uint(2)).Return(nil, dbError).Once()

		// Act - Executa com cenário de falha no banco
		result, err := bookService.GetByID(2)

		// Assert - Verifica se o erro do banco é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando há falha no banco de dados")
		assert.Nil(t, result, "Não deve retornar livro quando há erro de banco")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestBookService_List testa todos os cenários do método List do BookService
func TestBookService_List(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockBookRepository)
	bookService := services.NewBookService(mockRepo)

	t.Run("Deve_Listar_Livros_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista de livros
		mockBooks := []*entities.Book{
			{
				Model:       gorm.Model{ID: 1},
				Title:       "Livro 1",
				Author:      "Autor 1",
				Description: "Descrição 1",
				Quantity:    10,
				Available:   8,
			},
			{
				Model:       gorm.Model{ID: 2},
				Title:       "Livro 2",
				Author:      "Autor 2",
				Description: "Descrição 2",
				Quantity:    5,
				Available:   5,
			},
		}
		mockRepo.On("List").Return(mockBooks, nil).Once()

		// Act - Executa listagem
		result, err := bookService.List()

		// Assert - Verifica resultado da listagem
		assert.NoError(t, err, "Não deve retornar erro ao listar livros")
		assert.Len(t, result, 2, "Deve retornar 2 livros")
		assert.Equal(t, "Livro 1", result[0].Title, "O título do primeiro livro deve corresponder")
		assert.Equal(t, "Autor 1", result[0].Author, "O autor do primeiro livro deve corresponder")
		assert.Equal(t, 10, result[0].Quantity, "A quantidade do primeiro livro deve corresponder")
		assert.Equal(t, 8, result[0].Available, "A disponibilidade do primeiro livro deve corresponder")
		assert.Equal(t, "Livro 2", result[1].Title, "O título do segundo livro deve corresponder")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Lista_Vazia_Quando_Sem_Livros", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista vazia
		emptyList := []*entities.Book{}
		mockRepo.On("List").Return(emptyList, nil).Once()

		// Act - Executa listagem
		result, err := bookService.List()

		// Assert - Verifica resultado da listagem
		assert.NoError(t, err, "Não deve retornar erro com lista vazia")
		assert.Empty(t, result, "Deve retornar lista vazia")
		assert.Len(t, result, 0, "Lista deve ter comprimento 0")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Listagem", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: connection timeout")
		mockRepo.On("List").Return(nil, dbError).Once()

		// Act - Executa listagem com erro de banco
		result, err := bookService.List()

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar lista de livros")
		assert.Equal(t, dbError, err, "Deve propagar o erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestBookService_Update testa todos os cenários do método Update do BookService
func TestBookService_Update(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockBookRepository)
	bookService := services.NewBookService(mockRepo)

	t.Run("Deve_Atualizar_Livro_Com_Sucesso_Quando_Dados_Validos", func(t *testing.T) {
		// Arrange - Configura mock para simular livro encontrado
		mockBook := &entities.Book{
			Model:       gorm.Model{ID: 1},
			Title:       "Título Original",
			Author:      "Autor Original",
			Description: "Descrição Original",
			Quantity:    5,
			Available:   3,
		}
		mockRepo.On("FindByID", uint(1)).Return(mockBook, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*entities.Book")).Return(nil).Once()

		// Act - Executa atualização de livro
		updateDTO := dtos.BookUpdateDTO{
			Title:       "Título Atualizado",
			Author:      "Autor Atualizado",
			Description: "Descrição Atualizada",
		}
		result, err := bookService.Update(1, updateDTO)

		// Assert - Verifica resultado da atualização
		assert.NoError(t, err, "Não deve retornar erro ao atualizar livro")
		assert.NotNil(t, result, "Deve retornar livro atualizado")
		assert.Equal(t, "Título Atualizado", result.Title, "O título deve ser atualizado")
		assert.Equal(t, "Autor Atualizado", result.Author, "O autor deve ser atualizado")
		assert.Equal(t, "Descrição Atualizada", result.Description, "A descrição deve ser atualizada")
		assert.Equal(t, 5, result.Quantity, "A quantidade não deve ser alterada")
		assert.Equal(t, 3, result.Available, "A disponibilidade não deve ser alterada")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Atualizar_Quantidade_E_Disponibilidade_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para simular livro encontrado
		mockBook := &entities.Book{
			Model:       gorm.Model{ID: 2},
			Title:       "Livro Quantidade",
			Author:      "Autor Teste",
			Description: "Descrição Teste",
			Quantity:    5,
			Available:   3,
		}
		mockRepo.On("FindByID", uint(2)).Return(mockBook, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*entities.Book")).Return(nil).Once()

		// Act - Executa atualização com nova quantidade
		updateDTO := dtos.BookUpdateDTO{
			Quantity: 10, // Aumentando em 5
		}
		result, err := bookService.Update(2, updateDTO)

		// Assert - Verifica resultado da atualização de quantidade
		assert.NoError(t, err, "Não deve retornar erro ao atualizar quantidade")
		assert.NotNil(t, result, "Deve retornar livro atualizado")
		assert.Equal(t, 10, result.Quantity, "A quantidade deve ser atualizada")
		assert.Equal(t, 8, result.Available, "A disponibilidade deve ser ajustada proporcionalmente")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Livro_Nao_Existe", func(t *testing.T) {
		// Arrange - Configura mock para simular livro não encontrado
		mockRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Tenta atualizar livro inexistente
		updateDTO := dtos.BookUpdateDTO{
			Title: "Não Existe",
		}
		result, err := bookService.Update(999, updateDTO)

		// Assert - Verifica que ocorreu erro
		assert.Error(t, err, "Deve retornar erro quando livro não existe")
		assert.Nil(t, result, "Não deve retornar livro")
		assert.Contains(t, err.Error(), "livro não encontrado", "A mensagem deve indicar livro não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: connection failed")
		mockRepo.On("FindByID", uint(3)).Return(nil, dbError).Once()

		// Act - Executa atualização com erro de banco
		updateDTO := dtos.BookUpdateDTO{
			Title: "Erro Banco",
		}
		result, err := bookService.Update(3, updateDTO)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar livro")
		assert.Equal(t, dbError, err, "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Atualização", func(t *testing.T) {
		// Arrange - Configura mock para simular falha na atualização
		mockBook := &entities.Book{
			Model:       gorm.Model{ID: 4},
			Title:       "Falha Update",
			Author:      "Autor Falha",
			Description: "Descrição Falha",
			Quantity:    5,
			Available:   5,
		}
		mockRepo.On("FindByID", uint(4)).Return(mockBook, nil).Once()
		updateError := errors.New("falha ao atualizar no banco")
		mockRepo.On("Update", mock.AnythingOfType("*entities.Book")).Return(updateError).Once()

		// Act - Executa atualização com falha
		updateDTO := dtos.BookUpdateDTO{
			Title: "Nunca Atualizado",
		}
		result, err := bookService.Update(4, updateDTO)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha na atualização")
		assert.Nil(t, result, "Não deve retornar livro")
		assert.Equal(t, updateError, err, "Deve propagar erro original da atualização")
		mockRepo.AssertExpectations(t)
	})
}

// TestBookService_Delete testa todos os cenários do método Delete do BookService
func TestBookService_Delete(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockBookRepository)
	bookService := services.NewBookService(mockRepo)

	t.Run("Deve_Excluir_Livro_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para simular exclusão bem-sucedida
		mockRepo.On("Delete", uint(1)).Return(nil).Once()

		// Act - Executa exclusão
		err := bookService.Delete(1)

		// Assert - Verifica resultado da exclusão
		assert.NoError(t, err, "Não deve retornar erro ao excluir livro existente")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Livro_Nao_Existe", func(t *testing.T) {
		// Arrange - Configura mock para simular erro de livro não encontrado
		notFoundError := errors.New("livro não encontrado")
		mockRepo.On("Delete", uint(999)).Return(notFoundError).Once()

		// Act - Tenta excluir livro inexistente
		err := bookService.Delete(999)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro ao excluir livro inexistente")
		assert.Equal(t, notFoundError, err, "Deve propagar erro original da exclusão")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Exclusao", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: constraint violation")
		mockRepo.On("Delete", uint(2)).Return(dbError).Once()

		// Act - Executa exclusão com erro de banco
		err := bookService.Delete(2)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Equal(t, dbError, err, "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}
