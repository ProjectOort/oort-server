package account

import (
	"testing"

	"github.com/ProjectOort/oort-server/conf"
	"github.com/stretchr/testify/assert"
)

func TestTokenHelper(t *testing.T) {
	var helper *tokenHelper
	assert.NotPanics(t, func() {
		helper = newTokenHelper(&conf.Account{
			TokenKey:       "+tXuqxOhMLT5IHqjZlGhLT1rzYCIqPpoxGk0Sj5HaGk=",
			TokenExpireSec: 7200,
		})
	})
	assert.Panics(t, func() {
		newTokenHelper(&conf.Account{
			TokenKey:       "",
			TokenExpireSec: 7200,
		})
	})
	assert.Panics(t, func() {
		newTokenHelper(&conf.Account{
			TokenKey:       "123",
			TokenExpireSec: 7200,
		})
	})
	assert.NotNil(t, helper, "helper should not be nil")

	const randomText = "TEST_RANDOM_TEXT"
	token, err := helper.MakeToken(map[string]interface{}{
		"meta": randomText,
	})
	assert.NoError(t, err, "MakeToken should not failed")

	payload, err := helper.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, randomText, payload["meta"])
}
