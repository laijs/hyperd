package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	GITCOMMIT string = "0"
	VERSION   string = "0.3.0"

	IAMSTATIC string = "true"
	INITSHA1  string = ""
	INITPATH  string = ""

	HYPER_ROOT   string
	HYPER_FILE   string
	HYPER_DAEMON interface{}
)

const (
	APIVERSION = "1.17"
)

func MatchesContentType(contentType, expectedType string) bool {
	mimetype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		fmt.Printf("Error parsing media type: %s error: %v", contentType, err)
	}
	return err == nil && mimetype == expectedType
}

func DownloadFile(uri, target string) error {
	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0666)
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Range", "bytes="+strconv.FormatInt(stat.Size(), 10)+"-")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func Base64Decode(fileContent string) (string, error) {
	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	decodeBytes, err := b64.DecodeString(fileContent)
	if err != nil {
		return "", err
	}
	return string(decodeBytes), nil
}

// FormatMountLabel returns a string to be used by the mount command.
// The format of this string will be used to alter the labeling of the mountpoint.
// The string returned is suitable to be used as the options field of the mount command.
// If you need to have additional mount point options, you can pass them in as
// the first parameter.  Second parameter is the label that you wish to apply
// to all content in the mount point.
func FormatMountLabel(src, mountLabel string) string {
	if mountLabel != "" {
		switch src {
		case "":
			src = fmt.Sprintf("context=%q", mountLabel)
		default:
			src = fmt.Sprintf("%s,context=%q", src, mountLabel)
		}
	}
	return src
}

func ConvertPermStrToInt(str string) int {
	var res = 0
	if str[0] == '0' {
		if len(str) == 1 {
			res = 0
		} else if str[1] == 'x' {
			// this is hex number
			for i := 2; i < len(str); i++ {
				res = res*16 + int(str[i]-'0')
			}
		} else {
			// this is a octal number
			for i := 1; i < len(str); i++ {
				res = res*8 + int(str[i]-'0')
			}
		}
	} else {
		res, _ = strconv.Atoi(str)
	}
	if res > 511 {
		res = 511
	}
	return res
}

func RandStr(strSize int, randType string) string {
	var dictionary string
	if randType == "alphanum" {
		dictionary = "0123456789abcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

type Data struct {
	Root string `json:root`
	ISO  string `json:iso`
}

func SetDaemon(d interface{}) {
	HYPER_DAEMON = d
}

func SetHyperEnv(file, rootpath, isopath string) error {
	HYPER_ROOT = rootpath
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	var d = Data{
		Root: rootpath,
		ISO:  isopath,
	}
	var str []byte
	str, err = json.Marshal(d)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, string(str))
	defer f.Close()
	return nil
}

func GetAvailableDriver(drivers []string) string {
	for _, d := range drivers {
		if strings.Contains(d, "kvm") {
			if _, err := exec.LookPath("qemu-system-i386"); err == nil {
				return d
			}
		}
		if strings.Contains(d, "xen") {
			if _, err := exec.LookPath("xl"); err == nil {
				return d
			}
		}
		if strings.Contains(d, "vbox") {
			if _, err := exec.LookPath("vboxmanage"); err == nil {
				return d
			}
		}
	}
	return ""
}

func GetHostIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
	   // check the address type and if it is not a loopback the display it
	   if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
		  if ipnet.IP.To4() != nil {
			 return ipnet.IP.String()
		  }
	   }
	}
	return ""
}
