package safemime

import "strings"

// When serving arbitrary file uploads from our domain, browsers will interpret some content-types
// differently - for example, HTML. This is affected by things like the Content Security Policy and
// can be a security hole. Because mime types are so varied, it doesn't seem feasible to maintain a
// comprehensive blacklist of potentially insecure mime types, so we must use a whitelist. Since Go
// doesn't have const maps, we use this function. Some mime types are super obscure, but if I were a
// user of an obscure type, I'd like to be supported.

//If there is a type missing from this list that you would like on it, drop me a PR or something and we'll see.
// Usage: safe_mime_type := SafeMime()(unsafe_mime_type)
func SafeMime() func(string) string {
	innerMap := map[string]string{
		"application/applefile":		"application/applefile",
		"application/mac-binhex40":		"application/mac-binhex40",
		"application/x-macbinary":		"application/x-macbinary",
		"application/x-compress":		"application/x-compress",
		"application/compress":			"application/compress",
		"application/x-fortezza-ckl":	"application/x-fortezza-ckl",
		"application/x-fortezza-krl":	"application/x-fortezza-krl",
		"application/x-gzip":			"application/x-gzip",
		"application/gzip":				"application/gzip",
		"application/x-gunzip":			"application/x-gunzip",
		"application/brotli":			"application/brotli",
		"application/zip":				"application/zip",
		"application/http-index-format":	"text/plain", // safe?
		"application/ecmascript":		"text/plain", // safe
		"application/javascript":		"text/plain", // safe
		"application/x-javascript":		"text/plain", // safe
		"application/json":				"text/plain", // safe
		"application/x-javascript-config":		"application/octet-stream", // safe
		"application/octet-stream":			"application/octet-stream",
		"application/pgp":				"application/pgp",
		"application/x-pgp-message":	"application/x-pgp-message",
		"application/postscript":		"application/postscript",
		"application/pdf":				"application/pdf",
		"application/pre-encrypted":	"application/pre-encrypted",
		"application/x-uuencode":		"application/x-uuencode",
		"application/x-uue":			"application/x-uue",
		"application/uuencode":			"application/uuencode",
		"application/uue":				"application/uue",
		"application/x-x509-ca-cert":	"application/x-x509-ca-cert",
		"application/x-x509-server-cert":		"application/x-x509-server-cert",
		"application/x-x509-email-cert": "application/x-x509-email-cert",
		"application/x-x509-user-cert":	"application/x-x509-user-cert",
		"application/x-pkcs7-crl":		"application/x-pkcs7-crl",
		"application/x-pkcs7-mime":		"application/x-pkcs7-mime",
		"application/pkcs7-mime":		"application/pkcs7-mime",
		"application/x-pkcs7-signature":	"application/x-pkcs7-signature",
		"application/pkcs7-signature":	"application/pkcs7-signature",
		"application/x-www-form-urlencoded":	"application/octet-stream", // safe?
		"application/oleobject":		"application/oleobject",
		"application/x-oleobject":		"application/x-oleobject",
		"application/java-archive":		"application/java-archive",
		"application/manifest+json":	"text/plain", // safe
		"application/x-xpinstall":		"application/octet-stream", // safe
		"application/xml":				"text/plain", // safe
		"application/xhtml+xml":		"text/plain", // safe
		"application/xslt+xml":			"text/plain", // safe
		"application/mathml+xml":		"text/plain", // safe
		"application/rdf+xml":			"application/octet-stream", // safe
		"application/vnd.wap.xhtml+xml":	"application/octet-stream", // safe
		"application/package":			"application/octet-stream", // safe

		"audio/basic":					"audio/basic",
		"audio/ogg":					"audio/ogg",
		"audio/x-wav":					"audio/x-wav",
		"audio/webm":					"audio/webm",
		"audio/mpeg":					"audio/mpeg",
		"audio/mp4":					"audio/mp4",
		"audio/amr":					"audio/amr",
		"audio/flac":					"audio/flac",
		"audio/3gpp":					"audio/3gpp",
		"audio/3gpp2":					"audio/3gpp2",
		"audio/x-midi":					"audio/x-midi",
		"audio/x-matroska":				"audio/x-matroska",
		"audio/aac":					"audio/aac",
		"binary/octet-stream":			"binary/octet-stream",

		"image/gif":					"image/gif",
		"image/jpeg":					"image/jpeg",
		"image/webp":					"image/webp",
		"image/jpg":					"image/jpg",
		"image/pjpeg":					"image/pjpeg",
		"image/png":					"image/png",
		"image/apng":					"image/apng",
		"image/x-png":					"image/x-png",
		"image/x-portable-pixmap":		"image/x-portable-pixmap",
		"image/x-xbitmap":				"image/x-xbitmap",
		"image/x-xbm":					"image/x-xbm",
		"image/xbm":					"image/xbm",
		"image/x-jg":					"image/x-jg",
		"image/tiff":					"image/tiff",
		"image/bmp":					"image/bmp",
		"image/x-ms-bmp":				"image/x-ms-bmp",
		"image/x-icon":					"image/x-icon",
		"image/vnd.microsoft.icon":		"image/vnd.microsoft.icon",
		"image/icon":					"image/icon",
		"video/x-mng":					"video/x-mng",
		"image/x-jng":					"image/x-jng",
		"image/svg+xml":				"application/octet-stream", // safe

		"text/enriched":				"text/enriched",
		"text/calendar":				"text/calendar",
		"text/x-python":				"text/plain",
		"text/html":					"text/plain", // safe
		"text/plain":					"text/plain",
		"text/richtext":				"text/richtext",
		"text/vcard":					"text/vcard",
		"text/css":						"text/plain", // safe
		"text/json":					"text/plain", // safe
		"text/xml":						"text/plain", // safe
		"text/rdf":						"text/rdf",
		"text/vtt":						"text/vtt",
		"application/vnd.mozilla.xul+xml":		"application/octet-stream", // safe
		"text/ecmascript":				"text/plain", // safe
		"text/javascript":				"text/plain", // safe
		"text/xsl":						"text/plain", // safe
		"video/mpeg":					"video/mpeg",
		"video/mp4":					"video/mp4",
		"video/quicktime":				"video/quicktime",
		"video/x-raw-yuv":				"video/x-raw-yuv",
		"video/ogg":					"video/ogg",
		"video/webm":					"video/webm",
		"video/3gpp":					"video/3gpp",
		"video/3gpp2":					"video/3gpp2",
		"video/mp2t":					"video/mp2t",
		"video/avi":					"video/avi",
		"video/x-matroska":				"video/x-matroska",
		"application/ogg":				"application/ogg",

	}

	return func(key string) string {
		// this might mess up non-utf8 text. oh well, no one uses that.
		safemime := innerMap[key]
		if safemime == "" {
			if strings.HasPrefix(key, "text/") {
				safemime = "text/plain"
			} else {
				safemime = "application/octet-stream"
			}
		}
		return safemime
	}
}
