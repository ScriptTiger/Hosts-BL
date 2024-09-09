package main

// Import dependency packages
import (
	"bufio"
	"bytes"
	"encoding/binary"
	//"hash"
	//"hash/fnv"
	"hash/maphash"
	"index/suffixarray"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Globals
var (
	index [][]uint64
	saIndex *suffixarray.Index
	saBuffer *bytes.Buffer
	lookupBuffer *bytes.Buffer
	hostID int
	glob strings.Builder
	cmpLvl int = -1
	wasHost bool
	hasher maphash.Hash
	//hasher hash.Hash64
)

// Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: hosts-bl [options...] <source> <destination>\n"+
		" -comments               Don't remove comments\n"+
		" -compression <number>   Number of domains per line, 1 to 9\n"+
		" -dupe                   Don't check for and remove duplicates\n"+
		" -f <format>             Destination format:\n"+
		"                         adblock,dnsmasq,dualserver,fqdn,\n"+
		"                         hosts,ipv6,privoxy,rfqdn,rpz,unbound\n"+
		" -from_blackhole <IPv4>  Black hole address in source\n"+
		" -i <file>               Source file\n"+
		" -o <file>               Destination file\n"+
		" -to_blackhole <IPv4>    Black hole address in destination\n"+
		" -to_blackhole_v6 <IPv6> IPv6 Black hole address in destination\n")
	os.Exit(err)
}

// Function to check if format is reducible or not
func isReducible(format string) bool {
	rformat := [6]string{
		"adblock",
		"rfqdn",
		"dnsmasq",
		"privoxy",
		"rpz",
		"unbound"}
	for i := 0; i < len(rformat); i++ {
		if format == rformat[i] {return true}
	}
	return false
}

// Function to check if format is valid
func isValidFormat(format string) bool {
	if isReducible(format) {return true}
	vformat := [4]string{
		"fqdn",
		"hosts",
		"ipv6",
		"dualserver",
	}
	for i := 0; i < len(vformat); i++ {
		if format == vformat[i] {return true}
	}
	return false
}

// Function to hash
func makeHash(host string) (hostHash uint64) {
	hasher.WriteString(host)
	//hasher.Write([]byte(host))
	hostHash = hasher.Sum64()
	hasher.Reset()
	return
}

// Function to build index
func buildIndex(hosts []string, format string) {
	for _, host := range hosts {
		hostHash := makeHash(host)
		index = append(index, []uint64{uint64(hostID), hostHash})
		if isReducible(format) {
			binary.Write(saBuffer, binary.LittleEndian, hostHash)
			binary.Write(saBuffer, binary.LittleEndian, uint64(0))
		}
		hostID++
	}
}

// Function to format output and write to file/stdout
func writeLine(hosts []string, writer *bufio.Writer, format string, dupe, cmts bool, cmp int, tbhPtr, tbh6Ptr *string) {
	for _, host := range hosts {
		if !strings.HasPrefix(host, "#") {
			if (dupe && !isReducible(format))|| index[hostID][1] != 0 {
				switch format {
					case "hosts":
						compress(host, format, cmp, true, tbhPtr, tbh6Ptr, writer)
					case "ipv6":
						compress(host, format, cmp, true, tbhPtr, tbh6Ptr, writer)
					case "dualserver":
						writer.WriteString(host+"="+*tbhPtr+eol)
					case "adblock":
						writer.WriteString("||"+host+"^"+eol)
					case "dnsmasq":
						writer.WriteString("address=/"+host+"/"+eol)
					case "privoxy":
						writer.WriteString(
							host+eol+
							"."+host+eol,
						)
					case "rpz":
						writer.WriteString(
							host+" CNAME ."+eol+
							"*."+host+" CNAME ."+eol,
						)
					case "unbound":
						writer.WriteString("local-zone: \""+host+"\" always_nxdomain"+eol)
					default:
						writer.WriteString(host+eol)
				}
			}
			hostID++
		} else if cmts {
			switch format {
				case "hosts":
					compress(host, format, cmp, false, tbhPtr, tbh6Ptr, writer)
				case "ipv6":
					compress(host, format, cmp, false, tbhPtr, tbh6Ptr, writer)
				case "adblock":
					writer.WriteString("!"+strings.TrimPrefix(host, "#")+eol)
				case "rpz":
					writer.WriteString(";"+strings.TrimPrefix(host, "#")+eol)
				default:
					writer.WriteString(host+eol)
			}
		}
	}
}

