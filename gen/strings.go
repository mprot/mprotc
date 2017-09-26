package gen

import (
	"strings"
	"unicode"
)

// RPad transforms the given string to a string with a minimum length of width.
// If s is shorter than the given width, spaces a appended. Otherwise the
// string is returned unchanged.
func RPad(s string, width int) string {
	if len(s) < width {
		return s + strings.Repeat(" ", width-len(s))
	}
	return s
}

// SnakeCase transforms the given camel case string into snake case. The given
// string is split into words which are concatenated using an underscore as a
// separator.
//
// The function also recognizes uppercase words, i.e. words that only contain
// uppercase characters. Example: userID will be transformed to user_id.
func SnakeCase(s string) string {
	var res string
	iterWords(s, func(w string) {
		if len(res) != 0 {
			res += "_"
		}
		res += strings.ToLower(w)
	})
	return res
}

// TitleFirstWord transforms the given string to a string which first word is
// title cased.
func TitleFirstWord(s string) string {
	return transformFirstWord(s, strings.Title)
}

// LowerFirstWord transforms the given string to a string which first word is
// lower cased.
func LowerFirstWord(s string) string {
	return transformFirstWord(s, strings.ToLower)
}

func transformFirstWord(s string, transform func(string) string) string {
	var res string
	iterWords(s, func(w string) {
		if len(res) == 0 {
			res = transform(w)
		} else {
			res += w
		}
	})
	return res
}

func iterWords(s string, iter func(string)) {
	if s == "" {
		return
	}

	upperCount := 0 // count consecutive upper case characters
	word := []rune{}
	for _, r := range s {
		if unicode.IsUpper(r) {
			// If the last character was not uppercase (i.e. upperCount == 0),
			// we start a new word. Otherwise we have an uppercase word.
			if upperCount == 0 {
				if len(word) > 0 {
					iter(string(word))
				}
				word = word[:0]
			}
			word = append(word, r)
			upperCount++
		} else {
			if upperCount <= 1 {
				// We are still in a word.
				word = append(word, r)
			} else {
				// An uppercase word ended. Take the last character and add it
				// to the current word.
				iter(string(word[:len(word)-1]))
				word = append(word[:0], word[len(word)-1], r)
			}
			upperCount = 0
		}
	}
	iter(string(word))
}
