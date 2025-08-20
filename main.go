package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"regexp"
	"strconv"
	"math"
	"io/ioutil"
)

const UPX string = "/usr/local/bin/upx"
const DS string = "/usr/local/bin/ds"
const STORE string = "/tmp/.ds"

func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func getTotal() int {
	data, err := ioutil.ReadFile(STORE)
	if err != nil{
		return 0
	}
	total, _ := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	return total
}

func addTotal(b int) {
	total := 0
	if _, err := os.Stat(STORE); err == nil {
		data, err := ioutil.ReadFile(STORE)
		if err == nil {
			total, _ = strconv.Atoi(string(data))
		}
	}
	total += b
	ioutil.WriteFile(STORE, []byte(strconv.Itoa(total)), os.ModePerm)
}

func main(){
	if(len(os.Args) > 1){
		args := os.Args[1:]
		if len(args) > 0 {
			file := args[0]
			if file == "--bye" {
				fmt.Fprintf(os.Stdout, "DS TOTAL SAVINGS: %s\n", prettyByteSize(getTotal()))
				os.Remove(UPX)
				os.Remove(DS)
				os.Remove(STORE)
			}else{
				if file != UPX && file != DS {
					cmd := exec.Command(UPX, "-q", "--no-backup", "-9", "--best", "--lzma", file)
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid:true}
					stdout, err := cmd.Output()
					if err == nil {
						matches := regexp.MustCompile(`(\d+)\s+->\s+(\d+)\s+(\S+)`).FindAllStringSubmatch(string(stdout), -1)
						if len(matches) > 0 {
							if len(matches[0]) > 2 {
								bytesBefore, _ := strconv.Atoi(matches[0][1])
								bytesAfter, _ := strconv.Atoi(matches[0][2])
								fmt.Fprintf(os.Stdout, "file %s shrunk by %s (~%s)\n", file, matches[0][3], prettyByteSize(bytesBefore - bytesAfter))
								addTotal(bytesBefore - bytesAfter)
							}
						}
					}
				}
			}
		}
	}
	os.Exit(0)
}