package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Duration wait UpdateNewObject From DS: sum = 4s
const (
	CountRetryConst   = 20
	TimeStepRetryCont = 200 // ms
)

// virtual manager DS
const (
	PreficManager       = "Manager"
	ManagerSubcribe     = "Subscribe"
	MangerSchedule      = "Schedule"
	ManagerRemoveItself = "RemoveItself"
	ManagerPutMethod    = "PUT"
	MangerDeleteMethod  = "DELETE"
)

const (
	AddressableName     = "manager-addressable"
	AddressableProtocol = "HTTP"
	AddressableAddress  = "manager-serivce"
	AddressablePort     = 49999
	AddressablePath     = "/api/v1/callback"
)

const (
	ScenarioProfileName = "ScenarioProfile"
	DSManagerName       = "manager-deviceservice"
	DSManagerAdminState = "unlocked"
	DSManagerOpState    = "enabled"
)

func createManagerAddressable() {
	a := models.Addressable{
		Name:     AddressableName,
		Protocol: AddressableProtocol,
		Address:  AddressableAddress,
		Port:     AddressablePort,
		Path:     AddressablePath,
	}
	_, err := clientMetaAddressable.AddressableForName(AddressableName, context.Background())
	if err != nil {
		clientMetaAddressable.Add(&a, context.Background())
	}
}

func createManagerDS() {
	ds := models.DeviceService{
		Name:           DSManagerName,
		AdminState:     DSManagerAdminState,
		OperatingState: DSManagerOpState,
		Addressable: models.Addressable{
			Name: AddressableName,
		},
	}
	_, err := clientMetaDS.DeviceServiceForName(DSManagerName, context.Background())
	if err != nil {
		clientMetaDS.Add(&ds, context.Background())
	}
}

func ManagerObjectInit() {
	createManagerAddressable()
	createManagerDS()
	if ok := CacheInit(); !ok {
		LoggingClient.Error("Khong the doc duoc danh sach cac rootObject")
	}
}

func getNewDSFromTwoList(src []string, des []string) []string {
	var newds []string
	for _, iDes := range des {
		var has = false
		for _, iSrc := range src {
			if iDes == iSrc {
				has = true
				break
			}
		}
		if has == false {
			newds = append(newds, iDes)
		}
	}
	return newds
}

func findCommandByLabel(ctx context.Context, label string, objectId string) (models.Command, bool) {
	commandList, err := GetCommandForDevice(ctx, objectId)
	if err != nil {
		return models.Command{}, false
	}
	for _, subCommand := range commandList {
		if strings.HasPrefix(subCommand.Name, label) {
			return subCommand, true
		}
	}
	return models.Command{}, false
}

