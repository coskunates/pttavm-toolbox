package helpers

import (
	"math/rand"
	"net/url"
	"regexp"
	"strings"
)

func RegexReplace(regex, replace, str string) string {
	var re = regexp.MustCompile(regex)
	return re.ReplaceAllString(str, replace)
}

func MakeProductSlug(str string) string {
	str = strings.Replace(str, "ş", "s", -1)
	str = strings.Replace(str, "Ş", "s", -1)
	str = strings.Replace(str, "ı", "i", -1)
	str = strings.Replace(str, "I", "i", -1)
	str = strings.Replace(str, "İ", "i", -1)
	str = strings.Replace(str, "ğ", "g", -1)
	str = strings.Replace(str, "Ğ", "g", -1)
	str = strings.Replace(str, "ü", "u", -1)
	str = strings.Replace(str, "Ü", "u", -1)
	str = strings.Replace(str, "ö", "o", -1)
	str = strings.Replace(str, "Ö", "o", -1)
	str = strings.Replace(str, "Ç", "c", -1)
	str = strings.Replace(str, "ç", "c", -1)

	str = strings.ToLower(str)

	str = RegexReplace(`/&amp;amp;amp;amp;amp;amp;amp;amp;amp;.+?;/`, "", str)

	str = strings.Replace(str, " ", "-", -1)
	str = RegexReplace(`[^a-zA-Z0-9-_]`, "-", str)
	str = RegexReplace(`-+`, "-", str)

	str = strings.Trim(str, "-")

	return str
}

func RandStrFromCharset(n int, charset []byte) string {

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func IsValidURL(input string) bool {
	parsedUrl, err := url.Parse(input)
	if err != nil {
		return false
	}

	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return false
	}

	return true
}

// Örneğin mongodb'de bir field içerisinde dict veri yapısı kullanırken, key içerisinde . $ gibi karakterler gelmesi durumunda hata veriyor. O karakterleri yazarken değiştirip, okurken o haliyle alabilmek için.
func SanitizeKey(key string) string {
	key = strings.ReplaceAll(key, ".", "__DOT__")
	key = strings.ReplaceAll(key, "$", "__DOLLAR__")
	return key
}

func RestoreKey(sanitizedKey string) string {
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "__DOT__", ".")
	sanitizedKey = strings.ReplaceAll(sanitizedKey, "__DOLLAR__", "$")
	return sanitizedKey
}
