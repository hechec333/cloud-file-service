package auth

// Auth Server

// 发布的token 的类别有 1.login 登录令牌 2.grant 授权令牌

// login 代表系统内用户访问，如果是资源归属用户，全部权限，如果是其他用户，按照权限配置访问
// grant 代表外部授权访问，由token中指定的访问位控制权限， r w c d
type GetLoginAuthRequest struct {
	UserId int
}

type GetLoginAuthResponce struct {
	AccessToken  string
	ExpireIn     int64
	RefreshToken string
}

// 根据userId 生成token，accesstoken是短时间有效
// @rpc
func GetLoginAuth(req *GetLoginAuthRequest, res *GetLoginAuthResponce) error {

}

type RefreshLoginAuthRequest struct {
	RefreshToken string
}

type RefreshLoginAuthResponce struct {
	AccessToken  string
	RefreshToken string
}

// refreshToken 也会刷新
// @api
func RefreshLoginAuth(req *RefreshLoginAuthRequest, res *RefreshLoginAuthResponce) error {

}

type GetAuthGrantRequest struct {
	GrantType    string `form:"grant_type" binding:"required"` //默认 authorization_code
	AuthCode     string `form:"code" binding:"required"`
	ClientId     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" binding:"required"`
}

type GetAuthGrantResponce struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expire       int
	Scope        string
}

func GetAuthGrant(req *GetAuthGrantRequest, res *GetAuthGrantResponce) error {

}

type GetAuthorizationPermitRequest struct {
	ClientId     string `form:"client_id" binding:"required"`
	ResponceType string `form:"responce_type" binding:"required"`
	RedirectUrl  string `form:"redirect_url" binding:"required"`
	Scope        string `form:"scope" binding:"required"`
	State        string `form:"state" binding:"required"`
}
type GetAuthorizationPermitResponce struct {
	AuthCode string
	State    string
}

func GetAuthorizationPermit() error {

}

type AddCredentialsClientRequest struct {
	RedirectUrl string
}

type AddCredentialsClientResponce struct {
	ClientId     string
	ClientSecret string
}

func AddCredentialsClient() error {

}
