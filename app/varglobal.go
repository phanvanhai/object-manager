package app

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

var (
	services = map[string]string{
		clients.SupportNotificationsServiceKey: "Notifications",
		clients.CoreCommandServiceKey:          "Command",
		clients.CoreDataServiceKey:             "Data",
		clients.CoreMetaDataServiceKey:         "Metadata",
		clients.ExportClientServiceKey:         "Export",
		clients.ExportDistroServiceKey:         "Distro",
		clients.SupportLoggingServiceKey:       "Logging",
		clients.SupportSchedulerServiceKey:     "Scheduler",
	}

	clientCoreReading         coredata.ReadingClient
	clientCoreValueDescriptor coredata.ValueDescriptorClient
	clientMetaDevice          metadata.DeviceClient
	clientMetaAddressable     metadata.AddressableClient
	clientMetaCommand         metadata.CommandClient
	clientMetaProfile         metadata.DeviceProfileClient
	clientMetaDS              metadata.DeviceServiceClient
	clientCommand             command.CommandClient

	registryClient  registry.Client
	registryErrors  chan error       //A channel for "config wait errors" sourced from Registry
	registryUpdates chan interface{} //A channel for "config updates" sourced from Registry

	Configuration *ConfigurationStruct
	LoggingClient logger.LoggingClient
)
