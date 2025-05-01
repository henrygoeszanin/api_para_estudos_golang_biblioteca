package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	// Importações internas do projeto
	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
	"github.com/henrygoeszanin/api_golang_estudos/config"
)

// TokenExtractor é um middleware que extrai o token JWT de diferentes fontes na requisição
// seguindo uma ordem de prioridade: cookies, header de autorização e por fim query parameters.
// Uma vez encontrado, o token é adicionado ao header Authorization para processamento pelos
// middlewares subsequentes.
func TokenExtractor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. Tenta obter do cookie - primeira fonte de prioridade
		token, _ = c.Cookie("jwt")
		if token == "" {
			token, _ = c.Cookie("token") // fallback para cookie alternativo
		}

		// 2. Tenta obter do header Authorization - segunda fonte de prioridade
		if token == "" {
			auth := c.GetHeader("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				// Remove o prefixo "Bearer " para extrair apenas o token
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		// 3. Tenta obter de query parameter - última fonte de prioridade
		if token == "" {
			token = c.Query("token")
		}

		// Se um token foi encontrado em qualquer fonte, adicionamos ao header de Autorização
		// para que possa ser processado pelos middlewares de autenticação JWT
		if token != "" {
			c.Request.Header.Set("Authorization", "Bearer "+token)
			fmt.Printf("TokenExtractor: Token encontrado\n")
		} else {
			fmt.Printf("TokenExtractor: Nenhum token encontrado\n")
		}

		// Continua a execução da cadeia de middlewares
		c.Next()
	}
}

// AdminRequired é um middleware que verifica se o usuário autenticado possui privilégios de administrador
// através da propriedade "is_admin" nas claims do JWT. Caso contrário, a requisição é abortada com
// status 403 Forbidden.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extrai as claims do token JWT da requisição atual
		claims := jwt.ExtractClaims(c)

		// Verifica se a claim "is_admin" existe e se seu valor é true
		isAdmin, exists := claims["is_admin"]
		if !exists || isAdmin != true {
			// Retorna erro 403 para usuários não administradores
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Acesso restrito a administradores",
			})
			c.Abort() // Interrompe a execução de middlewares subsequentes
			return
		}

		// Se for admin, permite a continuação da requisição
		c.Next()
	}
}

// login define a estrutura esperada para as requisições de autenticação
// contendo email e senha do usuário. As tags de binding garantem validação
// básica dos campos.
type login struct {
	Email    string `json:"email" binding:"required,email"` // Email validado pelo formato
	Password string `json:"password" binding:"required"`    // Senha obrigatória
}

