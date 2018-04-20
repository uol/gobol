package grpc

/*
import (
	"time"
	"net"
	"fmt"

	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/uol/gobol/stats"

	pb "github.com/uol/mycenae/lib/proto"
	"github.com/uol/mycenae/lib/structs"
)
*/
// TODO: First need to change all errors to the new gobol.Error. Than will be easier to create a "HandleGRPCpacket" on
// the collector package
/*
func New () error {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterTimeseriesServer(s, &server{})
	s.Serve(lis)

	return nil
}

type server struct {
	tstats        *stats.Stats
}

func (s *server) SavePoints(ctx context.Context, pnts *pb.Points) (*pb.SaveErrors, error) {

	start := time.Now()

	go s.tstats.Increment("grpcNumberRequests")

	//points := []structs.MsgV2{}

	poin := pnts.GetPoints()

	if len(poin) == 0 {
		return nil, fmt.Errorf("no points")
	}

	returnPoints := structs.RestErrors{}

	restChan := make(chan structs.RestError)
	defer close(restChan)

	for _, point := range points {

		go s.tstats.Increment("grpcReceivedFloat")

		go s.collector.HandleRESTpacket(rt, point, true, restChan)

	}

	for range points {
		re := <-restChan
		if re.Error != nil {
			go trest.tstats.Increment("v2httpNumberError")

			_, ep := trest.treatError(re.Error)

			reu := structs.RestErrorUser{
				Datapoint: re.Datapoint,
				Error:     ep.Error,
			}

			returnPoints.Errors = append(returnPoints.Errors, reu)
		}
	}

	if len(returnPoints.Errors) > 0 {

		returnPoints.Failed = len(returnPoints.Errors)
		returnPoints.Success = len(points) - len(returnPoints.Errors)

		rt.SendSuccess(http.StatusBadRequest, returnPoints)
		return
	}

	rt.SendSuccess(http.StatusNoContent, nil)
	go trest.tstats.ValueAdd(
		"v2httpNumberRequestsTime",
		float64(time.Since(start).Nanoseconds())/float64(time.Millisecond),
	)
	return

}

func (s *server) QueryExpression(ctx context.Context, *pb.Expression) (*pb.Tsdata, error) {

}
*/