// Function to compress hosts
func compress(host, format string, cmp int, isHost bool, tbhPtr, tbh6Ptr *string, writer *bufio.Writer) {
	if cmp == 1 {
		if isHost {
			if format == "hosts" {
				writer.WriteString(*tbhPtr+" "+host+eol)
			} else {
				writer.WriteString(
					*tbhPtr+" "+host+eol+
					*tbh6Ptr+" "+host+eol,
				)
			}
		} else {writer.WriteString(host+eol)}
		return
	}
	if cmpLvl == -1 {
		wasHost = isHost
		cmpLvl++
	} else if wasHost != isHost {
		flushGlob(format, wasHost, tbhPtr, tbh6Ptr, writer)
		wasHost = isHost
	}
	if isHost {
		glob.WriteString(" "+host)
		cmpLvl++
		if cmpLvl == cmp {flushGlob(format, isHost, tbhPtr, tbh6Ptr, writer)}
	} else {glob.WriteString(host)}
}

// Function to flush glob
func flushGlob(format string, isHost bool, tbhPtr, tbh6Ptr *string, writer *bufio.Writer) {
	if isHost {
		if format == "hosts" {
			writer.WriteString(*tbhPtr+glob.String()+eol)
		} else {
			writer.WriteString(
				*tbhPtr+glob.String()+eol+
				*tbh6Ptr+glob.String()+eol,
			)
		}
	} else {writer.WriteString(glob.String()+eol)}
	cmpLvl = 0
	glob.Reset()
}

// Function to zero out subdomains of domains already present
func deSub(hosts []string) {
	for _, host := range hosts {
		parentString := host
		parentCount := len(strings.Split(parentString, "."))
		for ; parentCount > 2; parentCount-- {
			parentString = strings.Join(strings.Split(parentString, ".")[1:], ".")
 			hostHash := makeHash(parentString)
			binary.Write(lookupBuffer, binary.LittleEndian, uint64(0))
			binary.Write(lookupBuffer, binary.LittleEndian, hostHash)
			binary.Write(lookupBuffer, binary.LittleEndian, uint64(0))
			offsets := saIndex.Lookup(lookupBuffer.Bytes(), 1)
			lookupBuffer.Reset()
			if offsets != nil {
				index[hostID][1] = 0
				break
			}
		}
		hostID++
	}
}

// Function to zero out duplicate hosts on index
func deDupe() {
	sort.SliceStable(index, func(i, j int) bool {
		return index[i][1] < index[j][1]
	})
	var lastLookup uint64
	for i, lookup := range index {
		if lookup[1] == lastLookup {index[i][1] = 0
		} else {lastLookup = lookup[1]}
	}
	sort.Slice(index, func(i, j int) bool {
		return index[i][0] < index[j][0]
	})
}

// Function to scrub input
func scrubInput(line string, fbhPtr *string, cmts bool) ([]string) {
	if strings.ContainsAny(line, " .") && strings.HasPrefix(line, *fbhPtr) {
		line = strings.SplitAfterN(line, " ", 2)[1]
		if strings.ContainsAny(line, "#") {line = strings.TrimSuffix(strings.SplitAfterN(line, "#", 2)[0], "#")}
		line = strings.TrimSpace(line)
		if line == "0.0.0.0" {return nil}
		return strings.Fields(line)
	} else if cmts && strings.HasPrefix(line, "#") {return []string{line}}
	return nil
}