func sendManagerPutCommandByLabel(ctx context.Context, ds string, objectId string, label string, managerCommandName string, managerMethod string, contentBody string) (rep string, err error) {
	idMaster := cacheGetDSByIDMaster(ds)
	cm, ok := findCommandByLabel(ctx, label, idMaster)
	if !ok {
		return "", fmt.Errorf("Khong tim thay lenh cho doi tuong:%s", idMaster)
	}

	param := cm.Put.ParameterNames
	if len(param) < 4 {
		LoggingClient.Error("Lenh khong phu hop")
		return "Lenh khong phu hop", fmt.Errorf("Lenh khong phu hop")
	}

	var mapBody = make(map[string]string, 4)
	mapBody[param[0]] = convertIdName(objectId)
	mapBody[param[1]] = managerCommandName
	mapBody[param[2]] = managerMethod
	mapBody[param[3]] = contentBody
	body, err := json.Marshal(&mapBody)
	if err != nil {
		return "false", err
	}

	rep, err = clientCommand.Put(idMaster, cm.Id, string(body), ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	return
}

type action struct {
	Command string `json:"command,omitempty"`
	Body    string `json:"body,omitempty"`
}
type contentElementType struct {
	OwnerId    string `json:"ownerID,omitempty"`
	ObjectType string `json:"type,omitempty"`
	ElementId  string `json:"elementID,omitempty"`
	action
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------
type contentScheduleType struct {
	OwnerId      string `json:"ownerID,omitempty"`
	ScheduleName string `json:"name,omitempty"`
	Time         int32  `json:"time,omitempty"`
	// Start        string `json:"start,omitempty"`
	// End          string `json:"end,omitempty"`
	// Frequency    string `json:"frequency,omitempty"`
	// RunOnce      bool   `json:"runOnce,omitempty"`
	action
}

func getMapScheduleOf(rootObject models.Device) map[string]string {
	result, ok := rootObject.Protocols[PROTOCOLSSCHEDULENAME]
	if !ok {
		return nil
	}
	return result
}

// String returns a JSON encoded string representation of this contentScheduleType
func (i contentScheduleType) String() string {
	out, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// String returns a JSON encoded string representation of this contentScheduleType
func (i contentElementType) String() string {
	out, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// ----------------------------> write sendCommand <--------------------------------
func sendScheduleSetCommand(ctx context.Context, receiverObjectID string, body contentScheduleType) (rep string, err error) {
	i := 0
	mapSubElementSubObject, sliceDs := createMapElementSubObject(receiverObjectID, body.OwnerId)
	for subElement, subObject := range mapSubElementSubObject {
		// set: schedule of subObject to subElement
		subBody := body
		subBody.OwnerId = subObject

		rep, err = sendManagerPutCommandByLabel(ctx, sliceDs[i], subElement, PreficManager, MangerSchedule, ManagerPutMethod, subBody.String())
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error: sendSetCommandTo(%s, %s)", subElement, subBody.String()))
		}
		i++
	}
	return
}

// ----------------------------> write sendCommand <--------------------------------
func sendScheduleDeleteCommand(ctx context.Context, receiverObjectID string, body contentScheduleType) (rep string, err error) {
	i := 0
	mapSubElementSubObject, sliceDs := createMapElementSubObject(receiverObjectID, body.OwnerId)
	for subElement, subObject := range mapSubElementSubObject {
		// delete: schedule of subObject in subElement
		subBody := body
		subBody.OwnerId = subObject

		rep, err = sendManagerPutCommandByLabel(ctx, sliceDs[i], subElement, PreficManager, MangerSchedule, MangerDeleteMethod, subBody.String())
		if err != nil {
			LoggingClient.Info("Error: sendSetCommandTo(%s, %s)", subElement, subBody.String())
		}
		i++
	}
	return
}

/*
	Put Schedule to Object
*/

func PutScheduleToObject(ctx context.Context, objectId string, body contentScheduleType) (rep string, err error) {
	rootObject, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return "false", err
	}

	body.OwnerId = objectId
	// 1. Update Protocols.Schedule cua Object trong MetaData
	_, ok := rootObject.Protocols[PROTOCOLSSCHEDULENAME]
	if !ok {
		rootObject.Protocols[PROTOCOLSSCHEDULENAME] = make(map[string]string)
	}
	rootObject.Protocols[PROTOCOLSSCHEDULENAME][body.ScheduleName] = body.String()

	err = clientMetaDevice.Update(rootObject, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}

	// 2. send request to ObjectItself
	rep, err = sendScheduleSetCommand(ctx, objectId, body)
	return
}

func rePutScheduleToObject(ctx context.Context, objectId string, fromObjectId string) (rep string, err error) {
	rootObject, err := clientMetaDevice.Device(fromObjectId, ctx)
	if err != nil {
		return "false", err
	}

	scheduleList := getMapScheduleOf(rootObject)
	if scheduleList != nil {
		for _, valueSchedule := range scheduleList {
			body := contentScheduleType{}
			err := json.Unmarshal([]byte(valueSchedule), &body)
			if err == nil {
				rep, err = sendScheduleSetCommand(ctx, objectId, body)
			}
		}
	}
	return
}

/*
	Delete Schedule in Object
*/

func DeleteScheduleInObject(ctx context.Context, objectId string, body contentScheduleType) (rep string, err error) {
	rootObject, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return "false", err
	}

	// 1. Update Protocols.Schedule cua Object trong MetaData
	_, ok := rootObject.Protocols[PROTOCOLSSCHEDULENAME][body.ScheduleName]
	if ok {
		delete(rootObject.Protocols[PROTOCOLSSCHEDULENAME], body.ScheduleName)
	}

	err = clientMetaDevice.Update(rootObject, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
	}

	// 2. send request to ObjectItself
	rep, err = sendScheduleDeleteCommand(ctx, objectId, body)
	return
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------
func putElementToProtocols(protocols map[string]models.ProtocolProperties, elementId string, content contentElementType) {
	content.ElementId = elementId
	ct, _ := json.Marshal(content)
	pp := make(models.ProtocolProperties)
	json.Unmarshal(ct, &pp)
	elementName := convertIdName(elementId)
	protocols[elementName] = pp
}

func deleteElementInProtocols(protocols map[string]models.ProtocolProperties, elementId string) {
	elementName := convertIdName(elementId)
	delete(protocols, elementName)
}

// DS phai trien khai callback: Delete, Add
/* REST /manager/object:
POST:
PUT:
DEL:
*** Note : khong cho phep thay doi DeviceService. Muon thay doi = xoa va them moi lai
/*
	Add new Object
*/

func CreateRootObject(ctx context.Context, object *models.Device) (rep string, err error) {
	objectType := labelsType(object.Labels).getType()
	if objectType == "" {
		return "Khong biet loai doi tuong", fmt.Errorf("Khong biet loai doi tuong")
	}

	if id := convertNameId(object.Name); object.Name == "" || id != "" {
		return "Ten doi tuong da ton tai", fmt.Errorf("Ten doi tuong da ton tai")
	}
	rootId := uuid.New().String()
	object.Id = rootId

	if objectType == DEVICETYPE {
		object.Labels = make([]string, 5)
		object.Labels[0] = objectType
		object.Labels[1] = ROOTOBJECT
		object.Labels[2] = SUBOBJECT
		object.Labels[3] = rootId
		object.Labels[4] = UNINITIALIZIED
	} else {
		object.Labels = make([]string, 3)
		object.Labels[0] = objectType
		object.Labels[1] = ROOTOBJECT
		object.Labels[2] = INITIALIZIED // truong hop nay tam cho = da dc khoi tao
	}

	p := make(map[string]models.ProtocolProperties)
	p[PROTOCOLSNETWORKNAME] = object.Protocols[PROTOCOLSNETWORKNAME]
	object.Protocols = p

	dsName := object.Service.Name

	// rootObject se thuoc ve deviceSerivce: manager-service
	object.Service = models.DeviceService{
		Name: DSManagerName,
	}
	if objectType == DEVICETYPE {
		object.Service.Name = dsName
	}

	rep, err = clientMetaDevice.Add(object, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return rep, err
	}

	isInit := false
	var newObject models.Device
	for count := 0; (isInit == false) && (count <= CountRetryConst); count++ {
		newObject, err = clientMetaDevice.Device(rootId, ctx)
		if labelsType(newObject.Labels).isInitializied() {
			isInit = true
			break
		}
		time.Sleep(TimeStepRetryCont * time.Millisecond)
	}
	if isInit == false {
		LoggingClient.Warn("Object:" + newObject.Name + "chua duoc khoi tao boi Device Service:" + newObject.Service.Name)
	}

	cacheAddUpdateRoot(newObject)

	if objectType == DEVICETYPE {
		// doi voi Device, root va sub hop la 1
		cacheAddSubId(rootId, dsName, rootId)
		cacheAddMapNameId(object.Name, rootId)
		cacheUpdateMapMaster(dsName)
	}
	if isInit == false {
		rep = rep + "\nBut Object is uninitialized by DS:" + newObject.Service.Name
	}
	return rep, err
}

func createSubObject(ctx context.Context, rootObject *models.Device, dsName string) (string, error) {
	var rep string
	var err error
	sub := *rootObject

	sub.Id = ""
	subName := uuid.New().String()
	sub.Name = subName
	objectType := labelsType(rootObject.Labels).getType()
	sub.Labels = make([]string, 4)
	sub.Labels[0] = objectType
	sub.Labels[1] = SUBOBJECT
	sub.Labels[2] = rootObject.Id
	sub.Labels[3] = UNINITIALIZIED

	sub.Description = ""

	p := make(map[string]models.ProtocolProperties)
	p[PROTOCOLSNETWORKNAME] = rootObject.Protocols[PROTOCOLSNETWORKNAME]
	sub.Protocols = p
	sub.Service = models.DeviceService{
		Name: dsName,
	}

	rep, err = clientMetaDevice.Add(&sub, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}
	subId := rep

	isInit := false
	var newObject models.Device
	for count := 0; (isInit == false) && (count <= CountRetryConst); count++ {
		newObject, err = clientMetaDevice.Device(subId, ctx)
		if labelsType(newObject.Labels).isInitializied() {
			isInit = true
			break
		}
		time.Sleep(TimeStepRetryCont * time.Millisecond)
	}
	if isInit == false {
		LoggingClient.Warn("SubObject cua Object:" + convertIdName(rootObject.Id) + "chua duoc khoi tao boi Device Service:" + newObject.Service.Name)
	}

	cacheAddSubId(rootObject.Id, dsName, subId)
	cacheAddMapNameId(subName, subId)

	if isInit == false {
		rep = rep + "\nBut SubObject for DS:" + newObject.Service.Name + " is uninitialized"
	}
	return rep, err
}

func UpdateRootObject(ctx context.Context, from *models.Device) (rep string, err error) {
	rootId := from.Id
	if rootId == "" {
		LoggingClient.Error("Khong co ID cua doi tuong")
		return "fasle", fmt.Errorf("Khong co ID cua doi tuong")
	}

	to, err := clientMetaDevice.Device(rootId, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "false", err
	}
	// chi cho phep thay doi Profile cua loai Device
	// vi khi update Profile se thay doi thong tin Device, nen se tien hanh update Profile truoc
	// sau do lay lai thong tin moi nhat cho Deivce
	rootType := cacheGetType(rootId)
	if rootType == DEVICETYPE {
		// kiem tra khac Profile nil
		if (from.Profile.String() != models.DeviceProfile{}.String()) {
			// kiem tra Profile co khac Profile cu khong
			if from.Profile.Name != to.Profile.Name {
				to.Profile.Name = from.Profile.Name
				// call: xu ly xoa Device trong Group
				for parentId, parentType := range cacheGetParents(rootId) {
					if parentType == GROUPTYPE {
						rep, err = DeleteElementInObject(ctx, rootId, parentId)
					}
				}
				// to, err = clientMetaDevice.Device(rootId, ctx)
				// if err != nil {
				// 	LoggingClient.Error(err.Error())
				// 	return "fasle", err
				// }
			}
		}
	}
	rep = "true"
	oldname := to.Name
	newname := from.Name
	if len(from.Name) > 0 {
		to.Name = from.Name
	}
	if len(from.Protocols) > 0 {
		if pp, ok := from.Protocols[PROTOCOLSNETWORKNAME]; ok {
			to.Protocols[PROTOCOLSNETWORKNAME] = pp
		}
	}
	if len(from.AutoEvents) > 0 {
		to.AutoEvents = from.AutoEvents
	}
	if from.AdminState != "" {
		to.AdminState = from.AdminState
	}
	if from.Description != "" {
		to.Description = from.Description
	}

	if from.LastConnected != 0 {
		to.LastConnected = from.LastConnected
	}
	if from.LastReported != 0 {
		to.LastReported = from.LastReported
	}
	if from.Location != nil {
		to.Location = from.Location
	}
	if from.OperatingState != models.OperatingState("") {
		to.OperatingState = from.OperatingState
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}

	err = clientMetaDevice.Update(to, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		// return err
	} else {
		if oldname != newname {
			for parentId := range cacheGetParents(rootId) {
				renameElementInParent(ctx, oldname, parentId, newname)
			}
			cacheUpdateName(rootId, newname)
			cacheDeleteMapHasName(oldname)
			cacheAddMapNameId(newname, rootId)
		}
	}
	// update cac SubObjet, vi loai DeviceType, Sub = Root nen se bo qua phan sau
	if rootType == DEVICETYPE {
		return
	}
	for _, subId := range cacheGetSubIdsOfRootObject(rootId) {
		sub, err := clientMetaDevice.Device(subId, ctx)

		sub.AutoEvents = to.AutoEvents
		sub.AdminState = to.AdminState
		sub.LastConnected = to.LastConnected
		sub.LastReported = to.LastReported
		sub.Location = to.Location
		sub.OperatingState = to.OperatingState
		sub.Origin = to.Origin
		sub.Profile.Name = to.Profile.Name

		err = clientMetaDevice.Update(sub, ctx)
		if err != nil {
			LoggingClient.Error(err.Error())
			// return err
		}
	}
	return
}

func createMapElementSubObject(elementId string, objectId string) (map[string]string, []string) {
	sliceDS := cacheGetDSsOfRootObject(elementId)
	mapEO := make(map[string]string, len(sliceDS))
	sliceDs := make([]string, 0, len(sliceDS))
	for _, ds := range sliceDS {
		subObject, _ := cacheGetSubIdByDS(objectId, ds)
		subElement, err := cacheGetSubIdByDS(elementId, ds)
		if err == true {
			mapEO[subElement] = subObject
			sliceDs = append(sliceDs, ds)
		}
	}
	return mapEO, sliceDs
}

// ----------------------------> write sendCommand <--------------------------------
func sendRemoveItselfCommand(ctx context.Context, rootId string) (rep string, err error) {
	mapSub := cacheGetMapSub(rootId)
	for ds, subId := range mapSub {
		// send to each subObject
		// LoggingClient.Info(fmt.Sprintf("sendRemoveItselfCommand to %s", subId)) // <--------------------- wirte sendCommand()
		rep, err = sendManagerPutCommandByLabel(ctx, ds, subId, PreficManager, ManagerRemoveItself, MangerDeleteMethod, subId)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error: sendRemoveItselfCommand(%s)", subId))
		}
	}
	return
}

/*
	Delete Object -> DeleteElementInObject -> deleteElementInProtocols -> map.deleteRoot
*/

func DeleteObject(ctx context.Context, rootId string) (rep string, err error) {
	// 1. gui lenh RemoveItself toi chinh no, khi do, cac child cua no se phai tu unsubscribe
	rep, err = sendRemoveItselfCommand(ctx, rootId)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	// 2. xoa Object nay trong Procols cua cac Parent. Chi can xoa trong MetaData, khong can gui lenh
	for parentId := range cacheGetParents(rootId) {
		parentObject, err := clientMetaDevice.Device(parentId, ctx)
		if err == nil {
			deleteElementInProtocols(parentObject.Protocols, rootId)
			err = clientMetaDevice.Update(parentObject, ctx)
			if err != nil {
				LoggingClient.Error(err.Error())
			}
		}
	}
	// 3. xoa cac SubObject trong MetaData & xoa rootObject
	// doi voi DeviceType, Sub trung voi Root nen bo qua phan xoa SubObject
	if cacheGetType(rootId) != DEVICETYPE {
		listOb, _ := clientMetaDevice.DevicesByLabel(rootId, ctx)
		for _, sub := range listOb {
			fmt.Println("delete subObject ", sub.Id)
			clientMetaDevice.Delete(sub.Id, ctx)
			cacheDeleteMapHasName(sub.Name)
		}
	}
	// xoa RootObject
	err = clientMetaDevice.Delete(rootId, ctx)
	// 4. xoa Object trong MapRoot & xoa trong MapID
	cacheDeleteRoot(rootId)
	return
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------
/*
	Add/Update Element -> subscribeToObject -> sendSubscribeCommand + sendScheduleSetCommand
*/
// ----------------------------> write sendCommand <--------------------------------
func sendSubscribeCommand(ctx context.Context, receiverId string, body contentElementType) (rep string, err error) {
	i := 0
	mapSubElementSubObject, sliceDs := createMapElementSubObject(receiverId, body.OwnerId)
	for subElementId, subObjectId := range mapSubElementSubObject {
		// set: subElement subscribe to subObject
		subBody := body
		subBody.OwnerId = subObjectId
		rep, err = sendManagerPutCommandByLabel(ctx, sliceDs[i], subElementId, PreficManager, ManagerSubcribe, ManagerPutMethod, subBody.String())
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error: sendSubscribeCommand(%s, %s)", subElementId, subBody.String()))
		}
		i++
	}
	return
}

func renameElementInParent(ctx context.Context, oldName string, parentId string, newName string) {
	d, err := clientMetaDevice.Device(parentId, ctx)
	if err != nil {
		return
	}
	pp, ok := d.Protocols[oldName]
	if !ok {
		return
	}
	delete(d.Protocols, oldName)
	d.Protocols[newName] = pp
	clientMetaDevice.Update(d, ctx)
}

func getElementInParent(ctx context.Context, elementId string, parentId string) (contentElementType, error) {
	d, err := clientMetaDevice.Device(parentId, ctx)
	if err != nil {
		return contentElementType{}, err
	}
	elementName := convertIdName(elementId)
	pp, ok := d.Protocols[elementName]
	if !ok {
		return contentElementType{}, fmt.Errorf("khong tim thay ten cho Element ", elementId)
	}
	b, err := json.Marshal(pp)
	if err != nil {
		return contentElementType{}, err
	}

	var ct contentElementType
	err = json.Unmarshal(b, &ct)
	if err != nil {
		return contentElementType{}, err
	}
	return ct, err
}

func setSubscribeForElement(ctx context.Context, elementId string, content contentElementType) (rep string, err error) {
	objectId := content.OwnerId
	rep, err = sendSubscribeCommand(ctx, elementId, content)
	rep, err = rePutScheduleToObject(ctx, objectId, objectId)

	// update lai cac schedule da cai dat cho objecId
	for obId := range cacheGetParents(objectId) {
		// rep, err = rePutScheduleToObject(ctx, objectId, obId)
		element, err := getElementInParent(ctx, objectId, obId)
		if err == nil {
			PutElementToObject(ctx, obId, element, false)
		}
	}
	return
}

// them vao objectName 1 element co noi dung la "body"

func PutElementToObject(ctx context.Context, objectId string, body contentElementType, hasStore bool) (rep string, err error) {
	// 1. check new DS
	elementId := body.ElementId
	ok := checkExit(elementId)
	if !ok {
		return "false", fmt.Errorf("Khong ton tai Element")
	}
	body.ObjectType = cacheGetType(elementId)
	// body.UserName = infoElement.getUserName()

	ok = checkExit(objectId)
	if !ok {
		return "false", fmt.Errorf("Khong ton tai Object")
	}

	if cacheGetType(objectId) == GROUPTYPE {
		// 2 truong command-body phai trong
		body.Command = ""
		body.Body = ""
	}

	if cacheGetType(objectId) == SCENARIOTYPE {
		// 2 truong command-body phai khac rong
		if body.Command == "" || body.Body == "" {
			return "false", fmt.Errorf("Thieu thong tin Lenh cho Element trong kich ban")
		}
	}

	body.OwnerId = objectId
	dsObjectList := cacheGetDSsOfRootObject(objectId)
	dsElementList := cacheGetDSsOfRootObject(elementId)
	newds := getNewDSFromTwoList(dsObjectList, dsElementList)
	rootObject, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return "false", err
	}

	objectType := cacheGetType(objectId)
	for _, inewds := range newds {
		createSubObject(ctx, &rootObject, inewds)
	}

	if hasStore {
		// 4. Update in MetaData
		// update MapRoot.Parent of element
		cacheAddParents(elementId, objectId, objectType)

		putElementToProtocols(rootObject.Protocols, elementId, body)
		err = clientMetaDevice.Update(rootObject, ctx)
		if err != nil {
			LoggingClient.Error(err.Error())
			// return "false", err
		}
	}
	// 2. send to Element: subscribe to RootObject & body = content
	// 3. update Schedule of Parent, superParent

	setSubscribeForElement(ctx, elementId, body)

	return "true", err
}

