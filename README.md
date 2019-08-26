# api

Contains the API definitions used by [Operator Lifecycle Manager][olm] (OLM) and Marketplace

## `pkg/validation`: Operator Manifest Validation

`pkg/validation` exposes a convenient set of interfaces to validate Kubernetes object manifests, primarily for use in an Operator project.

Additionally this package defines a valid Operator [manifests format][registry-manifests] by exposing a set of `interfaces.Validator`'s to verify them and their constituent manifests. These bundles typically consist of a [Package Manifest][pkg-man] and a set of versioned [bundles][registry-manifests]. Each bundle contains a [ClusterServiceVersion][csv] and one or more [CustomResourceDefinition][crd]'s.

This library reports errors and warnings for missing mandatory and optional fields, respectively. It also supports validation of `ClusterServiceVersion` YAML manifests for any data types inconsistent with those in OLM's [`ClusterServiceVersion`][csv-type] type.

[olm]:https://github.com/operator-framework/operator-lifecycle-manager
[registry-manifests]:https://github.com/operator-framework/operator-registry/#manifest-format
[csv]:https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md
[pkg-man]:https://github.com/operator-framework/operator-lifecycle-manager#discovery-catalogs-and-automated-upgrades
[crd]:https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/
[csv-type]:https://github.com/operator-framework/operator-lifecycle-manager/blob/master/pkg/api/apis/operators/v1alpha1/clusterserviceversion_types.go#L359:6
