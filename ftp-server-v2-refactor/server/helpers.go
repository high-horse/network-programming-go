package server

import(
	"bufio"
	"net"
	"os"
	"strings"
	"fmt"
)


func sendLine(w *bufio.Writer, line string) {
	// fmt.Fprintln(w, line)
	w.WriteString(line + "\r\n")
	w.Flush()
}


func parseCmd(line string) (string, string) {
	parts := strings.SplitN(line, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}
	return cmd, arg
}

func getLANIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return strings.ReplaceAll(ipnet.IP.String(), ".", ",")
		}
	}
	return "127,0,0,1" // fallback
}



func fileModeToStr(mode os.FileMode) string {
	// Simplified version of ls -l mode string
	var str strings.Builder
	if mode.IsDir() {
		str.WriteByte('d')
	} else {
		str.WriteByte('-')
	}
	perms := []struct {
		bit  os.FileMode
		char byte
	}{
		{0400, 'r'}, {0200, 'w'}, {0100, 'x'},
		{0040, 'r'}, {0020, 'w'}, {0010, 'x'},
		{0004, 'r'}, {0002, 'w'}, {0001, 'x'},
	}
	for _, p := range perms {
		if mode&p.bit != 0 {
			str.WriteByte(p.char)
		} else {
			str.WriteByte('-')
		}
	}
	return str.String()
}

func humanReadableSize(size int64) string {
    const unit = 1024
    if size < unit {
        return fmt.Sprintf("%d B", size)
    }
    div, exp := int64(unit), 0
    for n := size / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    value := float64(size) / float64(div)
    units := []string{"KB", "MB", "GB", "TB"}
    if exp >= len(units) {
        exp = len(units) - 1
    }
    return fmt.Sprintf("%.2f %s", value, units[exp])
}