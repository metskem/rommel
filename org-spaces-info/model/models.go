package model

type SpaceConfig struct {
	Org                        string         `yaml:"org"`
	Space                      string         `yaml:"space"`
	SpaceDeveloper             SpaceDeveloper `yaml:"space-developer"`
	SpaceManager               SpaceManager   `yaml:"space-manager"`
	SpaceAuditor               SpaceAuditor   `yaml:"space-auditor"`
	AllowSSH                   bool           `yaml:"allow-ssh"`
	EnableRemoveUsers          bool           `yaml:"enable-remove-users"`
	EnableSpaceQuota           bool           `yaml:"enable-space-quota"`
	MemoryLimit                int            `yaml:"memory-limit"`
	InstanceMemoryLimit        int            `yaml:"instance-memory-limit"`
	TotalRoutes                int            `yaml:"total-routes"`
	TotalServices              int            `yaml:"total-services"`
	PaidServicePlansAllowed    bool           `yaml:"paid-service-plans-allowed"`
	EnableSecurityGroup        bool           `yaml:"enable-security-group"`
	TotalPrivateDomains        int            `yaml:"total_private_domains"`
	TotalReservedRoutePorts    int            `yaml:"total_reserved_route_ports"`
	TotalServiceKeys           int            `yaml:"total_service_keys"`
	AppInstanceLimit           int            `yaml:"app_instance_limit"`
	LogRateLimitBytesPerSecond int            `yaml:"log_rate_limit_bytes_per_second"`
	Metadata                   Metadata       `yaml:"metadata"`
}
type SpaceDeveloper struct {
	LdapUsers  []interface{} `yaml:"ldap_users"`
	Users      []string      `yaml:"users"`
	SpnUsers   []string      `yaml:"spn_users"`
	LdapGroup  string        `yaml:"ldap_group"`
	LdapGroups []string      `yaml:"ldap_groups"`
	SamlUsers  []interface{} `yaml:"saml_users"`
	AadGroups  []string      `yaml:"aad_groups"`
}
type SpaceManager struct {
	LdapUsers  []interface{} `yaml:"ldap_users"`
	Users      []interface{} `yaml:"users"`
	SpnUsers   []string      `yaml:"spn_users"`
	LdapGroup  string        `yaml:"ldap_group"`
	LdapGroups []interface{} `yaml:"ldap_groups"`
	SamlUsers  []interface{} `yaml:"saml_users"`
	AadGroups  []string      `yaml:"aad_groups"`
}
type SpaceAuditor struct {
	LdapUsers  []interface{} `yaml:"ldap_users"`
	Users      []string      `yaml:"users"`
	SpnUsers   []string      `yaml:"spn_users"`
	LdapGroup  string        `yaml:"ldap_group"`
	LdapGroups []interface{} `yaml:"ldap_groups"`
	SamlUsers  []interface{} `yaml:"saml_users"`
	AadGroups  []string      `yaml:"aad_groups"`
}
type Metadata struct {
	Manager              string   `yaml:"manager"`
	CostCenter           string   `yaml:"cost_center"`
	Contact              string   `yaml:"contact"`
	OwnedFunctionalUsers []string `yaml:"owned_functional_users"`
	Approvers            []string `yaml:"approvers"`
	Area                 string   `yaml:"area"`
	Tribe                string   `yaml:"tribe"`
}
