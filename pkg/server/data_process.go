package server

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"prober/pkg/common"
	"prober/pkg/pb"
)

const (
	MetricCollectInterval      = 30 * time.Second
	TargetFlushManagerInterval = 60 * time.Second
	MetricOriginSeparator      = `_`
	MetricUniqueSeparator      = `#`
)

var (
	IcmpDataMap = sync.Map{}

	PingLatencyGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingLatency,
		Help: "Duration of ping prober ",
	}, []string{"source_region", "target_region", "slice_name"})
	PingPackageDropGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingPackageDrop,
		Help: "rate of ping packagedrop ",
	}, []string{"source_region", "target_region", "slice_name"})

	PingTargetSuccessGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingTargetSuccess,
		Help: "target success",
	}, []string{"source_region", "target_region", "slice_name"})
	//RTT
	PingTargetRTTGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTAvg,
		Help: "rtt avg",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMinGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMin,
		Help: "rtt min",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMaxGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMax,
		Help: "rtt max",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMdevGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMdev,
		Help: "rtt mdev",
	}, []string{"source_region", "target_region", "slice_name"})
	// BaseTargetMTUGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Name: common.MetricsNameBaseMTU,
	// 	Help: "MTU",
	// }, []string{"source_region", "target_region", "slice_name"})
)

func NewGauge() {
	PingLatencyGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingLatency,
		Help: "Duration of ping prober ",
	}, []string{"source_region", "target_region", "slice_name"})
	PingPackageDropGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingPackageDrop,
		Help: "rate of ping packagedrop ",
	}, []string{"source_region", "target_region", "slice_name"})

	PingTargetSuccessGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingTargetSuccess,
		Help: "target success",
	}, []string{"source_region", "target_region", "slice_name"})
	//RTT
	PingTargetRTTGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTAvg,
		Help: "rtt avg",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMinGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMin,
		Help: "rtt min",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMaxGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMax,
		Help: "rtt max",
	}, []string{"source_region", "target_region", "slice_name"})
	PingTargetRTTMdevGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: common.MetricsNamePingRTTMdev,
		Help: "rtt mdev",
	}, []string{"source_region", "target_region", "slice_name"})
	// BaseTargetMTUGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Name: common.MetricsNameBaseMTU,
	// 	Help: "MTU",
	// }, []string{"source_region", "target_region", "slice_name"})
}

func DeleteMetrics() {
	prometheus.DefaultRegisterer.Unregister(PingLatencyGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingPackageDropGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingTargetSuccessGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingTargetRTTGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingTargetRTTMinGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingTargetRTTMaxGaugeVec)
	prometheus.DefaultRegisterer.Unregister(PingTargetRTTMdevGaugeVec)
	//prometheus.DefaultRegisterer.Unregister(BaseTargetMTUGaugeVec)

}

func NewMetrics() {

	prometheus.DefaultRegisterer.Register(PingLatencyGaugeVec)
	prometheus.DefaultRegisterer.Register(PingPackageDropGaugeVec)
	prometheus.DefaultRegisterer.Register(PingTargetSuccessGaugeVec)
	prometheus.DefaultRegisterer.Register(PingTargetRTTGaugeVec)
	prometheus.DefaultRegisterer.Register(PingTargetRTTMinGaugeVec)
	prometheus.DefaultRegisterer.Register(PingTargetRTTMaxGaugeVec)
	prometheus.DefaultRegisterer.Register(PingTargetRTTMdevGaugeVec)
	//prometheus.DefaultRegisterer.Register(BaseTargetMTUGaugeVec)

}

func DataProcess(ctx context.Context, logger log.Logger) error {

	ticker := time.NewTicker(MetricCollectInterval)
	level.Info(logger).Log("msg", "DataProcessManager start....")
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			DeleteMetrics()
			NewGauge()
			NewMetrics()
			go IcmpDataProcess(logger)

		case <-ctx.Done():
			level.Info(logger).Log("msg", "DataProcessManager exit....")
			return nil
		}
	}

	return nil
}

