package main

import (
	"flag"
	"fmt"
	"github.com/akira/go-puppetdb"
	"github.com/newrelic/go_nagios"
	"time"
)

var (
	daysw, daysc int
	pdbhost      string
)

func checkNode(node puppetdb.NodeJson) *nagios.NagiosStatus {
	t, _ := time.Parse(time.RFC3339Nano, node.CatalogTimestamp)
	twarn := time.Now().AddDate(0, 0, -daysw)
	tcrit := time.Now().AddDate(0, 0, -daysc)

	if t.Before(tcrit) {
		return &nagios.NagiosStatus{fmt.Sprintf("Node: %s checked in more than %d days ago: %s.\n", node.Name, daysc, t), nagios.NAGIOS_CRITICAL}
	}
	if t.Before(twarn) && t.After(tcrit) {
		return &nagios.NagiosStatus{fmt.Sprintf("Node: %s checked in more than %d days ago: %s.\n", node.Name, daysw, t), nagios.NAGIOS_WARNING}
	}

	return &nagios.NagiosStatus{"", nagios.NAGIOS_OK}
}

func main() {

	flag.IntVar(&daysw, "dw", 2, "Days node hasn't checked in to warn about")
	flag.IntVar(&daysc, "dc", 4, "Days node hasn't checked in to crit about")
	flag.StringVar(&pdbhost, "pdbhost", "localhost", "Hostname or IP of puppetdb host")

	flag.Parse()

	statuses := make([]*nagios.NagiosStatus, 0)

	client := puppetdb.NewClient(fmt.Sprintf("http://%s:8080", pdbhost), true)
	nodes, err := client.Nodes()
	if err != nil {
		errStatus := &nagios.NagiosStatus{fmt.Sprintf("Couldn't check nodes: %s", err), nagios.NAGIOS_UNKNOWN}
		nagios.ExitWithStatus(errStatus)
	}

	for _, node := range nodes {
		nodeStatus := checkNode(node)
		if nodeStatus.Value != nagios.NAGIOS_OK {
			statuses = append(statuses, checkNode(node))
		}
	}

	baseStatus := &nagios.NagiosStatus{fmt.Sprintf("Total Nodes: %v, broken nodes: %v\n", len(nodes), len(statuses)), nagios.NAGIOS_OK}
	baseStatus.Aggregate(statuses)
	nagios.ExitWithStatus(baseStatus)

}
