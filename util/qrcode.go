package util

import (
	"github.com/skip2/go-qrcode"
	"os"
)

func PrintTerminalQRCode(content string) error {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString(qr.ToSmallString(false))
	if err != nil {
		return err
	}
	return nil
}
