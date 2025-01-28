package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	uploadURL     = "http://localhost:30000" // URL сервера Tus
	chunkSize     = 200 * 1024 * 1024        // Размер блока (1MB)
	tusVersion    = "1.0.0"                  // Версия протокола Tus
	maxRetryCount = 3                        // Максимальное количество попыток
)

func main() {
	filePath := "./video.mp4" // Путь к загружаемому файлу

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	// Получаем размер файла
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Ошибка получения информации о файле: %v\n", err)
		return
	}
	fileSize := fileInfo.Size()

	// Создаём загрузку на сервере Tus
	uploadID, err := createTusUpload(fileSize, fileInfo.Name())
	if err != nil {
		fmt.Printf("Ошибка создания загрузки: %v\n", err)
		return
	}
	fmt.Printf("Загрузка создана с ID: %s\n", uploadID)

	// Загружаем файл по частям
	err = uploadFileChunks(uploadID, file)
	if err != nil {
		fmt.Printf("Ошибка загрузки файла: %v\n", err)
		return
	}
	fmt.Println("Файл успешно загружен!")
}

// createTusUpload создаёт новую загрузку на сервере Tus.
func createTusUpload(fileSize int64, fileName string) (string, error) {
	client := &http.Client{}

	// Создаём запрос на создание загрузки
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", uploadURL, "/files"), nil)
	if err != nil {
		return "", err
	}

	// Добавляем заголовки
	req.Header.Set("Tus-Resumable", tusVersion)
	req.Header.Set("Upload-Length", fmt.Sprintf("%d", fileSize))
	req.Header.Set("Upload-Metadata", encodeMetadata(map[string]string{
		"filename": fileName,
	}))

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("ошибка создания загрузки: %s", resp.Status)
	}

	// Получаем ID загрузки из заголовка Location
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("не указан заголовок Location в ответе")
	}

	return fmt.Sprintf("%s%s", uploadURL, location), nil
}

// uploadFileChunks загружает файл по частям на сервер Tus.
func uploadFileChunks(uploadID string, file *os.File) error {
	client := &http.Client{}

	offset := int64(0)
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	for {
		// Читаем следующий блок данных
		buffer := make([]byte, chunkSize)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		buffer = buffer[:n]

		// Генерируем контрольную сумму
		checksum := generateChecksum(buffer)

		// Создаём PATCH-запрос для загрузки блока
		req, err := http.NewRequest("PATCH", uploadID, bytes.NewReader(buffer))
		if err != nil {
			return err
		}

		// Добавляем заголовки
		req.Header.Set("Tus-Resumable", tusVersion)
		req.Header.Set("Upload-Offset", fmt.Sprintf("%d", offset))
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		req.Header.Set("Upload-Checksum", fmt.Sprintf("md5 %s", checksum))

		// Отправляем запрос
		var resp *http.Response
		for retry := 0; retry < maxRetryCount; retry++ {
			resp, err = client.Do(req)
			if err == nil && resp.StatusCode == http.StatusNoContent {
				break
			}

			if resp != nil {
				resp.Body.Close()
			}

			if retry == maxRetryCount-1 {
				return fmt.Errorf("не удалось загрузить блок: %v", err)
			}
		}

		// Обновляем смещение
		offset += int64(n)
		fmt.Printf("Блок загружен: %d/%d байт\n", offset, fileSize)

		// Проверяем статус ответа
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("ошибка загрузки блока: %s", resp.Status)
		}
	}

	return nil
}

// encodeMetadata кодирует метаданные в формат Base64.
func encodeMetadata(metadata map[string]string) string {
	var result string
	for key, value := range metadata {
		encoded := base64.StdEncoding.EncodeToString([]byte(value))
		result += fmt.Sprintf("%s %s,", key, encoded)
	}
	return result[:len(result)-1] // Убираем последнюю запятую
}

// generateChecksum генерирует MD5-хэш данных.
func generateChecksum(data []byte) string {
	hash := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}
