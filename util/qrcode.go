package util

import (
	"github.com/skip2/go-qrcode"
	"os"
	"os/exec"
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

func ShowQRCode(fileName string, content string) error {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return err
	}
	err = qr.WriteFile(512, fileName)
	if err != nil {
		return err
	}
	return exec.Command("cmd", "/c", "start", "", fileName).Run()
}
