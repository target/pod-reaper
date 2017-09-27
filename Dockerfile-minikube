# including this because default minikube drivers do not yet support multistage docker builds
# this will only be used for minikube
FROM scratch
COPY pod-reaper /
CMD ["/pod-reaper"]