func IcmpDataProcess(logger log.Logger) {

	level.Info(logger).Log("msg", "IcmpDataProcess run....")

	var expireds []string

	latencyMap := make(map[string][]int)
	packagedropMap := make(map[string][]int)
	targetSuccMap := make(map[string][]int)
	RttAvgMap := make(map[string][]int)
	RttMinMap := make(map[string][]int)
	RttMaxMap := make(map[string][]int)
	RttMdevMap := make(map[string][]int)
	//MTUMap := make(map[string][]int)

	f := func(k, v interface{}) bool {
		key := k.(string)
		va := v.(*pb.ProberResultOne)

		// check item expire
		now := time.Now().Unix()
		if now-va.TimeStamp > 600 {
			expireds = append(expireds, key)
		} else {
			if strings.Contains(va.MetricName, MetricOriginSeparator) {
				metricType := strings.Split(va.MetricName, MetricOriginSeparator)[1]

				uniqueKey := va.MetricName + MetricUniqueSeparator + va.SourceRegion + MetricUniqueSeparator + va.TargetRegion + MetricUniqueSeparator + va.SliceName

				//fmt.Println(metricType)
				switch metricType {
				case "latency":
					old := latencyMap[uniqueKey]
					if len(old) == 0 {
						latencyMap[uniqueKey] = []int{int(va.Value)}
					} else {
						latencyMap[uniqueKey] = append(latencyMap[uniqueKey], int(va.Value))
					}
				case "packageDrop":
					old := packagedropMap[uniqueKey]
					if len(old) == 0 {
						packagedropMap[uniqueKey] = []int{int(va.Value)}
					} else {
						packagedropMap[uniqueKey] = append(packagedropMap[uniqueKey], int(va.Value))
					}
				case "target":
					old := targetSuccMap[uniqueKey]
					if len(old) == 0 {
						targetSuccMap[uniqueKey] = []int{int(va.Value)}
					} else {
						targetSuccMap[uniqueKey] = append(targetSuccMap[uniqueKey], int(va.Value))
					}
				case "rttavg":
					old := RttAvgMap[uniqueKey]
					if len(old) == 0 {
						RttAvgMap[uniqueKey] = []int{int(va.Value)}
					} else {
						RttAvgMap[uniqueKey] = append(RttAvgMap[uniqueKey], int(va.Value))
					}
				case "rttmin":
					old := RttMinMap[uniqueKey]
					if len(old) == 0 {
						RttMinMap[uniqueKey] = []int{int(va.Value)}
					} else {
						RttMinMap[uniqueKey] = append(RttMinMap[uniqueKey], int(va.Value))
					}
				case "rttmax":
					old := RttMaxMap[uniqueKey]
					if len(old) == 0 {
						RttMaxMap[uniqueKey] = []int{int(va.Value)}
					} else {
						RttMaxMap[uniqueKey] = append(RttMaxMap[uniqueKey], int(va.Value))
					}
				case "rttmdev":
					old := RttMdevMap[uniqueKey]
					if len(old) == 0 {
						RttMdevMap[uniqueKey] = []int{int(va.Value)}
					} else {
						RttMdevMap[uniqueKey] = append(RttMdevMap[uniqueKey], int(va.Value))
					}
				}
			}

		}

		return true
	}
	IcmpDataMap.Range(f)
	// delete  expireds
	length := 0
	for _, e := range expireds {
		IcmpDataMap.Delete(e)
		length++
	}
	fmt.Println("【ResultDelete】", length)
	// compute data with avg or pct99
	dealWithDataMapAvg(latencyMap, PingLatencyGaugeVec)
	dealWithDataMapAvg(packagedropMap, PingPackageDropGaugeVec)
	dealWithDataMapAvg(RttAvgMap, PingTargetRTTGaugeVec)
	dealWithDataMapAvg(RttMinMap, PingTargetRTTMinGaugeVec)
	dealWithDataMapAvg(RttMaxMap, PingTargetRTTMaxGaugeVec)
	dealWithDataMapAvg(RttMdevMap, PingTargetRTTMdevGaugeVec)
	//dealWithDataMapAvg(MTUMap, BaseTargetMTUGaugeVec)

	dealWithDataMapBool(targetSuccMap, PingTargetSuccessGaugeVec)

	latencyMap = make(map[string][]int)
	packagedropMap = make(map[string][]int)
	targetSuccMap = make(map[string][]int)
	RttAvgMap = make(map[string][]int)
	RttMinMap = make(map[string][]int)
	RttMaxMap = make(map[string][]int)
	RttMdevMap = make(map[string][]int)
	//MTUMap = make(map[string][]int)
}

func dealWithDataMapAvg(dataM map[string][]int, promeVec *prometheus.GaugeVec) {

	for uniqueKey, datas := range dataM {
		sourceRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[1]
		targetRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[2]
		sliceName := strings.Split(uniqueKey, MetricUniqueSeparator)[3]

		var sum, avg int
		num := len(datas)
		for _, ds := range datas {
			sum += ds
		}
		avg = sum / int(num)

		promeVec.With(prometheus.Labels{"source_region": sourceRegion, "target_region": targetRegion, "slice_name": sliceName}).Set(float64(avg))

	}
}

func dealWithDataMapBool(dataM map[string][]int, promeVec *prometheus.GaugeVec) {

	for uniqueKey, datas := range dataM {
		sourceRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[1]
		targetRegion := strings.Split(uniqueKey, MetricUniqueSeparator)[2]
		sliceName := strings.Split(uniqueKey, MetricUniqueSeparator)[3]

		thisFailNum := 0

		for _, ds := range datas {
			if ds == -1 {
				thisFailNum += 1
			}
		}

		if thisFailNum == len(datas) {
			promeVec.With(prometheus.Labels{"source_region": sourceRegion, "target_region": targetRegion, "slice_name": sliceName}).Set(0)
		} else {
			promeVec.With(prometheus.Labels{"source_region": sourceRegion, "target_region": targetRegion, "slice_name": sliceName}).Set(1)
		}

	}
}