/*
	DeleteElement: -> unSubscribeInObject -> sendUnSubscribeCommand
*/
// ----------------------------> write sendCommand <--------------------------------
func sendUnSubscribeCommand(ctx context.Context, receiverObjectId string, objectId string) (rep string, err error) {
	i := 0
	mapSubElementSubObject, sliceDs := createMapElementSubObject(receiverObjectId, objectId)
	for subElement, subObject := range mapSubElementSubObject {
		// subElement Unsubscribe to subObject
		_, err := sendManagerPutCommandByLabel(ctx, sliceDs[i], subElement, PreficManager, ManagerSubcribe, MangerDeleteMethod, subObject)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error: sendSubscribeCommand(%s, %s)", subElement, subObject))
		}
		i++
	}
	return
}

func unSubscribeForElement(ctx context.Context, elementId string, objectId string) (rep string, err error) {
	// cac device thuc phai tu xoa cac schedule lien quan den object huy dang ky
	rep, err = sendUnSubscribeCommand(ctx, elementId, objectId)
	return
}

func DeleteElementInObject(ctx context.Context, elementId string, objectId string) (rep string, err error) {
	// 1. send request
	unSubscribeForElement(ctx, elementId, objectId)
	rootObject, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return "false", err
	}
	// update MapRoot.Parent of element
	cacheDeleteParents(elementId, objectId)

	// 2. Update Protocols of Object in MetaData
	deleteElementInProtocols(rootObject.Protocols, elementId)
	err = clientMetaDevice.Update(rootObject, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "false", err
	}
	return "true", err
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------
/*
	Command
*/

