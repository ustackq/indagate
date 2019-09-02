package http

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"

	"github.com/julienschmidt/httprouter"
	icontext "github.com/ustackq/indagate/pkg/context"
	"github.com/ustackq/indagate/pkg/service"

	"github.com/ustackq/indagate/pkg/utils/errors"
)

type UserBackend struct {
	Logger *zap.Logger

	UserService service.UserService
	// TODO add user service log
	PasswordsService service.PasswordsService
}

const (
	usersPath         = "/api/v1/users"
	usersIDPath       = "/api/v1/users/:id"
	usersPasswordPath = "/api/v1/users/:id/password"

	mePath         = "/api/v1/me"
	mePasswordPath = "/api/v1/me/password"
)

func NewUserBackend(ab *APIBackend) *UserBackend {
	return &UserBackend{
		Logger:           ab.Logger.With(zap.String("handler", "user")),
		UserService:      ab.UserService,
		PasswordsService: ab.PasswordsService,
	}
}

type UserHandler struct {
	*httprouter.Router
	Logger           *zap.Logger
	UserService      service.UserService
	PasswordsService service.PasswordsService
}

func NewUserHandler(ab *UserBackend) *UserHandler {
	uh := &UserHandler{
		Router: httprouter.New(),
		Logger: ab.Logger,

		UserService:      ab.UserService,
		PasswordsService: ab.PasswordsService,
	}

	uh.POST(usersPath, uh.handlePostUser)
	uh.GET(usersPath, uh.handleGetUsers)
	uh.GET(usersIDPath, uh.handleGetUser)
	uh.PATCH(usersIDPath, uh.handlePatchUser)
	uh.DELETE(usersIDPath, uh.handleDeleteUser)
	uh.PUT(usersPasswordPath, uh.handlePutUserPassword)

	uh.GET(mePath, uh.handleGetMe)
	uh.PUT(mePasswordPath, uh.handlePutUserPassword)
	return uh
}

type postUserRequest struct {
	User *service.User
}

func decodePostUserRequest(r *http.Request, ps httprouter.Params) (*postUserRequest, error) {
	u := &service.User{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return nil, err
	}

	return &postUserRequest{
		User: u,
	}, nil
}

func (uh *UserHandler) handlePostUser(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodePostUserRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := uh.UserService.CreateUser(ctx, req.User); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusCreated, newUserResponse(req.User)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

type getUsersRequest struct {
	filter service.UserFilter
}

func decodeGetUsersRequest(ctx context.Context, r *http.Request) (*getUsersRequest, error) {
	query := r.URL.Query()
	req := &getUsersRequest{}

	if userID := query.Get("id"); userID != "" {
		id, err := service.IDFromString(userID)
		if err != nil {
			return nil, err
		}

		req.filter.ID = id
	}

	if name := query.Get("name"); name != "" {
		req.filter.Name = &name
	}

	return req, nil
}

func (uh *UserHandler) handleGetUsers(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeGetUsersRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	users, _, err := uh.UserService.FindUsers(ctx, req.filter)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newUsersResponse(users)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

type userRequest struct {
	UserID service.ID
	Method string
}

func decodeUserRequest(r *http.Request, ps httprouter.Params) (*userRequest, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing id",
		}
	}

	var i service.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	return &userRequest{
		UserID: i,
		Method: r.Method,
	}, nil
}

func (uh *UserHandler) handleGetUser(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeUserRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	user, err := uh.UserService.FindUserByID(ctx, req.UserID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newUserResponse(user)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

}

type userResponse struct {
	Links map[string]string `json:"links"`
	service.User
}

func newUserResponse(user *service.User) *userResponse {
	return &userResponse{
		Links: map[string]string{
			"self": fmt.Sprintf("/api/v1/users/%s", user.ID),
			// TODO: add logs
		},
		User: *user,
	}
}

type usersResponse struct {
	Links map[string]string `json:"links"`
	Users []*userResponse
}

func newUsersResponse(users []*service.User) *usersResponse {
	res := &usersResponse{
		Links: map[string]string{
			"self": "/api/v1/users",
		},
		Users: make([]*userResponse, len(users)),
	}

	for _, user := range users {
		res.Users = append(res.Users, newUserResponse(user))
	}
	return res
}

type patchUserRequest struct {
	Update service.UserUpdate
	UserID service.ID
}

func decodePatchUserRequest(r *http.Request, ps httprouter.Params) (*patchUserRequest, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing id",
		}
	}
	var i service.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	var update service.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return nil, err
	}

	return &patchUserRequest{
		Update: update,
		UserID: i,
	}, nil
}

func (uh *UserHandler) handlePatchUser(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodePatchUserRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	user, err := uh.UserService.UpdateUser(ctx, req.UserID, req.Update)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newUserResponse(user)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

func (uh *UserHandler) handleDeleteUser(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	req, err := decodeUserRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := uh.UserService.DeleteUser(ctx, req.UserID); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

type passwordResetRequest struct {
	Username    string
	PasswordOld string
	PasswordNew string
}

type passwordResetRequestBody struct {
	Password string `json:"password"`
}

func (uh *UserHandler) handlePutUserPassword(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	username, old, ok := r.BasicAuth()
	if !ok {
		EncodeError(ctx, &errors.Error{
			Code: errors.Invalid,
			Msg:  "invalid basic auth",
		}, rw)
		return
	}

	pr := &passwordResetRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(pr); err != nil {
		EncodeError(ctx, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}, rw)
		return
	}

	if err := uh.PasswordsService.CompareAndSetPassword(ctx, username, old, pr.Password); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) handleGetMe(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	auth, err := icontext.GetAuthorizer(ctx)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	var id service.ID
	switch t := auth.(type) {
	case *service.Session:
		id = t.UserID
	case *service.Authorization:
		id = t.UserID
	}

	user, err := uh.UserService.FindUserByID(ctx, id)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newUserResponse(user)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}
