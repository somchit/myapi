package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"myapi/thqr"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/linebot"

	gim "myapi/goimagemerge"

	"github.com/pbnjay/pixfont"
)

type  Intent struct {
//{"intent": "\u0e04\u0e38\u0e22\u0e17\u0e31\u0e48\u0e27\u0e44\u0e1b", "category": "\u0e04\u0e38\u0e22\u0e17\u0e31\u0e48\u0e27\u0e44\u0e1b",
//	"confidence": 0.25108722782374965}

	//Type               MessageType      `json:"type"`
	//OriginalContentURL string           `json:"originalContentUrl"`
	//PreviewImageURL    string           `json:"previewImageUrl"`
	//QuickReply         *QuickReplyItems `json:"quickReply,omitempty"`
	//Sender             *Sender          `json:"sender,omitempty"`

    Intent  string `json:"intent"`
	Category string `json:"category"`
	Confidence float64 `json:"confidence"`
}

type Order struct {
	UserId string
	Name  string
	Price float64
	Quantity int
}

var orders =make(map[string]map[string]*Order)

func getUserOrder(userId string) map[string]*Order  {
	if o,ok:=orders[userId]; ok {
		return o
	}else{
		orders[userId] = make(map[string]*Order)
		return orders[userId]
	}
}

func addProduct(order *Order)  {
	o:=getUserOrder(order.UserId)

	if p, ok := o[order.Name]; ok {
		p.Quantity = p.Quantity + 1
		o[order.Name] = p
		orders[order.UserId] = o
	} else {
		o[order.Name] = order
		orders[order.UserId] = o
	}

}

func addOrdder(order *Order)  {
	if o,ok:=orders[order.UserId]; ok {
		if p, ok := o[order.Name]; ok {
			p.Quantity = p.Quantity + 1
			o[order.Name] = p
			orders[order.UserId] = o
		} else {
			o[order.Name] = order
			orders[order.UserId] = o
		}
	}
}

func removeOrder(order *Order)  {
	if o,ok:=orders[order.UserId]; ok {
		if _, ok := o[order.Name]; ok {
			delete(o, order.Name)
			orders[order.UserId] = o
		}
	}
}

func deleteOrder(order *Order)  {
	if o,ok:=orders[order.UserId]; ok {
		if p, ok := o[order.Name]; ok {
			if p.Quantity > 1 {
				p.Quantity = p.Quantity - 1
				o[order.Name] = p
				orders[order.UserId] = o
			} else {
				delete(o, order.Name)
				orders[order.UserId] = o
			}
		}
	}
}

func clearOrder(order *Order)  {
	if _,ok:=orders[order.UserId]; ok {
		delete(orders,order.UserId)
	}
}

