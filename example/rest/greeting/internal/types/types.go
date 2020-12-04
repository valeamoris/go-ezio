package types

type Request struct {
	Name string `json:"name" form:"name" path:"name"`
}

type Response struct {
	Message string `json:"message"`
}
