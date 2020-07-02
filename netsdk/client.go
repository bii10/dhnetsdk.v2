package netsdk

// #cgo CFLAGS: -I .
// #cgo LDFLAGS: -L . -ldhnetsdk

// #include <stdio.h>
// #include <stdlib.h>
// #include "dhnetsdk.h"
// extern int export_fAnalyzerDataCallBack2(long lAnalyzerHandle, unsigned int dwAlarmType, long pAlarmInfo, long pBuffer,unsigned int dwBufSize, long dwUser, int nSequence,long reserved);
import "C"

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type (
	ReconnectFunc  func(ip string, port int)
	DisconnectFunc func(ip string, port int)
	// DVRMessageFunc func(cmd DhAlarmType, buf []byte, ip string, port int) bool
	PictureExFunc func(client *Client, AlarmType EventIvs, alarmInfo interface{}, frame []byte, seq int) int
)

type Client struct {
	LoginID int

	realloadHandle int64

	DeviceInfo NET_DEVICEINFO_Ex
}

func Init(cb DisconnectFunc) error {
	initParam := NETSDK_INIT_PARAM{}
	bRet := InitEx(func(lLoginID LLONG, pchDVRIP string, nDVRPort int, dwUser LLONG) {
		if cb != nil {
			cb(pchDVRIP, nDVRPort)
		}
	}, &initParam)
	if false == bRet {
		return fmt.Errorf("Init NetSDK failed")
	}
	return nil
}

func NewClient() *Client {
	return &Client{}
}

func Login(addr string, user, pass string) (ncli *Client, err error) {
	var (
		port     int
		inParam  NET_IN_LOGIN_WITH_HIGHLEVEL_SECURITY
		outParam NET_OUT_LOGIN_WITH_HIGHLEVEL_SECURITY
	)

	addrs := strings.SplitN(addr, ":", 2)
	if len(addrs) == 2 {
		addr = addrs[0]
		if port, err = strconv.Atoi(addrs[1]); err != nil {
			return nil, err
		}
	} else {
		return nil, ErrInvalidAddress
	}
	copy(inParam.ST_szIP[:], []byte(addr))
	inParam.ST_nPort = int32(port)
	copy(inParam.ST_szUserName[:], []byte(user))
	copy(inParam.ST_szPassword[:], []byte(pass))

	id := LoginWithHighLevelSecurity(&inParam, &outParam)

	ncli = &Client{
		LoginID:    id,
		DeviceInfo: outParam.ST_stDeviceInfo,
	}
	return ncli, nil
}

func (client *Client) Logout() bool {
	return Logout(client.LoginID)
}

func (client *Client) StartListen() bool {
	return StartListenEx(client.LoginID)
}

func (client *Client) StopListen() bool {
	return StopListen(client.LoginID)
}

func (client *Client) RealLoadPictureEx(channel int, evt EventIvs, callback PictureExFunc) error {
	var visior = PictureVisitor{
		Client:   client,
		Callback: callback,
	}
	// var userdata *LLONG = (*LLONG)(unsafe.Pointer(&visior))
	p := pointer.Save(&visior)

	lAnalyzerHandle := C.CLIENT_RealLoadPictureEx(
		C.long(client.LoginID),
		C.int(channel),
		C.uint(evt),
		C.int(1),
		C.fAnalyzerDataCallBack(C.export_fAnalyzerDataCallBack2),
		C.long(uintptr(p)),
		unsafe.Pointer(uintptr(0)))

	if lAnalyzerHandle != 0 {
		log.Println("CLIENT_RealLoadPictureEx success")
	} else {
		return errors.New("can't realloadPicture")
	}

	log.Println("lAnalyzerHandle=", lAnalyzerHandle)
	client.realloadHandle = LLONG(lAnalyzerHandle)

	return nil
}