func main() {

	// Declare variables
	var (
		// Flags
		fmtPtr *string
		iFilePtr *string
		oFilePtr *string
		cmp int
		fbhPtr *string
		tbhPtr *string
		tbh6Ptr *string
		dupe bool
		cmts bool

		// Common variables
		err error
		iFile *os.File
		oFile *os.File
		iData []byte
		iReader *bytes.Reader
		scanner *bufio.Scanner
	)

	// Initialize flag pointers
	iFilePtr = new(string)
	oFilePtr = new(string)
	fmtPtr = new(string)
	fbhPtr = new(string)
	tbh6Ptr = new(string)

	// Default flag values
	*iFilePtr = ""
	*oFilePtr = ""
	cmp = -1
	*fmtPtr = "hosts"
	*fbhPtr = "0.0.0.0"
	tbhPtr = fbhPtr
	*tbh6Ptr = "::"
	dupe = false
	cmts = false

	// Check if any data available from standard input and use it as default source if there is
	stdinStat, _ := os.Stdin.Stat()
	if stdinStat.Mode() & os.ModeNamedPipe != 0 {*iFilePtr = "-"}

	// Push arguments to flag pointers
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "f":
					i++
					if *fmtPtr != "hosts" {help(1)}
					fmtPtr = &os.Args[i]
					continue
				case "i":
					i++
					if *iFilePtr != "" {help(2)}
					iFilePtr = &os.Args[i]
					continue
				case "compression":
					i++
					if cmp != -1 {help(3)}
					cmp, err = strconv.Atoi(os.Args[i])
					if err != nil {help(4)}
					if cmp < 1 || cmp > 9 {help(5)}
					continue
				case "from_blackhole":
					i++
					if *fbhPtr != "0.0.0.0" {help(6)}
					fbhPtr = &os.Args[i]
					continue
				case "to_blackhole":
					i++
					if *tbhPtr != "0.0.0.0" {help(7)}
					tbhPtr = &os.Args[i]
					continue
				case "to_blackhole_v6":
					i++
					if *tbh6Ptr != "::" {help(8)}
					tbh6Ptr = &os.Args[i]
					continue
				case "comments":
					if cmts {help(9)}
					cmts = true
					continue
				case "dupe":
					if dupe {help(10)}
					dupe = true
					continue
				case "o":
					i++
					if *oFilePtr != "" {help(11)}
					oFilePtr = &os.Args[i]
					continue
				default:
					help(12)
			}
		} else if *iFilePtr == "" {iFilePtr = &os.Args[i]
		} else if *oFilePtr == "" {oFilePtr = &os.Args[i]
		} else {help(13)}
	}

	// Print help if no input available
	if *iFilePtr == "" {help(0)}

	// Initialize format string for quick reference
	format := strings.ToLower(*fmtPtr)

	// Exit if invalid format is given
	if !isValidFormat(format) {help(14)}

	// Set default compression if none specified
	if cmp == -1 {cmp = 9}

	// Set default output if none specified
	if *oFilePtr == "" {
		if *iFilePtr == "-" {*oFilePtr = "-"
		} else {*oFilePtr = *fmtPtr+"-"+filepath.Base(*iFilePtr)}
	}

	// Set file handles and associated buffered I/O
	if *iFilePtr == "-" {
		iData, err = io.ReadAll(os.Stdin)
		iReader = bytes.NewReader(iData)
	} else {
		iFile, err = os.Open(*iFilePtr)
		defer iFile.Close()
	}
	if err != nil {panic(err)}

	if *oFilePtr == "-" {oFile = os.Stdout
	} else {
		oFile, err = os.Create(*oFilePtr)
		if err != nil {panic(err)}
		defer oFile.Close()
	}
	writer := bufio.NewWriter(oFile)

	// Initialize hasher
	hasher.SetSeed(maphash.MakeSeed())
	//hasher = fnv.New64a()

	// Initialize buffers
	saBuffer = new(bytes.Buffer)
	lookupBuffer = new(bytes.Buffer)
	binary.Write(saBuffer, binary.LittleEndian, uint64(0))

	// Build index
	if !dupe || isReducible(format) {
		if *iFilePtr == "-" {scanner = bufio.NewScanner(iReader)
		} else {scanner = bufio.NewScanner(iFile)}
		for scanner.Scan() {buildIndex(scrubInput(scanner.Text(), fbhPtr, false), format)}
		if isReducible(format) {
			saIndex = suffixarray.New(saBuffer.Bytes())
		}
	}

	// Zero out duplicates as needed
	if !dupe {deDupe()}

	// Zero out subdomains as needed
	if isReducible(format) {
		hostID = 0
		if *iFilePtr == "-" {
			iReader.Seek(0, io.SeekStart)
			scanner = bufio.NewScanner(iReader)
		} else {
			iFile.Seek(0, io.SeekStart)
			scanner = bufio.NewScanner(iFile)
		}
		for scanner.Scan() {deSub(scrubInput(scanner.Text(), fbhPtr, false))}
	}

	// Format and write final list
	hostID = 0
	if *iFilePtr == "-" {
		iReader.Seek(0, io.SeekStart)
		scanner = bufio.NewScanner(iReader)
	} else {
		iFile.Seek(0, io.SeekStart)
		scanner = bufio.NewScanner(iFile)
	}
	for scanner.Scan() {writeLine(scrubInput(scanner.Text(), fbhPtr, true), writer, format, dupe, cmts, cmp, tbhPtr, tbh6Ptr)}
	if glob.Len() != 0 {flushGlob(format, wasHost, tbhPtr, tbh6Ptr, writer)}
	writer.Flush()

}