func GetCommandForDevice(ctx context.Context, objectId string) ([]models.Command, error) {
	// ok := checkExit(objectId)
	// if !ok {
	// 	LoggingClient.Error("khong ton tai doi tuong")
	// 	return nil, fmt.Errorf("khong ton tai doi tuong")
	// }
	return clientMetaCommand.CommandsForDeviceId(objectId, ctx)
}

func IssuePutCommandByObjectName(ctx context.Context, objectId string, commandName string, body string) (rep string, err error) {
	ok := checkExit(objectId)
	if !ok {
		return "fasle", fmt.Errorf("Error: Not found device")
	}
	listSub := cacheGetSubIdsOfRootObject(objectId)
	if len(listSub) == 0 {
		LoggingClient.Warn("Doi tuong khong co Element nao de dieu khien")
		return "false", fmt.Errorf("Doi tuong khong co Element nao de dieu khien")
	}
	for _, subObject := range listSub {
		rep, err = clientCommand.PutDeviceCommandByNames(convertIdName(subObject), commandName, body, ctx)
		if err != nil {
			LoggingClient.Error(err.Error())
		}
	}
	return
}
func getValueDescriptorByDeviceIdAndCommandName(ctx context.Context, objectId string, commandName string, isPut bool) (map[string]string, error) {
	listcm, err := clientMetaCommand.CommandsForDeviceId(objectId, ctx)
	if err != nil {
		return nil, err
	}
	for _, cm := range listcm {
		if cm.Name == commandName {
			listvl := make(map[string]string)
			if isPut {
				if &(cm.Put) == nil {
					return nil, fmt.Errorf("Khong ho tro lenh PUT")
				}
				cm.Put.AllAssociatedValueDescriptors(&listvl)
				return listvl, err
			}
			if &(cm.Get) == nil {
				return nil, fmt.Errorf("Khong ho tro lenh GET")
			}
			cm.Get.AllAssociatedValueDescriptors(&listvl)
			return listvl, err
		}
	}
	return nil, fmt.Errorf("Error: Not found Command")
}

