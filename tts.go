package main

import (
	"context"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"time"
)

var (
	apiKey = os.Getenv("OPENAI_API_KEY")
	client = openai.NewClient(apiKey)
)

func sendQueryToChatGpt(query string) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func convertTextToAudioAndSaveMp3ToLocation(text string, location string) error {
	resp, err := client.CreateSpeech(
		context.Background(),
		openai.CreateSpeechRequest{
			Model: openai.TTSModel1,
			Input: text,
			Voice: openai.VoiceEcho,
		},
	)
	if err != nil {
		return err
	}
	f, err := os.Create(location + "/response.mp3")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.ReadCloser)
	if err != nil {
		return err
	}
	return nil
}

func playMp3Response() error {
	f, err := os.Open("audio/response.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done

	return nil
}

func main() {
	query := "Opisz zasady gry \"Papier Kamień Nożyce\" najbardziej zwięźle jak potrafisz"
	answer, err := sendQueryToChatGpt(query)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(answer)
	err = convertTextToAudioAndSaveMp3ToLocation(answer, "audio")
	if err != nil {
		fmt.Println(err)
	}
	err = playMp3Response()
	if err != nil {
		fmt.Println(err)
	}
}
