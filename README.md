## Report GCP Custom Metrics using Go

This application reports custom metrics to GCP Monitoring and is written in Go. To export the metrics, it reads a list of key-value pairs from an input file and creates (or appends the data to) the intended custom metric.


For each pair in the file, the key becomes the name of the custom metric, and ***the value becomes part of the label of the custom metric*** (and not the actual value of the custom metric). Depending on your need, this can be easily changed in `metrics.go`

### Before Starting
1.  Install Golang and Git on your instance

        sudo apt install -y git-all
        wget -c https://go.dev/dl/go1.19.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        
2. Install Google Ops agent

        # install gcp ops agent 
        curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
        sudo bash add-google-cloud-ops-agent-repo.sh --also-install

        # for GPU monitoring
        sudo mkdir -p /opt/google 
        cd /opt/google
        sudo git clone https://github.com/GoogleCloudPlatform/compute-gpu-monitoring.git 
        cd /opt/google/compute-gpu-monitoring/linux
        sudo python3 -m venv venv
        sudo venv/bin/pip install wheel
        sudo venv/bin/pip install -Ur requirements.txt
        sudo cp /opt/google/compute-gpu-monitoring/linux/systemd/google_gpu_monitoring_agent_venv.service /lib/systemd/system
        sudo systemctl daemon-reload
        sudo systemctl --no-reload --now enable /lib/systemd/system/google_gpu_monitoring_agent_venv.service
        cd ~

3. Install and build repo
        
        git clone https://github.com/garg02/custom-metric-reporter-gcp.git
        cd custom-metric-reporter-gcp
        go mod init metrics
        go mod tidy
        go build metrics.go
        chmod +x metrics
        sudo cp metrics /usr/local/bin/
        cd ~

### Add custom metrics and required metadata to file

```
echo "PROJECT_ID=$(gcloud info --format='value(config.project)')" > ~/batch_info.txt
echo "INSTANCE_ID=$(curl http://metadata.google.internal/computeMetadata/v1/instance/id -H Metadata-Flavor:Google)" >> ~/batch_info.txt
echo "ZONE_ID=$(curl http://metadata.google.internal/computeMetadata/v1/instance/zone -H Metadata-Flavor:Google | rev | cut -d/ -f1 | rev)" >> ~/batch_info.txt
        
# in this case the Unix timestamp is the batch_number
echo "cpu_batch_num=$(date +%s)" >> ~/batch_info.txt
echo "gpu_batch_num=$(date +%s)" >> ~/batch_info.txt
```
        
### Add cron job that reports the custom metric every minute
 
```
echo -e "* * * * * bash -c metrics -f ~/batch_info.txt >> metrics_stdout.log 2>> metrics_stderr.log" | crontab -u $USER -
```
       
#### Editing textfile to modify values of custom metrics
```
sed -i "s/cpu_batch_num=[[:digit:]]\+/cpu_batch_num=$(date +%s)/" ~/batch_info.txt
sed -i "s/gpu_batch_num=[[:digit:]]\+/gpu_batch_num=$(date +%s)/" ~/batch_info.txt
```

#### Note: Cron job requires storing key-value pairs in persistant memory as environment variables get reset each time.

### Appendix: 
1. `FinalMonitoring.json` has been included as an example for aligning and joining MQL metrics.
2. `stress.sh` and `minimap.sh` are wrapper scripts that use `metrics` to report the current batch being run
