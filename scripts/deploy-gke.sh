#!/bin/bash
set -e PROJECT_ID=$(gcloud config 
get-value project) 
REGION="us-central1" 
CLUSTER_NAME="anatel-cluster"
# Criar cluster se não existir
if ! gcloud container clusters 
describe $CLUSTER_NAME 
--region=$REGION &>/dev/null; then
  gcloud container clusters create 
  $CLUSTER_NAME --region=$REGION 
  --machine-type=e2-standard-4 
  --num-nodes=3
fi gcloud container clusters 
get-credentials $CLUSTER_NAME 
--region=$REGION
# Aplicar manifests
kubectl apply -f 
deployments/k8s/namespace.yaml 
kubectl apply -f 
deployments/k8s/deployment.yaml 
kubectl apply -f 
deployments/k8s/service.yaml 
kubectl apply -f 
deployments/k8s/ingress.yaml echo 
"Deploy concluído. Aguardando 
LoadBalancer IP..."
kubectl -n anatel-realtime get svc gateway-svc -w
