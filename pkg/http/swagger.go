package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	ierrors "github.com/ustackq/indagate/pkg/utils/errors"
	"go.uber.org/zap"
)

var _ http.Handler = (*swaggerLoader)(nil)

type swaggerLoader struct {
	logger *zap.Logger
	// ensure call initialize once.
	once sync.Once
	// the swagger converted yaml to json.
	content []byte
	// the err loading the swagger asset.
	err error
}

func newSwaggerLoader(logger *zap.Logger) *swaggerLoader {
	return &swaggerLoader{logger: logger}
}

// The functions defined in this file are placeholders when the binary is compiled
// without assets.

// Asset returns an error stating no assets were included in the binary.
func Asset(string) ([]byte, error) {
	return nil, errors.New("no assets included in binary")
}

func (sl *swaggerLoader) init() {
	swagger, err := sl.asset(Asset("swagger.yml"))
	if err != nil {
		sl.err = err
		return
	}
	content, err := yaml.YAMLToJSON(swagger)
	if err != nil {
		sl.err = err
	} else {
		sl.content = content
	}
}

func (sl *swaggerLoader) asset(data []byte, _ error) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}

	path := findSwaggerPath(sl.logger)
	if path == "" {
		return nil, errors.New("this binary not built with assets, and could not locate swagger.yml")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		sl.logger.Warn("Unable to load swagger.yml", zap.String("path", path), zap.Error(err))
		return nil, errors.New("this binary not built with assets, and could not locate swagger.yml")
	}

	sl.logger.Info("Loaded swagger.yml successfully.", zap.String("path", path))
	return b, nil

}
func findSwaggerPath(logger *zap.Logger) string {
	// 1. get path from environment
	path := os.Getenv("Indagate_SWAGGER_PATH")
	if path != "" {
		return path
	}
	logger.Info("Indagate_SWAGGER_PATH not set; falling back to checking relative paths")
	// 2. get current process execute path
	execPath, err := os.Executable()
	if err != nil {
		logger.Info("Can not detect path of indagate running path", zap.Error(err))
		return ""
	}
	execDir := filepath.Dir(execPath)
	path = filepath.Join(execDir, "..", "..", "http", "swagger.yml")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	path = filepath.Join(execDir, "http", "swagger.yml")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	wd, err := os.Getwd()
	if err == nil {
		path := filepath.Join(wd, "http", "swagger.yml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	logger.Warn("Could not find path to swagger definition.")
	return ""
}

func (sl *swaggerLoader) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	sl.once.Do(sl.init)
	if sl.err != nil {
		EncodeError(r.Context(), &ierrors.Error{
			Err:  sl.err,
			Msg:  "current binary not built with assets",
			Code: ierrors.Internal,
		}, rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(sl.content)
}
