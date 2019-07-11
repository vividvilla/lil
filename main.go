package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"
	"gitlab.zerodha.tech/commons/lil/store"
	redisstore "gitlab.zerodha.tech/commons/lil/store/redis"
)

var (
	// Version of the build.
	// This is injected at build-time.
	// Be sure to run the provided run script to inject correctly.
	buildVersion   = "unknown"
	buildDate      = "unknown"
	kf             *koanf.Koanf
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

func init() {
	// Init koanf.
	kf = koanf.New(".")

	// Initialize commandline flags.
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringSlice("conf", []string{"config.toml"},
		"Path to one or more config files (will be merged in order)")
	f.StringSliceP("datasource", "d", []string{}, "Path to data source plugin. Can specify multiple values.")
	f.StringSliceP("messenger", "m", []string{}, "Path to messenger plugin. Can specify multiple values.")
	f.Bool("version", false, "Current version of the build.")
	f.Bool("worker", false, "Run in worker mode.")
	f.Parse(os.Args[1:])

	// Read config from files.
	cFiles, _ := f.GetStringSlice("conf")
	for _, c := range cFiles {
		if err := kf.Load(file.Provider(c), toml.Parser()); err != nil {
			log.Fatalf("error loading file: %v", err)
		}
	}

	// Overide config with flag values.
	if err := kf.Load(posflag.Provider(f, ".", kf), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
}

func main() {
	// Display version and exit
	if kf.Bool("version") {
		fmt.Printf("Commit: %v\nBuild: %v\n", buildVersion, buildDate)
		return
	}
	log.Printf("Version: %v | Build: %v\n", buildVersion, buildDate)

	// Set base config
	baseURL = kf.String("base_url")
	shortURLLength = kf.Int("url_length")

	// Init redis pool
	str = redisstore.New(getRedisPool(
		kf.String("cache.address"),
		kf.String("cache.password"),
		kf.Int("cache.max_idle"),
		kf.Int("cache.max_active"),
		time.Duration(kf.Duration("cache.timeout"))*time.Millisecond,
	))

	// Routing
	router := chi.NewRouter()
	router.Get("/", http.HandlerFunc(handleWelcome))
	router.Get("/{uri}", http.HandlerFunc(handleRedirect))
	router.Delete("/api/{uri}", http.HandlerFunc(handleDelete))
	router.Post("/api/new", http.HandlerFunc(handleCreate))

	// Run server
	server := &http.Server{
		Addr:         kf.String("server.address"),
		Handler:      router,
		ReadTimeout:  kf.Duration("server.read_timeout") * time.Millisecond,
		WriteTimeout: kf.Duration("server.write_timeout") * time.Millisecond,
		IdleTimeout:  kf.Duration("server.idle_timeout") * time.Millisecond,
	}
	log.Printf("listening on - %v", kf.String("server.address"))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
