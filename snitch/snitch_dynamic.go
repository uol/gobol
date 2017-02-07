package snitch

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
)

func keyFromMetricID(metric string, tags map[string]string) string {
	merged := []string{}
	for k, v := range tags {
		merged = append(merged, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(merged)
	merged = append(merged, metric)
	return strings.Join(merged, ", ")
}

var aggregationsContants = map[string]bool{
	"avg": true,
	"min": true,
	"max": true,
	"sum": true,
}

func (st *Stats) getPoint(
	metricName string,
	tags map[string]string,
	aggregation, interval string,
	keep, nullable bool,
) (*CustomPoint, error) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	key := keyFromMetricID(metricName, tags)
	if metric, ok := st.points[key]; ok {
		return metric, nil
	}

	if _, err := cron.Parse(interval); err != nil {
		return nil, err
	}
	if _, ok := aggregationsContants[aggregation]; !ok {
		return nil, errors.New("Unknown aggregation")
	}

	metric := &CustomPoint{
		metric:      metricName,
		tags:        make(map[string]string),
		aggregation: aggregation,
		keepValue:   keep,
		sendOnNull:  nullable,
		valNull:     true,
		interval:    interval,
		sender:      st.receiver,
		pre: func(p *CustomPoint) bool {
			if !p.sendOnNull && p.IsValueNull() {
				return false
			}
			switch p.aggregation {
			case "avg":
				if p.GetValue() != 0 && p.GetCount() != 0 {
					p.SetValue(p.GetValue() / float64(p.GetCount()))
				}
			}
			return true
		},
		post: func(p *CustomPoint) {
			st.logger.WithFields(
				logrus.Fields{
					"metric":   p.metric,
					"interval": p.interval,
					"value":    p.GetValue(),
					"null":     p.IsValueNull(),
				},
			).Info("collected")
			if p.aggregation != "" {
				p.SetValue(0)
				p.SetCount(0)
				p.SetValueNull()
			} else if !p.keepValue {
				p.SetValue(0)
				p.SetCount(0)
				p.SetValueNull()
			}
		},
	}

	for k, v := range st.tags {
		metric.tags[k] = v
	}
	for k, v := range tags {
		metric.tags[k] = v
	}

	st.points[key] = metric
	st.cron.AddJob(interval, metric)
	return metric, nil
}
