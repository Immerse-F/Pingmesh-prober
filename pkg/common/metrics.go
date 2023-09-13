package common

/*
   metics name should be `module_type_unit`
*/
const (
	//ping -f 泛洪模式下 ewma  （ms）
	MetricsNamePingLatency = `ping_latency_millonseconds`

	//Loss丢包率 （-）
	MetricsNamePingPackageDrop = `ping_packageDrop_rate`

	//ping 执行是否成功 （bool）
	MetricsNamePingTargetSuccess = `ping_target_success`

	//ping 命令 avg rtt （ms）
	MetricsNamePingRTTAvg = `ping_rttavg`

	//ping 命令 min rtt （ms）
	MetricsNamePingRTTMin = `ping_rttmin`

	//ping 命令 max rtt （ms）
	MetricsNamePingRTTMax = `ping_rttmax`

	//ping 命令 mdev rtt （ms）
	MetricsNamePingRTTMdev = `ping_rttmdev`

	//基础网络MTU (B)
	MetricsNameBaseMTU = `Base_MTU`
)
