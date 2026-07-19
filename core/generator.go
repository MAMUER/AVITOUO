package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// GenerateUniqueID генерирует уникальный идентификатор до 100 знаков
// Разрешены: цифры, буквы RU/EN, символы , \ / ( ) [ ] - =
// Пример: xjfdge4735202
func GenerateUniqueID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, 16)
	rand.Read(bytes)

	id := make([]byte, 0, 20)
	for _, b := range bytes {
		id = append(id, charset[int(b)%len(charset)])
	}

	return string(id)
}

// ShuffleWords перемешивает слова в названии для уникальности
// Использует алгоритм Фишера-Йейтса
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
