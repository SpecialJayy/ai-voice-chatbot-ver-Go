package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hegedustibor/htgo-tts/voices"
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/go-resty/resty/v2"
	"github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/handlers"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

var (
	winmm         = syscall.MustLoadDLL("winmm.dll")
	mciSendString = winmm.MustFindProc("mciSendStringW")
)

func MCIWorker(lpstrCommand string, lpstrReturnString string, uReturnLength int, hwndCallback int) uintptr {
	i, _, _ := mciSendString.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpstrCommand))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpstrReturnString))),
		uintptr(uReturnLength), uintptr(hwndCallback))
	return i
}

func record() {
	fmt.Println("winmm.dll Record Audio to .wav file")

	i := MCIWorker("open new type waveaudio alias capture", "", 0, 0)
	if i != 0 {
		log.Fatal("Error Code A: ", i)
	}

	i = MCIWorker("record capture", "", 0, 0)
	if i != 0 {
		log.Fatal("Error Code B: ", i)
	}

	fmt.Println("Listening...")
	fmt.Println("Press any key to stop listening")
	fmt.Scanln()

	//time.Sleep(10 * time.Second)

	i = MCIWorker("save capture mic.wav", "", 0, 0)
	if i != 0 {
		log.Fatal("Error Code C: ", i)
	}

	i = MCIWorker("close capture", "", 0, 0)
	if i != 0 {
		log.Fatal("Error Code D: ", i)
	}

	fmt.Println("Audio saved to mic.wav")
}

func transcribe(fileName string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	connection := openai.NewClient(apiKey)
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: fileName,
	}
	resp, err := connection.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return ""
	}
	fmt.Println(resp.Text)
	return resp.Text
}

func sendQueryToChatGpt(query string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := resty.New()

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []interface{}{map[string]interface{}{
				"role":    "system",
				"content": query,
			}},
		}).
		Post(apiEndpoint)

	if err != nil {
		return "", err
	}

	body := response.Body()

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", err
	}

	content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return content, nil
}

func convertTextToAudioAndSaveMp3ToLocation(text string, location string) error {
	speech := htgotts.Speech{Folder: location, Language: voices.English, Handler: &handlers.Native{}}
	err := speech.Speak(text)
	return err
}

func main() {
	record()
	query := transcribe("mic.wav")
	answer, err := sendQueryToChatGpt(query)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(answer)
	err = convertTextToAudioAndSaveMp3ToLocation(answer, "audio")
	if err != nil {
		fmt.Println(err)
	}
}