func issueGetCommandOfDevice(ctx context.Context, objectId string, commandName string) (models.Event, error) {
	mVl, err := getValueDescriptorByDeviceIdAndCommandName(ctx, objectId, commandName, false)
	if err != nil {
		return models.Event{}, err
	}
	if len(mVl) == 0 {
		return models.Event{}, fmt.Errorf("Khong co resource nao lien quan toi lenh")
	}
	var ev = models.Event{
		Origin:  -1,
		Created: -1,
	}

	ev.Device = convertIdName(objectId)
	var rds []models.Reading
	for nvl := range mVl {
		r, err := clientCoreReading.ReadingsForNameAndDevice(nvl, ev.Device, 1, ctx)
		if err != nil {
			return ev, err
		}
		if len(r) == 0 {
			rds = append(rds, models.Reading{})
		} else {
			if r[0].Origin < ev.Origin || ev.Origin == -1 {
				ev.Origin = r[0].Origin
				ev.Created = r[0].Created
			}
			rds = append(rds, r[0])
		}
	}
	ev.Readings = rds
	return ev, err
}

func issueGetCommandOfGroup(ctx context.Context, objectId string, commandName string) ([]models.Event, error) {
	ob, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return nil, err
	}
	var result []models.Event
	p := ob.Protocols
	for pp := range p {
		if pp != PROTOCOLSNETWORKNAME && pp != PROTOCOLSSCHEDULENAME {
			elId := convertNameId(pp)
			elEv, e := issueGetCommandOfDevice(ctx, elId, commandName)
			if e != nil {
				return nil, e
			}
			result = append(result, elEv)
		}
	}
	return result, err
}

