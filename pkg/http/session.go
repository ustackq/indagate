package http

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"net/http"

	"go.uber.org/zap"
)

const cookieSessionName = "session"

func decodeCookieSession(ctx context.Context, r *http.Request) (string, *errors.Error) {
	c, err := r.Cookie(cookieSessionName)
	if err != nil {
		return "", &errors.Error{
			Err:  err,
			Code: errors.Invalid,
		}
	}
	return c.Value, nil
}

// SessionBackend is all services authz module
type SessionBackend struct {
	Logger *zap.Logger

	PasswordsService service.PasswordsService
	SessionService   service.SessionService
}

func NewSessionBackend(ab *APIBackend) *SessionBackend {
	return &SessionBackend{
		Logger: ab.Logger.With(zap.String("handler", "session")),

		PasswordsService: ab.PasswordsService,
		SessionService:   ab.SessionService,
	}
}

type SessionHandler struct {
	*httprouter.Router
	Logger *zap.Logger

	PasswordsService service.PasswordsService
	SessionService   service.SessionService
}

func NewSessionHandler(sb *SessionBackend) *SessionHandler {
	sh := &SessionHandler{
		Router: NewRouter(),
		Logger: sb.Logger,

		PasswordsService: sb.PasswordsService,
		SessionService:   sb.SessionService,
	}

	sh.HandlerFunc(http.MethodPost, "/api/v1/signin", sh.handleSignin)
	sh.HandlerFunc(http.MethodGet, "/api/v1/signout", sh.handleSignout)

	return sh
}

type signinRequest struct {
	Username string
	Password string
}

func decodeSigninRequets(ctx context.Context, r *http.Request) (*signinRequest, error) {
	u, p, ok := r.BasicAuth()
	if !ok {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "invalid basic auth",
		}
	}

	return &signinRequest{
		Username: u,
		Password: p,
	}, nil
}

func encodeCookieSession(rw http.ResponseWriter, s *service.Session) {
	c := &http.Cookie{
		Name:  cookieSessionName,
		Value: s.Key,
	}

	http.SetCookie(rw, c)
}

func (sh *SessionHandler) handleSignin(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := decodeSigninRequets(ctx, r)
	if err != nil {
		UnauthorizedError(ctx, rw)
		return
	}

	if err := sh.PasswordsService.ComparePassword(ctx, req.Username, req.Password); err != nil {
		UnauthorizedError(ctx, rw)
		return
	}

	s, e := sh.SessionService.CreateSession(ctx, req.Username)
	if e != nil {
		UnauthorizedError(ctx, rw)
		return
	}

	encodeCookieSession(rw, s)
	rw.WriteHeader(http.StatusNoContent)

}

type signoutRequets struct {
	Key string
}

func decodeSignoutRequest(ctx context.Context, r *http.Request) (*signoutRequets, *errors.Error) {
	key, err := decodeCookieSession(ctx, r)
	if err != nil {
		return nil, err
	}

	return &signoutRequets{
		Key: key,
	}, nil
}

func (sh *SessionHandler) handleSignout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeSignoutRequest(ctx, r)
	if err != nil {
		UnauthorizedError(ctx, rw)
		return
	}

	if err := sh.SessionService.ExpireSession(ctx, req.Key); err != nil {
		UnauthorizedError(ctx, rw)
		return
	}

	http.Redirect(rw, r, "/login", http.StatusFound)
}
