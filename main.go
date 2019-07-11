package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.zerodha.tech/commons/lil/store"
	redisstore "gitlab.zerodha.tech/commons/lil/store/redis"
)

var (
	// Version of the build.
	// This is injected at build-time.
	// Be sure to run the provided run script to inject correctly.
	buildVersion   = "unknown"
	buildDate      = "unknown"
	str            store.Store
	baseURL        string
	shortURLLength = 8
)

func getRedisPool(address string, password string, maxActive int, maxIdle int, timeout time.Duration) *redis.Pool {
	return &redis.Pool{
		Wait:      true,
		MaxActive: maxActive,
		MaxIdle:   maxIdle,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				address,
				redis.DialPassword(password),
				redis.DialConnectTimeout(timeout),
				redis.DialReadTimeout(timeout),
				redis.DialWriteTimeout(timeout),
			)

			return c, err
		},
	}
}

// Package initialisation.
func init() {
	// Command line flags.
	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Println(flagSet.FlagUsages())
		os.Exit(0)
	}

	// Setup the default configuration.
	viper.SetConfigName("config")
	viper.SetDefault("server.address", ":8085")

	// Get config path from flag.
	flagSet.StringSlice("config", []string{}, "Path to a config file to load. This can be specified multiple times and the config files will be merged in order")

	// Override flags.
	flagSet.String("server.address", ":8085", "Address to bind HTTP server")

	// Other flags.
	flagSet.BoolP("version", "v", false, "Current version of the build")

	flagSet.Parse(os.Args[1:])
	viper.BindPFlags(flagSet)

	// Read default config file. Won't throw the error yet.
	viper.AddConfigPath(".")
	vErr := viper.ReadInConfig()

	// Read explicit configs, if there are any.
	cfgs := viper.GetStringSlice("config")
	for _, c := range cfgs {
		log.Printf("reading config: %s", c)
		viper.SetConfigFile(c)
		if err := viper.MergeInConfig(); err != nil {
			log.Printf("error reading config: %v", err)
		}
	}

	// Was there an error reading the default config.toml?
	// It's okay as long as an additional config was read.
	if vErr != nil {
		if len(cfgs) == 0 {
			log.Fatalf("no config was read: %v", vErr)
		}
		log.Println("WARNING: no default config was read")
	}
}

func main() {
	// Display version and exit
	if viper.GetBool("version") {
		fmt.Printf("Commit: %v\nBuild: %v\n", buildVersion, buildDate)
		return
	}
	log.Printf("Version: %v | Build: %v\n", buildVersion, buildDate)

	// Set base config
	baseURL = viper.GetString("base_url")
	shortURLLength = viper.GetInt("url_length")

	// Init redis pool
	str = redisstore.New(getRedisPool(
		viper.GetString("cache.address"),
		viper.GetString("cache.password"),
		viper.GetInt("cache.max_idle"),
		viper.GetInt("cache.max_active"),
		time.Duration(viper.GetDuration("cache.timeout"))*time.Millisecond,
	))

	// Routing
	router := chi.NewRouter()
	router.Get("/", http.HandlerFunc(handleWelcome))
	router.Get("/{uri}", http.HandlerFunc(handleRedirect))
	router.Delete("/api/{uri}", http.HandlerFunc(handleDelete))
	router.Post("/api/new", http.HandlerFunc(handleCreate))

	// Run server
	server := &http.Server{
		Addr:         viper.GetString("server.address"),
		Handler:      router,
		ReadTimeout:  viper.GetDuration("server.read_timeout") * time.Millisecond,
		WriteTimeout: viper.GetDuration("server.write_timeout") * time.Millisecond,
		IdleTimeout:  viper.GetDuration("server.idle_timeout") * time.Millisecond,
	}
	log.Printf("listening on - %v", viper.GetString("server.address"))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
