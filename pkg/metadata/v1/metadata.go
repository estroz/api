package v1

import (
	"strings"

	"github.com/operator-framework/api/pkg/metadata/annotations"
)

const (
	// MediaTypeKey defines what other annotation keys
	MediaTypeKey annotations.Key = "operators.operatorframework.io.bundle.mediatype.v1"

	// Keys defined by what MediaTypeKey's value is in an annotations file.
	ManifestsKey      annotations.Key = "operators.operatorframework.io.bundle.manifests.v1"
	MetadataKey       annotations.Key = "operators.operatorframework.io.bundle.metadata.v1"
	PackageKey        annotations.Key = "operators.operatorframework.io.bundle.package.v1"
	ChannelsKey       annotations.Key = "operators.operatorframework.io.bundle.channels.v1"
	DefaultChannelKey annotations.Key = "operators.operatorframework.io.bundle.channel.default.v1"
)

type RegistryV1Type annotations.MediaType

func (t RegistryV1Type) String() string                    { return "registry+v1" }
func (t RegistryV1Type) SetAnnotations(a annotations.File) { t.Annotations = a }
func (t RegistryV1Type) GetManifestsDir() string {
	manifestsDir, _ := t.Annotations.GetValue(ManifestsKey)
	return manifestsDir
}
func (t RegistryV1Type) GetMetadataDir() string {
	metadataDir, _ := t.Annotations.GetValue(MetadataKey)
	return metadataDir
}
func (t RegistryV1Type) GetPackage() string {
	pkg, _ := t.Annotations.GetValue(PackageKey)
	return pkg
}
func (t RegistryV1Type) GetChannels() []string {
	channelsStr, _ := t.Annotations.GetValue(ChannelsKey)
	return strings.Split(channelsStr, ",")
}
func (t RegistryV1Type) GetDefaultChannel() string {
	defaultChannel, _ := t.Annotations.GetValue(DefaultChannelKey)
	return defaultChannel
}
func (t RegistryV1Type) GetRequiredKeys() []annotations.Key {
	return []annotations.Key{
		ManifestsKey,
		MetadataKey,
		PackageKey,
		ChannelsKey,
		DefaultChannelKey,
	}
}

type PlainType annotations.MediaType

func (t PlainType) String() string                    { return "plain" }
func (t PlainType) SetAnnotations(a annotations.File) { t.Annotations = a }
func (t PlainType) GetManifestsDir() string {
	manifestsDir, _ := t.Annotations.GetValue(ManifestsKey)
	return manifestsDir
}
func (t PlainType) GetMetadataDir() string {
	metadataDir, _ := t.Annotations.GetValue(MetadataKey)
	return metadataDir
}
func (t PlainType) GetPackage() string {
	pkg, _ := t.Annotations.GetValue(PackageKey)
	return pkg
}
func (t PlainType) GetChannels() []string {
	channelsStr, _ := t.Annotations.GetValue(ChannelsKey)
	return strings.Split(channelsStr, ",")
}
func (t PlainType) GetDefaultChannel() string {
	defaultChannel, _ := t.Annotations.GetValue(DefaultChannelKey)
	return defaultChannel
}
func (t PlainType) GetRequiredKeys() []annotations.Key {
	return []annotations.Key{
		ManifestsKey,
		MetadataKey,
		PackageKey,
		ChannelsKey,
		DefaultChannelKey,
	}
}

type HelmType annotations.MediaType

func (t HelmType) String() string                    { return "helm" }
func (t HelmType) SetAnnotations(a annotations.File) { t.Annotations = a }
func (t HelmType) GetManifestsDir() string {
	manifestsDir, _ := t.Annotations.GetValue(ManifestsKey)
	return manifestsDir
}
func (t HelmType) GetMetadataDir() string {
	metadataDir, _ := t.Annotations.GetValue(MetadataKey)
	return metadataDir
}
func (t HelmType) GetPackage() string {
	pkg, _ := t.Annotations.GetValue(PackageKey)
	return pkg
}
func (t HelmType) GetChannels() []string {
	channelsStr, _ := t.Annotations.GetValue(ChannelsKey)
	return strings.Split(channelsStr, ",")
}
func (t HelmType) GetDefaultChannel() string {
	defaultChannel, _ := t.Annotations.GetValue(DefaultChannelKey)
	return defaultChannel
}
func (t HelmType) GetRequiredKeys() []annotations.Key {
	return []annotations.Key{
		ManifestsKey,
		MetadataKey,
		PackageKey,
		ChannelsKey,
		DefaultChannelKey,
	}
}
