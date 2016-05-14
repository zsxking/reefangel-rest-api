package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

func startServer() *echo.Echo {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.Get("/ra/v1/stats", handleGetRAStatus())

	// Start server
	e.Run(standard.New(":1323"))

	return e
}

func handleGetRAStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, err := http.Get("http://192.168.1.20:2000/r99")
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return err
		}
		raStatusRaw := RAStatusRAW{}
		err = xml.Unmarshal([]byte(body), &raStatusRaw)
		if err != nil {
			fmt.Println(err)
			return err
		}
		raStatus := RAStatus{}
		_ = xml.Unmarshal([]byte(body), &raStatus)
		if err != nil {
			fmt.Println(err)
			return err
		}
		raStatus.Relays = NewRelays(
			raStatusRaw.RelayData,
			raStatusRaw.RelayMaskOn,
			raStatusRaw.RelayMaskOff,
			nil,
		)

		// raStatus.ExpModules = NewExpModules(raStatusRaw.ExpModule + raStatusRaw.ExpModule1*256)
		newExpModules := &ExpModules{}
		ReadBits(raStatusRaw.ExpModule + raStatusRaw.ExpModule1*256, newExpModules, true)
		raStatus.ExpModules = newExpModules

		alertFlag := &AlertFlag{}
		ReadBits(raStatusRaw.AlertFlags, alertFlag, false)
		raStatus.AlertFlag = alertFlag

		statusFlag := &StatusFlag{}
		ReadBits(raStatusRaw.StatusFlags, statusFlag, false)
		raStatus.StatusFlag = statusFlag

		return c.JSON(http.StatusOK, raStatus)
	}
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
