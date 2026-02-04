package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	pb "github.com/s0n1cAK/yandex-metrics/internal/proto"
	"github.com/s0n1cAK/yandex-metrics/internal/service/metrics"

	"google.golang.org/grpc/metadata"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	svc metrics.Service
}

func NewMetricsServer(svc metrics.Service) *MetricsServer {
	return &MetricsServer{svc: svc}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return &pb.UpdateMetricsResponse{}, nil
	}

	ip := realIPFromMD(ctx)

	converted := make([]models.Metrics, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		if m == nil {
			continue
		}

		switch m.Type {
		case pb.Metric_GAUGE:
			converted = append(converted, models.Metrics{
				ID:    m.Id,
				MType: models.Gauge,
				Value: lib.FloatPtr(m.Value),
			})
		case pb.Metric_COUNTER:
			converted = append(converted, models.Metrics{
				ID:    m.Id,
				MType: models.Counter,
				Delta: lib.IntPtr(m.Delta),
			})
		default:
			return nil, fmt.Errorf("unknown metric type: %v", m.Type)
		}
	}

	if err := s.svc.SetBatch(ctx, converted, ip); err != nil {
		return nil, err
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func realIPFromMD(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("x-real-ip")
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(vals[0])
}
