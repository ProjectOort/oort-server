package repo

import (
	"context"

	"github.com/ProjectOort/oort-server/biz/account"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// compile-time interface implementation check.
var _ account.Repo = (*AccountRepo)(nil)

type AccountRepo struct {
	_mongo *mongo.Database
}

// NewAccountRepo creates a AccountRepo object.
func NewAccountRepo(_mongo *mongo.Database) *AccountRepo {
	return &AccountRepo{_mongo: _mongo}
}

// Create a new account record for the repository. And return an error if occurred.
func (r *AccountRepo) Create(ctx context.Context, account *account.Account) (err error) {
	_, err = r._mongo.Collection("account").InsertOne(ctx, account)
	return
}

// GetByGiteeID finds an account record that matches the given Gitee ID.
func (r *AccountRepo) GetByGiteeID(ctx context.Context, gid int) (acc *account.Account, err error) {
	err = r._mongo.Collection("account").FindOne(ctx, bson.D{
		{"state", true},
		{"gitee_id", gid},
	}).Decode(&acc)
	return acc, err
}

// GetByUserName finds an account record that matches the given user_name.
func (r *AccountRepo) GetByUserName(ctx context.Context, uname string) (acc *account.Account, err error) {
	acc = new(account.Account)
	err = r._mongo.Collection("account").FindOne(ctx, bson.D{
		{"state", true},
		{"user_name", uname},
	}).Decode(&acc)
	return
}

// GetByEmail finds an account record that matches the given email.
func (r *AccountRepo) GetByEmail(ctx context.Context, email string) (acc *account.Account, err error) {
	acc = new(account.Account)
	err = r._mongo.Collection("email").FindOne(ctx, bson.D{
		{"state", true},
		{"bind_status", bson.D{{"email", true}}},
		{"email", email},
	}).Decode(&acc)
	return
}

// GetByMobile  finds an account record that matches the given mobile.
func (r *AccountRepo) GetByMobile(ctx context.Context, mobile string) (acc *account.Account, err error) {
	acc = new(account.Account)
	err = r._mongo.Collection("mobile").FindOne(ctx, bson.D{
		{"status", true},
		{"bind_status", bson.D{{"mobile", true}}},
		{"mobile", mobile},
	}).Decode(&acc)
	return
}
