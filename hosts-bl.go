package main

//Import dependency packages
import (
	"index/suffixarray"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//Function to display help text and exit
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

//Function to check if format is compressable or not
func compressable(format string) bool {
	if format == "hosts" || format == "ipv6" {return true}
	return false
}

//Function to check if format is reducible or not
func reducible(format string) bool {
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

func main() {

	//Declare variables
	var (
		//Flag pointers
		fmtPtr *string
		ifilePtr *string
		cmpPtr *int
		fbhPtr *string
		tbhPtr *string
		tbh6Ptr *string
		cmtsPtr *bool
		dPtr *bool
		ofilePtr *string

		//Common variables
		err error
		rawData	[]byte
		index *suffixarray.Index
		iData []string
		oData []string
		count int
		line string
		tokens []string
		offsets []int
		domainLine bool
	)

	//Initialize flag pointers
	ifilePtr = new(string)
	cmpPtr = new(int)
	fmtPtr = new(string)
	fbhPtr = new(string)
	tbh6Ptr = new(string)
	cmtsPtr = new(bool)
	ofilePtr = new(string)

	//Default flag values
	*cmpPtr = 9
	*fmtPtr = "hosts"
	*fbhPtr = "0.0.0.0"
	tbhPtr = fbhPtr
	*tbh6Ptr = "::"
	*cmtsPtr = false
	dPtr = cmtsPtr

	//Check if any data available from standard input and use it as default source if there is
	stdinStat, _ := os.Stdin.Stat()
	if stdinStat.Mode() & os.ModeNamedPipe != 0 {*ifilePtr = "-"}

	//Push arguments to flag pointers
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "f":
					i++
					fmtPtr = &os.Args[i]
					continue
				case "i":
					i++
					ifilePtr = &os.Args[i]
					continue
				case "compression":
					i++
					*cmpPtr, err = strconv.Atoi(os.Args[i])
					if err != nil {help(1)}
					if *cmpPtr < 1 || *cmpPtr > 9 {help(2)}
					continue
				case "from_blackhole":
					i++
					fbhPtr = &os.Args[i]
					continue
				case "to_blackhole":
					i++
					tbhPtr = &os.Args[i]
					continue
				case "to_blackhole_v6":
					i++
					tbh6Ptr = &os.Args[i]
					continue
				case "comments":
					*cmtsPtr = true
					continue
				case "dupe":
					*dPtr = true
					continue
				case "o":
					i++
					ofilePtr = &os.Args[i]
					continue
				default:
					help(3)
			}
		} else if *ifilePtr == "" {ifilePtr = &os.Args[i]
		} else if *ofilePtr == "" {ofilePtr = &os.Args[i]
		} else {help(4)}
	}

	//Print help if no input available
	if *ifilePtr == "" {help(0)}

	//Set default output if none specified
	if *ofilePtr == "" {
		if *ifilePtr == "-" {*ofilePtr = "-"
		} else {*ofilePtr = *fmtPtr+"-"+filepath.Base(*ifilePtr)}
	}

	//Initialize format string for quick reference
	format := strings.ToLower(*fmtPtr)

	//Initialize data from either stdin or file
	if *ifilePtr == "-" {
		rawData, err = io.ReadAll(os.Stdin)
		if err != nil {help(5)}
	} else {rawData, err = os.ReadFile(*ifilePtr)}
	if err != nil {help(6)}

	//Convert data to string and initialize variable for cleaning
	filteredData := string(rawData)

	//Conform line endings
	for strings.Contains(filteredData, "\r\n") || strings.Contains(filteredData, "\n\n") {
		for strings.Contains(filteredData, "\r\n") {
			filteredData = strings.Replace(filteredData, "\r\n", "\n", -1)
		}
		for strings.Contains(filteredData, "\n\n") {
			filteredData = strings.Replace(filteredData, "\n\n", "\n", -1)
		}
	}

	//Initialize scanner to scan data
	iData = strings.Split(filteredData, "\n")
	filteredData = ""

	//Scan through and extract clean data into array
	for _, line := range iData {
		tokens = strings.Fields(line)
		if len(tokens) == 0 {continue}
		if tokens[0] == *fbhPtr {
			if tokens[1] != "0.0.0.0" {oData = append(oData, tokens[1])}
		} else if *cmtsPtr {
			if strings.HasPrefix(line, "#") {oData = append(oData, line)}
		}
	}

	iData = oData
	oData = nil

	//Initialize search index if needed
	if !*dPtr || reducible(format) {index = suffixarray.New([]byte("\n"+strings.Join(iData, "\n")+"\n"))}

	//Remove duplicates if needed
	if !*dPtr {
		for i := 0; i < len(iData); i++ {
			if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
			} else {
				offsets = index.Lookup([]byte("\n"+iData[i]+"\n"), 2)
				if len(offsets) == 1 {oData = append(oData, iData[i])
				} else {
					count = 0
					for a := 0; a < len(oData); a++ {
						if oData[a] == iData[i] {
							count++
							break
						}
					}
					if count == 0 {oData = append(oData, iData[i])}
				}
			}
		}
		iData = oData
		oData = nil
	}

	//If requested format is FQDN, just dump to file or stdout and exit
	if format == "fqdn" {
		if *ofilePtr == "-" {os.Stdout.WriteString(strings.Join(iData, eol)+eol)
		} else {os.WriteFile(*ofilePtr, []byte(strings.Join(iData, eol)+eol), 0644)}
		os.Exit(0)
	}

	//Reduce domains if needed
	if reducible(format) {
		for i := 0; i < len(iData); i++ {
			if strings.HasPrefix(iData[i], "#") {
				oData = append(oData, iData[i])
			} else {
				tokens = strings.Split(iData[i], ".")
				if len(tokens) <= 2 {oData = append(oData, iData[i])
				} else {
					count = 0
					for a := 1; a <= len(tokens)-2; a++ {
						offsets = index.Lookup([]byte("\n"+strings.Join(tokens[a:], ".")+"\n"), 1)
						if offsets != nil {count++}
					}
					if count == 0 {oData = append(oData, iData[i])}
				}
			}	
		}
		iData = oData
		oData = nil
	}

	//If requested format is reduced FQDN, just dump to file or stdout and exit
	if format == "rfqdn" {
		if *ofilePtr == "-" {os.Stdout.WriteString(strings.Join(iData, eol)+eol)
		} else {os.WriteFile(*ofilePtr, []byte(strings.Join(iData, eol)+eol), 0644)}
		os.Exit(0)
	}

	//Compress data if needed
	if compressable(format) && *cmpPtr > 1 {
		line = ""
		count = 0
		for i := 0; i < len(iData); i++ {
			if strings.HasPrefix(iData[i], "#") {
				if i == 0 || domainLine == false {line = line+iData[i]
				} else if line != "" {
					oData = append(oData, strings.TrimPrefix(line, " "))
					line = iData[i]
				}
				domainLine = false
			} else {
				count++
				if i == 0 {line = iData[i]
				} else if domainLine == true {line = line+" "+iData[i]
				} else {
					if line != "" {
						oData = append(oData, line)
						line = iData[i]
					}
					count = 1
				}
				if count == *cmpPtr {
					if line != "" {
						oData = append(oData, strings.TrimPrefix(line, " "))
						line = ""
					}
					count = 0
				}
				domainLine = true
			}
		}
		if line != "" {oData = append(oData, strings.TrimPrefix(line, " "))}
		iData = oData
		oData = nil
	}

	//Format data
	for i := 0; i < len(iData); i++ {
		switch format {
			case "hosts":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {oData = append(oData, *tbhPtr+" "+iData[i])}
			case "dualserver":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {oData = append(oData, iData[i]+"="+*tbhPtr)}
			case "ipv6":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {
					oData = append(oData, *tbhPtr+" "+iData[i])
					oData = append(oData, *tbh6Ptr+" "+iData[i])
				}
			case "adblock":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, "!"+strings.TrimPrefix(iData[i], "#"))
				} else {
					oData = append(oData, "||"+iData[i]+"^")
				}
			case "dnsmasq":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {
					oData = append(oData, "address=/"+iData[i]+"/")
				}
			case "privoxy":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {
					oData = append(oData, iData[i])
					oData = append(oData, "."+iData[i])
				}
			case "rpz":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, ";"+strings.TrimPrefix(iData[i], "#"))
				} else {
					oData = append(oData, iData[i]+" CNAME .")
					oData = append(oData, "*."+iData[i]+" CNAME .")
				}
			case "unbound":
				if strings.HasPrefix(iData[i], "#") {
					oData = append(oData, iData[i])
				} else {
					oData = append(oData, "local-zone: \""+iData[i]+"\" always_nxdomain")
				}
			default:
				help(7)
		}
	}

	iData = oData
	oData = nil

	//Write formatted data to output file or stdout and exit
	if *ofilePtr == "-" {os.Stdout.WriteString(strings.Join(iData, eol)+eol)
	} else {os.WriteFile(*ofilePtr, []byte(strings.Join(iData, eol)+eol), 0644)}
	os.Exit(0)
}
