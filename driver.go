package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type ledControllerProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}

type ProtocolConfigData struct {
	DeviceID int `json:"deviceID,omitempty"`
}

type ledControllerProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}

type CommonCustomizedValues struct {
	ProtocolID int `json:"protocolID"`
}
type ledControllerVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	DataType string `json:"dataType"`
}

// ledController Realize the structure of random number
type ledController struct {
	mutex                 			sync.Mutex
	ledControllerProtocolConfig 	ledControllerProtocolConfig
	protocolCommonConfig  			ledControllerProtocolCommonConfig
	visitorConfig         			ledControllerVisitorConfig
	client                			map[int]int64
}

// InitDevice Sth that need to do in the first
// If you need mount a persistent connection, you should provIDe parameters in configmap's protocolCommon.
// and handle these parameters in the following function
func (vd *ledController) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &vd.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return err
		}
	}
	fmt.Printf("InitDevice%d...\n", vd.protocolCommonConfig.ProtocolID)
	return nil
}

// SetConfig Parse the configmap's raw json message
func (vd *ledController) SetConfig(protocolCommon, visitor, protocol []byte) (dataType string, deviceID int, err error) {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()
	vd.NewClient()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &vd.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return "", 0, err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &vd.visitorConfig); err != nil {
			fmt.Printf("Unmarshal visitorConfig error: %v\n", err)
			return "", 0, err
		}
	}

	if protocol != nil {
		if err = json.Unmarshal(protocol, &vd.ledControllerProtocolConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolConfig error: %v\n", err)
			return "", 0, err
		}
	}
	dataType = vd.visitorConfig.DataType
	deviceID = vd.ledControllerProtocolConfig.DeviceID
	return
}

// ReadDeviceData  is an interface that reads data from a specific device, data's dataType is consistent with configmap
func (vd *ledController) ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error) {
	// Parse raw json message to get a ledController instance
	DataTye, DeviceID, err := vd.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return nil, err
	}
	if DataTye == "string" {
		if vd.client[DeviceID] == 0 {
			return 0, errors.New("vd.limit should not be 0")
		}
		return "OFF", nil
	}else {
		return "", errors.New("dataType don't exist")
	}
}

// WriteDeviceData is an interface that write data to a specific device, data's dataType is consistent with configmap
func (vd *ledController) WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error) {
	// Parse raw json message to get a ledController instance
	_, DeviceID, err := vd.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return err
	}
	vd.client[DeviceID] = data.(int64)
	return nil
}

// StopDevice is an interface to disconnect a specific device
// This function is called when mapper stops serving
func (vd *ledController) StopDevice() (err error) {
	// in this func, u can get ur device-instance in the client map, and give a safety exit
	fmt.Println("----------Stop LED Controller Successful----------")
	return nil
}

// NewClient create device-instance, if device-instance exit, set the limit as 100.
// Control a group of devices through singleton pattern
func (vd *ledController) NewClient() {
	if vd.client == nil {
		vd.client = make(map[int]int64)
	}
	if _, ok := vd.client[vd.ledControllerProtocolConfig.DeviceID]; ok {
		if vd.client[vd.ledControllerProtocolConfig.DeviceID] == 0 {
			vd.client[vd.ledControllerProtocolConfig.DeviceID] = 100
		}
	}
}

// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
func (vd *ledController) GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool) {
	_, _, err := vd.SetConfig(protocolCommon, visitor, protocol)
	return err == nil
}
