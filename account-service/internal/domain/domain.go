package domain

import "time"

type Key string

const (
	Logger Key = "logger"

	RequestID Key = "request_id"
	Uuid      Key = "uuid"
)

type User struct {
	Guid string `json:"id"`
	Name string `json:"name"`

	Surname    string `json:"surname"`
	Avatar_url string `json:"avatar_url"`

	Age     int32  `json:"age"`
	Country string `json:"country"`
}

type Business struct {
	Guid string
	Name string
}

const (
	RedisTLl = 15 * time.Minute
)