func generateImgBase()(image.Image,error)  {
	base := image.NewRGBA(image.Rectangle{Max: image.Point{X: 512, Y: 800}})
	green := color.RGBA{11, 53, 102, 255}
	draw.Draw(base, base.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
  	return base,nil
}

func generateImgQr(qr []byte)(image.Image,error)  {
	//var img image.Image
	readqr:= bytes.NewReader(qr)
	//img, _ = png.Decode(readqr)
	return png.Decode(readqr)
}

func addLabel(img *image.RGBA, x, y int, label string) {
	//col := color.RGBA{200, 100, 0, 255}
	//point := fixed.Point26_6{fixed.Int26_6(x * 128), fixed.Int26_6(y * 128)}
	//
	//d := &font.Drawer{
	//	Dst:  img,
	//	Src:  image.NewUniform(col),
	//	Face: basicfont.Face7x13,
	//	Dot:  point,
	//}
	//d.DrawString(label)

	pixfont.DrawString(img, x, y, label, color.Black)
}
//
//func generateHeader() image.Image {
//	base := image.NewRGBA(image.Rectangle{Max: image.Point{X: 512, Y: 800}})
//	green := color.RGBA{11, 53, 102, 0}
//	draw.Draw(base, base.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
//	return base
//}


func generateFooter(name string,id string,amt float64) image.Image {
	base := image.NewRGBA(image.Rectangle{Max: image.Point{X: 512, Y: 100}})
	green := color.RGBA{255, 255, 255, 255}
	draw.Draw(base, base.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
	addLabel(base,20,20,"Merchant: "+ name)
	addLabel(base,20,40,"ACC/PPID: "+ id)
	addLabel(base,20,60,"Amount: "+ fmt.Sprintf("%0.2f",amt))
	return base
}


func generateImage(qr []byte,amount float64) ([]byte,error) {
	//m := image.NewRGBA(image.Rect(0, 0, 512, 800))
	//blue := color.RGBA{255, 255, 255, 255}
	//draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	//draw.
	var qrimg image.Image
	var baseimg image.Image
	qrimg,err:=generateImgQr(qr)
	if err!=nil{
		fmt.Println("step1")
		fmt.Println("%s",err.Error())
		return nil,err
	}

	baseimg,err=generateImgBase()
	if err!=nil{
		fmt.Println("step2")
		fmt.Println("%s",err.Error())
		return nil,err
	}

	footerimg:=generateFooter("keedi","0639749444",amount)
	//header:=generateHeader()

	grids := []*gim.Grid{
		{
			Image: &baseimg,
			Grids: []*gim.Grid{
				//{
				//	Image:&header,
				//},
				{
					ImageFilePath:   "./qr-header.png",
					OffsetX: 90,OffsetY: 0,
					//113566
					//BackgroundColor: color.RGBA{R: 0x11, G: 0x35, B: 0x66},
				},
				{
					Image:&qrimg,
					OffsetX: 0,OffsetY: 130,

				},
				{
					Image: &footerimg,
					OffsetX: 0,OffsetY: 642 ,
				},
			},
		},

	}

	rgba, err := gim.New(grids, 1, 2,
		//gim.OptGridSizeFromNthImageSize(1)
	).Merge()

	if err != nil {
		fmt.Println("step3")
		return nil,err
	}
	var b bytes.Buffer
	wr := bufio.NewWriter(&b)
	err = png.Encode(wr, rgba)
	if err != nil {
		fmt.Println("step4")
		return nil,err
	}

	return b.Bytes(),nil
}



func main() {
	e := echo.New()
	//xxx:="#add [eeeee] to cart price: [11.00] quantity: [2]"
	//xxx=strings.TrimSpace(xxx)
	//fmt.Println( strings.HasPrefix(xxx,"#"))
	//if strings.Index(xxx,"#")==1 {
	//	fmt.Println("xxxxxxxxx")
	//}
	//xxx=strings.ReplaceAll(xxx," ","")
	//xxx=strings.ReplaceAll(xxx,"to cart price:","")
	//xxx=strings.ReplaceAll(xxx,"quantity:","")
	//xxx=strings.ReplaceAll(xxx,"quantity:","")
	//xxx=strings.ReplaceAll(xxx,"]","")
	//xzz:=strings.Split(xxx,"[")
	//fmt.Println(xxx)
	//fmt.Println(len(xzz))


	bot, err := linebot.New("dda1710de68044f0703cb1a94c68f5fd", "Vj4z/S1ZiAD/7266F9z9OeWiJJj85PvreiPHt0ARHM/NSA4G4EKV+iSlLCitbwcouoNAFAVw283sMJYd9F3yRBUikBoHvZxJ90kAcu4w2CSEGKxV+1/kpJXkPLKwiCQbnooTNQrXddy57V7G2s6IvQdB04t89/1O/w1cDnyilFU=")
	if err!=nil {
		panic(err)
	}
	e.GET("/api", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")

	})

	e.GET("/api/qr/:ppid", func(c echo.Context) error {
		ppid:=c.Param("ppid")
		qr:=thqr.GeneratePayload(ppid, 0)
		fmt.Println(qr)
		pngqr, err := qrcode.Encode(qr, qrcode.High, 512)
		if err!=nil{
			return err
		}
		//var img image.Image
		//readqr:= bytes.NewReader(pngqr)
		//img, _ = png.Decode(readqr)
		//grids := []*gim.Grid{
		//	{
		//		ImageFilePath:   "./qr-header.png",
		//		BackgroundColor: color.White,
		//		OffsetX: 50, OffsetY: 20,
		//	},
		//	{
		//		Image:&img,
		//		OffsetX: 0, OffsetY: 0,
		//		//ImageFilePath:   "./cmd/gim/input/ginger.png",
		//		//BackgroundColor: color.RGBA{R: 0x8b, G: 0xd0, B: 0xc6},
		//	},
		//}
		//rgba, err := gim.New(grids, 1, 2,
		//	//gim.OptGridSizeFromNthImageSize(1)
		//).Merge()
		//if err != nil {
		//	panic(err)
		//}
		//var b bytes.Buffer
		//wr := bufio.NewWriter(&b)
		//
		////var buffs []byte
		////wr:= io.ByteWriter(buffs)
		//err = png.Encode(wr, rgba)
		//if err != nil {
		//	panic(err)
		//}

	    out,err:=	generateImage(pngqr,0)
		if err!=nil{
			return err
		}
		c.Response().Header().Set("Content-Type", "image/png")
		c.Response().Write(out)
		return nil
	})

	e.GET("/api/qr/:ppid/:amount", func(c echo.Context) error {
		ppid:=c.Param("ppid")
		//amount:=c.Param("amount")
		amount, err := strconv.ParseFloat(c.Param("amount"), 32)
		if err != nil {
			return err
		}
		qr:=thqr.GeneratePayload(ppid, float32(amount))
		fmt.Println(qr)
		png, err := qrcode.Encode(qr, qrcode.High, 512)
		out,err:=	generateImage(png,amount)
		if err!=nil{
			return err
		}
		c.Response().Header().Set("Content-Type", "image/png")
		c.Response().Write(out)
		return nil
	})

	e.POST("/api/bot", func(c echo.Context) error {
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

		//fmt.Println(events)

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				fmt.Println(event.Message)
				fmt.Println(event.Source)
				//fmt.Println(event.)
				//fmt.Println(event.Members.)
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					userMsg:=message.Text
					//event.Source.UserID
					userMsg=strings.TrimSpace(userMsg)
					if strings.HasPrefix(userMsg,"#") {
						//#add ["+payload.name+"] to cart price: ["+payload.price+"] quantity: ["+payload.quantity+"]
						//#add["+payload.name+"["+payload.price+"["+payload.quantity+"
						userMsg=strings.ReplaceAll(userMsg," ","")
						userMsg=strings.ReplaceAll(userMsg,"tocart","")
						userMsg=strings.ReplaceAll(userMsg,"price:","")
						userMsg=strings.ReplaceAll(userMsg,"quantity:","")
						userMsg=strings.ReplaceAll(userMsg,"]","")
						cmds:=strings.Split(userMsg,"[")

						order:=&Order{
							UserId: event.Source.UserID,
							Name: "",
							Price: 0.0,
							Quantity: 1,
						}

						if cmds[0]=="#add"{
						    order.Name=cmds[1]
							if price, err := strconv.ParseFloat(cmds[2], 32); err == nil {
								order.Price=price
							}
							if q, err := strconv.Atoi(cmds[3]); err == nil {
								order.Quantity=q
							}
							addOrdder(order)
						}else if cmds[0]=="#addproduct"{
							order.Name=cmds[1]
							if price, err := strconv.ParseFloat(cmds[2], 32); err == nil {
								order.Price=price
							}
							if q, err := strconv.Atoi(cmds[3]); err == nil {
								order.Quantity=q
							}
							addProduct(order)
						} else if cmds[0]=="#del"{
							order.Name=cmds[1]
							if price, err := strconv.ParseFloat(cmds[2], 32); err == nil {
								order.Price=price
							}
							if q, err := strconv.Atoi(cmds[3]); err == nil {
								order.Quantity=q
							}
							deleteOrder(order)
						} else if cmds[0]=="#remove"{
							order.Name=cmds[1]
							removeOrder(order)
						}else if cmds[0]=="#clear"{
							clearOrder(order)
						}else if cmds[0]=="#checkout"{
							userOrder:=getUserOrder(event.Source.UserID)
							detail,amount:= sendOrder(userOrder)
							fmt.Println(detail)
							//linebot.F
							bd:=[]byte(detail)
							var content linebot.FlexContainer
							cx:=linebot.BubbleContainer{}
							json.Unmarshal(bd,&cx)
							content=&cx
							//cx:=content.(*linebot.FlexContainer)
							//linebot.NewFlexMessage()
							//vp := reflect.New(reflect.TypeOf(content))
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage("your order",content)).Do(); err != nil {
								fmt.Print(err)
							}


							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("https://2546d9k6x1.execute-api.ap-southeast-1.amazonaws.com/api/qr/0639749444/"+fmt.Sprintf("%0.2f",amount))).Do(); err != nil {
								fmt.Print(err)
							}
							delete(orders,event.Source.UserID)

						}

					}else{
						instan, err := callbotnoi(message.Text)
						msg := ""
						if err != nil {
							msg = message.Text
						} else {
							msg = fmt.Sprintf("%s  %s  %f", instan.Intent, instan.Category, instan.Confidence)
						}

						if fmt.Sprintf("%s",instan.Category)=="สินค้าและบริการ" {
							//linebot.NewURIAction("สินค้าและบริการ","https://liff.line.me/1655374042-rm1G3A8M")
							//linebot.new
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("https://liff.line.me/1655374042-rm1G3A8M")).Do(); err != nil {
								fmt.Print(err)
							}
						}else {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
								fmt.Print(err)
							}
						}
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

func sendOrder(userOrder map[string]*Order)  (string,float64) {
	orderDetail:=`{
  "type": "bubble",
  "body": {
    "type": "box",
    "layout": "vertical",
    "contents": [
      {
        "type": "text",
        "text": "You Order",
        "weight": "bold",
        "color": "#1DB446",
        "size": "sm"
      },
      {
        "type": "text",
        "text": "Keedi Store",
        "weight": "bold",
        "size": "xxl",
        "margin": "md"
      },
      {
        "type": "text",
        "text": "keedi Tower, 4-1-6 , bangkok",
        "size": "xs",
        "color": "#aaaaaa",
        "wrap": true
      },
      {
        "type": "separator",
        "margin": "xxl"
      },
      {
        "type": "box",
        "layout": "vertical",
        "margin": "xxl",
        "spacing": "sm",
        "contents": [
			%s
          ,{
            "type": "separator",
            "margin": "xxl"
          },
    	   %s
      ,{
        "type": "separator",
        "margin": "xxl"
      },
      {
        "type": "box",
        "layout": "horizontal",
        "margin": "md",
        "contents": [
          {
            "type": "text",
            "text": "Order ID",
            "size": "xs",
            "color": "#aaaaaa",
            "flex": 0
          },
          {
            "type": "text",
            "text": "#743289384279",
            "color": "#aaaaaa",
            "size": "xs",
            "align": "end"
          }
        ]
      }
    ]
	}
]
  },
  "styles": {
    "footer": {
      "separator": true
    }
  }
}`
	var sum = 0.0
	var details=""
	var items=0
	for _, value := range userOrder {
		detail,total:=generateDetail(value)
		sum=sum+total
		items=items+value.Quantity
		if details!="" {
			details = detail +","+ details
		}else{
			details=detail
		}
	}

	return fmt.Sprintf(orderDetail,details,footerbox(items,sum)),sum
}


func generateDetail(order *Order) (string,float64) {
	detail:=`{
		"type": "box",
		"layout": "horizontal",
		"contents": [
		{
			"type": "text",
			"text": "%s",
			"size": "sm",
			"color": "#555555",
			"flex": 0
		},
		{
			"type": "text",
			"text": "$%0.2f",
			"size": "sm",
			"color": "#111111",
			"align": "end"
		}
		]
	}`
	var t = order.Price * float64(order.Quantity)
	return fmt.Sprintf(detail,order.Name,t),t
}

func footerbox(item int,total float64) string  {
	footer:=`{
            "type": "box",
            "layout": "horizontal",
            "margin": "xxl",
            "contents": [
              {
                "type": "text",
                "text": "ITEMS",
                "size": "sm",
                "color": "#555555"
              },
              {
                "type": "text",
                "text": "%d",
                "size": "sm",
                "color": "#111111",
                "align": "end"
              }
            ]
          },{
            "type": "box",
            "layout": "horizontal",
            "contents": [
              {
                "type": "text",
                "text": "TOTAL",
                "size": "sm",
                "color": "#555555"
              },
              {
                "type": "text",
                "text": "$%0.2f",
                "size": "sm",
                "color": "#111111",
                "align": "end"
              }
            ]
     }`
	return fmt.Sprintf(footer,item,total)
}

func  callbotnoi(keyword string) (*Intent,error) {
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
	//fmt.Println(string(body))

	if err!=nil {
		return nil,err
	}

	var dat Intent

	if err := json.Unmarshal(body, &dat); err != nil {
		return nil,err
	}
	fmt.Println(dat)
	return &dat,nil
}
