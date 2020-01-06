package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fissionlabsio/tmcrawl/config"
	"github.com/fissionlabsio/tmcrawl/crawl"
	"github.com/fissionlabsio/tmcrawl/db"
	"github.com/fissionlabsio/tmcrawl/server"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "tmcrawl [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "tmcrawl implements a Tendermint p2p network crawler utility and API.",
	Long: `tmcrawl implements a Tendermint p2p network crawler utility and API.

The utility will capture geolocation information and node metadata such as network
name, node version, RPC information, and node ID for each crawled node.`,
	RunE: tmcrawlCmdHandler,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", zerolog.InfoLevel.String(), "logging level")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", logLevelJSON, "logging format; must be either json or text")

	rootCmd.AddCommand(getVersionCmd())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func tmcrawlCmdHandler(cmd *cobra.Command, args []string) error {
	logLvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(logLvl)

	switch logFormat {
	case logLevelJSON:
		// JSON is the default logging format

	case logLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}

	cfg, err := config.ParseConfig(args[0])
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		if err := os.Mkdir(cfg.DataDir, os.ModePerm); err != nil {
			return err
		}
	}

	// create and open key/value DB
	db, err := db.NewBoltDB(cfg.DataDir, "tmcrawl.db", &bolt.Options{Timeout: 15 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	crawler := crawl.NewCrawler(cfg, db)
	go func() { crawler.Crawl() }()

	// create HTTP router and mount routes
	router := mux.NewRouter()
	server.RegisterRoutes(db, router)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.ListenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Str("address", cfg.ListenAddr).Msg("starting API server...")
	return srv.ListenAndServe()
}