func compareEvent(rq map[string]string, ev models.Event) bool {
	if ev.String() == (models.Event{}).String() {
		return false
	}
	rds := ev.Readings
	if len(rds) != len(rq) {
		return false
	}
	for _, r := range rds {
		name := r.Name
		mv, ok := rq[name]
		if !ok {
			return false
		}
		if r.Value != mv {
			return false
		}
	}
	return true
}

func checkStatusOfElement(ctx context.Context, objectId string, commandName string, body string) (bool, int64, int64, error) {
	elementType := cacheGetType(objectId)
	if elementType == "" {
		return false, -1, -1, fmt.Errorf("Khong biet loai cua Element")
	} else {
		if elementType == DEVICETYPE {
			return checkStatusOfDevice(ctx, objectId, commandName, body)
		}
		return checkStatusOfGroup(ctx, objectId, commandName, body)
	}
}

func checkStatusOfDevice(ctx context.Context, objectId string, commandName string, body string) (result bool, originMin int64, createdMin int64, err error) {
	rq := make(map[string]string)
	err = json.Unmarshal([]byte(body), &rq)
	if err != nil {
		return false, -1, -1, err
	}
	ev, err := issueGetCommandOfDevice(ctx, objectId, commandName)
	if err != nil {
		return false, -1, -1, err
	}
	result = compareEvent(rq, ev)
	return result, ev.Origin, ev.Created, err
}

