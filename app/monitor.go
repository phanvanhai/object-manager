package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

// RestGetReadingByDeviceNameInTimeRange :get all readings by nameID in the time range
func RestGetReadingByDeviceNameInTimeRange(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	devicename := vars[DEVICENAME]
	start := vars[START]
	end := vars[END]
	limit := vars[LIMIT]

	startInt, _ := strconv.ParseInt(start, 10, 64)
	endInt, _ := strconv.ParseInt(end, 10, 64)
	limitInt, _ := strconv.ParseInt(limit, 10, 64)

	ctx := r.Context()
	body, err := getReadingByDeviceNameInTimeRange(ctx, devicename, startInt, endInt, limitInt)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	reponseHTTPrequest(w, body, err)
}

// RestGetReadingByReadingNameInTimeRange :get all readings by nameID in the time range
func RestGetReadingByReadingNameInTimeRange(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	readingname := vars[READINGNAME]
	start := vars[START]
	end := vars[END]
	limit := vars[LIMIT]

	startInt, _ := strconv.ParseInt(start, 10, 64)
	endInt, _ := strconv.ParseInt(end, 10, 64)
	limitInt, _ := strconv.ParseInt(limit, 10, 64)

	ctx := r.Context()
	body, err := getReadingByReadingNameInTimeRange(ctx, readingname, startInt, endInt, limitInt)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	reponseHTTPrequest(w, body, err)
}

// getReadingByDeviceNameInTimeRange :get all readings by nameID in the time range
func getReadingByDeviceNameInTimeRange(ctx context.Context, nameID string, from int64, to int64, limit int64) ([]byte, error) {
	const batchSize = 10
	const maxRequest = 1000 // max = maxRequest * batchSize = 10 000
	if limit > (maxRequest * batchSize) {
		limit = maxRequest * batchSize
	}
	ids := make(map[string]bool)
	result := make([]models.Reading, batchSize*maxRequest)
	var pos int64
	var count int

	for ok := true; ok; ok = (count == batchSize) && (limit > 0) {
		readings, err := clientCoreReading.ReadingsForInterval(int(from), int(to), int(batchSize), ctx)
		if err != nil {
			// LoggingClient.Error(err.Error())
			return nil, err
		}

		count = len(readings)
		if count > 0 {
			from = readings[count-1].Created
		}

		for _, reading := range readings {
			if reading.Device != nameID {
				continue
			}
			id := reading.Id
			if ids[id] == true {
				continue
			}
			ids[id] = true

			result[pos] = reading
			// result[pos].UserName = MapRoot[reading.Device].getUserName()
			pos++
			if pos == limit {
				break
			}
		}
		limit = limit - pos
	}

	return json.Marshal(result[:pos])
}

// getReadingByReadingNameInTimeRange :get all readings of all Devices by ReadingName in the time range
func getReadingByReadingNameInTimeRange(ctx context.Context, readingname string, from int64, to int64, limit int64) ([]byte, error) {
	const batchSize = 10
	const maxRequest = 1000 // max = maxRequest * batchSize = 10 000
	if limit > (maxRequest * batchSize) {
		limit = maxRequest * batchSize
	}

	ids := make(map[string]bool)
	result := make([]models.Reading, batchSize*maxRequest)
	var pos int64
	var count int

	for ok := true; ok; ok = (count == batchSize) && (limit > 0) {
		readings, err := clientCoreReading.ReadingsForInterval(int(from), int(to), int(batchSize), ctx)
		if err != nil {
			// LoggingClient.Error(err.Error())
			return nil, err
		}

		count = len(readings)
		if count > 0 {
			from = readings[count-1].Created
		}

		for _, reading := range readings {
			if reading.Name != readingname {
				continue
			}
			id := reading.Id
			if ids[id] == true {
				continue
			}
			ids[id] = true

			result[pos] = reading
			// result[pos].UserName = MapRoot[reading.Device].getUserName()

			pos++
			if pos == limit {
				break
			}
		}
		limit = limit - pos
	}

	return json.Marshal(result[:pos])
}

func RestGetValueDescriptorByName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	name := vars[NAME]
	body, err := getValueDescriptorByName(name)
	reponseHTTPrequest(w, body, err)
}

func RestGetValueDescriptorByDeviceName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	devicename := vars[DEVICENAME]
	body, err := getValueDescriptorByDeviceName(devicename)
	reponseHTTPrequest(w, body, err)
}

//GetValueDescriptorByName : get ValueDescriptor by name
func getValueDescriptorByName(name string) ([]byte, error) {
	vd, err := clientCoreValueDescriptor.ValueDescriptorForName(name, context.Background())
	if err != nil {
		return nil, err
	}

	return json.Marshal(vd)
}

//GetValueDescriptorByDeviceName : get ValueDescriptors by devicename
func getValueDescriptorByDeviceName(name string) ([]byte, error) {
	vd, err := clientCoreValueDescriptor.ValueDescriptorsForDeviceByName(name, context.Background())
	if err != nil {
		return nil, err
	}

	return json.Marshal(vd)
}
