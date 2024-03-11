package pkg

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type SlaveZone struct {
	Name    string
	Masters []string
	File    string
}

type AxfrResult struct {
	Name  string
	Error error
}

func parseNamedConf(filePath string) ([]SlaveZone, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var zones []SlaveZone
	var currentZone *SlaveZone

	scanner := bufio.NewScanner(file)
	zonePattern := regexp.MustCompile(`zone\s+"(.+)"\s+in\s+\{`)
	masterPattern := regexp.MustCompile(`masters\s+\{`)
	filePattern := regexp.MustCompile(`file\s+"(.+)";`)
	masterLinePattern := regexp.MustCompile(`[\s\t]+(((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4})`)
	inZone := false
	inMasters := false
	// isSlave := false

	// This is a very basic parser, eventually we need to improve it
	for scanner.Scan() {
		line := scanner.Text()
		if zonePattern.MatchString(line) {
			// fmt.Println(line)
			matches := zonePattern.FindStringSubmatch(line)
			domain := matches[1]
			if !strings.HasSuffix(domain, ".") {
				domain += "."
			}
			currentZone = &SlaveZone{Name: domain}
			inZone = true
			// } else if strings.Contains(line, "type slave") {
			// 	isSlave = true
		} else if inZone && masterPattern.MatchString(line) {
			// fmt.Println(line)
			inMasters = true
		} else if inZone && inMasters && masterLinePattern.MatchString(line) {
			// fmt.Println(line)
			matches := masterLinePattern.FindStringSubmatch(line)
			currentZone.Masters = append(currentZone.Masters, matches[1])
		} else if inZone && inMasters && strings.Contains(line, "};") {
			inMasters = false
		} else if inZone && filePattern.MatchString(line) {
			matches := filePattern.FindStringSubmatch(line)
			currentZone.File = matches[1]
			zones = append(zones, *currentZone)
			inZone = false
			inMasters = false
			// isSlave = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// return nil, nil
	return zones, nil
}

func axfrzone(zone SlaveZone, master string) AxfrResult {
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
			result := axfrzone(zone, master)
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

	exitCode := 0

	for result := range resultChan {
		if len(result) > 0 {
			// fmt.Printf("Zone")
			for _, i := range result {
				fmt.Printf("Zone: %s %s\n", i.Name, i.Error.Error())
			}
			exitCode = 1
		}
	}
	fmt.Println(exitCode)

	return []AxfrResult{}

}

func CheckMasters(filename string) error {

	zones, err := parseNamedConf(filename)
	if err != nil {
		fmt.Printf("Error parsing %s\n%s\n", filename, err)
		return err
	}

	if len(zones) > 0 {
		result := processZones(zones, 200)

		if len(result) > 1 {
			fmt.Println("error")
		}
	} else {
		fmt.Println("no zones found")
	}

	return nil

}
