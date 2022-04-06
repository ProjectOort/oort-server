package account

import (
	"time"

	"github.com/ProjectOort/oort-server/biz/account"
)

type Account struct {
	Token string `json:"token"`

	ID          string     `json:"id"`
	BindStatus  BindStatus `json:"bind_status"`
	AvatarURL   string     `json:"avatar_url"`
	NickName    string     `json:"nick_name"`
	Description string     `json:"description"`

	UserName string `json:"user_name"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`

	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

type BindStatus struct {
	Mobile bool `json:"mobile"`
	Email  bool `json:"email"`
	Weixin bool `json:"weixin"`
	QQ     bool `json:"qq"`
	Gitee  bool `json:"gitee"`
	GitHub bool `json:"github"`
}

func MakeAccountPresenter(acc *account.Account, token string) *Account {
	return &Account{
		Token: token,
		ID:    acc.ID.Hex(),
		BindStatus: BindStatus{
			Mobile: acc.BindStatus.Mobile,
			Email:  acc.BindStatus.Email,
			Weixin: acc.BindStatus.Weixin,
			QQ:     acc.BindStatus.QQ,
			Gitee:  acc.BindStatus.Gitee,
			GitHub: acc.BindStatus.GitHub,
		},
		AvatarURL:   acc.AvatarURL,
		NickName:    acc.NickName,
		Description: acc.Description,
		UserName:    acc.UserName,
		Mobile:      acc.Mobile,
		Email:       acc.Email,
		CreatedTime: acc.CreatedTime,
		UpdatedTime: acc.UpdatedTime,
	}
}
