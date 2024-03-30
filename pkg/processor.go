package pkg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type AxfrResult struct {
	Name  string
	Error error
}

type SlaveZone struct {
	Name    string
	Masters []string
	File    string
}

func axfrZone(zone SlaveZone, master string) AxfrResult {
	var returnValue AxfrResult
	var rrSlice []dns.RR
	transfer := new(dns.Transfer)
	transfer.ReadTimeout = time.Duration(10 * time.Second)

	returnValue.Name = zone.Name
	returnValue.Error = nil

	msg := new(dns.Msg)
	msg.SetAxfr(zone.Name)

	masterPort := ""
	if strings.Contains(master, ":") {
		masterPort = master
	} else {
		masterPort = master + ":53"
	}

	axfrChan, err := transfer.In(msg, masterPort)
	if err != nil {
		// log.Fatalln(err.Error())

		returnValue.Error = err

		// log.Printf("master %s %s\n", master, err.Error())
		return returnValue
	}

	for x := range axfrChan {
		for _, y := range x.RR {
			y.Header().Rdlength = 0
			rrSlice = append(rrSlice, y)
		}
	}

	// fmt.Printf("%v\n", rrSlice)
	if len(rrSlice) == 0 {
		returnValue.Error = fmt.Errorf("master %s empty zone", master)
	}
	// log.Printf("master %s ok\n", master)
	return returnValue

}

func zoneWorker(taskChan <-chan SlaveZone, resultChan chan<- []AxfrResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for zone := range taskChan {
		var returnValue []AxfrResult
		// fmt.Printf("Zone: %s, Masters: %v, File: %s\n", zone.Name, zone.Masters, zone.File)
		// flag to detect if there's a good response
		// this should be a conf variable
		goodAnswers := 0
		for _, master := range zone.Masters {
			result := axfrZone(zone, master)
			if result.Error != nil {
				returnValue = append(returnValue, result)
			} else {
				goodAnswers++
			}
		}
		if goodAnswers == 0 {
			resultChan <- returnValue
		}
	}

}

func processZones(zones []SlaveZone, workers int) []AxfrResult {
	var axfresult []AxfrResult
	var wg sync.WaitGroup

	zoneChan := make(chan SlaveZone, len(zones))
	resultChan := make(chan []AxfrResult, len(zones))

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go zoneWorker(zoneChan, resultChan, &wg)
	}

	for _, zone := range zones {
		zoneChan <- zone
	}
	close(zoneChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if len(result) > 0 {
			// fmt.Printf("Zone")
			axfresult = append(axfresult, result...)
		}
	}

	return axfresult

}

func CheckMasters(zones []SlaveZone) error {
	var err error
	if len(zones) > 0 {
		result := processZones(zones, 200)

		if len(result) > 0 {
			strErr := fmt.Sprintf("ERROR in transfer:")
			for _, rslt := range result {
				strErr = fmt.Sprintf("%s\nZone %s %s", strErr, rslt.Name, rslt.Error)

			}
			err = fmt.Errorf(strErr)
		}
	} else {
		err = fmt.Errorf("no zones found")
	}

	return err

}
