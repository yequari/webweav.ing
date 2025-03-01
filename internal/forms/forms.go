package forms

import "git.32bit.cafe/32bitcafe/guestbook/internal/validator"

type UserRegistrationForm struct {
    Name        string  `schema:"username"`
    Email       string  `schema:"email"`
    Password    string  `schema:"password"`
    validator.Validator `schema:"-"`
}

type UserLoginForm struct {
    Email       string  `schema:"email"`
    Password    string  `schema:"password"`
    validator.Validator `schema:"-"`
}

