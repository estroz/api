# api

Contains the API definitions used by [Operator Lifecycle Manager][olm] (OLM) and Marketplace

## `pkg/validate`: Operator Manifest Verification

`pkg/validate` defines a valid operator [manifest bundle][registry-bundle] by providing functions to verify such bundles. These bundles typically consist of a [ClusterServiceVersion][csv], a [Package Manifest][pkg-man], and one or more [CustomResourceDefinitions][crd].

This library reports errors and warnings for missing mandatory and optional fields, respectively. It also supports validation of `ClusterServiceVersion` YAML manifests for any data types inconsistent with those in OLM's [`ClusterServiceVersion`][csv-type] type.

[olm]:https://github.com/operator-framework/operator-lifecycle-manager
[registry-bundle]:https://github.com/operator-framework/operator-registry/#manifest-format
[csv]:https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md
[pkg-man]:https://github.com/operator-framework/operator-lifecycle-manager#discovery-catalogs-and-automated-upgrades
[crd]:https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/
[csv-type]:https://github.com/operator-framework/operator-lifecycle-manager/blob/master/pkg/api/apis/operators/v1alpha1/clusterserviceversion_types.go#L359:6
