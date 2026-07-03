package main

import (
	"fmt"
	"strings"
)

type ArtifactEntry struct {
	Filename     string       `json:"filename"`
	Size         int64        `json:"size"`
	Sha256       string       `json:"sha256"`
	ArtifactInfo ArtifactInfo `json:"artifactInfo,omitempty"`
}

type ArtifactInfo struct {
	GroupId    string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version    string `json:"version"`
	Classifier string `json:"classifier,omitempty"`
}

func (artifactInfo ArtifactInfo) String() string {
	if artifactInfo.Classifier != "" {
		return fmt.Sprintf("%s:%s:%s:%s", artifactInfo.GroupId, artifactInfo.ArtifactId, artifactInfo.Version, artifactInfo.Classifier)
	}
	return fmt.Sprintf("%s:%s:%s", artifactInfo.GroupId, artifactInfo.ArtifactId, artifactInfo.Version)
}

func (artifactInfo ArtifactInfo) Compare(otherInfo ArtifactInfo) int {
	result := strings.Compare(artifactInfo.GroupId, otherInfo.GroupId)
	if result != 0 {
		return result
	}
	result = strings.Compare(artifactInfo.ArtifactId, otherInfo.ArtifactId)
	if result != 0 {
		return result
	}
	result = strings.Compare(artifactInfo.Version, otherInfo.Version)
	if result != 0 {
		return result
	}
	result = strings.Compare(artifactInfo.Classifier, otherInfo.Classifier)
	if result != 0 {
		return result
	}
	return 0
}
