package loc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rmrfslashbin/goawsloc/pkg/awslocation/placesvc"

	"github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

type PositionSummaryResults struct {
	Summary *types.SearchPlaceIndexForPositionSummary
	Results []types.SearchForPositionResult
}

type SuggestionSummaryResults struct {
	Summary *types.SearchPlaceIndexForSuggestionsSummary
	Results []types.SearchForSuggestionsResult
}

type TextSummaryResults struct {
	Summary *types.SearchPlaceIndexForTextSummary
	Results []types.SearchForTextResult
}

func runCreatePlaceIndex() error {
	tags := make(map[string]string, len(flags.tags))
	for _, tag := range flags.tags {
		parts := strings.Split(tag, "=")
		if len(parts) != 2 {
			log.WithFields(logrus.Fields{
				"tag": tag,
			}).Error("invalid tag")
			return fmt.Errorf("invalid tag: %s", tag)
		}
		tags[parts[0]] = parts[1]
	}
	if ret, err := svc.location.CreatePlaceIndex(flags.description, &tags); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error creating index")
		return err
	} else {
		log.WithFields(logrus.Fields{
			"createTime": ret.CreateTime,
			"indexARN":   *ret.IndexArn,
			"indexName":  *ret.IndexName,
		}).Info("Created index")
	}
	return nil
}

func runDeletePlaceIndex() error {
	if _, err := svc.location.DeletePlaceIndex(); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error deleting index")
		return err
	} else {
		log.Info("Deleted index")
	}
	return nil
}

func runDescribeIndex() error {
	if ret, err := svc.location.DescribePlaceIndex(flags.indexName); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error describing index")
		return err
	} else {
		if flags.json {
			if data, err := json.Marshal(ret); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Error("error marshalling json")
				return err
			} else {
				fmt.Println(string(data))
			}
		} else {
			fmt.Printf("Index Name:   %s\n", *ret.IndexName)
			fmt.Printf("Description:  %s\n", *ret.Description)
			fmt.Printf("Pricing Plan: %s\n", ret.PricingPlan)
			fmt.Printf("Data Source:  %s\n", *ret.DataSource)
			fmt.Printf("Data Storage: %s\n", ret.DataSourceConfiguration.IntendedUse.Values())
			fmt.Printf("Create Time:  %s\n", ret.CreateTime)
			fmt.Printf("Update Time:  %s\n", ret.UpdateTime)
			fmt.Printf("Index ARN:    %s\n", *ret.IndexArn)
			if len(ret.Tags) > 0 {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
				fmt.Fprintln(w, "Tags\tValue")
				for k, v := range ret.Tags {
					fmt.Fprintf(w, "%s\t%s\n", k, v)
				}
				w.Flush()
			} else {
				fmt.Println("Tags:        (none)")
			}
		}

	}
	return nil
}

func runListIndexes() error {
	if ret, err := svc.location.ListPlaceIndexes(); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error listing indexes")
		return err
	} else {
		if flags.json {
			if data, err := json.Marshal(ret); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Error("error marshalling json")
				return err
			} else {
				fmt.Println(string(data))
			}
		} else {
			log.WithFields(logrus.Fields{
				"count": len(ret.Entries),
			}).Info("Listed indexes")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
			fmt.Fprintln(w, "CTime\tMTime\tIndex\tPricing\tDataSource\tDescription")
			for _, entry := range ret.Entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", entry.CreateTime, entry.UpdateTime, *entry.IndexName, entry.PricingPlan, *entry.DataSource, *entry.Description)
			}
			w.Flush()
			fmt.Println()
		}
	}
	return nil
}

func runSearchPosition() error {
	if ret, err := svc.location.SearchPlaceIndexForPosition(&placesvc.LatLon{Latitude: flags.lat, Longitude: flags.lon}); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error searching position")
		return err
	} else {
		log.Info("Searched position")
		if flags.json {
			if data, err := json.Marshal(&PositionSummaryResults{Summary: ret.Summary, Results: ret.Results}); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Error("error marshalling json")
				return err
			} else {
				fmt.Println(string(data))
			}
		} else {
			spew.Dump(ret.Summary)
			spew.Dump(ret.Results)
		}
	}
	return nil
}

func runSearchSuggestion() error {
	fmt.Println(flags.countries)
	if ret, err := svc.location.SearchPlaceIndexForSuggestions(
		&placesvc.SuggestionSearch{
			Text:            &flags.text,
			BiasPosition:    &placesvc.LatLon{Latitude: flags.lat, Longitude: flags.lon},
			FilterBBox:      &placesvc.Box{X1: flags.x1, Y1: flags.y1, X2: flags.x2, Y2: flags.y2},
			FilterCountries: flags.countries,
		}); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error searching suggestion")
		return err
	} else {
		if flags.json {
			if data, err := json.Marshal(&SuggestionSummaryResults{Summary: ret.Summary, Results: ret.Results}); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Error("error marshalling json")
				return err
			} else {
				fmt.Println(string(data))
			}
		} else {
			log.Info("Searched suggestion")
			spew.Dump(ret.Summary)
			spew.Dump(ret.Results)
		}
	}
	return nil
}

func runSearchText() error {
	if ret, err := svc.location.SearchPlaceIndexForText(&placesvc.SuggestionSearch{
		Text:            &flags.text,
		BiasPosition:    &placesvc.LatLon{Latitude: flags.lat, Longitude: flags.lon},
		FilterBBox:      &placesvc.Box{X1: flags.x1, Y1: flags.y1, X2: flags.x2, Y2: flags.y2},
		FilterCountries: flags.countries,
	}); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error searching text")
		return err
	} else {
		if flags.json {
			if data, err := json.Marshal(&TextSummaryResults{Summary: ret.Summary, Results: ret.Results}); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
				}).Error("error marshalling json")
				return err
			} else {
				fmt.Println(string(data))
			}
		} else {
			log.Info("Searched text")
			spew.Dump(ret.Summary)
			spew.Dump(ret.Results)
		}
	}
	return nil
}

func runUpdatePlaceIndex() error {
	if _, err := svc.location.UpdatePlaceIndex(flags.description); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error updating index")
		return err
	} else {
		log.Info("Updated index")
	}
	return nil
}
