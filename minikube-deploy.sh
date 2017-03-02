# setup the minikube environment (don't need to run every time)
# minikube start
# eval $(minikube docker-env)

# build the local binary
rm pod-reaper
go fmt
go test
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

# build the docker container
docker rmi pod-reaper
docker build --tag=pod-reaper .

# delete any lingering deployment, create a new deployment
kubectl --context=minikube delete --filename deployment.yml
kubectl --context=minikube apply --filename deployment.yml
