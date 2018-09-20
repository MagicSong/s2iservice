#/bin/bash

helm repo update
helm install stable/redis-ha --name=redis-ha --namespace=kube-devops
