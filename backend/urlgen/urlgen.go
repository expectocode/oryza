package urlgen

import (
	"fmt"
	"math/rand"
	"mime"
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
	//"Calculating", "Synthesizing", "Transmitting", "Programming", "Rebooting",
	//"Parsing", "Leeching", "Consuming", "Sniffing", "Decrypting", "Designing",
	//"Compiling", "Interpreting", "Serializing", "Torrenting", "Encrypting",
	//"Patching"}
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

func GenLongUri() string {
	return fmt.Sprintf("%s%s%s%s", choice(&verbs),
		choice(&adjs),
		choice(&abbrs),
		choice(&nouns))
}

const alnumBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a 62-letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// thanks stackoverflow
func RandAlphanum(n int) string {
	// Generate a decent random string of length n from alnumBytes
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(alnumBytes) {
			b[i] = alnumBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func GetExtension(mimetype string) string {
	exts, _ := mime.ExtensionsByType(mimetype)
	var ext string
	if exts != nil {
		// If we have some values, use the first
		ext = string(exts[0])
	} else {
		// No detected extension
		extras := map[string]string{
			"image/webp":    ".webp",
			"text/x-python": ".py",
		}
		ext = extras[mimetype]
	}
	if ext == ".asc" {
		// Who uses asc??
		ext = ".txt"
	}
	// TODO consider adding custom eg x-log=.log, x-compressed-tar=.tar.gz
	return ext
}
