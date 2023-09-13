package server

import (
	"strconv"
	"sync"

	"github.com/go-kit/kit/log"

	"fmt"
	"prober/pkg/pb"
)

var (
	IcmpRegionProberMap = sync.Map{}
)

type TargetFlushManager struct {
	Logger     log.Logger
	ConfigFile string
}

type PointInfo struct {
	Zid   int
	YwNic string
}

func GerTargetsByIP(sourceNodeName string, sourceVLANId int, sourceIP string, sourceRounds int) (res []*pb.TargetConfig) {
	VirtualID, err := strconv.Atoi(sourceNodeName)
	if err != nil {
		fmt.Println("target_pool.go:String conversion to int failed.", err)
		return res
	}
	for i := 1; i <= 4096; i++ {
		sid := ""
		var tres *pb.TargetConfig
		//Number of agents on the same site
		var Count int
		Count = 3
		//0 40 80
		if i%Count != VirtualID/40 {
			continue
		}
		if i >= 1 && i <= 4096 {
			sid = fmt.Sprintf("%03x", i-1)
			var ttres []*pb.Targets
			for j := 1; j <= 40; j++ {
				temptarget := &pb.Targets{
					TargetIp:     "",
					TargetRegion: strconv.Itoa(j),
					//ip netns exec slice" + fmt.Sprintf("%d", i) + " ping6 XXX -c 5 -i 0.01 -f -W 0.2
					PingCmd: "",
				}
				ttres = append(ttres, temptarget)
			}
			tres = &pb.TargetConfig{
				Targets:   ttres,
				SliceName: fmt.Sprintf("slice%04d", i),
			}

		}
		res = append(res, tres)

	}
	return res

}
