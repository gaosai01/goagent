package main

/*
重构代码后分数为3736,对应本地跑分为：1865.54
删除日志打印输出后和使用了部分sync.pool线上分数为：3962.3600，本地为：1884.05
使用http2好像不行
修改http的idle时间和setkeepalive评分3400，换成fasthttp测试结果：4100
tcp传输数据write写整体，跑分4200
 */

const (
	VersionName = "1.1.0"
)