// SetupJWTMiddleware configura e retorna uma instância do middleware JWT para autenticação
// recebendo uma instância do serviço de usuário e configurações do sistema.
// Este middleware gerencia todo o ciclo de vida do JWT incluindo autenticação, geração,
// verificação e renovação de tokens.
func SetupJWTMiddleware(userService *services.UserService, cfg *config.Config) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "library-api",         // Nome do domínio de autenticação
		Key:         []byte(cfg.JWTSecret), // Chave secreta para assinatura dos tokens
		Timeout:     time.Hour * 24,        // Duração de validade do token: 24 horas
		MaxRefresh:  time.Hour * 24 * 7,    // Período máximo em que o token pode ser renovado: 7 dias
		IdentityKey: "id",                  // Chave que identifica o usuário nas claims

		// Configurações de cookies de autenticação
		SendCookie:     true,                     // Envia token como cookie
		CookieName:     "jwt",                    // Nome do cookie
		CookieMaxAge:   24 * time.Hour,           // Tempo de vida do cookie
		CookieDomain:   "",                       // Domínio do cookie (vazio = domínio atual)
		SecureCookie:   false,                    // Cookie não requer HTTPS (alterar para true em produção)
		CookieHTTPOnly: true,                     // Cookie não acessível via JavaScript (proteção XSS)
		CookieSameSite: http.SameSiteDefaultMode, // Política de SameSite do cookie

		// Configuração de lookup do token
		TokenLookup:   "cookie:jwt,header:Authorization", // Ordem de busca do token
		TokenHeadName: "Bearer",                          // Prefixo esperado no header
		TimeFunc:      time.Now,                          // Função para obter a hora atual

		// Authenticator: função responsável por validar as credenciais do usuário
		// e retornar os dados do usuário para geração do token JWT
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			// Lê e valida os dados de login do corpo da requisição
			if err := c.ShouldBind(&loginVals); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			fmt.Printf("Login - Tentativa para email: %s\n", loginVals.Email)

			// Autentica o usuário usando o serviço
			user, err := userService.AuthenticateUser(loginVals.Email, loginVals.Password)
			if err != nil {
				fmt.Printf("Login - Falha: %v\n", err)
				return nil, jwt.ErrFailedAuthentication
			}

			fmt.Printf("Login - Sucesso para usuário: %s (ID: %d)\n", user.Email, user.ID)
			fmt.Printf("Tipo do usuário retornado: %T\n", user)

			// Cria um DTO com os dados necessários para o token JWT
			userDTO := &dtos.UserResponseDTO{
				ID:      user.ID,
				Email:   user.Email,
				IsAdmin: user.IsAdmin,
			}

			fmt.Printf("Criando token com user info: ID=%d, Email=%s, IsAdmin=%v\n",
				userDTO.ID, userDTO.Email, userDTO.IsAdmin)

			return userDTO, nil
		},

		// PayloadFunc: função que define quais dados do usuário serão incluídos no token JWT
		// como claims personalizadas
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			fmt.Printf("PayloadFunc recebeu dados do tipo: %T\n", data)

			// Tenta converter o objeto para o tipo esperado
			if user, ok := data.(*dtos.UserResponseDTO); ok {
				fmt.Printf("Convertido com sucesso para UserResponseDTO: ID=%d, Email=%s\n",
					user.ID, user.Email)

				// Define as claims que serão adicionadas ao token JWT
				return jwt.MapClaims{
					"id":       user.ID,
					"email":    user.Email,
					"is_admin": user.IsAdmin,
				}
			}

			// Tratamento de erro para tipo inesperado
			fmt.Printf("AVISO: Falha na conversão para UserResponseDTO. Tentando alternativas...\n")
			fmt.Printf("Conteúdo do objeto recebido: %+v\n", data)

			// Retorna claims vazias se a conversão falhar
			return jwt.MapClaims{}
		},

		// IdentityHandler: extrai a identidade do usuário das claims do JWT
		// durante a validação do token
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			fmt.Printf("IdentityHandler - Claims recebidas: %+v\n", claims)

			// Verifica se todas as claims necessárias estão presentes
			idVal, idExists := claims["ID"]
			emailVal, emailExists := claims["Email"]
			isAdminVal, isAdminExists := claims["IsAdmin"]

			if !idExists || !emailExists || !isAdminExists {
				fmt.Printf("ALERTA: JWT incompleto ou inválido. Claims: %+v\n", claims)
				return nil
			}

			// Conversão segura do ID para o tipo correto
			var id uint
			if idFloat, ok := idVal.(float64); ok {
				id = uint(idFloat)
			} else {
				fmt.Printf("ERRO: Campo 'id' não é um número: %T = %v\n", idVal, idVal)
				return nil
			}

			// Reconstrói o objeto de usuário a partir das claims
			return &dtos.UserResponseDTO{
				ID:      id,
				Email:   emailVal.(string),
				IsAdmin: isAdminVal.(bool),
			}
		},

		// Authorizator: define regras de autorização após a autenticação
		// Aqui estamos permitindo acesso a qualquer usuário autenticado
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// Nesta implementação básica, qualquer usuário autenticado está autorizado
			// Para regras de autorização específicas, este é o lugar para implementá-las
			return true
		},

		// LoginResponse: personaliza a resposta HTTP em caso de login bem-sucedido
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			fmt.Println("==== Login Bem-Sucedido ====")
			fmt.Printf("Token gerado: %s\n", token)
			fmt.Printf("Expira em: %v\n", expire)

			// Retorna o token e sua data de expiração no corpo da resposta
			c.JSON(code, gin.H{
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		},

		// Unauthorized: personaliza a resposta HTTP em caso de falha na autenticação
		Unauthorized: func(c *gin.Context, code int, message string) {
			fmt.Printf("==== Falha na Autenticação ====\n")
			fmt.Printf("Rota: %s %s\n", c.Request.Method, c.Request.URL.Path)
			fmt.Printf("Código: %d, Mensagem: %s\n", code, message)

			// Retorna um erro JSON com o código e mensagem apropriados
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
	})
}
