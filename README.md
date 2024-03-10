docker build -t xubiao05/prometheus-kafka-adapter:latest .

docker-compose -f docker-compose-adapter.yml up -d

helm install prometheus-kafka-adapter /root/monitor/prometheus-kafka-adapter --namespace monitor 

remote_write:
- url: "http://prometheus-kafka-adapter:8080/receive"