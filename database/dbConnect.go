// Package database handles connections to different
// types of databases
package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	//"github.com/olivere/elastic"
	"cloud.google.com/go/firestore"
	"github.com/prakashjegan/golangexercise/config"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Import MySQL database driver
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"

	// Import PostgreSQL database driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"

	// Import SQLite3 database driver
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"

	// Import Redis Driver
	"github.com/mediocregopher/radix/v4"

	// Import Mongo driver
	aero "github.com/aerospike/aerospike-client-go"

	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/event"
	opts "go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"

	logSer "log"

	"github.com/olivere/elastic/v7"

	"net/http"

	temporal "go.temporal.io/sdk/client"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	//"google.golang.org/api/option"
)

// dbClient variable to access gorm
var dbClient *gorm.DB

var sqlDB *sql.DB
var err error

// redisClient variable to access the redis client
var redisClient *radix.Client

var aerospikeClient *aero.Client

var elasticClient *elastic.Client

var temporalClient *temporal.Client
var firebaseClient *firestore.Client
var firebaseAuthClient *auth.Client

var zapLogger *zap.Logger

// RedisConnTTL - context deadline in second
var RedisConnTTL int

// mongoClient instance
var mongoClient *qmgo.Client

// InitDB - function to initialize db
func InitDB() *gorm.DB {
	var db = dbClient

	configureDB := config.GetConfig().Database.RDBMS

	driver := configureDB.Env.Driver
	username := configureDB.Access.User
	password := configureDB.Access.Pass
	database := configureDB.Access.DbName
	host := configureDB.Env.Host
	port := configureDB.Env.Port
	sslmode := configureDB.Ssl.Sslmode
	timeZone := configureDB.Env.TimeZone
	maxIdleConns := configureDB.Conn.MaxIdleConns
	maxOpenConns := configureDB.Conn.MaxOpenConns
	connMaxLifetime := configureDB.Conn.ConnMaxLifetime
	logLevel := configureDB.Log.LogLevel

	switch driver {
	case "mysql":
		dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
		if sslmode != "disable" {
			dsn += "&tls=custom"
			err = InitTLSMySQL()
			if err != nil {
				log.WithError(err).Panic("panic code: 150")
			}
		}
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 151")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 152")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "postgres":
		dsn := "host=" + host + " port=" + port + " user=" + username + " dbname=" + database + " password=" + password + " sslmode=" + sslmode + " TimeZone=" + timeZone
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 153")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 154")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(database), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Silent),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 155")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	default:
		log.Fatal("The driver " + driver + " is not implemented yet")
	}

	dbClient = db

	return dbClient
}

// GetDB - get a connection
func GetDB() *gorm.DB {
	return dbClient
}

func GetAerospike() *aero.Client {
	return aerospikeClient
}

// InitAerospike - function to initialize redis client
func InitAerospike() (*aero.Client, error) {
	configureAerospike := config.GetConfig().Database.AEROSPIKE

	aClient, err := aero.NewClient(
		configureAerospike.Env.Host,
		configureAerospike.Env.Port)
	if err != nil {
		log.WithError(err).Panic("panic code: 161")
		return aClient, err
	}
	// Only for debugging
	if err == nil {
		fmt.Println("REDIS pool connection successful!")
	}

	aerospikeClient = aClient

	return aerospikeClient, nil
}

func InitTemporal() (temporal.Client, error) {
	fmt.Printf("Temporal Initialization\n")
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	logger.Info("Zap logger created")
	fmt.Printf("Temporal Zap Logger created\n")
	c, err := temporal.NewClient(temporal.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		return nil, err
	}
	//defer c.Close()
	temporalClient = &c
	zapLogger = logger
	return c, nil
}

func InitFireBase() (*firestore.Client, error) {
	configureFireBase := config.GetConfig().FireBaseConfig

	fmt.Printf("FireBase Initialization\n")
	ctx := context.Background()
	sa := option.WithCredentialsFile(configureFireBase.Path)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	authClient, err := app.Auth(context.Background())

	if err != nil {
		log.Fatalln(err)
	}
	firebaseClient = client
	firebaseAuthClient = authClient
	return client, nil
}

func NewTemporal() temporal.Client {
	fmt.Printf("Temporal Initialization\n")
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil
	}
	logger.Info("Zap logger created")
	fmt.Printf("Temporal Zap Logger created\n")
	c, err := temporal.NewClient(temporal.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		return nil
	}
	//defer c.Close()
	temporalClient = &c
	zapLogger = logger
	return c
}