func checkStatusOfGroup(ctx context.Context, objectId string, commandName string, body string) (result bool, originMin int64, createdMin int64, err error) {
	rq := make(map[string]string)
	err = json.Unmarshal([]byte(body), &rq)
	if err != nil {
		return false, -1, -1, err
	}
	originMin = -1
	createdMin = -1
	result = true
	listEv, err := issueGetCommandOfGroup(ctx, objectId, commandName)
	if err != nil {
		return false, -1, -1, err
	}
	for _, ev := range listEv {
		if compareEvent(rq, ev) == false {
			result = false
		}
		if originMin > ev.Origin || originMin == -1 {
			originMin = ev.Origin
			createdMin = ev.Created
		}
	}
	return result, originMin, createdMin, err
}

func createEventFromBool(ctx context.Context, objectId string, commandName string, value bool, origin int64, created int64) models.Event {
	device := convertIdName(objectId)
	rd := models.Reading{
		Device:  device,
		Name:    commandName,
		Origin:  origin,
		Created: created,
		Value:   strconv.FormatBool(value),
	}
	rds := make([]models.Reading, 1)
	rds[0] = rd
	result := models.Event{
		Device:   device,
		Origin:   origin,
		Created:  created,
		Readings: rds,
	}
	return result
}

func getCommandFromProperty(pp models.ProtocolProperties) string {
	n, ok := pp[PROTOCOLSCOMMANDNAME]
	if !ok {
		return ""
	}
	return n
}

func getBodyFromProperty(pp models.ProtocolProperties) string {
	b, ok := pp[PROTOCOLSBODYNAME]
	if !ok {
		return ""
	}
	return b
}

func issueGetCommandOfScenario(ctx context.Context, objectId string, commandName string) ([]models.Event, error) {
	ob, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		return nil, err
	}
	mVl, err := getValueDescriptorByDeviceIdAndCommandName(ctx, objectId, commandName, false)
	if err != nil {
		return nil, err
	}
	if len(mVl) == 0 {
		return nil, fmt.Errorf("Khong co resource nao lien quan toi lenh")
	}

	var result []models.Event
	p := ob.Protocols
	for pp, vpp := range p {
		if pp != PROTOCOLSNETWORKNAME && pp != PROTOCOLSSCHEDULENAME {
			elId := convertNameId(pp)
			cm := getCommandFromProperty(vpp)
			if cm == "" {
				return nil, fmt.Errorf("khong tim thay lenh cho element ", pp)
			}
			body := getBodyFromProperty(vpp)
			if body == "" {
				return nil, fmt.Errorf("khong tim thay noi dung lenh cho element ", pp)
			}
			check, origin, created, err := checkStatusOfElement(ctx, elId, cm, body)
			if err != nil {
				return nil, err
			}
			ev := createEventFromBool(ctx, elId, commandName, check, origin, created)
			result = append(result, ev)
		}
	}
	return result, err
}

func IssueGetCommandByObjectName(ctx context.Context, objectId string, commandName string) (string, error) {
	ok := checkExit(objectId)
	if !ok {
		return "", fmt.Errorf("Error: Not found Object")
	}
	listSub := cacheGetSubIdsOfRootObject(objectId)
	if len(listSub) == 0 {
		LoggingClient.Warn("Doi tuong khong co Element nao de dieu khien")
		return "", fmt.Errorf("Doi tuong khong co Element nao de dieu khien")
	}

	var err error
	var str string
	var rs []byte
	var listEv []models.Event

	obType := cacheGetType(objectId)
	if obType == "" {
		return "", fmt.Errorf("Khong xac dinh duoc loai doi tuong")
	}

	switch obType {
	case DEVICETYPE:
		str, err = clientCommand.GetDeviceCommandByNames(convertIdName(objectId), commandName, ctx)
		if err != nil {
			LoggingClient.Error(err.Error())
			return "", err
		}
		var ev models.Event
		err = json.Unmarshal([]byte(str), &ev)
		if err != nil {
			LoggingClient.Error(err.Error())
			return "", err
		}
		if ev.String() != (models.Event{}).String() {
			ev.Origin = -1
			// ev.Created = -1
			for _, r := range ev.Readings {
				if ev.Origin == -1 || ev.Origin > r.Origin {
					ev.Origin = r.Origin
				}
			}
		}
		listEv = append(listEv, ev)
	case GROUPTYPE:
		listEv, err = issueGetCommandOfGroup(ctx, objectId, commandName)
		if err != nil {
			LoggingClient.Error(err.Error())
			return "", err
		}
	case SCENARIOTYPE:
		listEv, err = issueGetCommandOfScenario(ctx, objectId, commandName)
		if err != nil {
			LoggingClient.Error(err.Error())
			return "", err
		}
	}
	rs, err = json.Marshal(listEv)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}
	return string(rs), err
}

