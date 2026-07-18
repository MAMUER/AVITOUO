package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func GenerateUniqueID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "av-" + hex.EncodeToString(bytes)
}

func FormatDescription(desc string) string {
	formatted := strings.ReplaceAll(desc, "\n", "<br>")
	formatted = strings.ReplaceAll(formatted, "\r\n", "<br>")
	return fmt.Sprintf("<![CDATA[%s]]>", formatted)
}

func ShuffleWords(title string, usedTitles map[string]bool) (string, error) {
	words := strings.Fields(title)
	if len(words) < 2 {
		return title, fmt.Errorf("недостаточно слов для перемешивания")
	}
	for i := len(words) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		words[i], words[j.Int64()] = words[j.Int64()], words[i]
	}
	newTitle := strings.Join(words, " ")
	if len(newTitle) > 0 {
		newTitle = strings.ToUpper(string(newTitle[0])) + newTitle[1:]
	}
	if usedTitles[newTitle] {
		return "", fmt.Errorf("уникальные комбинации исчерпаны для этого набора слов")
	}
	usedTitles[newTitle] = true
	return newTitle, nil
}