// InitElasticSearch - function to initialize Elastic Search client
func InitElasticSearch() (*elastic.Client, error) {
	configureElasticSearch := config.GetConfig().Database.ELASTICSEARCH

	// Elasticsearch cluster settings
	username := configureElasticSearch.Env.USERNAME
	password := configureElasticSearch.Env.PASSWORD

	caCertPath := configureElasticSearch.Env.CACERTPATH

	// Set up the TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs: func() *x509.CertPool {
			// Load CA certificate
			caCert, err := ioutil.ReadFile(caCertPath)
			if err != nil {
				log.Fatalf("Error reading CA certificate: %s", err)
			}

			// Create CertPool and add CA certificate
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			return caCertPool
		}(),
	}

	// // Create a new HTTP client with the username and password
	// httpClient := &http.Client{
	// 	Timeout: 10 * time.Second,
	// 	Transport: &http.Transport{
	// 		TLSHandshakeTimeout: 10 * time.Second,
	// 		Proxy:               http.ProxyFromEnvironment,
	// 	},
	// }

	// Set up the Elasticsearch client options
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(configureElasticSearch.Env.Host + ":" + configureElasticSearch.Env.Port),
		elastic.SetBasicAuth(username, password),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetInfoLog(logSer.New(os.Stdout, "", logSer.LstdFlags)),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(time.Millisecond*100, time.Second))),
		elastic.SetMaxRetries(10),
		elastic.SetGzip(true),
		elastic.SetHealthcheckInterval(10 * time.Second),
		elastic.SetHealthcheckTimeoutStartup(30 * time.Second),
		elastic.SetHealthcheckTimeout(5 * time.Second),
		elastic.SetSnifferTimeout(5 * time.Second),
		elastic.SetSnifferInterval(10 * time.Minute),
		//elastic.SetHttpClient(httpClient),
		elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}),
	}

	// Create the Elasticsearch client
	fmt.Printf("Creation Elastic Client \n")
	client, err := elastic.NewClient(options...)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s \n", err)
		fmt.Printf("%s \n", err)
		return nil, err
	}
	elasticClient = client
	fmt.Printf("Completed Creating Elastic Client %s \n", elasticClient)

	// indexSettings := `
	// {
	// 	"settings": {
	// 		"number_of_shards": 1,
	// 		"number_of_replicas": 0
	// 	},
	// 	"mappings": {
	// 		"properties": {
	// 			{
	// 				"id": {
	// 				  "type": "long"
	// 				},
	// 				"documentType": {
	// 				  "type": "keyword"
	// 				},
	// 				"createdAt": {
	// 				  "type": "date",
	// 				  "format": "yyyy-MM-dd HH:mm:ss"
	// 				},
	// 				"updatedAt": {
	// 				  "type": "date",
	// 				  "format": "yyyy-MM-dd HH:mm:ss"
	// 				},
	// 				"deletedAt": {
	// 				  "type": "date",
	// 				  "format": "yyyy-MM-dd HH:mm:ss"
	// 				},
	// 				"platformTags": {
	// 				  "type": "keyword"
	// 				},
	// 				"partnerId": {
	// 				  "type": "long"
	// 				},
	// 				"partnerType": {
	// 				  "type": "keyword"
	// 				},
	// 				"platformId": {
	// 				  "type": "long"
	// 				},
	// 				"platformType": {
	// 				  "type": "keyword"
	// 				},
	// 				"platformName": {
	// 				  "type": "keyword"
	// 				},
	// 				"verificationLink": {
	// 				  "type": "keyword"
	// 				},
	// 				"stakeHolderType": {
	// 				  "type": "keyword"
	// 				},
	// 				"logoLink": {
	// 				  "type": "keyword"
	// 				},
	// 				"organizationName": {
	// 				  "type": "keyword"
	// 				},
	// 				"organizationDescription": {
	// 				  "type": "keyword"
	// 				},
	// 				"preferredLanguage": {
	// 				  "type": "keyword"
	// 				},
	// 				"partnerCountryCode": {
	// 				  "type": "keyword"
	// 				},
	// 				"userName": {
	// 				  "type": "keyword"
	// 				},
	// 				"emailId": {
	// 				  "type": "keyword"
	// 				},
	// 				"skillSet": {
	// 				  "type": "keyword"
	// 				},
	// 				"totalEmployees": {
	// 				  "type": "integer"
	// 				},
	// 				"organizationLifeSpan": {
	// 				  "type": "integer"
	// 				},
	// 				"totalBudgetAmountSpent": {
	// 				  "type": "long"
	// 				},
	// 				"totalFollowers": {
	// 				  "type": "long"
	// 				},
	// 				"totalContent": {
	// 				  "type": "long"
	// 				},
	// 				"totalView": {
	// 				  "type": "long"
	// 				},
	// 				"averageViewPerContent": {
	// 				  "type": "long"
	// 				},
	// 				"averageLikesPerContent": {
	// 				  "type": "long"
	// 				},
	// 				"totalLikes": {
	// 				  "type": "long"
	// 				},
	// 				"averageTimePerContentSpentbyViewer": {
	// 				  "type": "long"
	// 				},
	// 				"totalCommentsPerView": {
	// 				  "type": "long"
	// 				},
	// 				"jobDefinitionID": {
	// 				  "type": "long"
	// 				},
	// 				"jobName": {
	// 				  "type": "keyword"
	// 				},
	// 				"jobType": {
	// 				  "type": "keyword"
	// 				},
	// 				"AdType": {
	// 				  "type": "integer"
	// 				},
	// 				"jobDescription": {
	// 				  "type": "keyword"
	// 				},
	// 				"JobDescriptionLink": {
	// 				  "type": "keyword"
	// 				},
	// 				"JobReferLink": {
	// 				  "type": "keyword"
	// 				},
	// 				"posterUserId": {
	// 				  "type": "long"
	// 				},
	// 				"posterUserType": {
	// 				  "type": "keyword"
	// 				},
	// 				"posterPartnerId": {
	// 				  "type": "long"
	// 				},
	// 				"posterPartnerType": {
	// 				  "type": "keyword"
	// 				},
	// 				"tentativeStartDate": {
	// 				  "type": "date",
	// 				  "format": "yyyy-MM-dd HH:mm:ss"
	// 				},
	// 				"tentativeEndDate": {
	// 				  "type": "date",
	// 				  "format": "yyyy-MM-dd HH:mm:ss"
	// 				},
	// 				"maxBufferPeriodInDays": {
	// 				  "type": "integer"
	// 				},
	// 				"acceptorUserId": {
	// 				  "type": "long"
	// 				},
	// 				"acceptorUserTypes": {
	// 				  "type": "keyword"
	// 				},
	// 				"acceptorPartnerId": {
	// 				  "type": "long"
	// 				},
	// 				"acceptorPartnerType": {
	// 				  "type": "keyword"
	// 				},
	// 				"jobStatus": {
	// 				  "type": "keyword"
	// 				},
	// 				"totalBudgetAmountForJob": {
	// 				  "type": "long"
	// 				}
	// 			  }
	// 		}
	// 	}
	// }`

	// _, err = client.CreateIndex("search_index").BodyString(indexSettings).Do(context.Background())
	// if err != nil {
	// 	log.Fatalf("Error creating index: %s", err)
	// }

	return elasticClient, nil
}

