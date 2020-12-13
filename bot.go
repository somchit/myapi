package main


import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	e := echo.New()
	bot, err := linebot.New("4a53e6351aeac37573996c5a6c40e7da", "y0IkmiDnaEfL9auhlP4jPVZWL9GDnxYeiRbdVBF9SsIdta/6/iYooyZRWMhiJPWZiyFttXJrnsOjI/gs2HnNK/NiZEiNdeQMvYnsuePLWNzT0KhekbBRDtNQAPBZvxipIDGiz1/7f4We8X2stY3qM1GUYhWQfeY8sLGRXgo3xvw=")
	if err!=nil {
		panic(err)
	}
	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")

	})
	e.POST("/hello", func(c echo.Context) error {
		r:=c.Request()
		w:=c.Response()
		events, err := bot.ParseRequest(r)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}

		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						fmt.Print(err)
					}
				case *linebot.StickerMessage:
					replyMessage := fmt.Sprintf(
						"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						fmt.Print(err)
					}
				}
			}
		}
		return nil
	})
	e.Logger.Fatal(e.Start(":3000"))
}

func  callbotnoi(keyword string) string {
	url := "https://openapi.botnoi.ai/botnoi/ecommerce?keyword="+keyword
	method := "GET"

	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDgwOTQwMDEsImlkIjoiNDIzMTQ5YTYtYWU4Yy00NDAyLWFmMTgtMGFhYjJjN2ExYmEzIiwiaXNzIjoiMUxRTm9rQ0xWOER6UmdudnozakpoSkNDU0RDVlk0VTQiLCJuYW1lIjoiQShzb21jaGl0KSIsInBpYyI6Imh0dHBzOi8vcHJvZmlsZS5saW5lLXNjZG4ubmV0LzBoVE54Z3RYS0hDMkZGSGlLd3dnQjBObmxiQlF3eU1BMHBQWEJBRG1oT0JWRTdmRWxsZjNwRUJtQWZYQVJzTEVsaWVueE1CbUFjVUFSZyJ9.Kg-74W-zEYiTzhZN2fGuAKm3U4Fi_fGDqsG7jdY0A_k")
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	if(err!=nil){
		return  "unkwon"
	}
	return string(body)
}
