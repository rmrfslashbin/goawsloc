package loc

import (
	"errors"
	"os"
	"path"

	"github.com/rmrfslashbin/goawsloc/pkg/awslocation/placesvc"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags struct contains settings for the root command
type Flags struct {
	countries   []string
	description string
	dotenvPath  string
	indexName   string
	json        bool
	lat         float64
	loglevel    string
	lon         float64
	text        string
	x1          float64
	x2          float64
	y1          float64
	y2          float64
	tags        []string
}

type Sercices struct {
	location *placesvc.Config
}

var (
	flags *Flags
	log   *logrus.Logger
	svc   *Sercices

	// rootCmd is the Viper root command
	RootCmd = &cobra.Command{
		Use: "loc-main",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set the log level
			switch flags.loglevel {
			case "error":
				log.SetLevel(logrus.ErrorLevel)
			case "warn":
				log.SetLevel(logrus.WarnLevel)
			case "info":
				log.SetLevel(logrus.InfoLevel)
			case "debug":
				log.SetLevel(logrus.DebugLevel)
			case "trace":
				log.SetLevel(logrus.TraceLevel)
			default:
				log.SetLevel(logrus.InfoLevel)
			}
			setup()
		},
	}

	cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "create location services",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runCreatePlaceIndex(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdDelete = &cobra.Command{
		Use:   "delete",
		Short: "delete location services",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runDeletePlaceIndex(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdDescribe = &cobra.Command{
		Use:   "describe",
		Short: "describe an index",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runDescribeIndex(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdList = &cobra.Command{
		Use:   "list",
		Short: "list indexes",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runListIndexes(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdPosition = &cobra.Command{
		Use:   "position",
		Short: "search coordinate, get a legible address",
		Long:  "Reverse geocodes a given coordinate and returns a legible address. Allows you to search for Places or points of interest near a given position",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runSearchPosition(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdSuggestion = &cobra.Command{
		Use:   "suggestion",
		Short: "search free-form text",
		Long:  "Generates suggestions for addresses and points of interest based on partial or misspelled free-form text. This operation is also known as autocomplete, autosuggest, or fuzzy matching",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.lat != 0 && flags.lon == 0 {
				return errors.New("latitude is set but longitude is not")
			}
			if flags.lat == 0 && flags.lon != 0 {
				return errors.New("longitude is set but latitude is not")
			}
			if flags.x1 != 0 && (flags.x2 == 0 || flags.y1 == 0 || flags.y2 == 0) {
				return errors.New("x1 is set but x2 or y1 or y2 is not")
			}
			if flags.x2 != 0 && (flags.x1 == 0 || flags.y1 == 0 || flags.y2 == 0) {
				return errors.New("x2 is set but x1 or y1 or y2 is not")
			}
			if flags.y1 != 0 && (flags.x1 == 0 || flags.x2 == 0 || flags.y2 == 0) {
				return errors.New("y1 is set but x1 or x2 or y2 is not")
			}
			if flags.y2 != 0 && (flags.x1 == 0 || flags.x2 == 0 || flags.y1 == 0) {
				return errors.New("y2 is set but x1 or x2 or y1 is not")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runSearchSuggestion(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdText = &cobra.Command{
		Use:   "text",
		Short: "geocode free-form text",
		Long:  "Geocodes free-form text, such as an address, name, city, or region to allow you to search for Places or points of interest",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.lat != 0 && flags.lon == 0 {
				return errors.New("latitude is set but longitude is not")
			}
			if flags.lat == 0 && flags.lon != 0 {
				return errors.New("longitude is set but latitude is not")
			}
			if flags.x1 != 0 && (flags.x2 == 0 || flags.y1 == 0 || flags.y2 == 0) {
				return errors.New("x1 is set but x2 or y1 or y2 is not")
			}
			if flags.x2 != 0 && (flags.x1 == 0 || flags.y1 == 0 || flags.y2 == 0) {
				return errors.New("x2 is set but x1 or y1 or y2 is not")
			}
			if flags.y1 != 0 && (flags.x1 == 0 || flags.x2 == 0 || flags.y2 == 0) {
				return errors.New("y1 is set but x1 or x2 or y2 is not")
			}
			if flags.y2 != 0 && (flags.x1 == 0 || flags.x2 == 0 || flags.y1 == 0) {
				return errors.New("y2 is set but x1 or x2 or y1 is not")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runSearchText(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}

	cmdUpdate = &cobra.Command{
		Use:   "update",
		Short: "update location services",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
			if err := runUpdatePlaceIndex(); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	flags = &Flags{}
	svc = &Sercices{}
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	RootCmd.PersistentFlags().StringVarP(&flags.loglevel, "loglevel", "", "info", "[error|warn|info|debug|trace]")
	RootCmd.PersistentFlags().StringVarP(&flags.dotenvPath, "dotenv", "", "", "dotenv path")
	RootCmd.PersistentFlags().BoolVarP(&flags.json, "json", "j", false, "output json")

	cmdCreate.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdCreate.Flags().StringVarP(&flags.description, "description", "", "", "index description")
	cmdCreate.Flags().StringSliceVarP(&flags.tags, "tags", "", []string{}, "index tags (key,value)")
	cmdCreate.MarkFlagRequired("index")

	cmdDelete.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdDelete.MarkFlagRequired("index")

	cmdDescribe.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdDelete.MarkFlagRequired("index")

	cmdPosition.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdPosition.Flags().Float64VarP(&flags.lat, "lat", "", 0, "latitude")
	cmdPosition.Flags().Float64VarP(&flags.lon, "lon", "", 0, "longitude")
	cmdPosition.MarkFlagRequired("index")
	cmdPosition.MarkFlagRequired("lat")
	cmdPosition.MarkFlagRequired("lon")

	cmdSuggestion.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdSuggestion.Flags().StringVarP(&flags.text, "text", "", "", "text")
	cmdSuggestion.Flags().StringSliceVarP(&flags.countries, "country", "", []string{}, "one or more countries to limit the search to")
	cmdSuggestion.Flags().Float64VarP(&flags.lat, "lat", "", 0, "latitude")
	cmdSuggestion.Flags().Float64VarP(&flags.lon, "lon", "", 0, "longitude")
	cmdSuggestion.Flags().Float64VarP(&flags.x1, "x1", "", 0, "x1")
	cmdSuggestion.Flags().Float64VarP(&flags.x2, "x2", "", 0, "x2")
	cmdSuggestion.Flags().Float64VarP(&flags.y1, "y1", "", 0, "y1")
	cmdSuggestion.Flags().Float64VarP(&flags.y2, "y2", "", 0, "y2")
	cmdSuggestion.MarkFlagRequired("index")
	cmdSuggestion.MarkFlagRequired("text")
	cmdSuggestion.MarkFlagRequired("country")

	cmdText.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdText.Flags().StringVarP(&flags.text, "text", "", "", "text")
	cmdText.Flags().StringSliceVarP(&flags.countries, "country", "", []string{}, "one or more countries to limit the search to")
	cmdText.Flags().Float64VarP(&flags.lat, "lat", "", 0, "latitude")
	cmdText.Flags().Float64VarP(&flags.lon, "lon", "", 0, "longitude")
	cmdText.Flags().Float64VarP(&flags.x1, "x1", "", 0, "x1")
	cmdText.Flags().Float64VarP(&flags.x2, "x2", "", 0, "x2")
	cmdText.Flags().Float64VarP(&flags.y1, "y1", "", 0, "y1")
	cmdText.Flags().Float64VarP(&flags.y2, "y2", "", 0, "y2")
	cmdText.MarkFlagRequired("index")
	cmdText.MarkFlagRequired("text")

	cmdUpdate.Flags().StringVarP(&flags.indexName, "index", "", "", "index name")
	cmdUpdate.Flags().StringVarP(&flags.description, "description", "", "", "index description")
	cmdUpdate.MarkFlagRequired("index")

	RootCmd.AddCommand(
		cmdCreate,
		cmdDelete,
		cmdDescribe,
		cmdList,
		cmdPosition,
		cmdSuggestion,
		cmdText,
		cmdUpdate,
	)
}

func setup() {
	if flags.dotenvPath == "" {
		/*
			// get platform specific user config directory
			configHome, err := os.UserConfigDir()
			if err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("could not get user config directory and dotenv file not set")
			}
			viper.AddConfigPath(path.Join(configHome, "tndx"))
		*/
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	} else {
		flags.dotenvPath = path.Clean(flags.dotenvPath)
		viper.SetConfigFile(flags.dotenvPath)
		if _, err := os.Stat(flags.dotenvPath); err != nil {
			log.WithFields(logrus.Fields{
				"path":  flags.dotenvPath,
				"error": err,
			}).Fatal("unable to load dotenv")
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		log.WithFields(logrus.Fields{
			"path": flags.dotenvPath,
			"err":  err,
		}).Fatal("failed to read dotenv file")
	}

	awsProfile := viper.GetString("AwsProfile")
	awsRegion := viper.GetString("AwsRegion")

	if awsProfile == "" {
		log.Fatal("AwsProfile not set")
	}
	if awsRegion == "" {
		log.Fatal("AwsRegion not set in yaml config file")
	}

	var err error
	svc.location, err = placesvc.New(
		placesvc.SetLogger(log),
		placesvc.SetAWSProfile(awsProfile),
		placesvc.SetAWSRegion(awsRegion),
		placesvc.SetIndexName(flags.indexName),
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create location service")
	}
}
