package server

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"prober/pkg/pb"
)

var (
	mu  sync.Mutex
	GRM *GRpcServerManager
)

type GRpcServerManager struct {
	Logger            log.Logger
	GrpcListenAddress string
	Server            *grpc.Server
}

func NewManagager(logger log.Logger, addr string) {
	gp := GRpcServerManager{
		Logger:            logger,
		GrpcListenAddress: addr,
	}

	GRM = &gp
}

func (gs *GRpcServerManager) Run(ctx context.Context, logger log.Logger) error {
	//6001
	lis, err := net.Listen("tcp", gs.GrpcListenAddress)
	if err != nil {
		level.Error(gs.Logger).Log("msg", "grpc failed to listen: ", "err", err)
	}
	s := grpc.NewServer(
		//16MB
		grpc.MaxRecvMsgSize(1677721600),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			// 设置连接最大的空闲时间，超过就释放
			MaxConnectionIdle: 5 * time.Minute,
		}))
	gs.Server = s

	// register service
	pb.RegisterGetProberTargetServer(s, &PServer{logger: logger})
	pb.RegisterPushProberResultServer(s, &PResult{logger: logger})
	pb.RegisterProberAgentIpReportServer(s, &PAgentR{logger: logger})
	pb.RegisterVlanServiceServer(s, &PVlanService{logger: logger})
	pb.RegisterVlanSwitchNotifyServer(s, &PNotifyVlanSwitched{logger: logger})

	level.Info(gs.Logger).Log("msg", "grpc success to serve", "addr", gs.GrpcListenAddress)
	if err := s.Serve(lis); err != nil {
		level.Error(gs.Logger).Log("msg", "grpc failed to serve err", "err", err)
		return err
	}

	return nil
}

type PAgentR struct {
	pb.UnimplementedProberAgentIpReportServer
	logger log.Logger
}

type PServer struct {
	pb.UnimplementedGetProberTargetServer
	logger log.Logger
}

type PResult struct {
	pb.UnimplementedPushProberResultServer
	logger log.Logger
}

type PVlanService struct {
	pb.UnimplementedVlanServiceServer
	logger log.Logger
}

type PNotifyVlanSwitched struct {
	pb.UnimplementedVlanSwitchNotifyServer
	logger log.Logger
}

func (pr *PAgentR) ProberAgentIpReports(ctx context.Context, in *pb.ProberAgentIpReportRequest) (*pb.ProberAgentIpReportResponse, error) {
	level.Debug(pr.logger).Log("msg", "ProberAgentIpReports receive", "args", in)
	return &pb.ProberAgentIpReportResponse{IsSuccess: true}, nil
}

func (s *PServer) GetProberTargets(ctx context.Context, in *pb.ProberTargetsGetRequest) (*pb.ProberTargetsGetResponse, error) {
	level.Info(s.logger).Log("msg", "GetProberTargets receive", "ip", in.LocalIp)
	tgs := GerTargetsByIP(in.LocalNodeName, int(in.LocalVlanId), in.LocalIp, int(in.AgentRound))
	return &pb.ProberTargetsGetResponse{TargetConfigs: tgs}, nil
}

func GetProbeResultUid(prr *pb.ProberResultOne) (uid string) {
	uid = prr.MetricName + prr.SourceRegion + prr.TargetRegion + prr.SliceName
	return
}

func (pr *PResult) PushProberResults(ctx context.Context, in *pb.ProberResultPushRequest) (*pb.ProberResultPushResponse, error) {
	level.Debug(pr.logger).Log("msg", "PushProberResult receive", "args", in)
	suNum := 0
	for _, prr := range in.ProberResults {
		uid := GetProbeResultUid(prr)
		IcmpDataMap.Store(uid, prr)
		suNum += 1
	}
	return &pb.ProberResultPushResponse{SuccessNum: int32(suNum)}, nil
}
