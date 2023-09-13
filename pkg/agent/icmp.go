package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"prober/pkg/common"
	"prober/pkg/pb"
)

func execCmd(cmdStr string, logger log.Logger) (success bool, outStr string) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		level.Error(logger).Log("execCmdMsg", err, "cmd", cmdStr)

		if strings.Contains(err.Error(), "killed") {
			return false, "killed"
		}

		return false, string(stderr.Bytes())
	}
	outStr = string(stdout.Bytes())
	return true, outStr

}

func ProbeICMP(lt *LocalTarget) []*pb.ProberResultOne {

	defer func() {
		if r := recover(); r != nil {
			resultErr, _ := r.(error)
			level.Error(lt.logger).Log("msg", "ProbeICMP panic ...", "resultErr", resultErr)

		}
	}()
	//GetLocalIp(lt.logger)

	pingCmd := lt.PingCmd
	success, outPutStr := execCmd(pingCmd, lt.logger)
	prs := make([]*pb.ProberResultOne, 0)

	var (
		pkgdLine       string
		latenLinke     string
		rttLine        string
		pkgRateNum     float64
		pingEwmaNum    float64
		pingRttNum     float64
		pingRttMinNum  float64
		pingRttMaxNum  float64
		pingRttMdevNum float64
		pingSuccess    float64
		baseMTUNum     float64
	)
	pkgRateNum = -1
	pingEwmaNum = -1
	pingSuccess = -1
	pingRttNum = -1
	pingRttMinNum = -1
	pingRttMaxNum = -1
	pingRttMdevNum = -1
	pingRttNum = -1
	baseMTUNum = -1

	prSu := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingTargetSuccess,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingSuccess),
		SliceName:    lt.SliceName,
	}
	prDr := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingPackageDrop,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pkgRateNum),
		SliceName:    lt.SliceName,
	}

	prLaten := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingLatency,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingEwmaNum),
		SliceName:    lt.SliceName,
	}

	prRttAvg := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingRTTAvg,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingRttNum),
		SliceName:    lt.SliceName,
	}
	prRttMin := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingRTTMin,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingRttMinNum),
		SliceName:    lt.SliceName,
	}
	prRttMax := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingRTTMax,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingRttMaxNum),
		SliceName:    lt.SliceName,
	}
	prRttMdev := pb.ProberResultOne{
		MetricName:   common.MetricsNamePingRTTMdev,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(pingRttMdevNum),
		SliceName:    lt.SliceName,
	}
	prMTU := pb.ProberResultOne{
		MetricName:   common.MetricsNameBaseMTU,
		SourceRegion: lt.SourceRegion,
		TargetRegion: lt.TargetRegion,
		TimeStamp:    time.Now().Unix(),
		Value:        int32(baseMTUNum),
		SliceName:    lt.SliceName,
	}

	if success == false {

		prSu.Value = -1
		prs = append(prs, &prSu)
		prs = append(prs, &prDr)
		prs = append(prs, &prLaten)
		prs = append(prs, &prRttAvg)
		prs = append(prs, &prRttMin)
		prs = append(prs, &prRttMax)
		prs = append(prs, &prRttMdev)
		if lt.SliceName == "silce0000" {
			prs = append(prs, &prMTU)
		}
		return prs

	}

	for _, line := range strings.Split(outPutStr, "\n") {
		if strings.Contains(line, "packets transmitted") {
			pkgdLine = line
			continue
		}
		if strings.Contains(line, "min/avg/max/mdev") {
			latenLinke = line
			rttLine = line
			continue
		}

	}
	/*
		PING 10.21.45.237 (10.21.45.237) 100(128) bytes of data.
		--- 10.21.45.237 ping statistics ---
		50 packets transmitted, 0 received, 100% packet loss, time 499ms
		rtt min/avg/max/mdev = 0.002/0.003/0.042/0.005 ms, ipg/ewma 0.006/0.003 ms
	*/

	if len(pkgdLine) > 0 {

		pkgRate := strings.Split(pkgdLine, " ")[5]
		pkgRate = strings.Replace(pkgRate, "%", "", -1)
		pkgRateNum, _ = strconv.ParseFloat(pkgRate, 64)
		prDr.Value = int32(pkgRateNum)
	}

	if len(latenLinke) > 0 {
		pingEwmas := strings.Split(latenLinke, " ")
		pingEwma := pingEwmas[len(pingEwmas)-2]
		pingEwma = strings.Split(pingEwma, "/")[1]
		pingEwmaNum, _ = strconv.ParseFloat(pingEwma, 64)
		prLaten.Value = int32(pingEwmaNum)
	}
	//rtt avg
	if len(rttLine) > 0 {
		pingRttAvg := strings.Split(rttLine, " ")
		pingRtt := pingRttAvg[3]
		pingRttNum, _ = strconv.ParseFloat(strings.Split(pingRtt, "/")[1], 64)
		prRttAvg.Value = int32(pingRttNum)
		pingRttNum, _ = strconv.ParseFloat(strings.Split(pingRtt, "/")[0], 64)
		prRttMin.Value = int32(pingRttNum)
		pingRttNum, _ = strconv.ParseFloat(strings.Split(pingRtt, "/")[2], 64)
		prRttMax.Value = int32(pingRttNum)
		pingRttNum, _ = strconv.ParseFloat(strings.Split(pingRtt, "/")[3], 64)
		prRttMdev.Value = int32(pingRttNum)
	}

	//基础网络测试MTU
	if lt.SliceName == "silce0000" {
		mtu, err := getMTU(lt.TargetIP)
		if err != nil {
			fmt.Println("Error!:", err)
			prMTU.Value = -1
		}
		prMTU.Value = int32(mtu)

	}

	if pkgRateNum == 100 {
		prSu.Value = -1
	} else {
		prSu.Value = 1
	}

	prs = append(prs, &prSu)
	prs = append(prs, &prDr)
	prs = append(prs, &prLaten)
	prs = append(prs, &prRttAvg)
	prs = append(prs, &prRttMin)
	prs = append(prs, &prRttMax)
	prs = append(prs, &prRttMdev)
	if lt.SliceName == "silce0000" {
		prs = append(prs, &prMTU)
	}

	return prs
}
