#!/bin/bash
echo "export CURRTIME=$(date +%s)" > /tmp/curr_time
bash gcp_report_batch/batch_metrics.sh
stress-ng --cpu 4 --io 1 --vm 1 --vm-bytes 3G --timeout 1680s
