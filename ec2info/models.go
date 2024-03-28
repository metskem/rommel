package main

import "time"

type EC2Info struct {
	Reservations []struct {
		Groups    []interface{} `json:"Groups"`
		Instances []struct {
			AmiLaunchIndex int       `json:"AmiLaunchIndex"`
			ImageID        string    `json:"ImageId"`
			InstanceID     string    `json:"InstanceId"`
			InstanceType   string    `json:"InstanceType"`
			KeyName        string    `json:"KeyName"`
			LaunchTime     time.Time `json:"LaunchTime"`
			Monitoring     struct {
				State string `json:"State"`
			} `json:"Monitoring"`
			Placement struct {
				AvailabilityZone string `json:"AvailabilityZone"`
				GroupName        string `json:"GroupName"`
				Tenancy          string `json:"Tenancy"`
			} `json:"Placement"`
			PrivateDNSName   string        `json:"PrivateDnsName"`
			PrivateIPAddress string        `json:"PrivateIpAddress"`
			ProductCodes     []interface{} `json:"ProductCodes"`
			PublicDNSName    string        `json:"PublicDnsName"`
			State            struct {
				Code int    `json:"Code"`
				Name string `json:"Name"`
			} `json:"State"`
			StateTransitionReason string `json:"StateTransitionReason"`
			SubnetID              string `json:"SubnetId"`
			VpcID                 string `json:"VpcId"`
			Architecture          string `json:"Architecture"`
			BlockDeviceMappings   []struct {
				DeviceName string `json:"DeviceName"`
				Ebs        struct {
					AttachTime          time.Time `json:"AttachTime"`
					DeleteOnTermination bool      `json:"DeleteOnTermination"`
					Status              string    `json:"Status"`
					VolumeID            string    `json:"VolumeId"`
				} `json:"Ebs"`
			} `json:"BlockDeviceMappings"`
			ClientToken       string `json:"ClientToken"`
			EbsOptimized      bool   `json:"EbsOptimized"`
			EnaSupport        bool   `json:"EnaSupport"`
			Hypervisor        string `json:"Hypervisor"`
			NetworkInterfaces []struct {
				Attachment struct {
					AttachTime          time.Time `json:"AttachTime"`
					AttachmentID        string    `json:"AttachmentId"`
					DeleteOnTermination bool      `json:"DeleteOnTermination"`
					DeviceIndex         int       `json:"DeviceIndex"`
					Status              string    `json:"Status"`
					NetworkCardIndex    int       `json:"NetworkCardIndex"`
				} `json:"Attachment"`
				Description string `json:"Description"`
				Groups      []struct {
					GroupName string `json:"GroupName"`
					GroupID   string `json:"GroupId"`
				} `json:"Groups"`
				Ipv6Addresses      []interface{} `json:"Ipv6Addresses"`
				MacAddress         string        `json:"MacAddress"`
				NetworkInterfaceID string        `json:"NetworkInterfaceId"`
				OwnerID            string        `json:"OwnerId"`
				PrivateDNSName     string        `json:"PrivateDnsName"`
				PrivateIPAddress   string        `json:"PrivateIpAddress"`
				PrivateIPAddresses []struct {
					Primary          bool   `json:"Primary"`
					PrivateDNSName   string `json:"PrivateDnsName"`
					PrivateIPAddress string `json:"PrivateIpAddress"`
				} `json:"PrivateIpAddresses"`
				SourceDestCheck bool   `json:"SourceDestCheck"`
				Status          string `json:"Status"`
				SubnetID        string `json:"SubnetId"`
				VpcID           string `json:"VpcId"`
				InterfaceType   string `json:"InterfaceType"`
			} `json:"NetworkInterfaces"`
			RootDeviceName string `json:"RootDeviceName"`
			RootDeviceType string `json:"RootDeviceType"`
			SecurityGroups []struct {
				GroupName string `json:"GroupName"`
				GroupID   string `json:"GroupId"`
			} `json:"SecurityGroups"`
			SourceDestCheck bool `json:"SourceDestCheck"`
			Tags            []struct {
				Key   string `json:"Key"`
				Value string `json:"Value"`
			} `json:"Tags"`
			VirtualizationType string `json:"VirtualizationType"`
			CPUOptions         struct {
				CoreCount      int `json:"CoreCount"`
				ThreadsPerCore int `json:"ThreadsPerCore"`
			} `json:"CpuOptions"`
			CapacityReservationSpecification struct {
				CapacityReservationPreference string `json:"CapacityReservationPreference"`
			} `json:"CapacityReservationSpecification"`
			HibernationOptions struct {
				Configured bool `json:"Configured"`
			} `json:"HibernationOptions"`
			MetadataOptions struct {
				State                   string `json:"State"`
				HTTPTokens              string `json:"HttpTokens"`
				HTTPPutResponseHopLimit int    `json:"HttpPutResponseHopLimit"`
				HTTPEndpoint            string `json:"HttpEndpoint"`
			} `json:"MetadataOptions"`
			EnclaveOptions struct {
				Enabled bool `json:"Enabled"`
			} `json:"EnclaveOptions"`
		} `json:"Instances"`
		OwnerID       string `json:"OwnerId"`
		ReservationID string `json:"ReservationId"`
		RequesterID   string `json:"RequesterId,omitempty"`
	} `json:"Reservations"`
}
