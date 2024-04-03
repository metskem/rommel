package main

import (
	"encoding/json"
	"fmt"
	"github.com/metskem/rommel/cfsr-as3-generator/conf"
	"github.com/metskem/rommel/cfsr-as3-generator/model"
	"io"
	"math/rand"
	"os"
)

func main() {
	var err error
	var f5Config map[string]interface{}
	jsonFile, err := os.Open("resources/cfsr01.json")
	if err != nil {
		fmt.Println(err)
	}
	defer func() { _ = jsonFile.Close() }()
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &f5Config)
	if err != nil {
		fmt.Printf("failed to unmarshal json, error: %v", err)
		return
	}

	var declaration map[string]interface{}
	var declarationBytes []byte
	if declarationBytes, err = json.Marshal(f5Config["declaration"]); err != nil {
		fmt.Printf("failed to marshal declaration, error: %v", err)
		return
	}
	if err = json.Unmarshal(declarationBytes, &declaration); err != nil {
		fmt.Printf("failed to unmarshal declaration, error: %v", err)
		return
	}

	var tenant map[string]interface{}
	var tenantBytes []byte
	tenantBytes, err = json.Marshal(declaration["cfsr01_tenant"])
	if err != nil {
		fmt.Printf("failed to marshal declaration, error: %v", err)
		return
	}
	if err = json.Unmarshal(tenantBytes, &tenant); err != nil {
		fmt.Printf("failed to unmarshal tenant, error: %v", err)
		return
	}
	//fmt.Printf("tenant: %s\n", tenant)

	var application map[string]interface{}
	var applicationBytes []byte
	applicationBytes, err = json.Marshal(tenant["cfsr01_application"])
	if err != nil {
		fmt.Printf("failed to marshal application, error: %v", err)
		return
	}
	if err = json.Unmarshal(applicationBytes, &application); err != nil {
		fmt.Printf("failed to unmarshal application, error: %v", err)
		return
	}

	poolMembers := make([]model.PoolMember, 1)
	poolMembers[0] = model.PoolMember{ServicePort: 443, ServerAddresses: []string{"10.253.5.4", "10.253.21.4"}}
	dataGroupRecords := make([]model.DataGroupRecord, conf.NumApps)
	dataGroup := model.DataGroup{Class: "Data_Group", Label: "cfsr01_dg_urlMatch", KeyDataType: "string", Records: dataGroupRecords}

	for ix := 0; ix < conf.NumApps; ix++ {
		poolMonitors := make([]model.PoolMonitor, 1)
		poolMonitors[0].Use = fmt.Sprintf("monitor_%04d", ix)
		application[fmt.Sprintf("pool_%04d", ix)] = model.Pool{Class: "Pool", Label: fmt.Sprintf("Pool_%04d", ix), Monitors: poolMonitors, Members: poolMembers}
		randomOffset := rand.Intn(5)
		application[fmt.Sprintf("monitor_%04d", ix)] = model.Monitor{Class: "Monitor", Label: fmt.Sprintf("Mon_%04d", ix), MonitorType: "https", Interval: 60 + randomOffset, Timeout: 61 + randomOffset, Send: fmt.Sprintf("GET /health HTTP/1.1\r\nHost: app%d.apps.cfsrdev.sample-domain.com\r\nConnection: Close\r\n\r\n", ix), Receive: "200 OK"}
		dataGroupRecords[ix] = model.DataGroupRecord{Key: fmt.Sprintf("app%d.apps.cfsrdev.sample-domain.com", ix), Value: fmt.Sprintf("pool_%04d", ix)}
	}

	// hook it back up in the structure:
	dataGroup.Records = dataGroupRecords
	application["dg_urlMatch"] = dataGroup
	tenant["cfsr01_application"] = application
	declaration["cfsr01_tenant"] = tenant
	f5Config["declaration"] = declaration

	f5ConfigIndented, err := json.MarshalIndent(f5Config, "", "  ")
	fmt.Printf("%s\n", f5ConfigIndented)
}
