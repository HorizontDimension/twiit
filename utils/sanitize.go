package utils

import (
	"bytes"
)

// Replace a set of accented characters with ascii equivalents.
func Sanitize(text string) string {
	// Replace some common accent characters
	b := bytes.NewBufferString("")
	for _, c := range text {
		// Check transliterations first
		if transliterations[c] > 0 {
			b.WriteRune(transliterations[c])
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// A limited list of transliterations to catch common european names translated to urls.
// This set could be expanded with at least caps and many more characters.
var transliterations = map[rune]rune{
	'Š': 'S',
	'š': 's',
	'Ž': 'Z',
	'ž': 'z',
	'À': 'A',
	'Á': 'A',
	'Â': 'A',
	'Ã': 'A',
	'Ä': 'A',
	'Å': 'A',
	'Æ': 'A',
	'Ç': 'C',
	'È': 'E',
	'É': 'E',
	'Ê': 'E',
	'Ë': 'E',
	'Ì': 'I',
	'Í': 'I',
	'Î': 'I',
	'Ï': 'I',
	'Ñ': 'N',
	'Ò': 'O',
	'Ó': 'O',
	'Ô': 'O',
	'Õ': 'O',
	'Ö': 'O',
	'Ø': 'O',
	'Ù': 'U',
	'Ú': 'U',
	'Û': 'U',
	'Ü': 'U',
	'Ý': 'Y',
	'Þ': 'B',
	'à': 'a',
	'á': 'a',
	'â': 'a',
	'ã': 'a',
	'ä': 'a',
	'å': 'a',
	'æ': 'a',
	'ç': 'c',
	'è': 'e',
	'é': 'e',
	'ê': 'e',
	'ë': 'e',
	'ì': 'i',
	'í': 'i',
	'î': 'i',
	'ï': 'i',
	'ð': 'o',
	'ñ': 'n',
	'ò': 'o',
	'ó': 'o',
	'ô': 'o',
	'õ': 'o',
	'ö': 'o',
	'ø': 'o',
	'ù': 'u',
	'ú': 'u',
	'û': 'u',
	'ý': 'y',
	'þ': 'b',
	'ÿ': 'y',
}
