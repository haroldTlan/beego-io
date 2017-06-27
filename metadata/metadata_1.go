package metadata

import (
	"errors"
	"os"
	"speedio/util"
)

const (
	VERSION_VAL = "disk1.0"
)

var attrsSlice []string

type DiskMetadata_1 struct {
	Metadata
	Version string
	Attrs   map[string]string
}

func NewDiskMetadata_1() *DiskMetadata_1 {
	attrs := map[string]string{
		"host_uuid": util.HostUuid(),
		//"host_uuid": "",
		"disk_uuid": "",
		"raid_uuid": "",
	}
	attrsSlice = []string{"host_uuid", "disk_uuid", "raid_uuid"}

	return &DiskMetadata_1{
		Metadata: Metadata{Magic: MAGIC_VAL},
		Version:  VERSION_VAL,
		Attrs:    attrs,
	}
}

func (dm *DiskMetadata_1) Parse(message string) error {
	if !dm.checkMetadata(message) {
		return errors.New("Metadata check sum error!")
	}

	for _, k := range attrsSlice {
		_, v := dm.getMetadataAttr(message, k)
		dm.Attrs[k] = v
	}

	return nil
}

func (dm *DiskMetadata_1) Save(devPath string, off int) error {
	dev, err := os.OpenFile(devPath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer dev.Close()

	dev.Seek(int64(512*off), 0)
	data := dm.Bytes()
	dev.Write(data)
	err = dev.Sync()

	return err
}

func (dm *DiskMetadata_1) Bytes() []byte {
	var cnt int
	mc := make([][2]string, 0)
	for _, k := range attrsSlice {
		if dm.Attrs[k] != "" {
			mc = append(mc, [2]string{k, dm.Attrs[k]})
			cnt++
		}
	}

	if cnt > 0 {
		mc = dm.appendChecksum(mc)
	}
	ms := dm.format(mc)

	return dm.packHeader(dm.Version, []byte(ms))
}
