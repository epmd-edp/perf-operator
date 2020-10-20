package chain

import (
	"context"
	"github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/pkg/client/perf"
	"github.com/epmd-edp/perf-operator/pkg/controller/perfserver/chain/handler"
	"github.com/epmd-edp/perf-operator/pkg/util"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type CheckConnectionToPerf struct {
	next   handler.PerfServerHandler
	client client.Client
}

func (h CheckConnectionToPerf) ServeRequest(server *v1alpha1.PerfServer) error {
	log.Info("start checking connection to PERF", "url", server.Spec.RootUrl)
	if err := h.connectToPerf(server); err != nil {
		server.Status.Available = false
		err := errors.Wrapf(err, "couldn't connect to PERF instance", "url", server.Spec.RootUrl)
		server.Status.DetailedMessage = err.Error()
		h.updateStatus(server)
		return err
	}
	server.Status.Available = true
	log.Info("connection to PERF has been established", "url", server.Spec.RootUrl)
	return nil
}

func (h CheckConnectionToPerf) connectToPerf(server *v1alpha1.PerfServer) error {
	_, err := h.newPerfRestClient(server.Spec.ApiUrl, server.Spec.CredentialName, server.Namespace)
	if err != nil {
		return errors.Wrapf(err, "couldn't init PERF client", "url", server.Spec.RootUrl)
	}
	return nil
}

func (h CheckConnectionToPerf) newPerfRestClient(url, secretName, namespace string) (*perf.PerfClientAdapter, error) {
	s, err := util.GetSecret(h.client, secretName, namespace)
	if err != nil {
		return nil, err
	}

	c, err := perf.NewRestClient(url, string(s.Data["username"]), string(s.Data["password"]))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (h CheckConnectionToPerf) updateStatus(server *v1alpha1.PerfServer) {
	server.Status.LastTimeUpdated = time.Now()
	if err := h.client.Status().Update(context.TODO(), server); err != nil {
		_ = h.client.Update(context.TODO(), server)
	}
}
