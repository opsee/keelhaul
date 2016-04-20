package autocheck

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/opsee/bastion/config"
)

type cloudWatchRequest struct {
	target                 string
	metrics                []*string
	namespace              string
	statisticsIntervalSecs int
	statisticsPeriod       int
	statistics             []*string
}

type CloudWatchResponse struct {
	// TODO
}

func (cr *CloudWatchRequest) getStats(cloudwatchClient *cloudwatch.CloudWatch, stat string) (cloudwatchResponse, error) {
	for _, metric := range cr.metrics {
		// 1 minute lag.  otherwise we won't get stats
		endTime := time.Now().UTC().Add(time.Duration(-1) * time.Minute)
		startTime := endTime.Add(time.Duration(-1*cr.StatisticsIntervalSecs) * time.Second)

		log.WithFields(log.Fields{"startTime": startTime, "endTime": endTime}).Debug("Fetching cloudwatch metric statistics")

		dimensions := []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("DBInstanceIdentifier"),
				Value: aws.String(cr.target),
			},
		}

		params := &cloudwatch.GetMetricStatisticsInput{
			StartTime:  aws.Time(startTime),
			EndTime:    aws.Time(endTime),
			MetricName: aws.String(metric.Name),
			Namespace:  aws.String(metric.Namespace),
			Period:     aws.Int64(int64(cr.StatisticsPeriod)),
			Statistics: cr.statistics,
			Dimensions: dimensions,
		}

		resp, err := cloudwatchClient.GetMetricStatistics(params)
		if err != nil {
			log.WithError(err).Errorf("No datapoints for %s", metric.Name)
			continue
		}

		if len(resp.Datapoints) == 0 {
			log.WithError(err).Errorf("No datapoints for %s", metric.Name)
			continue
		}

		for _, datapoint := range resp.Datapoints {
			for _, statistic := range cr.statistics {
				value := float64(0.0)
				switch aws.StringValue(statistic) {
				case "Average":
					value = aws.Float64Value(datapoint.Average)
				case "Maximum":
					value = aws.Float64Value(datapoint.Maximum)
				case "Minimum":
					value = aws.Float64Value(datapoint.Minimum)
				case "SampleCount":
					value = aws.Float64Value(datapoint.SampleCount)
				case "Sum":
					value = aws.Float64Value(datapoint.Sum)
				default:
					log.Errorf("Unknown statistic type %s", statistic)
				}
				// TODO add to response
			}
			break
		}
	}

	return &cloudWatchResponse{}
}
