package placesvc

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/sirupsen/logrus"
)

type Option func(config *Config)

// Configuration structure.
type Config struct {
	region       string
	profile      string
	indexService string
	indexName    string
	intendedUse  string
	language     string
	pricingPlan  string
	log          *logrus.Logger
	svc          *location.Client
}

type LatLon struct {
	Latitude  float64
	Longitude float64
}

type Box struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

type SuggestionSearch struct {
	// The free-form partial text to use to generate place suggestions. For example,
	// eiffel tow.
	//
	// This member is required.
	Text *string

	// An optional parameter that indicates a preference for place suggestions that are
	// closer to a specified position. If provided, this parameter must contain a pair
	// of numbers. The first number represents the X coordinate, or longitude; the
	// second number represents the Y coordinate, or latitude. For example, [-123.1174,
	// 49.2847] represents the position with longitude -123.1174 and latitude 49.2847.
	// BiasPosition and FilterBBox are mutually exclusive. Specifying both options
	// results in an error.
	BiasPosition *LatLon

	// An optional parameter that limits the search results by returning only
	// suggestions within a specified bounding box. If provided, this parameter must
	// contain a total of four consecutive numbers in two pairs. The first pair of
	// numbers represents the X and Y coordinates (longitude and latitude,
	// respectively) of the southwest corner of the bounding box; the second pair of
	// numbers represents the X and Y coordinates (longitude and latitude,
	// respectively) of the northeast corner of the bounding box. For example,
	// [-12.7935, -37.4835, -12.0684, -36.9542] represents a bounding box where the
	// southwest corner has longitude -12.7935 and latitude -37.4835, and the northeast
	// corner has longitude -12.0684 and latitude -36.9542. FilterBBox and BiasPosition
	// are mutually exclusive. Specifying both options results in an error.
	FilterBBox *Box

	// An optional parameter that limits the search results by returning only
	// suggestions within the provided list of countries.
	//
	// * Use the ISO 3166
	// (https://www.iso.org/iso-3166-country-codes.html) 3-digit country code. For
	// example, Australia uses three upper-case characters: AUS.
	FilterCountries []string

	// The preferred language used to return results. The value must be a valid BCP 47
	// (https://tools.ietf.org/search/bcp47) language tag, for example, en for English.
	// This setting affects the languages used in the results. It does not change which
	// results are returned. If the language is not specified, or not supported for a
	// particular result, the partner automatically chooses a language for the result.
	// Used only when the partner selected is Here.
	Language *string
}

