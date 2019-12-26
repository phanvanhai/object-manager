package app

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

const ManagerProfileNameConst = "ManagerProfile"

// lable init:
const (
	INITIALIZIED   = "initializied"
	UNINITIALIZIED = "uninitializied"
)

// type
const (
	DEVICETYPE   = "DeviceType"
	GROUPTYPE    = "GroupType"
	SCENARIOTYPE = "ScenarioType"
)

// Level
const (
	ROOTOBJECT = "RootObject"
	SUBOBJECT  = "SubObject"
)

// Protocols.network, .Schedule
const (
	PROTOCOLSNETWORKNAME  = "Network"
	PROTOCOLSSCHEDULENAME = "Schedule"
	PROTOCOLSCOMMANDNAME  = "command"
	PROTOCOLSBODYNAME     = "body"
)

type labelsType []string

func (l labelsType) getType() string {
	for _, s := range l {
		if (s == DEVICETYPE) || (s == GROUPTYPE) || (s == SCENARIOTYPE) {
			return s
		}
	}
	return ""
}

func (l labelsType) getRootName() string {
	if l == nil {
		return ""
	}
	for _, s := range l {
		if (s != DEVICETYPE) && (s != GROUPTYPE) && (s != SCENARIOTYPE) &&
			(s != ROOTOBJECT) && (s != SUBOBJECT) &&
			(s != INITIALIZIED) && (s != UNINITIALIZIED) {
			return s
		}
	}
	return ""
}

func (l labelsType) isRootObject() bool {
	if l == nil {
		return false
	}
	for _, s := range l {
		if s == ROOTOBJECT {
			return true
		}
	}
	return false
}

func (l labelsType) isSubObject() bool {
	if l == nil {
		return false
	}
	for _, s := range l {
		if s == SUBOBJECT {
			return true
		}
	}
	return false
}

func (l labelsType) isInitializied() bool {
	if l == nil {
		return false
	}
	for _, s := range l {
		if s == INITIALIZIED {
			return true
		}
	}
	return false
}

// func (l labelsType) getLevel() string {
// 	if l == nil {
// 		return ""
// 	}
// 	for _, s := range l {
// 		if (s == ROOTOBJECT) || (s == SUBOBJECT) {
// 			return s
// 		}
// 	}
// 	return ""
// }

type mapNameIdType map[string]string
type mapIdNameType map[string]string
type infoObject struct {
	ObjectType string
	SubDsIds   map[string]string //map[DsName]SubId
	Parents    map[string]string //map[RootId]TypeObject
}
type mapIdType map[string]infoObject
type mapDSMasterType map[string]string

// bien global
var (
	mapNameId   mapNameIdType
	mapIdName   mapIdNameType
	mapId       mapIdType
	mapDsMaster mapDSMasterType
)

func (m infoObject) addSubId(ds string, subId string) {
	if m.SubDsIds == nil {
		m.SubDsIds = make(map[string]string)
	}
	m.SubDsIds[ds] = subId
}

func cacheAddSubId(rootId string, ds string, subId string) {
	ob := mapId[rootId]
	ob.addSubId(ds, subId)
	mapId[rootId] = ob
}

func (m infoObject) deleteSubObjectByDs(ds string) {
	if _, ok := m.SubDsIds[ds]; ok {
		delete(m.SubDsIds, ds)
	}
}

func cacheDeleteSubObjectByDs(rootId string, ds string) {
	ob := mapId[rootId]
	ob.deleteSubObjectByDs(ds)
	mapId[rootId] = ob
}

func (m infoObject) deleteSubObjectsBySubId(subId string) {
	for ds, id := range m.SubDsIds {
		if id == subId {
			delete(m.SubDsIds, ds)
			break
		}
	}
}

func cacheDeleteSubObjectsBySubId(rootId string, subId string) {
	ob := mapId[rootId]
	ob.deleteSubObjectsBySubId(subId)
	mapId[rootId] = ob
}

func (m infoObject) addParents(rootId string, objectType string) {
	if m.Parents == nil {
		m.Parents = make(map[string]string)
	}
	m.Parents[rootId] = objectType
}

func cacheAddParents(rootId string, parentId string, objectType string) {
	ob := mapId[rootId]
	ob.addParents(parentId, objectType)
	mapId[rootId] = ob
}

func cacheGetParents(rootId string) map[string]string {
	ob := mapId[rootId]
	return ob.Parents
}

func (m infoObject) deleteParents(rootId string) {
	if _, ok := m.Parents[rootId]; ok {
		delete(m.Parents, rootId)
	}
}

func cacheDeleteParents(rootId string, parentId string) {
	ob := mapId[rootId]
	ob.deleteParents(parentId)
	mapId[rootId] = ob
}

func (m infoObject) setType(objectType string) {
	m.ObjectType = objectType
}

func cacheSetType(rootId string, objectType string) {
	ob := mapId[rootId]
	ob.setType(objectType)
	mapId[rootId] = ob
}

func (m infoObject) getType() string {
	return m.ObjectType
}

func cacheGetType(rootId string) string {
	ob := mapId[rootId]
	return ob.ObjectType
}

func cacheGetMapSub(rootId string) map[string]string {
	r := mapId[rootId]

	return r.SubDsIds
}

func cacheGetSubIdByDS(rootId string, ds string) (string, bool) {
	r := mapId[rootId]
	sub, ok := r.SubDsIds[ds]
	if !ok {
		return "", false
	}
	return sub, true
}

