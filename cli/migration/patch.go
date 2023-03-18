package migration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	nurl "net/url"
	"os"
	"strconv"
	"strings"
)
import "github.com/go-sql-driver/mysql"

var patchCli = &cobra.Command{
	Use: "patch",
	Run: func(cmd *cobra.Command, args []string) {
		patchDatabase(databaseUrl)
	},
	Args: cobra.MaximumNArgs(1),
}

func patchDatabase(databaseUrl string)  {
	config, err := urlToMySQLConfig(databaseUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	query := `SHOW TABLES LIKE 'flyway_schema_history'`
	var result string
	if err := conn.QueryRowContext(context.Background(), query).Scan(&result); err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println("not table flyway_schema_history; exit 0")
			os.Exit(0)
		}
	}

	var rowNum int
	query = "SELECT count(*) as rowNum FROM flyway_schema_history"
	if err := conn.QueryRowContext(context.Background(), query).Scan(&rowNum); err != nil {
		fmt.Println("count flyway_schema_history error", err)
		os.Exit(1)
	}
	fmt.Println("count of flyway_schema_history is ", rowNum)
	query = `SHOW TABLES LIKE 'schema_migrations'`
	if err := conn.QueryRowContext(context.Background(), query).Scan(&result); err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			os.Exit(1)
		} else {
			query = "CREATE TABLE `" + "schema_migrations" + "` (version bigint not null primary key, dirty boolean not null)"
			if _, err := conn.ExecContext(context.Background(), query); err != nil {
				fmt.Println("fail to create table schema_migrations, error: ", err)
				os.Exit(1)
			}
		}
	}

	var version, dirty int
	query = "SELECT version, dirty FROM `" + "schema_migrations" + "` LIMIT 1"
	err = conn.QueryRowContext(context.Background(), query).Scan(&version, &dirty)
	switch {
	case err == sql.ErrNoRows:
		query = "INSERT INTO `" + "schema_migrations" + "` (version, dirty) VALUES (?, ?)"
		fmt.Println("update ", rowNum-1, dirty)
		_, err = conn.ExecContext(context.Background(), query, rowNum-1, 0)
		if err != nil {
			fmt.Println("fail to init schema_migrations, error: ", err)
			os.Exit(1)
		}
	case err != nil:
		fmt.Println(err)
		os.Exit(1)

	default:
		fmt.Println("exist version: ", version, "dirty: ", dirty)
		fmt.Println("update ", rowNum - 1, dirty)
		if rowNum -1 > version {
			query = "UPDATE `schema_migrations` SET version=?"
			conn.ExecContext(context.Background(), query, rowNum-1)
			if err != nil {
				fmt.Println("fail to update version schema_migrations, error: ", err)
				os.Exit(1)
			}
		}
	}
}

func readBool(input string) (value bool, valid bool) {
	switch input {
	case "1", "true", "TRUE", "True":
		return true, true
	case "0", "false", "FALSE", "False":
		return false, true
	}

	// Not a valid bool value
	return
}

func urlToMySQLConfig(url string) (*mysql.Config, error) {
	// Need to parse out custom TLS parameters and call
	// mysql.RegisterTLSConfig() before mysql.ParseDSN() is called
	// which consumes the registered tls.Config
	// Fixes: https://github.com/golang-migrate/migrate/issues/411
	//
	// Can't use url.Parse() since it fails to parse MySQL DSNs
	// mysql.ParseDSN() also searches for "?" to find query parameters:
	// https://github.com/go-sql-driver/mysql/blob/46351a8/dsn.go#L344
	if idx := strings.LastIndex(url, "?"); idx > 0 {
		rawParams := url[idx+1:]
		parsedParams, err := nurl.ParseQuery(rawParams)
		if err != nil {
			return nil, err
		}

		ctls := parsedParams.Get("tls")
		if len(ctls) > 0 {
			if _, isBool := readBool(ctls); !isBool && strings.ToLower(ctls) != "skip-verify" {
				rootCertPool := x509.NewCertPool()
				pem, err := ioutil.ReadFile(parsedParams.Get("x-tls-ca"))
				if err != nil {
					return nil, err
				}

				if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
					return nil, fmt.Errorf("ErrAppendPEM")
				}

				clientCert := make([]tls.Certificate, 0, 1)
				if ccert, ckey := parsedParams.Get("x-tls-cert"), parsedParams.Get("x-tls-key"); ccert != "" || ckey != "" {
					if ccert == "" || ckey == "" {
						return nil, fmt.Errorf("ErrTLSCertKeyConfig")
					}
					certs, err := tls.LoadX509KeyPair(ccert, ckey)
					if err != nil {
						return nil, err
					}
					clientCert = append(clientCert, certs)
				}

				insecureSkipVerify := false
				insecureSkipVerifyStr := parsedParams.Get("x-tls-insecure-skip-verify")
				if len(insecureSkipVerifyStr) > 0 {
					x, err := strconv.ParseBool(insecureSkipVerifyStr)
					if err != nil {
						return nil, err
					}
					insecureSkipVerify = x
				}

				err = mysql.RegisterTLSConfig(ctls, &tls.Config{
					RootCAs:            rootCertPool,
					Certificates:       clientCert,
					InsecureSkipVerify: insecureSkipVerify,
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	config, err := mysql.ParseDSN(strings.TrimPrefix(url, "mysql://"))
	if err != nil {
		return nil, err
	}

	config.MultiStatements = true

	// Keep backwards compatibility from when we used net/url.Parse() to parse the DSN.
	// net/url.Parse() would automatically unescape it for us.
	// See: https://play.golang.org/p/q9j1io-YICQ
	user, err := nurl.QueryUnescape(config.User)
	if err != nil {
		return nil, err
	}
	config.User = user

	password, err := nurl.QueryUnescape(config.Passwd)
	if err != nil {
		return nil, err
	}
	config.Passwd = password

	return config, nil
}
