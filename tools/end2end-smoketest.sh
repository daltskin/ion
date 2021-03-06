#!/bin/sh -e
cd "$(dirname "$0")"
cd ..

BUILD=$1
TERRAFORM=$2
ION_IMAGE_TAG=$3

echo "--------------------------------------------------------"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "WARNING: This script will deploy into your currently selected Azure Subscription, Kubernetes clusters and Docker hub user"
echo "You must have already:"
echo " MUST: Run terraform init in the ./deployment folder"
echo " MUST: have kubectl installed and available in your path"
echo " Must: Be logged into Azure CLI and have the right subscription set as your default"
echo " Must: Be logged into docker cli and have set $DOCKER_USER to your username"
echo "--------------------------------------------------------"

sleep 5

if [ -z "$LOCAL_IP" ]; then
    export LOCAL_IP=localhost
fi

if [ -z "$DOCKER_USER" ]; then
    echo "You must specify a $DOCKER_USER environment variable to which the ion images can be pushed"
    exit 1
fi

if [ "$BUILD" = true ]; then
    echo "--------------------------------------------------------"
    echo "Building source and pushing images"
    echo "--------------------------------------------------------"

    make
    ./build/pushimages.sh
    export ION_IMAGE_TAG=$(cat imagetag.temp)
fi

if [ -z "$ION_IMAGE_TAG" ]; then
    echo "Skipped local container build, existing container image tag must be provided!"
    exit 1
fi

echo "-> Using tag $ION_IMAGE_TAG"

if [ "$TERRAFORM" = true ]; then
    #Refresh the azurecli token
    az group list >> /dev/null

    echo "--------------------------------------------------------"
    echo "Cleaning up k8s, removing all deployments"
    echo "--------------------------------------------------------"

    if [ -f "./kubeconfig.private.yaml" ]
    then
        echo "Kubeconfig found cleaning up cluster."
        export KUBECONFIG=./kubeconfig.private.yaml
        kubectl delete deployments --all || true
        kubectl delete jobs --all || true
        kubectl delete pods --all || true
        kubectl delete secrets --all || true
    else
        echo "Kubeconfig not found, no cluster created skipping cleanup..."
    fi


    echo "--------------------------------------------------------"
    echo "Deploying terraform"
    echo "--------------------------------------------------------"

    cd ./deployment
    if [ ! -f ./vars.private.tfvars ]; then
        echo "vars.private.tfvars not found in deployment file!"
        echo "WARNING.... you'll need to create it some of the fields in ./deployment/vars.private.tfvars without it the terraform deployment will fail"
        return
    fi

    sed -i "s/docker_root.*/docker_root=\"$DOCKER_USER\"/g" vars.private.tfvars
    sed -i "s/docker_user.*/docker_user=\"$ION_IMAGE_TAG\"/g" vars.private.tfvars
    terraform init
    terraform apply -var-file ./vars.private.tfvars -auto-approve
    terraform output kubeconfig > ../kubeconfig.private.yaml

    echo "--------------------------------------------------------"
    echo "Setting kubectl context to new cluster"
    echo "--------------------------------------------------------"
    az aks get-credentials -n $(terraform output cluster_name) -g $(terraform output resource_group_name)
    cd -

    echo "--------------------------------------------------------"
    echo "Wait for the pods to start"
    echo "--------------------------------------------------------"

    sleep 15

    export KUBECONFIG=./kubeconfig.private.yaml
    kubectl get pods || true

else
    echo "--------------------------------------------------------"
    echo "Cleaning up k8s, removing all jobs and pods"
    echo "--------------------------------------------------------"
    kubectl delete jobs --all || true
fi

echo "--------------------------------------------------------"
echo "Forwarding ports for management api and front api"
echo "--------------------------------------------------------"

#Cleanup any leftover listeners
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill || true

kubectl get pods | grep ion-front | awk '{print $1}' | xargs -I % kubectl port-forward % 9001:9001 &
FORWARD_PID1=$!
kubectl get pods | grep ion-management | awk '{print $1}' | xargs -I % kubectl port-forward % 9000:9000 &
FORWARD_PID2=$!


echo "--------------------------------------------------------"
echo "Deploying downloader and transcoder module with tag $ION_IMAGE_TAG"
echo "--------------------------------------------------------"

if [ "$BUILD" = false ]; then
    docker pull "$DOCKER_USER/ion-cli:$ION_IMAGE_TAG"
    docker tag "$DOCKER_USER/ion-cli:$ION_IMAGE_TAG" ion-cli:latest
fi

docker run --network host ion-cli module create -i frontapi.new_link -o file_downloaded -n downloader -m $DOCKER_USER/ion-module-download-file:$ION_IMAGE_TAG -p kubernetes --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG

#docker run --network host -v ${PWD}:/src ion-cli module create -i file_downloaded -o file_transcoded -n transcode -m $DOCKER_USER/ion-module-transcode:$ION_IMAGE_TAG -p kubernetes --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG --config-map-file /src/tools/transcoder.env
docker run --network host -v ${PWD}:/src ion-cli module create -i file_downloaded -o file_transcoded -n transcode -m $DOCKER_USER/ion-module-transcode:$ION_IMAGE_TAG -p azurebatch --handler-image $DOCKER_USER/ion-handler:$ION_IMAGE_TAG --config-map-file /src/tools/transcoder.env

sleep 30


echo "--------------------------------------------------------"
echo "Submitting a video for processing to the frontapi"
echo "--------------------------------------------------------"

curl --header "Content-Type: application/json"   --request POST   --data '{"url": "http://download.blender.org/peach/bigbuckbunny_movies/BigBuckBunny_320x180.mp4"}'   http://localhost:9001/

if [ -x "$(command -v beep)" ]; then
    beep
fi

if [ -x "$(command -v notify-send)" ]; then
    notify-send -u critical ion-end2end "Ion ready for testing"
fi

read -p "Press enter to to stop forwarding ports to management api and front api and exit..." key
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill || true

