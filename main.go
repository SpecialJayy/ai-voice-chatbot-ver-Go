package main

import (
	"bytes"
	"context"
	"fmt"
	darkman "github.com/JKGplay/piper-voice-darkman"
	"github.com/amitybell/piper"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

var (
	apiKey        = os.Getenv("OPENAI_API_KEY")
	client        = openai.NewClient(apiKey)
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

	time.Sleep(1 * time.Second)

	i = MCIWorker("save capture audio/mic.wav", "", 0, 0)
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
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "audio/" + fileName,
	}
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return ""
	}
	fmt.Println(resp.Text)
	return resp.Text
}

func sendQueryToChatGpt(query string) (string, error) {
	modifiedQuery := query + "Odpowiedz po polsku. Używaj prostego języka. Nie rób błędów ortograficznych. Odpowiedź napisz jak najzwięźlej potrafisz. Odpowiedź podaj maksymalnie w trzech zdaniach."
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Nazywasz się \"Miś Mądrala\". Jesteś przyjaznym misiem, który odpowiada dzieciom na różne pytania. Zapytany o to kim jesteś odpowiesz, że jesteś \"Misiem Mądralą\".",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: modifiedQuery,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func piperAndPlayResponse(text string) error {
	tts, err := piper.New("", darkman.Asset)
	if err != nil {
		return err
	}
	wavBytes, err := tts.Synthesize(text)
	if err != nil {
		return err
	}
	r := bytes.NewReader(wavBytes)
	streamer, format, err := wav.Decode(r)
	if err != nil {
		return err
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done

	return nil
}

func convert(inputFileName, outputFileName string) error {
	cmd := exec.Command("LAME/lame.exe", inputFileName, outputFileName)
	err := cmd.Run()
	return err
}

func main() {
	record()
	err := convert("audio/mic.wav", "audio/mic.mp3")
	if err != nil {
		fmt.Println(err)
		return
	}
	query := transcribe("mic.mp3")
	answer, err := sendQueryToChatGpt(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(answer)
	err = piperAndPlayResponse(answer)
	if err != nil {
		fmt.Println(err)
		return
	}
}
