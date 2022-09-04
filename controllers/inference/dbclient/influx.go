package dbclient

import (
	client "github.com/influxdata/influxdb1-client/v2"
	logger "github.com/sirupsen/logrus"
	"melody/controllers/const"
)

// influxdb demo

func connInflux() client.Client {
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     consts.DefaultMelodyInfluxDBAddress,
		Username: consts.DefaultMelodyInfluxDBUser,
		Password: "",
	})
	if err != nil {
		logger.Fatal(err)
	}
	return cli
}

// query
func queryDB(cli client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: consts.DefaultMelodyInferenceMetricDatabase,
	}
	if response, err := cli.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