func New(opts ...func(*Config)) *Config {
	config := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(config)
	}

	if config.region == "" {
		config.region = os.Getenv("AWS_REGION")
	}

	if config.indexService == "" {
		config.indexService = "Here"
	}

	if config.intendedUse == "" {
		config.intendedUse = "SingleUse"
	}

	if config.language == "" {
		config.language = "en"
	}

	if config.pricingPlan == "" {
		config.pricingPlan = "RequestBasedUsage"
	}

	c, err := awsconfig.LoadDefaultConfig(context.TODO(), func(o *awsconfig.LoadOptions) error {
		o.Region = config.region
		if config.profile != "" {
			o.SharedConfigProfile = config.profile
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
	config.svc = location.NewFromConfig(c)

	return config
}

func SetAWSRegion(region string) Option {
	return func(config *Config) {
		config.region = region
	}
}

func SetAWSProfile(profile string) Option {
	return func(config *Config) {
		config.profile = profile
	}
}

func SetIndexName(indexName string) Option {
	return func(config *Config) {
		config.indexName = indexName
	}
}

func SetIndexService(indexService string) Option {
	return func(config *Config) {
		config.indexService = indexService
	}
}

func SetLanguage(language string) Option {
	return func(config *Config) {
		config.language = language
	}
}

func SetLogger(log *logrus.Logger) Option {
	return func(config *Config) {
		config.log = log
	}
}

func (c *Config) sanity() error {
	if c.indexName == "" {
		return errors.New("indexName not set")
	}
	return nil
}

func (config *Config) CreatePlaceIndex(description string, tags *map[string]string) (*location.CreatePlaceIndexOutput, error) {
	if err := config.sanity(); err != nil {
		return nil, err
	}

	return config.svc.CreatePlaceIndex(
		context.TODO(),
		&location.CreatePlaceIndexInput{
			DataSource:              aws.String(config.indexService),
			DataSourceConfiguration: &types.DataSourceConfiguration{IntendedUse: types.IntendedUse(config.intendedUse)},
			Description:             aws.String(description),
			IndexName:               aws.String(config.indexName),
			PricingPlan:             types.PricingPlan(config.pricingPlan),
			Tags:                    *tags,
		},
	)
}

func (config *Config) DeletePlaceIndex() (*location.DeletePlaceIndexOutput, error) {
	if err := config.sanity(); err != nil {
		return nil, err
	}

	return config.svc.DeletePlaceIndex(
		context.TODO(),
		&location.DeletePlaceIndexInput{
			IndexName: aws.String(config.indexName),
		},
	)
}

func (config *Config) DescribePlaceIndex(indexName string) (*location.DescribePlaceIndexOutput, error) {
	if indexName == "" {
		if err := config.sanity(); err != nil {
			return nil, err
		}
		indexName = config.indexName
	}

	return config.svc.DescribePlaceIndex(
		context.TODO(),
		&location.DescribePlaceIndexInput{
			IndexName: aws.String(indexName),
		},
	)
}

func (config *Config) ListPlaceIndexes() (*location.ListPlaceIndexesOutput, error) {
	return config.svc.ListPlaceIndexes(
		context.TODO(),
		&location.ListPlaceIndexesInput{},
	)
}

func (config *Config) SearchPlaceIndexForPosition(latLon *LatLon) (*location.SearchPlaceIndexForPositionOutput, error) {
	return config.svc.SearchPlaceIndexForPosition(
		context.TODO(),
		&location.SearchPlaceIndexForPositionInput{
			IndexName: aws.String(config.indexName),
			Language:  aws.String(config.language),
			Position:  []float64{latLon.Longitude, latLon.Latitude},
		},
	)
}

func (config *Config) SearchPlaceIndexForSuggestions(search *SuggestionSearch) (*location.SearchPlaceIndexForSuggestionsOutput, error) {
	return config.svc.SearchPlaceIndexForSuggestions(
		context.TODO(),
		&location.SearchPlaceIndexForSuggestionsInput{
			IndexName: aws.String(config.indexName),
			Text:      search.Text,
			//BiasPosition:    []float64{search.BiasPosition.Longitude, search.BiasPosition.Latitude},
			//FilterBBox:      []float64{search.FilterBBox.X1, search.FilterBBox.Y1, search.FilterBBox.X2, search.FilterBBox.Y2},
			FilterCountries: search.FilterCountries,
			Language:        aws.String(config.language),
		},
	)
}

func (config *Config) SearchPlaceIndexForText(search *SuggestionSearch) (*location.SearchPlaceIndexForTextOutput, error) {
	return config.svc.SearchPlaceIndexForText(
		context.TODO(),
		&location.SearchPlaceIndexForTextInput{
			IndexName: aws.String(config.indexName),
			Text:      search.Text,
			//BiasPosition:    []float64{search.BiasPosition.Longitude, search.BiasPosition.Latitude},
			//FilterBBox:      []float64{search.FilterBBox.X1, search.FilterBBox.Y1, search.FilterBBox.X2, search.FilterBBox.Y2},
			FilterCountries: search.FilterCountries,
			Language:        aws.String(config.language),
		},
	)
}

func (config *Config) UpdatePlaceIndex(description string) (*location.UpdatePlaceIndexOutput, error) {
	if err := config.sanity(); err != nil {
		return nil, err
	}

	return config.svc.UpdatePlaceIndex(
		context.TODO(),
		&location.UpdatePlaceIndexInput{
			IndexName:               aws.String(config.indexName),
			DataSourceConfiguration: &types.DataSourceConfiguration{IntendedUse: types.IntendedUse(config.intendedUse)},
			Description:             aws.String(description),
			PricingPlan:             types.PricingPlan(config.pricingPlan),
		},
	)
}
