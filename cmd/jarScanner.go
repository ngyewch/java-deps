package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magiconair/properties"
)

func ScanJar(path string) (*ArtifactEntry, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	sha256Bytes, err := calculateSha256ForFile(path)
	if err != nil {
		return nil, err
	}

	artifactInfo, err := scanJarForArtifactInfo(path)
	if err != nil {
		return nil, err
	}

	if artifactInfo == nil {
		return nil, nil
	}

	entry := &ArtifactEntry{
		Filename:     stat.Name(),
		Size:         stat.Size(),
		Sha256:       hex.EncodeToString(sha256Bytes),
		ArtifactInfo: *artifactInfo,
	}

	return entry, nil
}

func calculateSha256ForFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func scanJarForArtifactInfo(path string) (*ArtifactInfo, error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer func(zipReader *zip.ReadCloser) {
		_ = zipReader.Close()
	}(zipReader)

	name := strings.TrimSuffix(filepath.Base(path), ".jar")
	regex, err := regexp.Compile(`-\d+(\.\d+(\.\d+)?)?`)
	if err != nil {
		return nil, err
	}
	p := regex.FindStringIndex(name)
	var fileArtifactId string
	var fileVersion string
	if p != nil {
		fileArtifactId = name[:p[0]]
		fileVersion = name[p[0]+1:]
	}

	var artifactInfo *ArtifactInfo
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "META-INF/maven/") {
			artifactInfo1, err := readArtifactInfoFromZipFile(f)
			if err != nil {
				return nil, err
			}
			if artifactInfo1 != nil {
				if artifactInfo1.ArtifactId == fileArtifactId {
					artifactInfo = artifactInfo1
					break
				}
			}
		}
	}
	if artifactInfo == nil {
		artifactInfo = &ArtifactInfo{
			ArtifactId: fileArtifactId,
			Version:    fileVersion,
		}
	}

	return artifactInfo, nil
}

func readArtifactInfoFromZipFile(zipFile *zip.File) (*ArtifactInfo, error) {
	if strings.HasPrefix(zipFile.Name, "META-INF/maven/") {
		if strings.HasSuffix(zipFile.Name, "/pom.properties") {
			r, err := zipFile.Open()
			if err != nil {
				return nil, err
			}
			defer func(r io.ReadCloser) {
				_ = r.Close()
			}(r)
			return readArtifactInfoFromPomProperties(r)
		} else if strings.HasSuffix(zipFile.Name, "/pom.xml") {
			r, err := zipFile.Open()
			if err != nil {
				return nil, err
			}
			defer func(r io.ReadCloser) {
				_ = r.Close()
			}(r)
			return readArtifactInfoFromPomXml(r)
		}
	}
	return nil, nil
}

type PropsArtifactInfo struct {
	GroupId    string `properties:"groupId"`
	ArtifactId string `properties:"artifactId"`
	Version    string `properties:"version"`
	Classifier string `properties:"classifier,default="`
}

func readArtifactInfoFromPomProperties(r io.Reader) (*ArtifactInfo, error) {
	props, err := properties.LoadReader(r, properties.UTF8)
	if err != nil {
		return nil, err
	}

	var propsArtifactInfo PropsArtifactInfo
	err = props.Decode(&propsArtifactInfo)
	if err != nil {
		return nil, err
	}

	return &ArtifactInfo{
		GroupId:    propsArtifactInfo.GroupId,
		ArtifactId: propsArtifactInfo.ArtifactId,
		Version:    propsArtifactInfo.Version,
		Classifier: propsArtifactInfo.Classifier,
	}, nil
}

type XmlParentArtifactInfo struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type XmlArtifactInfo struct {
	Parent     *XmlParentArtifactInfo `xml:"parent"`
	GroupId    string                 `xml:"groupId"`
	ArtifactId string                 `xml:"artifactId"`
	Version    string                 `xml:"version"`
	Classifier string                 `xml:"classifier"`
}

func readArtifactInfoFromPomXml(r io.Reader) (*ArtifactInfo, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var xmlArtifactInfo XmlArtifactInfo
	err = xml.Unmarshal(b, &xmlArtifactInfo)
	if err != nil {
		return nil, err
	}

	var artifactInfo ArtifactInfo
	if xmlArtifactInfo.Parent != nil {
		artifactInfo.GroupId = xmlArtifactInfo.Parent.GroupId
		artifactInfo.ArtifactId = xmlArtifactInfo.Parent.ArtifactId
		artifactInfo.Version = xmlArtifactInfo.Parent.Version
	}
	if xmlArtifactInfo.GroupId != "" {
		artifactInfo.GroupId = xmlArtifactInfo.GroupId
	}
	if xmlArtifactInfo.ArtifactId != "" {
		artifactInfo.ArtifactId = xmlArtifactInfo.ArtifactId
	}
	if xmlArtifactInfo.Version != "" {
		artifactInfo.Version = xmlArtifactInfo.Version
	}
	artifactInfo.Classifier = xmlArtifactInfo.Classifier
	return &artifactInfo, nil
}
