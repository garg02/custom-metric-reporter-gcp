package main

import (
        "context"
        "fmt"
        "log"
        "time"
        "flag"
        "bufio"
 	"io"
 	"os"
 	"strings"

        monitoring "cloud.google.com/go/monitoring/apiv3/v2"
        googlepb "github.com/golang/protobuf/ptypes/timestamp"
        metricpb "google.golang.org/genproto/googleapis/api/metric"
        monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
        monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type Config map[string]string

func ReadConfig(filename string) (Config, error) {
        // init with empty strings
        config := Config{
 		"INSTANCE_ID":       "",
 		"PROJECT_ID":       "",
 		"ZONE_ID":       "",
        }
 	if len(filename) == 0 {
 		return config, nil
 	}
 	file, err := os.Open(filename)
 	if err != nil {
 		return nil, err
 	}
 	defer file.Close()
 	
 	reader := bufio.NewReader(file)
 	
 	for {
                line, err := reader.ReadString('\n')
 		
 		// check if the line has = sign and process the line. 

 		if equal := strings.Index(line, "="); equal >= 0 {
 			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
 				value := ""
 				if len(line) > equal {
 					value = strings.TrimSpace(line[equal+1:])
 				}
                             // assign the config map
 				config[key] = value
 			}
 		}
 		if err == io.EOF {
 			break
 		}
 		if err != nil {
 			return nil, err
 		}
 	}
 	return config, nil
 }

func SubmitMetric(projID string, instID string, zoneID string, 
                  batchLabel string, batchVal string) (error){
// Creates a client.
        ctx := context.Background()
        client, err := monitoring.NewMetricClient(ctx)
        if err != nil {
                log.Fatalf("Failed to create client: %v", err)
                return err
        }

        // Prepares an individual data point
        dataPoint := &monitoringpb.Point{
                Interval: &monitoringpb.TimeInterval{
                        EndTime: &googlepb.Timestamp{
                                Seconds: time.Now().Unix(),
                        },
                },
                Value: &monitoringpb.TypedValue{
                        Value: &monitoringpb.TypedValue_Int64Value{
                                Int64Value: 0,
                        },
                },
        }

        // Writes time series data.
        if err := client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
                Name: "projects/" + projID,
                TimeSeries: []*monitoringpb.TimeSeries{
                        {
                                Metric: &metricpb.Metric{
                                        Type: "custom.googleapis.com/" + batchLabel,
                                        Labels: map[string]string{
                                                "batch_num": batchVal,
                                        },
                                },
                                Resource: &monitoredrespb.MonitoredResource{
                                        Type: "gce_instance",
                                        Labels: map[string]string{
                                                "instance_id": instID,
                                                        "zone": zoneID,
                                        },
                                },
                                Points: []*monitoringpb.Point{
                                        dataPoint,
                                },
                        },
                },
        }); err != nil {
                log.Fatalf("Failed to write time series data: %v", err)
                return err
        }

        // Closes the client and flushes the data to Stackdriver.
        if err := client.Close(); err != nil {
                log.Fatalf("Failed to close client: %v", err)
                return err
        }
        
        return nil
}


func main() {
        if len(os.Args) < 1 {
                log.Fatalf("need at least one argument")
        }
        file_ptr := flag.String("f", "", "file with metadata and batch label(s)/value(s)")
        flag.Parse()

                if len(*file_ptr) == 0 {
                        log.Fatalf("file argument may not be empty; ensure use of -f flag")
                }
        
        config, err := ReadConfig(*file_ptr)
        
        if err != nil {
 		fmt.Println(err)
 	}

        if len(config) <= 3 {
                log.Fatalf("Must have at least one metric to report")
        }
        if len(config["INSTANCE_ID"]) == 0 || len(config["PROJECT_ID"]) == 0 || len(config["ZONE_ID"]) == 0 {
                log.Fatalf("INSTANCE_ID, PROJECT_ID, and ZONE_ID values must be present in the file")
        }
        
        instID := config["INSTANCE_ID"]
        delete(config, "INSTANCE_ID")
        projID := config["PROJECT_ID"]
        delete(config, "PROJECT_ID")
        zoneID := config["ZONE_ID"]
        delete(config, "ZONE_ID")
        
        for batchLabel, batchVal := range config {
                for i := 1; i < 5; i++ {
                        err = SubmitMetric(projID, instID, zoneID, batchLabel, batchVal)
                        if err == nil {break}
                        time.Sleep(5 * time.Second)
                }
        }       
}
