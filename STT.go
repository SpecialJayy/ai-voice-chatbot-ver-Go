package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"unsafe"
	"syscall"
	"time"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
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

func main() {
	record()
	transcribe()
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

	time.Sleep(10 * time.Second)

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

func transcribe() {
	const API_KEY = "API_KEY"
	client := aai.NewClient(API_KEY)
	f, err := os.Open("mic.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	transcript, err := client.Transcripts.TranscribeFromReader(context.TODO(), f, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*transcript.Text)
}