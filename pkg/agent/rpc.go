package agent

import (
	"context"
	"time"

	"github.com/flyaways/pool"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/grpc"

	"prober/pkg/pb"
)

var (
	GrpcPool *pool.GRPCPool
)

const (
	RefreshInterval = 15 * time.Second
	PushInterval    = 15 * time.Second
	ReportInterval  = 60 * time.Second
	//连接
	retryInterval = time.Second * 5 // 连接重试间隔
	maxRetryCount = 3               // 最大连接重试次数

)

func InitRpcPool(serverAddr string, logger log.Logger) bool {

	options := &pool.Options{
		InitTargets:  []string{serverAddr},
		InitCap:      5,
		MaxCap:       100,
		DialTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		ReadTimeout:  time.Second * 50,
		WriteTimeout: time.Second * 50,
	}

	var err error
	var retryCount int

	for retryCount = 0; retryCount < maxRetryCount; retryCount++ {

		GrpcPool, err = pool.NewGRPCPool(options, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16777216)))
		if err == nil {
			break
		}

		level.Error(logger).Log("init_rpc_pool_failed_error", err, "retry_count", retryCount)
		time.Sleep(retryInterval)
	}

	if retryCount == maxRetryCount {
		level.Error(logger).Log("msg", "init_GrpcPool_failed")
		return false
	}

	return true

}
func reportAgentIp(logger log.Logger) {
	level.Info(logger).Log("msg", "reportAgentIp run...")

	var conn *grpc.ClientConn
	var err error
	for {
		conn, err = GrpcPool.Get()
		if err != nil {
			level.Error(logger).Log("get_rpc_conn_from_pool_err", err)
			time.Sleep(5 * time.Second) // 休眠 5 秒钟
			continue
		}
		break
	}

	defer conn.Close()
	c := pb.NewProberAgentIpReportClient(conn)
	t := pb.ProberAgentIpReportRequest{Ip: LocalIp}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.ProberAgentIpReports(ctx, &t)
	if err != nil {
		level.Error(logger).Log("msg", "could_not_reportAgentIp", "Ip", LocalIp, "Region:", LocalNodeName)
		return
	}

	level.Info(logger).Log("reportAgentIpResult", r)

}
func getProberTarget(logger log.Logger) {
	level.Info(logger).Log("msg", "getProberTarget run...")
	conn, err := GrpcPool.Get()
	if err != nil {
		level.Error(logger).Log("get_rpc_conn_from_pool_err", err)
		return
	}

	defer conn.Close()
	c := pb.NewGetProberTargetClient(conn)

	// Contact the server and print out its response.
	name := LocalNodeName
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := c.GetProberTargets(ctx, &pb.ProberTargetsGetRequest{AgentRound: AgentRound, LocalIp: LocalIp, LocalNodeName: LocalNodeName, LocalVlanId: LocalVLANID})
	if err != nil {
		level.Error(logger).Log("msg", "could_not_get_target", "name", name, "error:", err)
		return
	}

	if len(r.TargetConfigs) > 0 {

		TargetUpdateChan <- r
	} else {
		level.Info(logger).Log("msg", "receive_empty_targets")
	}

}

//问题
func pushPbResults(logger log.Logger) {

	var prs []*pb.ProberResultOne
	f := func(k, v interface{}) bool {

		va := v.([]*pb.ProberResultOne)
		//key := k.(string)
		//fmt.Println(key, va)
		prs = append(prs, va...)
		return true
	}
	PbResMap.Range(f)
	if len(prs) == 0 {
		level.Info(logger).Log("msg", "empty_result_list_not_to_push")
		return
	}

	conn, err := GrpcPool.Get()
	if err != nil {
		level.Error(logger).Log("get_rpc_conn_from_pool_err", err)
		return
	}

	defer conn.Close()
	c := pb.NewPushProberResultClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	r, err := c.PushProberResults(ctx, &pb.ProberResultPushRequest{ProberResults: prs})
	if err != nil {
		level.Error(logger).Log("msg", "could_not_push_result ", "prs", prs, "error:", err)
		numElements := 0
		PbResMap.Range(func(_, _ interface{}) bool {
			numElements++
			return true
		})
		if numElements > 1000 {
			PbResMap.Range(func(key, value interface{}) bool {
				PbResMap.Delete(key)
				return true
			})
		}
	} else {
		PbResMap.Range(func(key, value interface{}) bool {
			PbResMap.Delete(key)
			return true
		})
	}
	level.Info(logger).Log("pushPbResults", r)
}
