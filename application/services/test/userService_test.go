package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
	"github.com/henrygoeszanin/api_golang_estudos/domain/entities"
)

// MockUserRepository é uma implementação mock do UserRepository
// Usada para simular o comportamento do repositório real durante os testes
// sem depender de acesso ao banco de dados
type MockUserRepository struct {
	mock.Mock
}

// Create implementa o método Create da interface UserRepository
// Permite configurar o comportamento de criação de usuários nos testes
func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// FindByID implementa o método FindByID da interface UserRepository
// Permite configurar o comportamento de busca por ID nos testes
func (m *MockUserRepository) FindByID(id uint) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

// FindByEmail implementa o método FindByEmail da interface UserRepository
// Permite configurar o comportamento de busca por email nos testes
func (m *MockUserRepository) FindByEmail(email string) (*entities.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

// Update implementa o método Update da interface UserRepository
// Permite configurar o comportamento de atualização de usuários nos testes
func (m *MockUserRepository) Update(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Delete implementa o método Delete da interface UserRepository
// Permite configurar o comportamento de exclusão de usuários nos testes
func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// List implementa o método List da interface UserRepository
// Permite configurar o comportamento de listagem de usuários nos testes
func (m *MockUserRepository) List() ([]*entities.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

// PromoteToAdmin implementa o método PromoteToAdmin da interface UserRepository
// Permite configurar o comportamento de promoção a admin nos testes
func (m *MockUserRepository) PromoteToAdmin(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// IsFirstUser implementa o método IsFirstUser da interface UserRepository
// Permite configurar o comportamento de verificação de primeiro usuário nos testes
func (m *MockUserRepository) IsFirstUser() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// TestUserService_Create testa todos os cenários do método Create do UserService
// Verifica o comportamento do serviço ao criar usuários em diferentes situações
func TestUserService_Create(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Criar_Usuario_Com_Sucesso_Quando_Dados_Validos", func(t *testing.T) {
		// Arrange - Configuração específica para este caso de teste
		// Configura o mock para simular que o email não existe e a criação é bem-sucedida
		mockRepo.On("FindByEmail", "test@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil).Once()

		// Act - Execução da funcionalidade que está sendo testada
		userDTO := dtos.UserCreateDTO{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}
		result, err := userService.Create(userDTO)

		// Assert - Verificação dos resultados esperados
		assert.NoError(t, err, "Não deve retornar erro ao criar usuário com dados válidos")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, "Test User", result.Name, "O nome no DTO de resposta deve corresponder ao fornecido")
		assert.Equal(t, "test@example.com", result.Email, "O email no DTO de resposta deve corresponder ao fornecido")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Email_Ja_Cadastrado", func(t *testing.T) {
		// Arrange - Configura o mock para simular email já existente
		existingUser := &entities.User{
			Email: "existing@example.com",
			Name:  "Existing User",
		}
		mockRepo.On("FindByEmail", "existing@example.com").Return(existingUser, nil).Once()

		// Act - Tenta criar um usuário com email já existente
		userDTO := dtos.UserCreateDTO{
			Name:     "Duplicate Email",
			Email:    "existing@example.com",
			Password: "password123",
		}
		result, err := userService.Create(userDTO)

		// Assert - Verifica se o erro apropriado foi retornado
		assert.Error(t, err, "Deve retornar erro ao tentar criar usuário com email duplicado")
		assert.Nil(t, result, "Não deve retornar usuário quando há erro")
		assert.Contains(t, err.Error(), "email já está em uso", "A mensagem de erro deve indicar email duplicado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_No_Repositorio", func(t *testing.T) {
		// Arrange - Configura o mock para simular erro no repositório
		dbError := errors.New("database error: connection failed")
		mockRepo.On("FindByEmail", "error@example.com").Return(nil, dbError).Once()

		// Act - Executa com cenário de falha no banco
		userDTO := dtos.UserCreateDTO{
			Name:     "Error User",
			Email:    "error@example.com",
			Password: "password123",
		}
		result, err := userService.Create(userDTO)

		// Assert - Verifica se o erro do repositório é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando o repositório falha")
		assert.Nil(t, result, "Não deve retornar usuário quando há erro no repositório")
		assert.Contains(t, err.Error(), "database error", "Deve propagar o erro original do repositório")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_GetByID testa todos os cenários do método GetByID do UserService
// Verifica o comportamento do serviço ao buscar usuários por ID em diferentes situações
func TestUserService_GetByID(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Retornar_Usuario_Quando_ID_Existente", func(t *testing.T) {
		// Arrange - Configura mock para retornar um usuário específico
		mockUser := &entities.User{
			Model: gorm.Model{ID: 1},
			Name:  "Found User",
			Email: "found@example.com",
		}
		mockRepo.On("FindByID", uint(1)).Return(mockUser, nil).Once()

		// Act - Busca usuário com ID existente
		result, err := userService.GetByID(1)

		// Assert - Verifica se o usuário correto foi retornado
		assert.NoError(t, err, "Não deve retornar erro ao encontrar usuário existente")
		assert.NotNil(t, result, "Deve retornar um DTO de resposta não nulo")
		assert.Equal(t, "Found User", result.Name, "O nome deve corresponder ao usuário encontrado")
		assert.Equal(t, "found@example.com", result.Email, "O email deve corresponder ao usuário encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Configura mock para simular ID inexistente
		mockRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Busca usuário com ID inexistente
		result, err := userService.GetByID(999)

		// Assert - Verifica se o erro apropriado foi retornado
		assert.Error(t, err, "Deve retornar erro quando usuário não é encontrado")
		assert.Nil(t, result, "Não deve retornar usuário quando não encontrado")
		assert.Contains(t, err.Error(), "usuário não encontrado", "A mensagem deve indicar usuário não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_No_Acesso_Ao_Banco", func(t *testing.T) {
		// Arrange - Configura mock para simular erro de banco de dados
		dbError := errors.New("database error: connection timeout")
		mockRepo.On("FindByID", uint(2)).Return(nil, dbError).Once()

		// Act - Executa com cenário de falha no banco
		result, err := userService.GetByID(2)

		// Assert - Verifica se o erro do banco é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando há falha no banco de dados")
		assert.Nil(t, result, "Não deve retornar usuário quando há erro de banco")
		assert.Contains(t, err.Error(), "database error", "Deve propagar o erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_AuthenticateUser testa todos os cenários do método AuthenticateUser
// Verifica o comportamento do serviço ao autenticar usuários em diferentes situações
func TestUserService_AuthenticateUser(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Retornar_Erro_Quando_Email_Nao_Cadastrado", func(t *testing.T) {
		// Arrange - Configura mock para simular email não cadastrado
		mockRepo.On("FindByEmail", "notfound@example.com").Return(nil, nil).Once()

		// Act - Tenta autenticar com email inexistente
		result, err := userService.AuthenticateUser("notfound@example.com", "password123")

		// Assert - Verifica se o erro apropriado foi retornado
		assert.Error(t, err, "Deve retornar erro quando email não está cadastrado")
		assert.Nil(t, result, "Não deve retornar usuário quando autenticação falha")
		assert.Contains(t, err.Error(), "credenciais inválidas", "A mensagem deve indicar credenciais inválidas sem expor detalhes")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Consulta_DB", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco de dados
		dbError := errors.New("database error: connection dropped")
		mockRepo.On("FindByEmail", "error@example.com").Return(nil, dbError).Once()

		// Act - Tenta autenticar com erro de banco
		result, err := userService.AuthenticateUser("error@example.com", "password123")

		// Assert - Verifica se o erro do banco é propagado corretamente
		assert.Error(t, err, "Deve retornar erro quando há falha no banco de dados")
		assert.Nil(t, result, "Não deve retornar usuário quando há erro de banco")
		assert.Contains(t, err.Error(), "database error", "Deve propagar o erro original do banco")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Senha_Incorreta", func(t *testing.T) {
		// Arrange - Configura mock para retornar usuário com senha diferente
		// Precisamos usar senha real hasheada para testar bcrypt.CompareHashAndPassword
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("senha_correta"), bcrypt.DefaultCost)
		mockUser := &entities.User{
			Model:    gorm.Model{ID: 3},
			Name:     "Password Test User",
			Email:    "password@example.com",
			Password: string(hashedPassword),
		}
		mockRepo.On("FindByEmail", "password@example.com").Return(mockUser, nil).Once()

		// Act - Tenta autenticar com senha incorreta
		result, err := userService.AuthenticateUser("password@example.com", "senha_errada")

		// Assert - Verifica se o erro de senha incorreta é retornado
		assert.Error(t, err, "Deve retornar erro quando a senha está incorreta")
		assert.Nil(t, result, "Não deve retornar usuário quando a senha está incorreta")
		assert.Contains(t, err.Error(), "credenciais inválidas", "A mensagem deve indicar credenciais inválidas sem expor detalhes")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Autenticar_Com_Sucesso_Credenciais_Corretas", func(t *testing.T) {
		// Arrange - Configura mock para retornar usuário com senha correta
		senha := "senha_correta"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		mockUser := &entities.User{
			Model:    gorm.Model{ID: 4},
			Name:     "Valid User",
			Email:    "valid@example.com",
			Password: string(hashedPassword),
			IsAdmin:  true,
		}
		mockRepo.On("FindByEmail", "valid@example.com").Return(mockUser, nil).Once()

		// Act - Tenta autenticar com credenciais válidas
		result, err := userService.AuthenticateUser("valid@example.com", senha)

		// Assert - Verifica se a autenticação foi bem-sucedida
		assert.NoError(t, err, "Não deve retornar erro com credenciais corretas")
		assert.NotNil(t, result, "Deve retornar o usuário quando autenticado com sucesso")
		assert.Equal(t, uint(4), result.ID, "ID do usuário deve corresponder")
		assert.Equal(t, "valid@example.com", result.Email, "Email do usuário deve corresponder")
		assert.Equal(t, true, result.IsAdmin, "Status de admin deve corresponder")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_Update testa todos os cenários do método Update do UserService
// Verifica o comportamento do serviço ao atualizar usuários em diferentes situações
func TestUserService_Update(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Atualizar_Usuario_Com_Sucesso_Quando_Dados_Validos", func(t *testing.T) {
		// Arrange - Configura mock para simular usuário encontrado
		mockUser := &entities.User{
			Model: gorm.Model{ID: 1},
			Name:  "Nome Original",
			Email: "original@example.com",
		}
		mockRepo.On("FindByID", uint(1)).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*entities.User")).Return(nil).Once()

		// Act - Executa atualização de usuário
		updateDTO := dtos.UserUpdateDTO{
			Name: "Nome Atualizado",
		}
		result, err := userService.Update(1, updateDTO)

		// Assert - Verifica resultado da atualização
		assert.NoError(t, err, "Não deve retornar erro ao atualizar usuário")
		assert.NotNil(t, result, "Deve retornar usuário atualizado")
		assert.Equal(t, "Nome Atualizado", result.Name, "O nome deve ser atualizado")
		assert.Equal(t, "original@example.com", result.Email, "O email não deve ser alterado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Atualizar_Senha_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para simular usuário encontrado
		senhaOriginal, _ := bcrypt.GenerateFromPassword([]byte("senha_antiga"), bcrypt.DefaultCost)
		mockUser := &entities.User{
			Model:    gorm.Model{ID: 2},
			Name:     "Usuário Senha",
			Email:    "senha@example.com",
			Password: string(senhaOriginal),
		}
		mockRepo.On("FindByID", uint(2)).Return(mockUser, nil).Once()
		mockRepo.On("Update", mock.MatchedBy(func(u *entities.User) bool {
			// Verifica se a senha foi alterada (não podemos comparar diretamente por causa do hash)
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("nova_senha"))
			return err == nil
		})).Return(nil).Once()

		// Act - Executa atualização de senha
		updateDTO := dtos.UserUpdateDTO{
			Password: "nova_senha",
		}
		result, err := userService.Update(2, updateDTO)

		// Assert - Verifica resultado da atualização
		assert.NoError(t, err, "Não deve retornar erro ao atualizar senha")
		assert.NotNil(t, result, "Deve retornar usuário atualizado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Existe", func(t *testing.T) {
		// Arrange - Configura mock para simular usuário não encontrado
		mockRepo.On("FindByID", uint(999)).Return(nil, nil).Once()

		// Act - Tenta atualizar usuário inexistente
		updateDTO := dtos.UserUpdateDTO{
			Name: "Não Existe",
		}
		result, err := userService.Update(999, updateDTO)

		// Assert - Verifica que ocorreu erro
		assert.Error(t, err, "Deve retornar erro quando usuário não existe")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "usuário não encontrado", "A mensagem deve indicar usuário não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: connection failed")
		mockRepo.On("FindByID", uint(3)).Return(nil, dbError).Once()

		// Act - Executa atualização com erro de banco
		updateDTO := dtos.UserUpdateDTO{
			Name: "Erro Banco",
		}
		result, err := userService.Update(3, updateDTO)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "database error", "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Atualização", func(t *testing.T) {
		// Arrange - Configura mock para simular falha na atualização
		mockUser := &entities.User{
			Model: gorm.Model{ID: 4},
			Name:  "Falha Update",
			Email: "falhaupdate@example.com",
		}
		mockRepo.On("FindByID", uint(4)).Return(mockUser, nil).Once()
		updateError := errors.New("falha ao atualizar no banco")
		mockRepo.On("Update", mock.AnythingOfType("*entities.User")).Return(updateError).Once()

		// Act - Executa atualização com falha
		updateDTO := dtos.UserUpdateDTO{
			Name: "Nunca Atualizado",
		}
		result, err := userService.Update(4, updateDTO)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha na atualização")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "falha ao atualizar no banco", "Deve propagar erro original da atualização")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_Delete testa todos os cenários do método Delete do UserService
// Verifica o comportamento do serviço ao excluir usuários
func TestUserService_Delete(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Excluir_Usuario_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para simular exclusão bem-sucedida
		mockRepo.On("Delete", uint(1)).Return(nil).Once()

		// Act - Executa exclusão
		err := userService.Delete(1)

		// Assert - Verifica resultado da exclusão
		assert.NoError(t, err, "Não deve retornar erro ao excluir usuário existente")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Existe", func(t *testing.T) {
		// Arrange - Configura mock para simular erro de usuário não encontrado
		notFoundError := errors.New("usuário não encontrado")
		mockRepo.On("Delete", uint(999)).Return(notFoundError).Once()

		// Act - Tenta excluir usuário inexistente
		err := userService.Delete(999)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro ao excluir usuário inexistente")
		assert.Contains(t, err.Error(), "usuário não encontrado", "Deve propagar erro original da exclusão")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Exclusao", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: constraint violation")
		mockRepo.On("Delete", uint(2)).Return(dbError).Once()

		// Act - Executa exclusão com erro de banco
		err := userService.Delete(2)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Contains(t, err.Error(), "database error", "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_List testa o método List do UserService
// Verifica o comportamento do serviço ao listar todos os usuários
func TestUserService_List(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Listar_Usuarios_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista de usuários
		mockUsers := []*entities.User{
			{
				Model:   gorm.Model{ID: 1},
				Name:    "Usuário 1",
				Email:   "usuario1@example.com",
				IsAdmin: true,
			},
			{
				Model:   gorm.Model{ID: 2},
				Name:    "Usuário 2",
				Email:   "usuario2@example.com",
				IsAdmin: false,
			},
		}
		mockRepo.On("List").Return(mockUsers, nil).Once()

		// Act - Executa listagem
		result, err := userService.List()

		// Assert - Verifica resultado da listagem
		assert.NoError(t, err, "Não deve retornar erro ao listar usuários")
		assert.Len(t, result, 2, "Deve retornar 2 usuários")
		assert.Equal(t, "Usuário 1", result[0].Name, "O nome deve corresponder")
		assert.Equal(t, "usuario1@example.com", result[0].Email, "O email deve corresponder")
		assert.True(t, result[0].IsAdmin, "Status de admin deve corresponder")
		assert.Equal(t, "Usuário 2", result[1].Name, "O nome deve corresponder")
		assert.Equal(t, "usuario2@example.com", result[1].Email, "O email deve corresponder")
		assert.False(t, result[1].IsAdmin, "Status de admin deve corresponder")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Lista_Vazia_Quando_Sem_Usuarios", func(t *testing.T) {
		// Arrange - Configura mock para retornar lista vazia
		emptyList := []*entities.User{}
		mockRepo.On("List").Return(emptyList, nil).Once()

		// Act - Executa listagem
		result, err := userService.List()

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
		result, err := userService.List()

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar lista de usuários")
		assert.Contains(t, err.Error(), "database error", "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_PromoteToAdmin testa o método PromoteToAdmin do UserService
// Verifica o comportamento do serviço ao promover usuários a administradores
func TestUserService_PromoteToAdmin(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Promover_Usuario_Com_Sucesso", func(t *testing.T) {
		// Arrange - Configura mock para promoção bem-sucedida
		mockRepo.On("PromoteToAdmin", uint(1)).Return(nil).Once()
		mockUserPromoted := &entities.User{
			Model:   gorm.Model{ID: 1},
			Name:    "Usuário Promovido",
			Email:   "promovido@example.com",
			IsAdmin: true,
		}
		mockRepo.On("FindByID", uint(1)).Return(mockUserPromoted, nil).Once()

		// Act - Executa promoção
		result, err := userService.PromoteToAdmin(1)

		// Assert - Verifica resultado da promoção
		assert.NoError(t, err, "Não deve retornar erro ao promover usuário")
		assert.NotNil(t, result, "Deve retornar usuário promovido")
		assert.Equal(t, "Usuário Promovido", result.Name, "O nome deve corresponder")
		assert.Equal(t, "promovido@example.com", result.Email, "O email deve corresponder")
		assert.True(t, result.IsAdmin, "Deve estar marcado como admin")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Promocao", func(t *testing.T) {
		// Arrange - Configura mock para simular falha na promoção
		promoteError := errors.New("falha ao promover usuário")
		mockRepo.On("PromoteToAdmin", uint(2)).Return(promoteError).Once()

		// Act - Executa promoção com falha
		result, err := userService.PromoteToAdmin(2)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha na promoção")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "falha ao promover usuário", "Deve propagar erro original da promoção")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Usuario_Nao_Encontrado_Apos_Promocao", func(t *testing.T) {
		// Arrange - Configura promoção sucedida, mas usuário não encontrado depois
		mockRepo.On("PromoteToAdmin", uint(3)).Return(nil).Once()
		mockRepo.On("FindByID", uint(3)).Return(nil, nil).Once()

		// Act - Executa promoção com falha na busca posterior
		result, err := userService.PromoteToAdmin(3)

		// Assert - Verifica que o erro correto é retornado
		assert.Error(t, err, "Deve retornar erro quando usuário não encontrado após promoção")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "usuário não encontrado", "Deve indicar usuário não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca_Apos_Promocao", func(t *testing.T) {
		// Arrange - Configura promoção sucedida, mas erro na busca posterior
		mockRepo.On("PromoteToAdmin", uint(4)).Return(nil).Once()
		dbError := errors.New("database error: connection lost")
		mockRepo.On("FindByID", uint(4)).Return(nil, dbError).Once()

		// Act - Executa promoção com erro de banco após promoção
		result, err := userService.PromoteToAdmin(4)

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco após promoção")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "database error", "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserService_GetByEmail testa o método GetByEmail do UserService
// Verifica o comportamento do serviço ao buscar usuários por email
func TestUserService_GetByEmail(t *testing.T) {
	// Arrange - Configuração inicial comum para todos os testes
	mockRepo := new(MockUserRepository)
	userService := services.NewUserService(mockRepo)

	t.Run("Deve_Encontrar_Usuario_Por_Email", func(t *testing.T) {
		// Arrange - Configura mock para retornar usuário por email
		mockUser := &entities.User{
			Model:   gorm.Model{ID: 1},
			Name:    "Usuário Email",
			Email:   "emailteste@example.com",
			IsAdmin: false,
		}
		mockRepo.On("FindByEmail", "emailteste@example.com").Return(mockUser, nil).Once()

		// Act - Busca usuário por email
		result, err := userService.GetByEmail("emailteste@example.com")

		// Assert - Verifica resultado da busca
		assert.NoError(t, err, "Não deve retornar erro quando encontra usuário")
		assert.NotNil(t, result, "Deve retornar usuário encontrado")
		assert.Equal(t, "Usuário Email", result.Name, "O nome deve corresponder")
		assert.Equal(t, "emailteste@example.com", result.Email, "O email deve corresponder")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Nulo_Quando_Email_Nao_Encontrado", func(t *testing.T) {
		// Arrange - Configura mock para não encontrar o email
		mockRepo.On("FindByEmail", "naoexiste@example.com").Return(nil, nil).Once()

		// Act - Busca usuário por email inexistente
		result, err := userService.GetByEmail("naoexiste@example.com")

		// Assert - Verifica resultado da busca
		assert.NoError(t, err, "Não deve retornar erro quando email não existe")
		assert.Nil(t, result, "Deve retornar nulo quando email não encontrado")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Deve_Retornar_Erro_Quando_Falha_Na_Busca_Por_Email", func(t *testing.T) {
		// Arrange - Configura mock para simular erro no banco
		dbError := errors.New("database error: query failed")
		mockRepo.On("FindByEmail", "erro@example.com").Return(nil, dbError).Once()

		// Act - Busca com erro de banco
		result, err := userService.GetByEmail("erro@example.com")

		// Assert - Verifica que o erro é propagado
		assert.Error(t, err, "Deve retornar erro quando falha no banco")
		assert.Nil(t, result, "Não deve retornar usuário")
		assert.Contains(t, err.Error(), "database error", "Deve propagar erro original do banco")
		mockRepo.AssertExpectations(t)
	})
}
