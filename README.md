docker build -t prometheus-kafka-adapter:latest .

helm install prometheus-kafka-adapter /root/monitor/prometheus-kafka-adapter --namespace monitor 

remote_write:
- url: "http://prometheus-kafka-adapter:8080/receive"