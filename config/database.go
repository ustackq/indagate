package config

import (
	"database/sql"
	"net/url"
	"strings"

	// Load the MySQL driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DatabaseStore struct {
	driverName     string
	dataSourceName string
	db             *sqlx.DB
}

func NewDatabaseStore(cfg string) (store Store, err error) {
	driverName, dataSourceName, err := parseDB(cfg)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	ds := &DatabaseStore{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		db:             db,
	}

	if err = initial(ds.db); err != nil {
		return nil, err
	}

	if ds.Load(); err != nil {
		return nil, err
	}

	return ds, nil
}

// initial ensure db exist and backup.
func initial(db *sqlx.DB) error {
	return nil
}

// parseDB split cfg into a driver name and data source name.
func parseDB(cfg string) (string, string, error) {
	u, err := url.Parse(cfg)
	if err != nil {
		return "", "", err
	}
	scheme := u.Scheme
	switch scheme {
	case "mysql":
		u.Scheme = ""
		cfg = strings.TrimPrefix(u.String(), "//")
	case "postgres":
		// Do nothing
	default:
		return "", "", err
	}
	return scheme, cfg, nil
}

func (ds *DatabaseStore) Load() (err error) {

	var cfgData []byte

	// query db wether exist cfg table
	row := ds.db.QueryRow("SELECT Value FROM Configurations WHERE Active")
	if err = row.Scan(&cfgData); err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(cfgData) == 0 {
		return nil
	}
	return nil
}
