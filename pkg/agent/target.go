package agent

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"prober/pkg/pb"
)

var (
	TargetCache      = sync.Map{}
	PbResMap         = sync.Map{}
	Probers          map[string]ProbeFn
	LTM              *LocalTargetManger
	TargetUpdateChan = make(chan *pb.ProberTargetsGetResponse, 1)
	//SiteUpdateInterval          = 250 * time.Millisecond
	MaxConcurrentGoroutines int = 200
)

//type ProbeFn func(ctx context.Context, lt *LocalTarget, logger log.Logger) pb.ProberResultOne
type ProbeFn func(lt *LocalTarget) []*pb.ProberResultOne

type LocalTargetManger struct {
	logger log.Logger
	mux    sync.RWMutex
	Map    map[string]*LocalTarget
}

func (ltm *LocalTargetManger) GetMapKeys() []string {

	count := len(ltm.Map)
	keys := make([]string, count)
	i := 0
	for hostname := range ltm.Map {
		keys[i] = hostname
		i++
	}
	return keys
}

func (ltm *LocalTargetManger) realRefreshWork(tgs *pb.ProberTargetsGetResponse) {
	level.Info(ltm.logger).Log("msg", "realRefreshWork start")
	LTM.mux.Lock()
	defer LTM.mux.Unlock()
	remoteTargetIds := make(map[string]bool)

	localIds := LTM.GetMapKeys()
	limitGoroutine := make(chan int, MaxConcurrentGoroutines)
	var wg sync.WaitGroup
	for _, t := range tgs.TargetConfigs {
		pbFunc := Probers["icmp"]
		SliceName := t.SliceName
		for _, t := range t.Targets {
			thisId := t.TargetRegion + SliceName
			remoteTargetIds[thisId] = true
			nt := &LocalTarget{
				logger:       LTM.logger,
				SourceRegion: LocalRegion,
				TargetRegion: t.TargetRegion,
				PingCmd:      t.PingCmd,
				Prober:       pbFunc,
				SliceName:    SliceName,
				TargetIP:     t.TargetIp,
				QuitChan:     make(chan struct{}),
			}
			LTM.Map[thisId] = nt
			wg.Add(1)
			limitGoroutine <- 0
			go nt.Start(&wg, limitGoroutine)
			channelSize := len(limitGoroutine)
			fmt.Printf("当前通道内容的大小: %d\n", channelSize)
		}
		//time.Sleep(SiteUpdateInterval)
	}
	wg.Wait()
	for _, key := range localIds {
		if _, found := remoteTargetIds[key]; !found {
			LTM.Map[key].Stop()
			delete(LTM.Map, key)
		}
	}

}

func NewLocalTargetManger(logger log.Logger) {
	localM := make(map[string]*LocalTarget)
	LTM = &LocalTargetManger{}
	LTM.logger = logger
	LTM.Map = localM
}

type LocalTarget struct {
	logger       log.Logger
	SourceRegion string
	TargetRegion string
	SliceName    string
	PingCmd      string
	TargetIP     string
	Prober       ProbeFn
	QuitChan     chan struct{}
}

func PushWork(logger log.Logger) {
	for {
		pushPbResults(logger)
		time.Sleep(PushInterval)
	}
}
func ReportIp(logger log.Logger) {
	for {
		reportAgentIp(logger)
		time.Sleep(ReportInterval)
	}
}

func RefreshTarget(logger log.Logger) {
	//启动一个协程监听TargetUpdateChan
	go doRefreshWork(logger)
	level.Info(logger).Log("msg", "RefreshTarget start")
	for {
		getProberTarget(logger)
		time.Sleep(RefreshInterval)
	}

}

func doRefreshWork(logger log.Logger) {
	for {
		select {
		case tgs := <-TargetUpdateChan:
			// refresh local map
			LTM.realRefreshWork(tgs)

		}
	}
}

func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	return err
}

func InitNameSpace(zid int) {
	fmt.Println("Executing namespace initialization ... ")
	for i := 1; i <= 4096; i++ {
		var useNic, ywCidr, ywIP, gwIP, namespaceID, ywNic string
		var vlanID, n int
		var err error

		if i <= 2048 {
			//SliceName = fmt.Sprintf("slice%04d", i)
			vlanID = i + 1000
			ywNic = fmt.Sprintf("ens5")
			useNic = fmt.Sprintf("%s.%d", ywNic, vlanID)
			n = i - 1
			sid := fmt.Sprintf("%03x", n)
			ywCidr = fmt.Sprintf("XXXXX%s::/32", sid)
			ywIP = fmt.Sprintf("XXXXX%s:%d::XXXXX", sid, zid)
			gwIP = fmt.Sprintf("XXXXX%s:%d::XXXXX", sid, zid)
			namespaceID = fmt.Sprintf("slice%d", i)
		} else {
			//SliceName = fmt.Sprintf("slice%04d", i)
			ywNic = fmt.Sprintf("ens6")
			vlanID = i - 1048
			useNic = fmt.Sprintf("%s.%d", ywNic, vlanID)
			n = i - 1
			sid := fmt.Sprintf("%03x", n)
			ywCidr = fmt.Sprintf("XXXXX%s::/32", sid)
			ywIP = fmt.Sprintf("XXXXX%s:%d::100:2", sid, zid)
			gwIP = fmt.Sprintf("XXXXX%s:%d::1", sid, zid)
			namespaceID = fmt.Sprintf("slice%d", i)
		}

		// Execute shell commands to configure namespaces
		cmds := []string{
			fmt.Sprintf("ip link add link %s name %s type vlan id %d", ywNic, useNic, vlanID),
			fmt.Sprintf("ip netns add %s", namespaceID),
			fmt.Sprintf("ip link set %s netns %s", useNic, namespaceID),
			fmt.Sprintf("ip netns exec %s ip -6 addr add %s/64 dev %s", namespaceID, ywIP, useNic),
			fmt.Sprintf("ip netns exec %s ip link set dev %s up", namespaceID, useNic),
			fmt.Sprintf("ip netns exec %s ip -6 route add %s via %s", namespaceID, ywCidr, gwIP),
		}

		for _, cmd := range cmds {
			err = executeCommand(cmd)
			if err != nil {
				//fmt.Printf("Error executing  namespaces : %s\n", err.Error())

			}
		}
	}
	fmt.Println("Namespace initialization completed .")
}

func Init(logger log.Logger) {

	Probers = map[string]ProbeFn{
		"icmp": ProbeICMP,
	}
	NewLocalTargetManger(logger)
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "0" {
			return
		}
	}
	InitNameSpace(LocalZid)
}

func (lt *LocalTarget) Uid() string {

	return lt.SourceRegion + "_" + lt.TargetRegion + "_" + lt.SliceName
}

func (lt *LocalTarget) Start(wg *sync.WaitGroup, limitGoroutine chan int) {
	defer func() {
		<-limitGoroutine
		wg.Done()
	}()
	res := ProbeICMP(lt)
	if len(res) > 0 {
		if res[0].Value != -1 {
			PbResMap.Store(lt.Uid(), res)
		}
	}
	return
}

func (lt *LocalTarget) Stop() {
	close(lt.QuitChan)
}
