package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"regexp"
	"strconv"
	"math"
)

const UPX string = "/usr/local/bin/upx"
const DS string = "/usr/local/bin/ds"
const ENV string = "DS_UPX_TOTAL_SAVINGS"

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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
			return value
	}
	return fallback
}

func main(){
	if(len(os.Args) > 1){
		args := os.Args[1:]
		if len(args) > 0 {
			file := args[0]
			if file == "--bye" {
				bytesTotal, _ := strconv.Atoi(getEnv(ENV, "0"))
				fmt.Fprintf(os.Stdout, "DS TOTAL SAVINGS: %s\n", prettyByteSize(bytesTotal))
				os.Remove(UPX)
				os.Remove(DS)
			}else{
				cmd := exec.Command(UPX, "-q", "--no-backup", "-9", "--best", "--lzma", file)
				cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid:true}
				stdout, err := cmd.Output()
				if err == nil {
					matches := regexp.MustCompile(`(\d+)\s+->\s+(\d+)\s+(\S+)`).FindAllStringSubmatch(string(stdout), -1)
					if len(matches) > 0 {
						if len(matches[0]) > 2 {
							bytesBefore, _ := strconv.Atoi(matches[0][1])
							bytesAfter, _ := strconv.Atoi(matches[0][2])
							bytesTotal, _ := strconv.Atoi(getEnv(ENV, "0"))
							fmt.Fprintf(os.Stdout, "file %s shrunk by %s (~%s)\n", file, matches[0][3], prettyByteSize(bytesBefore - bytesAfter))
							os.Setenv(ENV, strconv.Itoa(bytesTotal + (bytesBefore - bytesAfter)))
						}
					}
				}
			}
		}
	}
	os.Exit(0)
}