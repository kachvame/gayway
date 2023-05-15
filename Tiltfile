load("ext://configmap", "configmap_create")
load("ext://namespace", "namespace_inject")
load("ext://restart_process", "docker_build_with_restart")

configmap_create("gayway", namespace="kekboard", from_env_file=".env")

docker_build_with_restart(
    "gayway",
    context=".",
    target="development",
    entrypoint="go run ./cmd/gayway",
    live_update=[
        sync(".", "/app"),
    ],
)

k8s_yaml(namespace_inject(kustomize("./manifests"), "kekboard"))

k8s_resource(workload="gayway", resource_deps=["etcd", "kafka"])
