variable "TAG" {
  default = "latest"
}

variable "IMAGE_REGISTRY" {
    default = "local"
}

target "daemon-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-daemon:${TAG}"]
}

target "daemoninit-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-daemon-init:${TAG}"]
}

target "controller-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-controller:${TAG}"]
}

target "cnimanager-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-cnimanager:${TAG}"]
}

target "cni-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-cni:${TAG}"]
}

target "cni-ipam-tags" {
    tags = ["${IMAGE_REGISTRY}/kube-egress-gateway-cni-ipam:${TAG}"]
}