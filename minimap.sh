#!/bin/bash
echo "export CURRTIME=$(date +%s)" > /tmp/curr_time
bash gcp_report_batch/batch_metrics.sh
time ./minimap2-2.24_x64-linux/minimap2 -ax map-ont --MD -t 7 chr17.fa n_guppy.fq | samtools view -F 0x904 -hb -@3 | samtools sort -@3 | samtools view -hb -@3 > all.bam
time samtools index -@7 all.bam
rm samtools.*
