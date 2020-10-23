package chain

import (
	"context"
	"github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/pkg/client/perf"
	"github.com/epmd-edp/perf-operator/pkg/controller/perfserver/chain/handler"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type CheckConnectionToPerf struct {
	next       handler.PerfServerHandler
	client     client.Client
	perfClient perf.PerfClient
}

func (h CheckConnectionToPerf) ServeRequest(server *v1alpha1.PerfServer) error {
	log.Info("start checking connection to PERF", "url", server.Spec.RootUrl)
	connected, err := h.perfClient.Connected()
	if err != nil {
		return err
	}

	if !connected {
		server.Status.Available = false
		err := errors.Wrapf(err, "couldn't connect to PERF instance with %v url", server.Spec.RootUrl)
		server.Status.DetailedMessage = err.Error()
		return err
	}
	server.Status.Available = true
	server.Status.DetailedMessage = "connected"

	h.updateStatus(server)

	log.Info("connection to PERF has been established", "url", server.Spec.RootUrl)
	return nextServeOrNil(h.next, server)
}

func (h CheckConnectionToPerf) updateStatus(server *v1alpha1.PerfServer) {
	server.Status.LastTimeUpdated = time.Now()
	if err := h.client.Status().Update(context.TODO(), server); err != nil {
		_ = h.client.Update(context.TODO(), server)
	}
}