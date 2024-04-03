package model

//
//type F5Config struct {
//	Class       string `json:"class"`
//	Action      string `json:"action"`
//	Persist     bool   `json:"persist"`
//	Declaration struct {
//		Class         string `json:"class"`
//		SchemaVersion string `json:"schemaVersion"`
//		Label         string `json:"label"`
//		Remark        string `json:"remark"`
//		Cfsr01Tenant  struct {
//			Class             string `json:"class"`
//			Cfsr01Application struct {
//				Class         string `json:"class"`
//				ServiceCfsr01 struct {
//					Class            string   `json:"class"`
//					Label            string   `json:"label"`
//					VirtualAddresses []string `json:"virtualAddresses"`
//					Pool             string   `json:"pool"`
//					ClientTLS        string   `json:"clientTLS"`
//					ServerTLS        string   `json:"serverTLS"`
//				} `json:"service_cfsr01"`
//				Cfsr01ClientTLS struct {
//					Class          string `json:"class"`
//					Label          string `json:"label"`
//					Ssl3Enabled    bool   `json:"ssl3Enabled"`
//					SslEnabled     bool   `json:"sslEnabled"`
//					TLS13Enabled   bool   `json:"tls1_3Enabled"`
//					SessionTickets bool   `json:"sessionTickets"`
//				} `json:"cfsr01_client_tls"`
//				Cfsr01ServerTLS struct {
//					Class        string `json:"class"`
//					Label        string `json:"label"`
//					Ssl3Enabled  bool   `json:"ssl3Enabled"`
//					SslEnabled   bool   `json:"sslEnabled"`
//					TLS10Enabled bool   `json:"tls1_0Enabled"`
//					TLS13Enabled bool   `json:"tls1_3Enabled"`
//					Certificates []struct {
//						Certificate string `json:"certificate"`
//					} `json:"certificates"`
//				} `json:"cfsr01_server_tls"`
//				Cfsr01ServerTLSCert struct {
//					Class       string `json:"class"`
//					Label       string `json:"label"`
//					Certificate string `json:"certificate"`
//					PrivateKey  string `json:"privateKey"`
//				} `json:"cfsr01_server_tls_cert"`
//				Cfsr01DgURLMatch struct {
//					Class       string `json:"class"`
//					Label       string `json:"label"`
//					KeyDataType string `json:"keyDataType"`
//					Records     []struct {
//						Key   string `json:"key"`
//						Value string `json:"value"`
//					} `json:"records"`
//				} `json:"cfsr01_dg_urlMatch"`
//				Cfsr01IruleURLMatch struct {
//					Class string `json:"class"`
//					Label string `json:"label"`
//					IRule string `json:"iRule"`
//				} `json:"cfsr01_irule_urlMatch"`
//				Cfsr01PoolDefault struct {
//					Class    string   `json:"class"`
//					Label    string   `json:"label"`
//					Monitors []string `json:"monitors"`
//					Members  []struct {
//						ServicePort     int      `json:"servicePort"`
//						ServerAddresses []string `json:"serverAddresses"`
//					} `json:"members"`
//				} `json:"cfsr01_pool_default"`
//				//Cfsr01Mon10 struct {
//				//	Class       string `json:"class"`
//				//	Label       string `json:"label"`
//				//	MonitorType string `json:"monitorType"`
//				//	Interval    int    `json:"interval"`
//				//	Timeout     int    `json:"timeout"`
//				//	Send        string `json:"send"`
//				//	Receive     string `json:"receive"`
//				//} `json:"cfsr01_mon_10"`
//				//Cfsr01Pool10 struct {
//				//	Class    string `json:"class"`
//				//	Label    string `json:"label"`
//				//	Monitors []struct {
//				//		Use string `json:"use"`
//				//	} `json:"monitors"`
//				//	Members []struct {
//				//		ServicePort     int      `json:"servicePort"`
//				//		ServerAddresses []string `json:"serverAddresses"`
//				//	} `json:"members"`
//				//} `json:"cfsr01_pool_10"`
//				//Cfsr01Mon20 struct {
//				//	Class       string `json:"class"`
//				//	Label       string `json:"label"`
//				//	MonitorType string `json:"monitorType"`
//				//	Interval    int    `json:"interval"`
//				//	Timeout     int    `json:"timeout"`
//				//	Send        string `json:"send"`
//				//	Receive     string `json:"receive"`
//				//} `json:"cfsr01_mon_20"`
//				//Cfsr01Pool20 struct {
//				//	Class    string `json:"class"`
//				//	Label    string `json:"label"`
//				//	Monitors []struct {
//				//		Use string `json:"use"`
//				//	} `json:"monitors"`
//				//	Members []struct {
//				//		ServicePort     int      `json:"servicePort"`
//				//		ServerAddresses []string `json:"serverAddresses"`
//				//	} `json:"members"`
//				//} `json:"cfsr01_pool_20"`
//			} `json:"cfsr01_application"`
//		} `json:"cfsr01_tenant"`
//	} `json:"declaration"`
//}

type Monitor struct {
	Class       string `json:"class"`
	Label       string `json:"label"`
	MonitorType string `json:"monitorType"`
	Interval    int    `json:"interval"`
	Timeout     int    `json:"timeout"`
	Send        string `json:"send"`
	Receive     string `json:"receive"`
}

type Pool struct {
	Class    string        `json:"class"`
	Label    string        `json:"label"`
	Monitors []PoolMonitor `json:"monitors"`
	Members  []PoolMember  `json:"members"`
}

type PoolMonitor struct {
	Use string `json:"use"`
}

type PoolMember struct {
	ServicePort     int      `json:"servicePort"`
	ServerAddresses []string `json:"serverAddresses"`
}

type DataGroup struct {
	Class       string            `json:"class"`
	Label       string            `json:"label"`
	KeyDataType string            `json:"keyDataType"`
	Records     []DataGroupRecord `json:"records"`
}

type DataGroupRecord struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
