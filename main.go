package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-lambda-go/events"
)

var (
	flOwner              string
	flLogGroup           string
	flLogStream          string
	flSubscriptionFilter string
	flLogEvents          string
)

var logEventsFileData []byte

var logEventsList []events.CloudwatchLogsLogEvent

const (
	// StatusInvalidArguments indicates specified invalid arguments.
	StatusInvalidArguments = 1
	// StatusTemplate CreationFailre indicates program cannot create original log data.
	StatusTemplateCreationFailure = 2
	// StatusGZIPCompressFailure indicate specified gzip compression
	StatusGZIPCompressFailure = 3
	// StatusEncodingResultFailure indicate specified event json Marshal
	StatusEncodingResultFailure = 4
)

func init() {
	log.SetFlags(0)
	flag.CommandLine.StringVar(&flOwner, "owner", "123456789123", "log owner accountID (default: \"123456789123\")")
	flag.CommandLine.StringVar(&flLogGroup, "log-group", "testLogGroup", "log group name (default: \"testLogGroup\")")
	flag.CommandLine.StringVar(&flLogStream, "log-stream", "testLogStream", "log stream name (default: \"testLogStream\")")
	flag.CommandLine.StringVar(&flSubscriptionFilter, "subscription-filter", "testFilter", "subscription filter name (default: \"testFilter\")")
	flag.CommandLine.StringVar(&flLogEvents, "log-events", "", "log events (default: equal to \"$sam local generate-event cloudwatch logs\")")

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		log.Println(err)
		os.Exit(StatusInvalidArguments)
	}

	var logEventRawData []byte

	if len(flLogEvents) == 0 {
		logEventRawData = []byte(`[{"id":"eventId1","message":"[ERROR] First test message","timestamp":1440442987000},{"id":"eventId2","message":"[ERROR] Second test message","timestamp":1440442987001}]`)
	} else {
		if strings.HasPrefix(flLogEvents, "file://") {
			path := strings.Replace(flLogEvents, "file://", "", 1)
			path, err := filepath.Abs(path)
			if err != nil {
				os.Exit(StatusInvalidArguments)
			}
			f, err := os.Open(path)
			if err != nil {
				os.Exit(StatusInvalidArguments)
			}
			logEventRawData, err = io.ReadAll(f)
			if err != nil {
				os.Exit(StatusInvalidArguments)
			}
		} else {
			logEventRawData = []byte(flLogEvents)
		}
	}

	if err := json.Unmarshal(logEventRawData, &logEventsList); err != nil {
		os.Exit(StatusInvalidArguments)
	}
}

//go:embed template.json
var templateJSON string

func main() {
	retcode := 0
	defer func() { os.Exit(retcode) }()

	tpl, err := template.New("").Parse(templateJSON)
	if err != nil {
		log.Printf("failed to parse template. error=%v", err)
		retcode = StatusTemplateCreationFailure
		return
	}

	v := struct {
		Owner              string
		LogGroup           string
		LogStream          string
		SubscriptionFilter string
		LogEvents          []events.CloudwatchLogsLogEvent
		LastItemIndex      int
	}{
		Owner:              flOwner,
		LogGroup:           flLogGroup,
		LogStream:          flLogStream,
		SubscriptionFilter: flSubscriptionFilter,
		LogEvents:          logEventsList,
		LastItemIndex:      len(logEventsList) - 1,
	}
	var cloudwatchLogsData bytes.Buffer
	if err := tpl.Execute(&cloudwatchLogsData, v); err != nil {
		log.Printf("failed to execute template. error=%v", err)
		retcode = StatusTemplateCreationFailure
		return
	}

	var flatCloudwatchLogsData bytes.Buffer
	if err := json.Compact(&flatCloudwatchLogsData, cloudwatchLogsData.Bytes()); err != nil {
		log.Printf("failed to compact json. error=%v", err)
		retcode = StatusTemplateCreationFailure
		return
	}

	var gzipData bytes.Buffer
	gw, err := gzip.NewWriterLevel(&gzipData, gzip.BestCompression)
	if err != nil {
		log.Printf("invalid gzip compression level. error=%v", err)
		retcode = StatusGZIPCompressFailure
		return
	}
	if _, err := gw.Write(flatCloudwatchLogsData.Bytes()); err != nil {
		log.Printf("fail to gzip compression. error=%v", err)
		retcode = StatusGZIPCompressFailure
		return
	}
	if err := gw.Close(); err != nil {
		log.Printf("fail to flash gzip compression. error=%v", err)
		retcode = StatusGZIPCompressFailure
		return
	}

	result := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: base64.StdEncoding.EncodeToString(gzipData.Bytes()),
		},
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("fail to Marshal AWSLogEvent. error=%v", err)
		retcode = StatusEncodingResultFailure
		return
	}
	var resultFlatJSON bytes.Buffer
	if err := json.Indent(&resultFlatJSON, resultJSON, "", "\t"); err != nil {
		log.Printf("fail to Indent AWSLogEvent. error=%v", err)
		retcode = StatusEncodingResultFailure
		return
	}
	fmt.Println(resultFlatJSON.String())
}
