package urlgen

import (
	"fmt"
	"math/rand"
	"time"
)

func choice(arr *[]string) string {
	return (*arr)[rand.Intn(len(*arr))]
}

var abbrs []string
var adjs []string
//var ings []string
var verbs []string
var nouns []string

func Setup() {
	rand.Seed(time.Now().Unix())
	abbrs = []string{"TCP", "HTTP", "SDD", "RAM", "GB", "CSS", "SSL", "AGP",
		"SQL", "FTP", "PCI", "AI", "ADP", "RSS", "XML", "EXE", "COM", "HDD",
		"THX", "SMTP", "SMS", "USB", "PNG", "SAS", "IB", "SCSI", "JSON", "XSS",
		"JBOD", "SASL", "DDL", "TLA", "NTP", "ADB", "LKML"}
	adjs = []string{"Auxiliary", "Primary", "Backend", "Digital", "Opensource",
		"Virtual", "Crossplatform", "Redundant", "Online", "Haptic",
		"Multibyte", "Bluetooth", "Wireless", "1080p", "Neural", "Optical",
		"Corroded", "Production", "Hacky", "Deterministic", "Binary",
		"Convolutional", "Driverless", "Proprietary", "Critical", "Cryptographic",
		"Simulated", "Smart", "4K"}
	//ings = []string{"Bypassing", "Hacking", "Overriding", "Compressing", "Copying",
	//"Navigating", "Indexing", "Connecting", "Generating", "Quantifying",
	//"Calculating", "Synthesizing", "Transmitting", "Programming", "Parsing",
	//"Leeching", "Consuming", "Sniffing", "Decrypting", "Designing", "Compiling",
	//"Interpreting", "Serializing", "Torrenting", "Encrypting", "Patching"}
	verbs = []string{"Bypass", "Hack", "Override", "Compress", "Copy", "Navigate",
		"Index", "Connect", "Generate", "Quantify", "Calculate", "Synthesize",
		"Input", "Transmit", "Program", "Reboot", "Parse", "Leech", "Consume",
		"Sniff", "Decrypt", "Design", "Compile", "Interpret", "Serialize",
		"Torrent", "Encrypt", "Patch"}
	nouns = []string{"Driver", "Protocol", "Bandwidth", "Panel", "Microchip",
		"Program", "Port", "Card", "Array", "Interface", "System", "Sensor",
		"Firewall", "Pixel", "Alarm", "Feed", "Monitor", "Application", "Bus",
		"Transmitter", "Circuit", "Capacitor", "Matrix", "Vector", "Voxel",
		"Architecture", "Bytecode", "Network", "Router", "Gateway", "Certificate",
		"Padding", "Message", "Signal", "Buffer", "Stack"}
}

func GenUrl() string {
	return fmt.Sprintf("%s%s%s%s", choice(&verbs),
		choice(&adjs),
		choice(&abbrs),
		choice(&nouns))
}

//func main() {
//rand.Seed(time.Now().Unix())
//log.Printf("Abbr %s", choice(&abbrs))
//log.Printf("Adj %s", choice(&adjs))
//log.Printf("Ingverb %s", choice(&ings))
//log.Printf("Verb %s", choice(&verbs))
//log.Printf("Noun %s", choice(&nouns))
//log.Printf("Full shorturl %s%s%s%s", choice(&verbs), choice(&adjs), choice(&abbrs), choice(&nouns))
//log.Printf("%d %d %d %d. Tot %d", len(verbs), len(adjs), len(abbrs), len(nouns), len(verbs) * len(adjs) * len(abbrs) * len(nouns))
//}
