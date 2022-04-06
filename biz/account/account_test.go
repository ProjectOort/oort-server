package account

import (
	"testing"

	"github.com/ProjectOort/oort-server/conf"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestHashPasswd(t *testing.T) {
	{
		acc := &Account{Password: "12345678"}
		assert.NoError(t, acc.HashPasswd())
	}
}

func TestPasswdEqual(t *testing.T) {
	const p1 = "12345678"
	const p2 = "87654321"

	acc := &Account{Password: p1}
	acc.HashPasswd()

	assert.True(t, acc.PasswdEqual(p1))
	assert.False(t, acc.PasswdEqual(p2))
}

func TestToken(t *testing.T) {
	{
		var helper *tokenHelper
		assert.NotPanics(t, func() {
			helper = newTokenHelper(&conf.Account{
				TokenKey:       "+tXuqxOhMLT5IHqjZlGhLT1rzYCIqPpoxGk0Sj5HaGk=",
				TokenExpireSec: 7200,
			})
		})
		sid := primitive.NewObjectID()
		acc := &Account{ID: sid}
		token, err := acc.Token(helper)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		tid, err := ValidateToken(token, helper)
		assert.NoError(t, err)
		assert.Equal(t, sid, tid)
	}
	{
		// var helper *tokenHelper
		// assert.NotPanics(t, func() {
		// 	helper = newTokenHelper(&conf.Account{
		// 		TokenKey:       "+tXuqxOhMLT5IHqjZlGhLT1rzYCIqPpoxGk0Sj5HaGk=",
		// 		TokenExpireSec: 1,
		// 	})
		// })
		// acc := &Account{ID: primitive.NewObjectID()}
		// token, err := acc.Token(helper)
		// assert.NoError(t, err)
		// assert.NotEmpty(t, token)

		// time.Sleep(2 * time.Second)
		// _, err = ValidateToken(token, helper)
		// assert.Error(t, err)
	}
}
