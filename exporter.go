package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "cloudmonitor"
)

// CloudmonitorExporter collects metrics from Aliyun via cms API
type CloudmonitorExporter struct {
	client *cms.Client

	// nat gateway
	netTxRate        *prometheus.Desc
	netTxRatePercent *prometheus.Desc
	snatConnections  *prometheus.Desc

	// slb dashbaord
	activeConnection                 *prometheus.Desc
	trafficRX                        *prometheus.Desc
	trafficTX                        *prometheus.Desc
	newConnection                    *prometheus.Desc
	maxConnection                    *prometheus.Desc
	dropConnection                   *prometheus.Desc
	dropPacketRX                     *prometheus.Desc
	dropPacketTX                     *prometheus.Desc
	dropTrafficRX                    *prometheus.Desc
	dropTrafficTX                    *prometheus.Desc
	instanceNewConnectionUtilization *prometheus.Desc
	instanceUpstreamCode5xx          *prometheus.Desc
	instanceStatusCode5xx            *prometheus.Desc
	instanceRt                       *prometheus.Desc
	instanceQps                      *prometheus.Desc
	instanceQpsUtilization           *prometheus.Desc
	instanceTrafficRX                *prometheus.Desc
	instanceTrafficTX                *prometheus.Desc

	// rds dashbaord
	cpuUsage        *prometheus.Desc
	connectionUsage *prometheus.Desc
	activeSessions  *prometheus.Desc
}

// NewExporter instantiate an CloudmonitorExport
func NewExporter(c *cms.Client) *CloudmonitorExporter {
	return &CloudmonitorExporter{
		client: c,

		// nat gateway
		netTxRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "net_tx_rate", "bytes"),
			"Outbound bandwith of gateway in bits/s",
			[]string{
				"id",
			},
			nil,
		),
		netTxRatePercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "net_tx_rate", "percent"),
			"Outbound bandwith of gateway used in percentage",
			[]string{
				"id",
			},
			nil,
		),
		snatConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "snat", "connections"),
			"Max number of snat connections per minute",
			[]string{
				"id",
			},
			nil,
		),

		// slb dashboard
		instanceQps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "instance_qps"),
			"Seven-layer protocol instance Queries-per-second",
			[]string{
				"id",
			},
			nil,
		),

		instanceQpsUtilization: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "instance_qps_utilization"),
			"Seven-layer protocol instance Queries-per-second used in percentage",
			[]string{
				"id",
			},
			nil,
		),

		instanceTrafficRX: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "traffic_rx_average"),
			"Average traffic received per second",
			[]string{
				"id",
			},
			nil,
		),

		instanceTrafficTX: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "traffic_tx_average"),
			"Average traffic sent per second",
			[]string{
				"id",
			},
			nil,
		),

		instanceRt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "request_time"),
			"slb request time",
			[]string{
				"id",
			},
			nil,
		),

		instanceUpstreamCode5xx: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "upstream_code_5xx"),
			"Backend server 5xx error",
			[]string{
				"id",
			},
			nil,
		),

		instanceStatusCode5xx: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "status_code_5xx"),
			"5xx error in the instance itself",
			[]string{
				"id",
			},
			nil,
		),

		instanceNewConnectionUtilization: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "slb", "new_connection_utilization"),
			"Average number of new connections created per second in percentage",
			[]string{
				"id",
			},
			nil,
		),

		// rds dashbaord
		cpuUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rds", "cpu_usage_average"),
			"CPU usage per minute",
			[]string{
				"id",
			},
			nil,
		),

		connectionUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rds", "connection_usage"),
			"Connection usage per minute",
			[]string{
				"id",
			},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the cloudmonitor exporter.
// It implements prometheus.Collector.
func (e *CloudmonitorExporter) Describe(ch chan<- *prometheus.Desc) {
	// nat gateway
	ch <- e.netTxRate
	ch <- e.netTxRatePercent
	ch <- e.snatConnections

	// slb dashboard
	ch <- e.instanceQps
	ch <- e.instanceQpsUtilization
	ch <- e.instanceTrafficRX
	ch <- e.instanceTrafficTX
	ch <- e.instanceStatusCode5xx
	ch <- e.instanceUpstreamCode5xx
	ch <- e.instanceRt
	ch <- e.instanceNewConnectionUtilization

	// rds dashboard
	ch <- e.cpuUsage
	ch <- e.connectionUsage
}

// Collect fetches the metrics from Aliyun cms
// It implements prometheus.Collector.
func (e *CloudmonitorExporter) Collect(ch chan<- prometheus.Metric) {
	natGateway := NewNatGateway(e.client)
	slbDashboard := NewSLBDashboard(e.client)
	rdsDashboard := NewRDSDashboard(e.client)

	for _, point := range natGateway.retrieveNetTxRate() {
		ch <- prometheus.MustNewConstMetric(
			e.netTxRate,
			prometheus.GaugeValue,
			float64(point.Value),
			point.InstanceId,
		)
	}

	for _, point := range natGateway.retrieveNetTxRatePercent() {
		ch <- prometheus.MustNewConstMetric(
			e.netTxRatePercent,
			prometheus.GaugeValue,
			float64(point.Value),
			point.InstanceId,
		)
	}

	for _, point := range natGateway.retrieveSnatConn() {
		ch <- prometheus.MustNewConstMetric(
			e.snatConnections,
			prometheus.GaugeValue,
			float64(point.Maximum),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceQps() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceQps,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceQpsUtilization() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceQpsUtilization,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveTrafficRX() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceTrafficRX,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveTrafficTX() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceTrafficTX,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceNewConnectionUtilization() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceNewConnectionUtilization,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceRt() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceRt,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceStatusCode5xx() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceStatusCode5xx,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range slbDashboard.retrieveInstanceUpstreamCode5xx() {
		ch <- prometheus.MustNewConstMetric(
			e.instanceUpstreamCode5xx,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range rdsDashboard.retrieveCPUUsage() {
		ch <- prometheus.MustNewConstMetric(
			e.cpuUsage,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

	for _, point := range rdsDashboard.retrieveConnectionUsage() {
		ch <- prometheus.MustNewConstMetric(
			e.connectionUsage,
			prometheus.GaugeValue,
			float64(point.Average),
			point.InstanceId,
		)
	}

}
