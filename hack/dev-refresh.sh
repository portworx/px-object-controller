make all && make container && make deploy 
kubectl -n kube-system delete pod -l app=px-object-controller
kubectl apply -f deploy/rbac.yaml