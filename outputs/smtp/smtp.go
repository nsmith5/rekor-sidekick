package smtp

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nsmith5/rekor-sidekick/outputs"
	"gopkg.in/gomail.v2"
)

const (
	driverName = `smtp`
)

type impl struct {
	Host string
	Port int

	Username string
	Password string

	From string
	To   []string
}

func (i *impl) Send(e outputs.Event) error {
	d := gomail.NewDialer(i.Host, i.Port, i.Username, i.Password)

	m := gomail.NewMessage()
	m.SetHeader("From", i.From)
	m.SetHeader("To", i.To...)
	m.SetHeader("Subject", "rekor-sidekick alert")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("smtp: err", err)
		return err
	}

	return nil
}

func (i *impl) Name() string {
	return driverName
}

func New(config map[string]interface{}) (outputs.Output, error) {
	var i impl
	err := mapstructure.Decode(config, &i)
	if err != nil {
		return nil, err
	}

	fmt.Printf("smtp: driver %#v\n", i)
	return &i, nil
}

func init() {
	outputs.RegisterDriver(driverName, outputs.CreatorFunc(New))
}
