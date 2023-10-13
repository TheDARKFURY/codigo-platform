package service

import "fmt"

type Response struct {
	Message string `json:"message"`
}

type BadRequest struct {
	Message string
}

func (v *BadRequest) Error() string {
	return fmt.Sprintf(v.Message)
}

type UnauthorizedRequest struct {
	Message string
}

func (v *UnauthorizedRequest) Error() string {
	return fmt.Sprintf(v.Message)
}

type ServerError struct {
	Message string
}

func (v *ServerError) Error() string {
	return fmt.Sprintf(v.Message)
}

type UnknownError struct {
	Message string
}

func (v *UnknownError) Error() string {
	return fmt.Sprintf(v.Message)
}