func cacheGetDSsOfRootObject(rootId string) []string {
	var ds []string
	ob := mapId[rootId]
	for dsname := range ob.SubDsIds {
		ds = append(ds, dsname)
	}
	return ds
}

func cacheGetSubIdsOfRootObject(rootId string) []string {
	var subIds []string
	ob, _ := mapId[rootId]
	for _, subId := range ob.SubDsIds {
		subIds = append(subIds, subId)
	}
	return subIds
}

func cacheGetDSByIDMaster(ds string) string {
	id := mapDsMaster[ds]
	return id
}

func cacheGetMapMaster() map[string]string {
	return mapDsMaster
}

func cacheUpdateMapMaster(dsCheck string) {
	_, ok := mapDsMaster[dsCheck]
	if ok {
		return
	}
	// create new masterDevice
	new := models.Device{}
	new.Id = uuid.New().String()
	new.Name = "MasterOf" + dsCheck
	new.Labels = make([]string, 1)
	new.Labels[0] = INITIALIZIED
	new.AdminState = UNLOCKED
	new.OperatingState = ENABLED
	new.Service = models.DeviceService{
		Name: dsCheck,
	}
	new.Profile = models.DeviceProfile{
		Name: ManagerProfileNameConst,
	}
	pp := make(models.ProtocolProperties)
	pp["MasterObject"] = ManagerProfileNameConst
	p := make(map[string]models.ProtocolProperties)
	p["MasterObject"] = pp
	new.Protocols = p

	id, err := clientMetaDevice.Add(&new, context.Background())
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	mapDsMaster[dsCheck] = id
}

func CacheInit() bool {
	sliceDevice, err := clientMetaDevice.DevicesByLabel(ROOTOBJECT, context.Background())
	mapId = make(mapIdType, len(sliceDevice))
	mapNameId = make(mapNameIdType, len(sliceDevice))
	mapIdName = make(mapIdNameType, len(sliceDevice))

	sliceMasterDevice, err := clientMetaDevice.DevicesForProfileByName(ManagerProfileNameConst, context.Background())
	mapDsMaster = make(mapDSMasterType, len(sliceMasterDevice))
	for _, master := range sliceMasterDevice {
		idMaster := master.Id
		dsMaster := master.Service.Name
		mapDsMaster[dsMaster] = idMaster
	}

	for _, d := range sliceDevice {
		cacheAddUpdateRoot(d)
		if labelsType(d.Labels).getType() == DEVICETYPE {
			cacheUpdateMapMaster(d.Service.Name)
		}
	}

	// update field Parent of elements
	for _, d := range sliceDevice {
		objectType := labelsType(d.Labels).getType()

		for element := range d.Protocols {
			if (element != PROTOCOLSNETWORKNAME) && (element != PROTOCOLSSCHEDULENAME) {
				elementId, ok := mapNameId[element]
				if ok {
					cacheAddParents(elementId, d.Id, objectType)
				} else {
					delete(d.Protocols, element)
					clientMetaDevice.Update(d, context.Background())
				}
			}
		}
	}
	sliceDevice, err = clientMetaDevice.DevicesByLabel(SUBOBJECT, context.Background())
	for _, d := range sliceDevice {
		cacheAddUpdateSub(d)
	}

	if err != nil {
		return false
	}

	return true
}

func cacheAddUpdateRoot(object models.Device) {
	if labelsType(object.Labels).isRootObject() {
		id := object.Id
		a := infoObject{
			ObjectType: labelsType(object.Labels).getType(),
			SubDsIds:   make(map[string]string),
			Parents:    make(map[string]string),
		}
		mapId[id] = a
		mapNameId[object.Name] = id
		mapIdName[id] = object.Name
		// get SubObject
		sliceSubDevice, err := clientMetaDevice.DevicesByLabel(id, context.Background())
		if err == nil && len(sliceSubDevice) != 0 {
			for _, sd := range sliceSubDevice {
				mapId[id].SubDsIds[sd.Service.Name] = sd.Id
			}
		}
	}
}

func cacheDeleteRoot(id string) {
	if _, ok := mapId[id]; ok {
		name := mapIdName[id]
		delete(mapIdName, id)
		delete(mapNameId, name)
		delete(mapId, id)
	}
}

func cacheAddUpdateSub(object models.Device) {
	mapNameId[object.Name] = object.Id
	mapIdName[object.Id] = object.Name
}

// func cacheDeleteSubObject(subId string, subName string) {
// 	cacheDeleteSubObjectsBySubId(subId)
// 	delete(mapNameId, subName)
// }

func convertNameId(name string) string {
	r, _ := mapNameId[name]
	return r
}
func convertIdName(id string) string {
	r, _ := mapIdName[id]
	return r
}

func cacheAddMapNameId(name string, id string) {
	mapNameId[name] = id
	mapIdName[id] = name
}

func cacheUpdateName(rootId string, name string) {
	old := mapIdName[rootId]
	mapIdName[rootId] = name
	delete(mapNameId, old)
	mapNameId[name] = rootId
}

func cacheDeleteMapHasName(name string) {
	id := mapNameId[name]
	delete(mapNameId, name)
	delete(mapIdName, id)
}

func cacheGetParentTypeByParentId(rootId string, parentId string) string {
	ob := mapId[rootId]
	t := ob.Parents[parentId]
	return t
}

func checkExit(rootId string) bool {
	_, ok := mapId[rootId]
	return ok
}

func checkExitByName(rootName string) bool {
	_, ok := mapNameId[rootName]
	return ok
}
