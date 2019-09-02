package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ustackq/indagate/pkg/service"

	"go.uber.org/zap"

	"github.com/ustackq/indagate/pkg/utils/errors"
)

type MemberBackend struct {
	Logger *zap.Logger

	ResourceType               service.ResourceType
	UserType                   service.UserType
	UserResourceMappingService service.UserResourceMappingService
	UserService                service.UserService
}

func newPostMemberHandler(mb MemberBackend) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		req, err := decodePostMemberRequest(r, ps)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		user, err := mb.UserService.FindUserByID(ctx, req.MemberID)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		mapping := &service.UserResourceMapping{
			ResourceID:   req.ResourceID,
			ResourceType: mb.ResourceType,
			UserID:       req.MemberID,
			UserType:     mb.UserType,
		}

		if err := mb.UserResourceMappingService.CreateUserResourceMapping(ctx, mapping); err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		if err := encodeResponse(ctx, rw, http.StatusCreated, newResourceUserResponse(user, mb.UserType)); err != nil {
			EncodeError(ctx, err, rw)
			return
		}
	}
}

type resourceUserResponse struct {
	Role service.UserType `json:"role"`
	*userResponse
}

func newResourceUserResponse(user *service.User, ut service.UserType) *resourceUserResponse {
	return &resourceUserResponse{
		Role:         ut,
		userResponse: newUserResponse(user),
	}
}

type postMemberRequest struct {
	MemberID   service.ID
	ResourceID service.ID
}

type postOrgRequest struct {
	Org *service.Organization
}

func decodePostOrgRequest(ctx context.Context, r *http.Request) (*postOrgRequest, error) {
	org := &service.Organization{}
	if err := json.NewDecoder(r.Body).Decode(org); err != nil {
		return nil, err
	}

	return &postOrgRequest{
		Org: org,
	}, nil
}

func decodePostMemberRequest(r *http.Request, ps httprouter.Params) (*postMemberRequest, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing id",
		}
	}

	var oid service.ID
	if err := oid.DecodeFromString(id); err != nil {
		return nil, err
	}

	user := &service.User{}
	if !user.ID.Valid() {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "user id missing or invalid",
		}
	}

	return &postMemberRequest{
		MemberID:   user.ID,
		ResourceID: oid,
	}, nil
}

type getMembersRequest struct {
	MemberID   service.ID
	ResourceID service.ID
}

func decodeGetMembersRequets(ctx context.Context, ps httprouter.Params) (*getMembersRequest, error) {
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

	req := &getMembersRequest{
		ResourceID: i,
	}

	return req, nil
}

type resourceUsersResponse struct {
	Links map[string]string       `json:"links"`
	Users []*resourceUserResponse `json:"users"`
}

func newResourceUsersResponse(opts service.FindOptions, filter service.UserResourceMappingFilter, users []*service.User) *resourceUsersResponse {
	rs := resourceUsersResponse{
		Links: map[string]string{
			"self": fmt.Sprintf("/api/v1/%s/%s/%ss", filter.ResourceType, filter.ResourceID, filter.UserType),
		},
		Users: make([]*resourceUserResponse, 0, len(users)),
	}

	for _, user := range users {
		rs.Users = append(rs.Users, newResourceUserResponse(user, filter.UserType))
	}
	return &rs
}

func newGetMembersHandler(mb MemberBackend) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		req, err := decodeGetMembersRequets(ctx, ps)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		filter := service.UserResourceMappingFilter{
			ResourceID:   req.ResourceID,
			ResourceType: mb.ResourceType,
			UserType:     mb.UserType,
		}

		opts := service.FindOptions{}
		mappings, _, err := mb.UserResourceMappingService.FindUserResourceMappings(ctx, filter)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		users := make([]*service.User, 0, len(mappings))
		for _, m := range mappings {
			if m.MappingType == service.OrgMappingType {
				continue
			}

			user, err := mb.UserService.FindUserByID(ctx, m.UserID)
			if err != nil {
				EncodeError(ctx, err, rw)
				return
			}

			users = append(users, user)
		}

		if err := encodeResponse(ctx, rw, http.StatusOK, newResourceUsersResponse(opts, filter, users)); err != nil {
			EncodeError(ctx, err, rw)
			return
		}

	}
}

type deleteMemberRequest struct {
	MemberID   service.ID
	ResourceID service.ID
}

func decodeDeleteMemberRequets(r *http.Request, ps httprouter.Params) (*deleteMemberRequest, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing resource id",
		}
	}
	var rid service.ID
	if err := rid.DecodeFromString(id); err != nil {
		return nil, err
	}

	uid := ps.ByName("userID")
	if uid == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing member id",
		}
	}

	var mid service.ID
	if err := mid.DecodeFromString(uid); err != nil {
		return nil, err
	}

	return &deleteMemberRequest{
		MemberID:   mid,
		ResourceID: rid,
	}, nil
}

func newDeleteMemberHandler(mb MemberBackend) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		req, err := decodeDeleteMemberRequets(r, ps)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		if err := mb.UserResourceMappingService.DeleteUserResourceMapping(ctx, req.ResourceID, req.MemberID); err != nil {
			EncodeError(ctx, err, rw)
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}
}
