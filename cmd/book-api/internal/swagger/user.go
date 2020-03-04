package swagger

import "github.com/book-library/internal/users"

// HTTP status code 200 and user model in data
// swagger:response userResp
type swaggUserResp struct {
	// in:body
	Body struct {
		// HTTP status code 200
		Code int `json:"code"`
		// User model
		Data users.User `json:"data"`
	}
}

// HTTP status code 200 and an array of user models in data
// swagger:response usersResp
type swaggUsersResp struct {
	// in:body
	Body struct {
		// HTTP status code 200 - Status OK
		Code int `json:"code"`
		// Array of user models
		Data []users.User `json:"data"`
	}
}
