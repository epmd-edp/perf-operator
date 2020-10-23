package perf

import (
	"github.com/epmd-edp/perf-operator/pkg/client/luminate"
	"github.com/epmd-edp/perf-operator/pkg/model/command"
	"github.com/epmd-edp/perf-operator/pkg/model/dto"
	"github.com/epmd-edp/perf-operator/pkg/util/cluster"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
)

type PerfClient interface {
	Connected() (bool, error)
	GetProject(name string) (ds *dto.PerfProject, err error)
	ProjectExists(name string) (bool, error)
	GetProjectDataSource(projectName, dataSourceName string) (*dto.DataSource, error)
	CreateDataSource(projectName string, command command.DataSourceCreateCommand) error
	ActivateDataSource(projectName string, dataSourceId int) error
	UpdateDataSource(ds dto.DataSource) error
}

type PerfClientAdapter struct {
	client resty.Client
}

var log = logf.Log.WithName("perf_client")

const luminatesecConfigMapName = "luminatesec-conf"

func NewRestClient(url, user, pwd, lumToken string) (*PerfClientAdapter, error) {
	rl := log.WithValues("url", url, "user", user)
	rl.Info("initializing new Perf REST client.")

	token, err := getAuthorizationToken(url, user, pwd, lumToken)
	if err != nil {
		return nil, err
	}

	cl := resty.New().
		SetHostURL(url).
		SetHeader("lum-api-token", lumToken).
		SetAuthToken(token)
	rl.Info("Perf REST client successfully has been created.")
	return &PerfClientAdapter{
		client: *cl,
	}, err
}

func GetPerfCredentials(client client.Client, secretName, namespace string) (*dto.PerfCredentials, error) {
	cm, err := cluster.GetConfigMap(client, luminatesecConfigMapName, namespace)
	if err != nil {
		return nil, err
	}

	lumClient := luminate.NewLuminateRestClient(cm.Data["apiUrl"])

	lumSecret, err := cluster.GetSecret(client, cm.Data["secretName"], namespace)
	if err != nil {
		return nil, err
	}

	lumToken, err := lumClient.GetApiToken(string(lumSecret.Data["username"]), string(lumSecret.Data["password"]))
	if err != nil {
		return nil, err
	}

	s, err := cluster.GetSecret(client, secretName, namespace)
	if err != nil {
		return nil, err
	}

	return &dto.PerfCredentials{
		Username:      string(s.Data["username"]),
		Password:      string(s.Data["password"]),
		LuminateToken: *lumToken,
	}, nil
}

func getAuthorizationToken(url, user, pwd, lumApiToken string) (string, error) {
	resp, err := resty.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"accept":        "text/plain",
			"lum-api-token": lumApiToken,
		}).
		SetFormData(map[string]string{
			"username":       user,
			"password":       pwd,
			"useExternalSSO": "false", // weird behaviour. at this moment should be false despite of using lumApiToken
		}).Post(url + "/api/v2/sso/token")
	if err != nil {
		return "", errors.Wrap(err, "couldn't get PERF token for %v user.")
	}
	if resp.IsError() {
		return "", errors.Errorf("couldn't get PERF token for %v user. Status - %v", user, resp.StatusCode())
	}
	return resp.String(), nil
}

func (c PerfClientAdapter) Connected() (bool, error) {
	log.Info("start checking connection to PERF", "url", c.client.HostURL)
	_, err := c.getProjects()
	if err != nil {
		return false, errors.Wrapf(err, "couldn't establish connection with PERF %v", c.client.HostURL)
	}
	log.Info("connection to PERF was established.", "url", c.client.HostURL)
	return true, nil
}

func (c PerfClientAdapter) GetProject(name string) (ds *dto.PerfProject, err error) {
	projects, err := c.getProjects()
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if p.Name == strings.ToUpper(name) {
			return &p, nil
		}
	}
	return nil, nil
}

func (c PerfClientAdapter) ProjectExists(name string) (bool, error) {
	log.Info("start checking project for existence", "name", name)
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("name", name).
		Get("/api/v2/nodes/check/name")
	if err != nil {
		return false, errors.Wrapf(err, "couldn't check %v project for existence.", name)
	}
	if resp.IsError() {
		if resp.StatusCode() == 400 {
			return true, nil
		}
		return false, errors.Errorf("couldn't check %v project for existence.. Status - %v", name, resp.StatusCode())
	}
	return false, nil
}

