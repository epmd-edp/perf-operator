package chain

import (
	"github.com/epmd-edp/perf-operator/pkg/controller/perfserver/chain/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("perf_server_handler")

func CreateDefChain(client client.Client) handler.PerfServerHandler {
	return CheckConnectionToPerf{
		client: client,
	}
}
