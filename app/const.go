/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package app

const (
	ClientData          = "data"
	ClientMetadata      = "metadata"
	ClientLogging       = "logging"
	ClientCommand       = "command"
	ClientExport        = "export"
	ClientNotifications = "notifications"
	ClientScheduler     = "scheduler"

	APIv1Prefix    = "/api/v1"
	Colon          = ":"
	HttpScheme     = "http://"
	HttpProto      = "HTTP"
	StatusResponse = "pong"
)

const (
	ID               = "id"
	NAME             = "name"
	DEVICEIDURLPARAM = "{deviceId}"
	OPSTATE          = "opstate"
	URLADMINSTATE    = "adminstate"
	ADMINSTATE       = "adminState"
	YAML             = "yaml"
	COMMANDLIST      = "commandlist"
	COMMAND          = "command"
	COMMANDID        = "commandid"
	COMMANDNAME      = "commandname"
	DEVICE           = "device"
	DEVICENAME       = "devicename"
	KEY              = "key"
	VALUE            = "value"
	PINGRESPONSE     = "pong"
	CONTENTTYPE      = "Content-Type"
	APPLICATIONJSON  = "application/json"
	TEXTPLAIN        = "text/plain"
	UNLOCKED         = "UNLOCKED"
	ENABLED          = "ENABLED"

	PING             = "ping"
	METRIC           = "metric"
	CONFIG           = "config"
	OBJECT           = "object"
	OBJECTNAME       = "objectname"
	ELEMENT          = "element"
	ELEMENTNAME      = "elementname"
	SCHEDULE         = "schedule"
	SCHEDULENAME     = "schedulename"
	LABEL            = "label"
	TYPE             = "type"
	TYPENAME         = "typename"
	OPERATION        = "operation"
	SERVICENAME      = "servicename"
	UPDATENAME       = "updatename"
	UPDATEPROFILE    = "updateprofile"
	UPDATEADMINSTATE = "updateadminstate"
	UPDATEPROTOCOL   = "updateprotocol"
	PROFILE          = "profile"
	PROFILENAME      = "profilename"
	UPLOAD           = "upload"
	READING          = "reading"
	READINGNAME      = "readingname"
	START            = "start"
	END              = "end"
	LIMIT            = "limit"
	VALUEDESCRIPTOR  = "valuedescriptor"
)

const (
	AppApiMonitor                            = "/monitor"
	AppApiMonitorByReadingName               = AppApiMonitor + "/" + READING + "/" + READINGNAME + "/{" + READINGNAME + "}/{" + START + "}/{" + END + "}/{" + LIMIT + "}"
	AppApiMonitorByDeviceName                = AppApiMonitor + "/" + READING + "/" + DEVICENAME + "/{" + DEVICENAME + "}/{" + START + "}/{" + END + "}/{" + LIMIT + "}"
	AppApiMonitorValueDescriptorByName       = AppApiMonitor + "/" + VALUEDESCRIPTOR + "/" + NAME + "/{" + NAME + "}"
	AppApiMonitorValueDescriptorByDeviceName = AppApiMonitor + "/" + VALUEDESCRIPTOR + "/" + DEVICENAME + "/{" + DEVICENAME + "}"

	AppApiObject             = "/object"
	AppApiObjectObjectName   = AppApiObject + "/" + NAME + "/{" + OBJECTNAME + "}"
	AppApiObjectElement      = AppApiObjectObjectName + "/" + ELEMENT + "/{" + ELEMENTNAME + "}"
	AppApiObjectSchedule     = AppApiObjectObjectName + "/" + SCHEDULE + "/{" + SCHEDULENAME + "}"
	AppApiObjectCommandList  = AppApiObjectObjectName + "/" + COMMANDLIST
	AppApiObjectIssueCommand = AppApiObjectObjectName + "/" + COMMAND + "/{" + COMMANDNAME + "}"

	AppApiProfile       = "/profile"
	AppApiProfileUpload = AppApiProfile + "/" + UPLOAD
	AppApiProfileName   = AppApiProfile + "/" + NAME + "/{" + PROFILENAME + "}"
)
