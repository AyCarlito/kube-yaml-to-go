# kube-yaml-to-go

kube-yaml-to-go generates Go source code from Kubernetes YAML.

## Prerequisites

Install the following:

1. [Go](https://go.dev/dl/)
2. [Docker](https://docs.docker.com/engine/install/)

## Install

- Build the binary:

```shell
git clone https://github.com/AyCarlito/kube-yaml-to-go.git
cd kube-yaml-to-go
make build
```

- Move the binary to the desired location e.g:

```shell
mv bin/kube-yaml-to-go /usr/local/bin/
```

- Alternatively, the application is containerised through the `Dockerfile` at the root of the repository, which can
be built and run through:

```shell
make docker-build docker-run
```

## Usage

- `kube-yaml-to-go` is a [Cobra](https://github.com/spf13/cobra) CLI application built on a structure of commands,
arguments & flags:

```shell
./bin/kube-yaml-to-go --help
Generate Go source code from Kubernetes YAML.

Usage:
  kube-yaml-to-go [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generates Go source code.
  help        Help about any command

Flags:
  -h, --help            help for kube-yaml-to-go
      --input string    Path to input file. Reads from stdin if unset.
      --output string   Path to output file. Writes to stdout if unset.
      --verbose         Enables verbose output. Generates a full source file.

Use "kube-yaml-to-go [command] --help" for more information about a command.
```

## Examples

- Read the [WordPress MySQL Application](https://kubernetes.io/docs/tutorials/stateful-application/mysql-wordpress-persistent-volume/)
from file and generate equivalent source:

```shell
./bin/kube-yaml-to-go generate --input docs/examples/wordpress.yaml
```

```go
&corev1.Service{
        TypeMeta: metav1.TypeMeta{
                Kind:       "Service",
                APIVersion: "v1",
        },
        ObjectMeta: metav1.ObjectMeta{
                Name: "wordpress-mysql",
                Labels: map[string]string{
                        "app": "wordpress",
                },
        },
        Spec: corev1.ServiceSpec{
                Ports: []corev1.ServicePort{
                        corev1.ServicePort{
                                Port: 3306,
                        },
                },
                Selector: map[string]string{
                        "app":  "wordpress",
                        "tier": "mysql",
                },
                ClusterIP: "None",
        },
}
&corev1.PersistentVolumeClaim{
        TypeMeta: metav1.TypeMeta{
                Kind:       "PersistentVolumeClaim",
                APIVersion: "v1",
        },
        ObjectMeta: metav1.ObjectMeta{
                Name: "mysql-pv-claim",
                Labels: map[string]string{
                        "app": "wordpress",
                },
        },
        Spec: corev1.PersistentVolumeClaimSpec{
                AccessModes: []corev1.PersistentVolumeAccessMode{
                        "ReadWriteOnce",
                },
                Resources: corev1.VolumeResourceRequirements{
                        Requests: corev1.ResourceList{
                                "storage": resource.Quantity{
                                        Format: "BinarySI",
                                },
                        },
                },
        },
}
&appsv1.Deployment{
        TypeMeta: metav1.TypeMeta{
                Kind:       "Deployment",
                APIVersion: "apps/v1",
        },
        ObjectMeta: metav1.ObjectMeta{
                Name: "wordpress-mysql",
                Labels: map[string]string{
                        "app": "wordpress",
                },
        },
        Spec: appsv1.DeploymentSpec{
                Selector: &metav1.LabelSelector{
                        MatchLabels: map[string]string{
                                "app":  "wordpress",
                                "tier": "mysql",
                        },
                },
                Template: corev1.PodTemplateSpec{
                        ObjectMeta: metav1.ObjectMeta{
                                Labels: map[string]string{
                                        "app":  "wordpress",
                                        "tier": "mysql",
                                },
                        },
                        Spec: corev1.PodSpec{
                                Volumes: []corev1.Volume{
                                        corev1.Volume{
                                                Name: "mysql-persistent-storage",
                                                VolumeSource: corev1.VolumeSource{
                                                        PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
                                                                ClaimName: "mysql-pv-claim",
                                                        },
                                                },
                                        },
                                },
                                Containers: []corev1.Container{
                                        corev1.Container{
                                                Name:  "mysql",
                                                Image: "mysql:8.0",
                                                Ports: []corev1.ContainerPort{
                                                        corev1.ContainerPort{
                                                                Name:          "mysql",
                                                                ContainerPort: 3306,
                                                        },
                                                },
                                                Env: []corev1.EnvVar{
                                                        corev1.EnvVar{
                                                                Name: "MYSQL_ROOT_PASSWORD",
                                                                ValueFrom: &corev1.EnvVarSource{
                                                                        SecretKeyRef: &corev1.SecretKeySelector{
                                                                                LocalObjectReference: corev1.LocalObjectReference{
                                                                                        Name: "mysql-pass",
                                                                                },
                                                                                Key: "password",
                                                                        },
                                                                },
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "MYSQL_DATABASE",
                                                                Value: "wordpress",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "MYSQL_USER",
                                                                Value: "wordpress",
                                                        },
                                                        corev1.EnvVar{
                                                                Name: "MYSQL_PASSWORD",
                                                                ValueFrom: &corev1.EnvVarSource{
                                                                        SecretKeyRef: &corev1.SecretKeySelector{
                                                                                LocalObjectReference: corev1.LocalObjectReference{
                                                                                        Name: "mysql-pass",
                                                                                },
                                                                                Key: "password",
                                                                        },
                                                                },
                                                        },
                                                },
                                                VolumeMounts: []corev1.VolumeMount{
                                                        corev1.VolumeMount{
                                                                Name:      "mysql-persistent-storage",
                                                                MountPath: "/var/lib/mysql",
                                                        },
                                                },
                                        },
                                },
                        },
                },
                Strategy: appsv1.DeploymentStrategy{
                        Type: "Recreate",
                },
        },
}
```

- Read the [Cassandra StatefulSet](https://kubernetes.io/docs/tutorials/stateful-application/cassandra/)
from stdin and generate a verbose source file:

```shell
./bin/kube-yaml-to-go generate --verbose < docs/examples/cassandra-statefulset.yaml
```

```go
package main

import (
        ""
        appsv1 "k8s.io/api/apps/v1"
        corev1 "k8s.io/api/core/v1"
        storagev1 "k8s.io/api/storage/v1"
        "k8s.io/apimachinery/pkg/api/resource"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var StatefulSetDocumentIndex0 = &appsv1.StatefulSet{
        TypeMeta: metav1.TypeMeta{
                Kind:       "StatefulSet",
                APIVersion: "apps/v1",
        },
        ObjectMeta: metav1.ObjectMeta{
                Name: "cassandra",
                Labels: map[string]string{
                        "app": "cassandra",
                },
        },
        Spec: appsv1.StatefulSetSpec{
                Replicas: &[]int32{3}[0],
                Selector: &metav1.LabelSelector{
                        MatchLabels: map[string]string{
                                "app": "cassandra",
                        },
                },
                Template: corev1.PodTemplateSpec{
                        ObjectMeta: metav1.ObjectMeta{
                                Labels: map[string]string{
                                        "app": "cassandra",
                                },
                        },
                        Spec: corev1.PodSpec{
                                Containers: []corev1.Container{
                                        corev1.Container{
                                                Name:  "cassandra",
                                                Image: "gcr.io/google-samples/cassandra:v13",
                                                Ports: []corev1.ContainerPort{
                                                        corev1.ContainerPort{
                                                                Name:          "intra-node",
                                                                ContainerPort: 7000,
                                                        },
                                                        corev1.ContainerPort{
                                                                Name:          "tls-intra-node",
                                                                ContainerPort: 7001,
                                                        },
                                                        corev1.ContainerPort{
                                                                Name:          "jmx",
                                                                ContainerPort: 7199,
                                                        },
                                                        corev1.ContainerPort{
                                                                Name:          "cql",
                                                                ContainerPort: 9042,
                                                        },
                                                },
                                                Env: []corev1.EnvVar{
                                                        corev1.EnvVar{
                                                                Name:  "MAX_HEAP_SIZE",
                                                                Value: "512M",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "HEAP_NEWSIZE",
                                                                Value: "100M",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "CASSANDRA_SEEDS",
                                                                Value: "cassandra-0.cassandra.default.svc.cluster.local",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "CASSANDRA_CLUSTER_NAME",
                                                                Value: "K8Demo",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "CASSANDRA_DC",
                                                                Value: "DC1-K8Demo",
                                                        },
                                                        corev1.EnvVar{
                                                                Name:  "CASSANDRA_RACK",
                                                                Value: "Rack1-K8Demo",
                                                        },
                                                        corev1.EnvVar{
                                                                Name: "POD_IP",
                                                                ValueFrom: &corev1.EnvVarSource{
                                                                        FieldRef: &corev1.ObjectFieldSelector{
                                                                                FieldPath: "status.podIP",
                                                                        },
                                                                },
                                                        },
                                                },
                                                Resources: corev1.ResourceRequirements{
                                                        Limits: corev1.ResourceList{
                                                                "memory": resource.Quantity{
                                                                        Format: "BinarySI",
                                                                },
                                                                "cpu": resource.Quantity{
                                                                        Format: "DecimalSI",
                                                                },
                                                        },
                                                        Requests: corev1.ResourceList{
                                                                "cpu": resource.Quantity{
                                                                        Format: "DecimalSI",
                                                                },
                                                                "memory": resource.Quantity{
                                                                        Format: "BinarySI",
                                                                },
                                                        },
                                                },
                                                VolumeMounts: []corev1.VolumeMount{
                                                        corev1.VolumeMount{
                                                                Name:      "cassandra-data",
                                                                MountPath: "/cassandra_data",
                                                        },
                                                },
                                                ReadinessProbe: &corev1.Probe{
                                                        ProbeHandler: corev1.ProbeHandler{
                                                                Exec: &corev1.ExecAction{
                                                                        Command: []string{
                                                                                "/bin/bash",
                                                                                "-c",
                                                                                "/ready-probe.sh",
                                                                        },
                                                                },
                                                        },
                                                        InitialDelaySeconds: 15,
                                                        TimeoutSeconds:      5,
                                                },
                                                Lifecycle: &corev1.Lifecycle{
                                                        PreStop: &corev1.LifecycleHandler{
                                                                Exec: &corev1.ExecAction{
                                                                        Command: []string{
                                                                                "/bin/sh",
                                                                                "-c",
                                                                                "nodetool drain",
                                                                        },
                                                                },
                                                        },
                                                },
                                                ImagePullPolicy: "Always",
                                                SecurityContext: &corev1.SecurityContext{
                                                        Capabilities: &corev1.Capabilities{
                                                                Add: []corev1.Capability{
                                                                        "IPC_LOCK",
                                                                },
                                                        },
                                                },
                                        },
                                },
                                TerminationGracePeriodSeconds: &[]int64{500}[0],
                        },
                },
                VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
                        corev1.PersistentVolumeClaim{
                                ObjectMeta: metav1.ObjectMeta{
                                        Name: "cassandra-data",
                                },
                                Spec: corev1.PersistentVolumeClaimSpec{
                                        AccessModes: []corev1.PersistentVolumeAccessMode{
                                                "ReadWriteOnce",
                                        },
                                        Resources: corev1.VolumeResourceRequirements{
                                                Requests: corev1.ResourceList{
                                                        "storage": resource.Quantity{
                                                                Format: "BinarySI",
                                                        },
                                                },
                                        },
                                        StorageClassName: &[]string{"fast"}[0],
                                },
                        },
                },
                ServiceName: "cassandra",
        },
}
var StorageClassDocumentIndex1 = &storagev1.StorageClass{
        TypeMeta: metav1.TypeMeta{
                Kind:       "StorageClass",
                APIVersion: "storage.k8s.io/v1",
        },
        ObjectMeta: metav1.ObjectMeta{
                Name: "fast",
        },
        Provisioner: "k8s.io/minikube-hostpath",
        Parameters: map[string]string{
                "type": "pd-ssd",
        },
}
```
