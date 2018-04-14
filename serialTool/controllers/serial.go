package controllers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/astaxie/beego"
	"github.com/jacobsa/go-serial/serial"
)

var Serials map[string]io.ReadWriteCloser

var ReceiverChan map[string](chan string)

func init() {
	Serials = make(map[string]io.ReadWriteCloser)
	ReceiverChan = make(map[string](chan string))

	go func() {
		tmp := make([]byte, 1024)
		for {
			for k, v := range Serials {
				if n, err := v.Read(tmp); (n > 0) && (err == nil) {
					if _, ok := ReceiverChan[k]; ok == false {
						ReceiverChan[k] = make(chan string, 100)
					}

					ReceiverChan[k] <- string(tmp[:n])
				}
			}
		}
	}()
}

type SerialOpenController struct {
	beego.Controller
}

type SerialCloseController struct {
	beego.Controller
}

type SerialSendController struct {
	beego.Controller
}

type SerialReceiveController struct {
	beego.Controller
}

func (s *SerialOpenController) Post() {
	id := s.Ctx.Input.Param(":id")
	fmt.Println(id)
	if _, ok := Serials[id]; ok == true {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "id error"}
		s.ServeJSON()
		return
	}

	var opt serial.OpenOptions
	var err error
	if err = json.Unmarshal(s.Ctx.Input.RequestBody, &opt); err != nil {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial open fail"}
		s.ServeJSON()
		return
	}

	var tserial io.ReadWriteCloser
	tserial, err = serial.Open(opt)
	if err != nil {
		fmt.Println(err)
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial open fail"}
		s.ServeJSON()
		return
	}
	Serials[id] = tserial
	s.Data["json"] = map[string]interface{}{"code": 0, "message": "serial open success"}
	s.ServeJSON()
}

func (s *SerialCloseController) Post() {
	id := s.Ctx.Input.Param(":id")
	if _, ok := Serials[id]; ok == false {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "id error"}
		s.ServeJSON()
		return
	}
	if err := Serials[id].Close(); err != nil {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial close fail"}
		s.ServeJSON()
		return
	}
	delete(Serials, id)
	s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial close success"}
	s.ServeJSON()
}

func (s *SerialSendController) Post() {
	id := s.Ctx.Input.Param(":id")
	if _, ok := Serials[id]; ok == false {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "id error"}
		s.ServeJSON()
		return
	}
	var data = struct {
		Type int //发送数据类型
		Data string
	}{}
	if err := json.Unmarshal(s.Ctx.Input.RequestBody, &data); err != nil {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial send fail"}
		s.ServeJSON()
		return
	}
	if _, err := Serials[id].Write([]byte(data.Data)); err != nil {
		delete(Serials, id)
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "serial send fail"}
		s.ServeJSON()
		return
	}
	s.Data["json"] = map[string]interface{}{"code": 0, "message": "serial send success"}
	s.ServeJSON()
}

func (s *SerialReceiveController) Post() {
	id := s.Ctx.Input.Param(":id")
	if _, ok := Serials[id]; ok == false {
		s.Data["json"] = map[string]interface{}{"code": 1, "message": "id error"}
		s.ServeJSON()
		return
	}

	var data string
	if _, ok := ReceiverChan[id]; ok == true {
		data = <-ReceiverChan[id]

	}
	s.Data["json"] = map[string]interface{}{"code": 0, "message": "serial receive success", "data": data}
	s.ServeJSON()
}
