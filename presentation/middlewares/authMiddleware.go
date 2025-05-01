package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	// Adicione esta importação
	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/services"
	"github.com/henrygoeszanin/api_golang_estudos/config"
)

// TokenExtractor é um middleware que extrai o token de diferentes fontes
func TokenExtractor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. Tenta obter do cookie
		token, _ = c.Cookie("jwt")
		if token == "" {
			token, _ = c.Cookie("token")
		}

		// 2. Tenta obter do header
		if token == "" {
			auth := c.GetHeader("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		// 3. Tenta obter do query parameter
		if token == "" {
			token = c.Query("token")
		}

		if token != "" {
			c.Request.Header.Set("Authorization", "Bearer "+token)
			fmt.Printf("TokenExtractor: Token encontrado\n")
		} else {
			fmt.Printf("TokenExtractor: Nenhum token encontrado\n")
		}

		c.Next()
	}
}

// AdminRequired verifica se o usuário é administrador
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		isAdmin, exists := claims["is_admin"]
		if !exists || isAdmin != true {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Acesso restrito a administradores",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Estrutura para o login
type login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SetupJWTMiddleware configura o middleware JWT
func SetupJWTMiddleware(userService *services.UserService, cfg *config.Config) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "library-api",
		Key:         []byte(cfg.JWTSecret),
		Timeout:     time.Hour * 24,     // 24 horas
		MaxRefresh:  time.Hour * 24 * 7, // 7 dias para refresh
		IdentityKey: "id",

		// Configurações de cookies
		SendCookie:     true,
		CookieName:     "jwt",
		CookieMaxAge:   24 * time.Hour,
		CookieDomain:   "",
		SecureCookie:   false,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteDefaultMode,

		// Configuração do token
		TokenLookup:   "cookie:jwt,header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		// Função para autenticar o usuário
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			fmt.Printf("Login - Tentativa para email: %s\n", loginVals.Email)

			user, err := userService.AuthenticateUser(loginVals.Email, loginVals.Password)
			if err != nil {
				fmt.Printf("Login - Falha: %v\n", err)
				return nil, jwt.ErrFailedAuthentication
			}

			fmt.Printf("Login - Sucesso para usuário: %s (ID: %d)\n", user.Email, user.ID)
			fmt.Printf("Tipo do usuário retornado: %T\n", user)

			// Garantindo que retornamos explicitamente um *dtos.UserResponseDTO
			userDTO := &dtos.UserResponseDTO{
				ID:      user.ID,
				Email:   user.Email,
				IsAdmin: user.IsAdmin,
			}

			fmt.Printf("Criando token com user info: ID=%d, Email=%s, IsAdmin=%v\n",
				userDTO.ID, userDTO.Email, userDTO.IsAdmin)

			return userDTO, nil
		},

		// Função para gerar o payload do token
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			fmt.Printf("PayloadFunc recebeu dados do tipo: %T\n", data)

			// Tente converter para UserResponseDTO
			if user, ok := data.(*dtos.UserResponseDTO); ok {
				fmt.Printf("Convertido com sucesso para UserResponseDTO: ID=%d, Email=%s\n",
					user.ID, user.Email)

				return jwt.MapClaims{
					"id":       user.ID,
					"email":    user.Email,
					"is_admin": user.IsAdmin,
				}
			}

			// Se falhar, tente converter direto do objeto retornado pelo serviço
			// Assumindo que o objeto tem campos ID, Email e IsAdmin
			fmt.Printf("AVISO: Falha na conversão para UserResponseDTO. Tentando alternativas...\n")

			// Adicione detalhes sobre o objeto recebido para debug
			fmt.Printf("Conteúdo do objeto recebido: %+v\n", data)

			return jwt.MapClaims{}
		},

		// Função para extrair a identidade do token
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			fmt.Printf("IdentityHandler - Claims recebidas: %+v\n", claims)

			// Verificar se os campos necessários existem
			idVal, idExists := claims["ID"]
			emailVal, emailExists := claims["Email"]
			isAdminVal, isAdminExists := claims["IsAdmin"]

			if !idExists || !emailExists || !isAdminExists {
				fmt.Printf("ALERTA: JWT incompleto ou inválido. Claims: %+v\n", claims)
				return nil
			}

			// Conversão segura de tipos
			var id uint
			if idFloat, ok := idVal.(float64); ok {
				id = uint(idFloat)
			} else {
				fmt.Printf("ERRO: Campo 'id' não é um número: %T = %v\n", idVal, idVal)
				return nil
			}

			return &dtos.UserResponseDTO{
				ID:      id,
				Email:   emailVal.(string),
				IsAdmin: isAdminVal.(bool),
			}
		},

		// Função para autorizar o acesso
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// Qualquer usuário autenticado está autorizado por padrão
			return true
		},

		// Função para resposta do login
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			fmt.Println("==== Login Bem-Sucedido ====")
			fmt.Printf("Token gerado: %s\n", token)
			fmt.Printf("Expira em: %v\n", expire)

			c.JSON(code, gin.H{
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		},

		// Função para erro de autenticação
		Unauthorized: func(c *gin.Context, code int, message string) {
			fmt.Printf("==== Falha na Autenticação ====\n")
			fmt.Printf("Rota: %s %s\n", c.Request.Method, c.Request.URL.Path)
			fmt.Printf("Código: %d, Mensagem: %s\n", code, message)

			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
	})
}
