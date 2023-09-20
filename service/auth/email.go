package auth

type GetEmailAuthRequest struct {
	EmailAddresss string
}

type GetEmailAuthResponce struct {
	Eid string
}

func GetEmailAuth(req *GetEmailAuthRequest, res *GetEmailAuthResponce) error {

}

type RefreshEmailAuthRequest struct {
	Eid          string
	EmailAddress string
}

func RefreshEmailAuth(req *RefreshEmailAuthRequest) error {

}

type CheckEmailAuthRequest struct {
	Eid  string
	Code string
}

type CheckEmailAuthRsponce struct {
	Success   bool
	AuthToken string
}

func CheckEmailAuth(req *CheckEmailAuthRequest, res *CheckEmailAuthRsponce) error {

}
