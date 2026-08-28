package main

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

// init initializes the global validator instance.
func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	// Register a new rule named "alpha_start", where the first character must be a letter
	Validate.RegisterValidation("alpha_start", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()

		if len(value) == 0 {
			return false
		}

		return unicode.IsLetter(rune(value[0]))
	})
}

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=5000"`
	Tags    []string `json:"tags" validate:"required,max=5,dive,min=1,max=10"`
}

type UpdatePostPayload struct {
	Title   *string   `json:"title" validate:"omitempty,max=100"`
	Content *string   `json:"content" validate:"omitempty,max=5000"`
	Tags    *[]string `json:"tags" validate:"omitempty,dive,min=1,max=10"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=4,max=25,alpha_start"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type ReqCreateUserToken struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type ResetPassword struct {
	OldPassword string `json:"old_password" validate:"required,min=3,max=72"`
	NewPassword string `json:"new_password" validate:"required,min=3,max=72"`
}

type CommentReq struct {
	Content  string `json:"content" validate:"required,max=500"`
	ParentID *int64 `json:"parent_id" validate:"omitempty,gt=0"`
}

type UpdateProfilePayload struct {
	Bio   string   `json:"bio" validate:"omitempty,max=500"`
	Links []string `json:"links" validate:"omitempty,max=5,dive,omitempty,http_url,max=255"`
}

type ForgetPasswordPayload struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordPayload struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=3,max=72"`
}

type RenamePayload struct {
	NewName string `json:"new_name" validate:"required,min=4,max=25,alpha_start"`
}
