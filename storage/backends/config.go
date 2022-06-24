package backends

import (
	"fmt"
	"melody/pkg/controllers/consts"
	"os"
	"strconv"
)

const (
	EnvDBHost     = "MYSQL_HOST"
	EnvDBPort     = "MYSQL_PORT"
	EnvDBDatabase = "MYSQL_DB_NAME"
	EnvDBUser     = "MYSQL_USER"
	EnvDBPassword = "MYSQL_PASSWORD"
	EnvLogMode    = "MYSQL_LOGMODE"
)

func GetMysqlDBSource() (dbSource, logMode string, err error) {
	host := consts.DefaultMelodyMySqlServiceName // GetEnvOrDefault(EnvDBHost, consts.DefaultMorphlingMySqlServiceName)
	port, err := strconv.Atoi(GetEnvOrDefault(EnvDBPort, consts.DefaultMelodyMySqlServicePort))
	if err != nil {
		return "", "", err
	}

	db := GetEnvOrDefault(EnvDBDatabase, "melody")
	user := GetEnvOrDefault(EnvDBUser, "root")
	password := GetEnvOrDefault(EnvDBPassword, "melody")

	// Expected: "root:morphling@tcp(morphling-mysql:3306)/morphling?timeout=5s"
	dbSource = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=35s", user, password, host, port, db)
	logMode = GetEnvOrDefault(EnvLogMode, "no")

	return dbSource, logMode, nil
}

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