//----------------------------------------------------------------> REST-API <-----------------------------------------------------------------------------
func RestPostObject(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var d models.Device
	err := json.NewDecoder(r.Body).Decode(&d)
	// Problem decoding
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding the object: " + err.Error())
		return
	}

	ctx := r.Context()

	body, err := CreateRootObject(ctx, &d)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestPutObject(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var d models.Device
	err := json.NewDecoder(r.Body).Decode(&d)
	// Problem decoding
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding the object: " + err.Error())
		return
	}

	ctx := r.Context()
	body, err := UpdateRootObject(ctx, &d)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestDeleteObject(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	var objectName = vars[OBJECTNAME]
	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)
	ctx := r.Context()

	body, err := DeleteObject(ctx, objectId)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestGetObjectForName(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	var objectName = vars[OBJECTNAME]
	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)
	ctx := r.Context()

	object, err := clientMetaDevice.Device(objectId, ctx)
	if err != nil {
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	body, err := json.Marshal(object)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestGetObjectsList(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	ctx := r.Context()
	list, err := clientMetaDevice.DevicesByLabel(ROOTOBJECT, ctx)
	if err != nil {
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	body, err := json.Marshal(list)
	reponseHTTPrequest(w, body, err)
}

func RestPutElement(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	elementName := vars[ELEMENTNAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	if checkExitByName(elementName) == false {
		err := fmt.Errorf("Element khong ton tai")
		LoggingClient.Error("Element khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	objectId := convertNameId(objectName)
	elementId := convertNameId(elementName)

	if cacheGetType(objectId) == DEVICETYPE {
		err := fmt.Errorf("Khong the them Element toi loai Device")
		LoggingClient.Warn("Khong the them Element toi loai Device")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	if cacheGetType(elementId) == SCENARIOTYPE {
		err := fmt.Errorf("Khong ho tro them Element loai Scenario")
		LoggingClient.Warn("Khong ho tro them Element loai Scenario")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	if cacheGetType(elementId) == GROUPTYPE && cacheGetType(objectId) == GROUPTYPE {
		err := fmt.Errorf("Khong ho tro them Element loai Group vao loai Group")
		LoggingClient.Warn("Khong ho tro them Element loai Group vao loai Group")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	var content contentElementType
	err := json.NewDecoder(r.Body).Decode(&content)
	// Problem decoding
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding the object: " + err.Error())
		return
	}

	ctx := r.Context()
	content.ElementId = elementId
	body, err := PutElementToObject(ctx, objectId, content, true)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestDeleteElement(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	elementName := vars[ELEMENTNAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	if checkExitByName(elementName) == false {
		err := fmt.Errorf("Element khong ton tai")
		LoggingClient.Error("Element khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)
	elementId := convertNameId(elementName)

	t := cacheGetParentTypeByParentId(elementId, objectId)
	if t == "" {
		err := fmt.Errorf("%s khong chua Element: %s", objectId, elementId)
		LoggingClient.Warn("%s khong chua Element: %s", objectId, elementId)
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	ctx := r.Context()

	body, err := DeleteElementInObject(ctx, elementId, objectId)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestPutSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	scheduleName := vars[SCHEDULENAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)

	var content contentScheduleType
	err := json.NewDecoder(r.Body).Decode(&content)
	// Problem decoding
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding the object: " + err.Error())
		return
	}

	ctx := r.Context()
	content.ScheduleName = scheduleName
	body, err := PutScheduleToObject(ctx, objectId, content)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	scheduleName := vars[SCHEDULENAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)

	ctx := r.Context()
	var content = contentScheduleType{
		OwnerId:      objectId,
		ScheduleName: scheduleName,
	}

	body, err := DeleteScheduleInObject(ctx, objectId, content)
	reponseHTTPrequest(w, []byte(body), err)
}

func RestGetCommandForObjectName(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)

	list, err := GetCommandForDevice(ctx, objectId)
	if err != nil {
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	body, err := json.Marshal(list)
	reponseHTTPrequest(w, body, err)
}

func RestIssueGetCommand(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	commandName := vars[COMMANDNAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)

	body, err := IssueGetCommandByObjectName(ctx, objectId, commandName)
	if err != nil {
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}
	reponseHTTPrequest(w, []byte(body), err)
}

func RestIssuePutCommand(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	objectName := vars[OBJECTNAME]
	commandName := vars[COMMANDNAME]

	if checkExitByName(objectName) == false {
		err := fmt.Errorf("Doi tuong khong ton tai")
		LoggingClient.Error("Doi tuong khong ton tai")
		reponseHTTPrequest(w, []byte("false"), err)
		return
	}

	objectId := convertNameId(objectName)

	content, err := ioutil.ReadAll(r.Body)
	// Problem decoding
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding the object: " + err.Error())
		return
	}

	ctx := r.Context()
	body, err := IssuePutCommandByObjectName(ctx, objectId, commandName, string(content))
	reponseHTTPrequest(w, []byte(body), err)
}
