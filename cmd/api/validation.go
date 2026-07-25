package main

import "github.com/go-playground/validator/v10"

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags" validate:"max=5"`
}

type UpdatePostPayload struct {
	Title   *string   `json:"title" validate:"omitempty,max=100"`
	Content *string   `json:"content" validate:"omitempty,max=1000"`
	Tags    *[]string `json:"tags" validate:"omitempty"`
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type ReqCreateUserToken struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}