func GetElasticClient() *elastic.Client {
	return elasticClient
}

func GetTemporalClient() (*temporal.Client, *zap.Logger) {
	return temporalClient, zapLogger
}

func GetFireBaseClient() *firestore.Client {
	return firebaseClient
}

func GetFireBaseAuthClient() *auth.Client {
	return firebaseAuthClient
}

// InitRedis - function to initialize redis client
func InitRedis() (*radix.Client, error) {
	configureRedis := config.GetConfig().Database.REDIS

	RedisConnTTL = configureRedis.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RedisConnTTL)*time.Second)
	defer cancel()

	rClient, err := (radix.PoolConfig{
		Size: configureRedis.Conn.PoolSize,
	}).New(ctx, "tcp", fmt.Sprintf("%v:%v",
		configureRedis.Env.Host,
		configureRedis.Env.Port))
	if err != nil {
		log.WithError(err).Panic("panic code: 161")
		return &rClient, err
	}
	// Only for debugging
	if err == nil {
		fmt.Println("REDIS pool connection successful!")
	}

	redisClient = &rClient

	return redisClient, nil
}

// GetRedis - get a connection
func GetRedis() *radix.Client {
	return redisClient
}

// InitMongo - function to initialize mongo client
func InitMongo() (*qmgo.Client, error) {
	configureMongo := config.GetConfig().Database.MongoDB

	// Connect to the database or cluster
	uri := configureMongo.Env.URI

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configureMongo.Env.ConnTTL)*time.Second)
	defer cancel()

	clientConfig := &qmgo.Config{
		Uri:         uri,
		MaxPoolSize: &configureMongo.Env.PoolSize,
	}
	serverAPIOptions := opts.ServerAPI(opts.ServerAPIVersion1)

	opt := opts.Client().SetAppName(configureMongo.Env.AppName)
	opt.SetServerAPIOptions(serverAPIOptions)

	// for monitoring pool
	if configureMongo.Env.PoolMon == "yes" {
		poolMonitor := &event.PoolMonitor{
			Event: func(evt *event.PoolEvent) {
				switch evt.Type {
				case event.GetSucceeded:
					fmt.Println("GetSucceeded")
				case event.ConnectionReturned:
					fmt.Println("ConnectionReturned")
				}
			},
		}
		opt.SetPoolMonitor(poolMonitor)
	}

	client, err := qmgo.NewClient(ctx, clientConfig, options.ClientOptions{ClientOptions: opt})
	if err != nil {
		return client, err
	}

	// Only for debugging
	fmt.Println("MongoDB pool connection successful!")

	mongoClient = client

	return mongoClient, nil
}

// GetMongo - get a connection
func GetMongo() *qmgo.Client {
	return mongoClient
}
