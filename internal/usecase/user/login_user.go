package user

import (
	"context"

	"github.com/mauvalente/go-bid/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "email is not valid")
	eval.CheckField(validator.NotBlank(req.Password), "password", "this field cannot be blanh")

	return eval
}
