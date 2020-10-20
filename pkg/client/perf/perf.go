package perf

import (
	"github.com/epmd-edp/perf-operator/pkg/model"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type PerfClient interface {
	GetDataSource(id string) (*model.DataSource, error)
}

type PerfClientAdapter struct {
	client resty.Client
}

var log = logf.Log.WithName("perf_client")

func NewRestClient(url, user, pwd string) (*PerfClientAdapter, error) {
	rl := log.WithValues("url", url, "user", user)
	rl.Info("initializing new Perf REST client.")
	token, err := getToken(url, user, pwd)
	if err != nil {
		return nil, err
	}
	cl := resty.New().
		SetHostURL(url).
		SetAuthToken(token)
	rl.Info("Perf REST client successfully has been created.")
	return &PerfClientAdapter{
		client: *cl,
	}, err
}

func getToken(url, user, pwd string) (string, error) {
	resp, err := resty.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"accept":       "text/plain",
		}).
		SetFormData(map[string]string{
			"username":       user,
			"password":       pwd,
			"useExternalSSO": "true",
		}).Post(url + "/api/v2/sso/token")
	if err != nil || resp.IsError() {
		return "", errors.Wrapf(err, "Couldn't get PERF token for %v user.", user)
	}
	return resp.String(), nil
}

func (c PerfClientAdapter) GetDataSource(id string) (ds *model.DataSource, err error) {
	log.Info("start retrieving PERF datasource", "id", id, "url", c.client.HostURL)
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&ds).
		SetPathParams(map[string]string{
			"id": id,
		}).
		Get("/api/v2/datasources/{id}")
	if err != nil || resp.IsError() {
		return nil, errors.Wrapf(err, "couldn't get PERF datasource by %v id", id)
	}
	return
}

func (c PerfClientAdapter) CheckConnection(id string) error {
	return nil
}