func (c PerfClientAdapter) getProjects() ([]dto.PerfProject, error) {
	var pp []dto.PerfProject
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&pp).
		Get("/api/v2/nodes")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get projects from PERF")
	}
	if resp.IsError() {
		return nil, errors.Errorf("couldn't get projects from PERF. Status - %v", resp.StatusCode())
	}
	return pp, nil
}

func (c PerfClientAdapter) GetProjectDataSource(projectName, dataSourceName string) (*dto.DataSource, error) {
	rlog := log.WithValues("projectName", projectName, "dataSourceName", dataSourceName)
	rlog.Info("start retrieving PERF datasource")
	project, err := c.GetProject(projectName)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.Errorf("PERF project %v wasn't found", projectName)
	}

	dss, err := c.getProjectDataSources(project.Id)
	if err != nil {
		return nil, err
	}

	for _, ds := range dss {
		if ds.Name == strings.ToUpper(dataSourceName) {
			rlog.Info("datasource has been found in PERF.")
			return &ds, nil
		}
	}
	rlog.Info("datasource has not been found in PERF.")
	return nil, nil
}

func (c PerfClientAdapter) getProjectDataSources(projectId int) ([]dto.DataSource, error) {
	var ds []dto.DataSource
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&ds).
		SetPathParams(map[string]string{
			"id": strconv.Itoa(projectId),
		}).
		Get("/api/v2/nodes/{id}/datasets/datasources")
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get datasources for %v project", projectId)
	}
	if resp.IsError() {
		return nil, errors.Errorf("couldn't get datasources %v project. Status - %v", projectId, resp.StatusCode())
	}
	return ds, nil
}

func (c PerfClientAdapter) CreateDataSource(projectName string, command command.DataSourceCreateCommand) error {
	rlog := log.WithValues("project name", projectName, "datasource name", command.Name)
	rlog.Info("start creating datasource under project")
	project, err := c.GetProject(projectName)
	if err != nil {
		return err
	}
	if project == nil {
		return errors.Errorf("PERF project %v wasn't found", projectName)
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParams(map[string]string{
			"id": strconv.Itoa(project.Id),
		}).
		SetBody(command).
		Post("/api/v2/datasources/node/{id}")
	if err != nil {
		return errors.Wrapf(err, "couldn't create %v datasource under %v project", command.Name, projectName)
	}
	if resp.IsError() {
		return errors.Errorf("couldn't create %v datasource under %v project. Status - %v",
			command.Name, projectName, resp.StatusCode())
	}

	rlog.Info("datasource has been created.")
	return nil
}

func (c PerfClientAdapter) ActivateDataSource(projectName string, dataSourceId int) error {
	rlog := log.WithValues("project name", projectName, "datasource id", dataSourceId)
	rlog.Info("try to activate data source")

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParams(map[string]string{
			"id": strconv.Itoa(dataSourceId),
		}).
		Put("/api/v2/datasources/{id}/activation")
	if err != nil {
		return errors.Wrapf(err, "couldn't activate %v datasource under %v project", dataSourceId, projectName)
	}
	if resp.IsError() {
		return errors.Errorf("couldn't activate %v datasource under %v project. Status - %v",
			dataSourceId, projectName, resp.StatusCode())
	}
	rlog.Info("data source has been activated")
	return nil
}

func (c PerfClientAdapter) UpdateDataSource(ds dto.DataSource) error {
	log.Info("start updating PERF datasource", "name", ds.Name)
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParams(map[string]string{
			"id": strconv.Itoa(ds.Id),
		}).
		SetBody(ds).
		Put("/api/v2/datasources/{id}")
	if err != nil {
		return errors.Wrapf(err, "couldn't update %v datasource", ds.Name)
	}
	if resp.IsError() {
		return errors.Errorf("couldn't update %v datasource. Status - %v", ds.Name, resp.StatusCode())
	}
	log.Info("PERF datasource has been update.", "name", ds.Name)
	return nil
}