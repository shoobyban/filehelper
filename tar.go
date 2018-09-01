package filehelper

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/shoobyban/slog"
)

// WriteTar will append to datafile with filename using buf data
func WriteTar(datafile, filename string, buf []byte) {
	f, err := os.OpenFile(datafile, os.O_RDWR, os.ModePerm)
	if err != nil {
		f, err = os.OpenFile(datafile, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	} else {
		fi, err := os.Stat(datafile)
		if err != nil {
			log.Fatalln(err)
		}
		if fi.Size() > 1024 {
			if _, err = f.Seek(-2<<9, os.SEEK_END); err != nil {
				log.Fatalln(err)
			}
		}
	}
	tw := tar.NewWriter(f)

	hdr := &tar.Header{
		Name:     filename,
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(buf)),
		ModTime:  time.Now(),
	}
	slog.Infof("Writing %s %d", filename, int64(len(buf)))
	if err := tw.WriteHeader(hdr); err != nil {
		slog.Infof("Error writing tar header %s", err.Error())
	}
	if _, err := tw.Write(buf); err != nil {
		slog.Infof("Error writing tar data %s", err.Error())
	}
	if err := tw.Close(); err != nil {
		slog.Infof("Error closing tar %s", err.Error())
	}
	f.Close()
}

// ListTar will return file list from given tar file
func ListTar(filename string) []string {
	var ret []string
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	// defer io.Copy(os.Stdout, tarReader)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeReg: // = regular file
			ret = append(ret, name)
		default:
			ret = append(ret, name)
		}
	}
	return ret
}

// ReadTar reads filename from given tarball and returns content
func ReadTar(tarfile, filename string) interface{} {
	f, err := os.Open(tarfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if header.Name == filename {

			bs, _ := ioutil.ReadAll(tarReader)
			return bs
		}

	}
	return nil
}

// FindInTar looks for search string in tarball, returns list of filenames and matches
func FindInTar(tarfile, search string) map[string]string {
	f, err := os.Open(tarfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	res := map[string]string{}
	tarReader := tar.NewReader(f)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bs, _ := ioutil.ReadAll(tarReader)
		if bytes.Contains(bs, []byte(search)) {
			begining := bytes.Index(bs, []byte(search))
			end := begining + len(search) + 3
			if begining > 3 {
				begining = begining - 3
			}
			if end > len(bs) {
				end = len(bs)
			}
			res[header.Name] = string(bs[begining:end])
		}
	}
	return res
}
