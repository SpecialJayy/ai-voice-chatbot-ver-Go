package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
)

func main() {
	record()
	transcribe()
}

func record() {
	//I dont know why it doesn't overrite the file, so imma just delete it since it makes a new one anyway
	os.Remove("sample.mp3")
	cmd := exec.Command("ffmpeg", "-f", "dshow", "-i", "audio=Mikrofon (2 — Realtek(R) Audio)", "sample.mp3")
	err := cmd.Run()
	fmt.Println("Mów teraz. Gdy skończysz naciśnik CTRL + C")
	if err != nil {
		log.Fatal(err)
	}
}

func transcribe() {
	const API_KEY = "API_KEY"
	client := aai.NewClient(API_KEY)
	f, err := os.Open("sample.mp3")
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
