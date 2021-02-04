// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"runtime"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

type resourceBool struct{}

func (rb *resourceBool) value(db *db, deviceName, deviceResourceName string) (*dsModels.CommandValue, error) {
	result := &dsModels.CommandValue{}
	/* To stress cpu*/
	done := make(chan int)
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	for i := 0; i < n; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}
	//time.Sleep(10 * time.Second)
	for i := 0; i < n; i++ {
		done <- 1
	}
	/*yep*/
	enableRandomization, currentValue, _, err := db.getVirtualResourceData(deviceName, deviceResourceName)
	if err != nil {
		return result, err
	}

	var newValueBool bool
	if enableRandomization {
		rand.Seed(time.Now().UnixNano())
		newValueBool = rand.Int()%2 == 0
	} else {
		if newValueBool, err = strconv.ParseBool(currentValue); err != nil {
			return result, err
		}
	}
	now := time.Now().UnixNano()
	if result, err = dsModels.NewBoolValue(deviceResourceName, now, newValueBool); err != nil {
		return result, err
	}
	if err := db.updateResourceValue(result.ValueToString(), deviceName, deviceResourceName, false); err != nil {
		return result, err
	}

	return result, nil
}

func (rb *resourceBool) write(param *dsModels.CommandValue, deviceName string, db *db) error {
	enableRandomizationPrefix := "EnableRandomization_"
	if strings.Contains(param.DeviceResourceName, enableRandomizationPrefix) {
		if v, err := param.BoolValue(); err == nil {
			return db.updateResourceRandomization(v, deviceName, param.DeviceResourceName[len(enableRandomizationPrefix):len(param.DeviceResourceName)])
		} else {
			return fmt.Errorf("resourceBool.write: %v", err)
		}
	} else {
		if _, err := param.BoolValue(); err == nil {
			return db.updateResourceValue(param.ValueToString(), deviceName, param.DeviceResourceName, true)
		} else {
			return fmt.Errorf("resourceBool.write: %v", err)
		}
	}
}
