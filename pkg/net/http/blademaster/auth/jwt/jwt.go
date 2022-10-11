package jwt

import (
	"github.com/mapgoo-lab/atreus/pkg/net/http/blademaster/auth"
	"fmt"
	"github.com/mapgoo-lab/atreus/pkg/log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	bm "github.com/mapgoo-lab/atreus/pkg/net/http/blademaster"
)

const (

	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// bearerFormat authorization token format
	bearerFormat string = "Bearer %s"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	authorizationKey string = "Authorization"

	// reason holds the error reason.
	reason string = "UNAUTHORIZED"
)

type Config struct{
	//加密秘钥
	Secret string

	//过期时间（单位：天）
	Expire int
}

type JwtAuth struct {
	config *Config
}

func New(cfg *Config) *JwtAuth {
	return &JwtAuth{
		config: cfg,
	}
}

type MapgooClaims struct {
	UserTitle *auth.UserTitle `json:"title,omitempty"`
	Claims *jwt.RegisteredClaims `json:"claims,omitempty"`
}

func (claims *MapgooClaims) Valid() error {
	if claims.UserTitle == nil {
		return fmt.Errorf("Invalid user title")
	}

	return nil
}

//校验JWT Token
func (jwtAuth *JwtAuth) Auth(c *bm.Context) {

	auths := strings.SplitN(c.Request.Header.Get(authorizationKey), " ", 2)
	if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
		log.Error("Error header format: %s", auths)
		c.Status(401)
		return
	}
	jwtToken := auths[1]
	tokenInfo, err := jwt.ParseWithClaims(jwtToken, &MapgooClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(jwtAuth.config.Secret), nil
	})
	if err != nil {
		log.Error("ParseWithClaims failed, error:%s", err)
		c.AbortWithStatus(401)
		return
	} else if !tokenInfo.Valid {
		log.Error("tokeninfo invalid")
		c.AbortWithStatus(401)
		return
	} else if tokenInfo.Method != jwt.SigningMethodHS256 {
		//return nil, ErrUnSupportSigningMethod
		log.Error("signingMethod error")
		c.AbortWithStatus(401)
		return
	} else {
		if mgClaims, ok := tokenInfo.Claims.(*MapgooClaims); ok {
			if mgClaims.Valid() != nil {
				log.Error("mgClaims invalid")
				c.AbortWithStatus(401)
				return
			} else if !mgClaims.Claims.VerifyExpiresAt(time.Now(), true) {
				//token过期
				log.Error("token expired")
				c.AbortWithStatus(401)
				return
			} else {
				c.Set("user_title", mgClaims.UserTitle)
			}
		}
	}
}

//生成JWT Token
func (jwtAuth *JwtAuth) GetToken(userTitle *auth.UserTitle) (string, error) {
	mgClaims := &MapgooClaims{
		UserTitle: userTitle,
		Claims: &jwt.RegisteredClaims{
			//设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(jwtAuth.config.Expire) * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mgClaims)
	tokenStr, err := token.SignedString([]byte(jwtAuth.config.Secret))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(bearerFormat, tokenStr), nil
}
