package metadata

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	MAGIC_VAL     = "bf2f21a8-80ca-4934-bdd4-950c79c6b5e7"
	MAGIC         = "magic"
	VERSION       = "version"
	LENGTH        = "length"
	CHECKSUM      = "checksum"
	HEADER_LENGTH = 128
)

type Metadata struct {
	Magic string
}

func NewMetadata() *Metadata {
	return &Metadata{Magic: MAGIC_VAL}
}

func (me *Metadata) getMetadataAttr(message string, attr string) (string, string) {
	reg := regexp.MustCompile(fmt.Sprintf("%s: ([\\w\\-]+)", attr))
	result := reg.FindStringSubmatch(message)
	if result == nil {
		return "", ""
	}
	return result[0], result[1]
}

func (me *Metadata) unpackHeader(dev *os.File) (string, string, error) {
	data := make([]byte, HEADER_LENGTH)
	n, err := dev.Read(data)
	if err != nil || n < HEADER_LENGTH {
		return "", "", err
	}

	header := strings.Trim(string(data), "\n")
	_, magic := me.getMetadataAttr(header, MAGIC)
	_, version := me.getMetadataAttr(header, VERSION)
	_, length := me.getMetadataAttr(header, LENGTH)

	if magic != MAGIC_VAL {
		return "", "", errors.New("Metadata Attr Not Found!")
	}

	if !me.checkMetadata(header) {
		return "", "", errors.New("Metadata check sum error!")
	}
	return version, length, nil
}

func (me *Metadata) unpack(pth string, off int) (string, []byte, error) {
	fd, err := os.Open(pth)
	defer fd.Close()
	if err != nil {
		return "", nil, err
	}

	fd.Seek(int64(512*off), 0)
	ver, leng, err := me.unpackHeader(fd)
	if err != nil {
		return "", nil, err
	}

	length, _ := strconv.Atoi(leng)
	data := make([]byte, length)
	_, err = fd.Read(data)
	if err != nil {
		return "", nil, err
	}

	return ver, data, nil
}

func (me *Metadata) checkMetadata(header string) bool {
	reg := regexp.MustCompile(fmt.Sprintf("%s: (.+)", CHECKSUM))
	result := reg.FindStringSubmatch(header)
	if result == nil {
		return false
	}

	header = strings.TrimSuffix(header, result[0])
	header = strings.Trim(header, "\n")

	//fmt.Printf("%q\n", header)
	cal := me.calcChecksum(header)
	_, chk := me.getMetadataAttr(result[0], CHECKSUM)

	//fmt.Printf("%q\n%q\n", cal, chk)
	return cal == chk
}

func (me *Metadata) calcChecksum(header string) string {
	data := []byte(header)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func (me *Metadata) appendChecksum(kv [][2]string) [][2]string {
	ms := me.format(kv)
	chksum := me.calcChecksum(ms)

	kv = append(kv, [2]string{CHECKSUM, chksum})
	return kv
}

func (me *Metadata) format(kv [][2]string) string {
	mc := make([]string, 0)
	for _, v := range kv {
		mc = append(mc, fmt.Sprintf("%s: %v", v[0], v[1]))
	}

	return strings.Join(mc, "\n")
}

func (me *Metadata) packHeader(version string, ms []byte) []byte {
	strLen := strconv.Itoa(len(ms))
	attrs := [][2]string{{MAGIC, MAGIC_VAL}, {VERSION, version}, {LENGTH, strLen}}

	attrs = me.appendChecksum(attrs)
	header := me.format(attrs)
	Nulls := make([]byte, (HEADER_LENGTH - len(header) - 1))

	bs := append([]byte(header), Nulls...)
	bs = append(bs, []byte("\n")...)
	bs = append(bs, ms...)
	return bs
}

func Parse(diskPath string, off int) (*DiskMetadata_1, error) {
	md := NewMetadata()
	_, mc, err := md.unpack(diskPath, off)
	if err != nil {
		return nil, err
	}

	dmd := NewDiskMetadata_1()
	if mc != nil {
		err := dmd.Parse(string(mc))
		if err != nil {
			return nil, err
		}
	}

	return dmd, nil
}

func Update(diskPath string, off int, kv map[string]string) (*DiskMetadata_1, error) {
	dmd, err := Parse(diskPath, off)
	if err != nil {
		return nil, err
	}

	for k, v := range kv {
		if _, ok := dmd.Attrs[k]; ok {
			dmd.Attrs[k] = v
		}
	}

	err = dmd.Save(diskPath, off)
	if err != nil {
		return nil, err
	}

	return dmd, nil
}

func Write(diskPath string, off int, kv map[string]string) (*DiskMetadata_1, error) {
	dmd, err := Parse(diskPath, off)
	if err != nil {
		dmd = NewDiskMetadata_1()
	}

	for k, v := range kv {
		if _, ok := dmd.Attrs[k]; ok {
			dmd.Attrs[k] = v
		}
	}

	err = dmd.Save(diskPath, off)
	if err != nil {
		return nil, err
	}

	return dmd, nil
}
