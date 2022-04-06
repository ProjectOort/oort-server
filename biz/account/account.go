package account

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID          primitive.ObjectID `bson:"_id"`
	State       bool               `bson:"state"`
	BindStatus  BindStatus         `bson:"bind_status"`
	CreatedTime time.Time          `bson:"created_time"`
	UpdatedTime time.Time          `bson:"updated_time"`

	NickName    string `bson:"nick_name"`
	AvatarURL   string `bson:"avatar_url"`
	Description string `bson:"description"`

	UserName string `bson:"user_name"`
	Mobile   string `bson:"mobile"`
	Email    string `bson:"email"`
	Password string `bson:"password"`

	GiteeID int `bson:"gitee_id"`
}

type BindStatus struct {
	Mobile bool `bson:"mobile"`
	Email  bool `bson:"email"`
	Weixin bool `bson:"weixin"`
	QQ     bool `bson:"qq"`
	Gitee  bool `bson:"gitee"`
	GitHub bool `bson:"github"`
}

func (x *Account) HashPasswd() error {
	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(x.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	x.Password = string(hashedPasswd)
	return nil
}

func (x *Account) PasswdEqual(otherPasswd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(x.Password), []byte(otherPasswd)) == nil
}

const (
	_AccountIDPayloadKey = "aid"
)

type TokenMaker interface {
	MakeToken(payload map[string]interface{}) (string, error)
}

func (x *Account) Token(maker TokenMaker) (string, error) {
	return maker.MakeToken(map[string]interface{}{
		_AccountIDPayloadKey: x.ID.Hex(),
	})
}

type TokenValidator interface {
	ValidateToken(token string) (map[string]interface{}, error)
}

func ValidateToken(token string, validator TokenValidator) (primitive.ObjectID, error) {
	var accID primitive.ObjectID
	payload, err := validator.ValidateToken(token)
	if err != nil {
		return accID, err
	}
	accIDHex, ok := payload[_AccountIDPayloadKey].(string)
	if !ok {
		return accID, errors.New("invalid token")
	}
	accID, err = primitive.ObjectIDFromHex(accIDHex)
	if err != nil {
		return accID, errors.New("invalid token")
	}
	return accID, nil
}
