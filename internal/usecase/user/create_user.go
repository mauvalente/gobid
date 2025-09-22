package user

import (
	"context"

	"github.com/mauvalente/go-bid/internal/validator"
)

type CreateUserReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.Username), "username", "this field cannot be empty")
	eval.CheckField(validator.NotBlank(req.Email), "email", "this field cannot be empty")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "must be a valid e-mail")
	eval.CheckField(validator.NotBlank(req.Bio), "bio", "this field cannot be empty")
	eval.CheckField(
		validator.MinChars(req.Bio, 10) &&
			validator.MaxChars(req.Bio, 255),
		"bio",
		"this field must have a length between 10 and 255",
	)
	eval.CheckField(
		validator.MinChars(req.Password, 8), "password", "this field must have more than 8 chars",
	)

	return eval
}
