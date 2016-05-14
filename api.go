package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/tarm/serial"
)

type RaApi struct {
	Echo         *echo.Echo
	SerialStatus *RAStatus
}

func startServer() *RaApi {
	// Echo instance
	e := echo.New()
	api := &RaApi{Echo: e}

	api.StartStatusGetter("ttyUSB0")

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.Get("/ra/v2/stats/wifi/:host", handleGetRAStatusFromWifi)
	e.Get("/ra/v2/stats/serial", api.handleGetCachedSerialStatus)
	e.Get("/:command", handleCommandBypass)

	// Start server
	e.Run(standard.New(":2000"))

	return api
}

func (api *RaApi) StartStatusGetter(name string) {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			response, err := getFromSerial(name, "sa")
			if err != nil {
				fmt.Println(err)
				continue
			}
			status, err := parseStatusResponse(response)
			if err != nil {
				fmt.Println(err)
				continue
			}
			status.Source = name
			api.SerialStatus = status
			fmt.Println(status)
		}
	}()
}

func (api *RaApi) handleGetCachedSerialStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, api.SerialStatus)
}

func handleCommandBypass(c echo.Context) error {
	command := c.Param("command")
	response, err := getFromSerial("ttyUSB0", command)
	if err != nil {
		return err
	}
	return c.XML(http.StatusOK, response)
}

func parseStatusResponse(response []byte) (*RAStatus, error) {

	raStatusRaw := RAStatusRAW{}
	err := xml.Unmarshal(response, &raStatusRaw)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	raStatus := RAStatus{}
	_ = xml.Unmarshal(response, &raStatus)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	raStatus.Relays = NewRelays(
		raStatusRaw.RelayData,
		raStatusRaw.RelayMaskOn,
		raStatusRaw.RelayMaskOff,
		nil,
	)

	// raStatus.ExpModules = NewExpModules(raStatusRaw.ExpModule + raStatusRaw.ExpModule1*256)
	newExpModules := &ExpModules{}
	ReadBits(raStatusRaw.ExpModule+raStatusRaw.ExpModule1*256, newExpModules, true)
	raStatus.ExpModules = newExpModules

	alertFlag := &AlertFlag{}
	ReadBits(raStatusRaw.AlertFlags, alertFlag, false)
	raStatus.AlertFlag = alertFlag

	statusFlag := &StatusFlag{}
	ReadBits(raStatusRaw.StatusFlags, statusFlag, false)
	raStatus.StatusFlag = statusFlag

	return &raStatus, nil
}

func handleGetRAStatusFromSerial(c echo.Context) error {
	command := "sa"
	name := "ttyUSB0"
	response, err := getFromSerial(name, command)
	if err != nil {
		return err
	}
	status, err := parseStatusResponse(response)
	if err != nil {
		return err
	}
	status.Source = name
	return c.JSON(http.StatusOK, status)
}

func handleGetRAStatusFromWifi(c echo.Context) error {
	command := "sa"
	host := c.Param("host")
	response, err := getFromWifi(host, command)
	if err != nil {
		return err
	}
	status, err := parseStatusResponse(response)
	if err != nil {
		return err
	}
	status.Source = host
	return c.JSON(http.StatusOK, status)
}

func getFromSerial(serialName string, command string) ([]byte, error) {
	name := "/dev/" + serialName
	config := &serial.Config{Name: name, Baud: 57600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	s.Flush()

	n, err := s.Write([]byte(fmt.Sprintf("GET /%s ", command)))
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	buf := make([]byte, 2048)
	resp := ""
	n = 1
	for n > 0 {
		n, _ = s.Read(buf)
		resp += string(buf[:n])
	}

	respSegs := strings.Split(resp, "\n")
	response := respSegs[len(respSegs)-1]

	return []byte(response), nil
}

func getFromWifi(host string, command string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:2000/%s", host, command))
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	return body, nil
}

func parseIntToBits(n int, numOfBits int) []int {
	bits := make([]int, numOfBits)
	var i uint
	for i = 0; i < uint(numOfBits); i++ {
		bit := n & (1 << i) / (1 << i)
		bits[i] = bit
	}
	return bits
}
