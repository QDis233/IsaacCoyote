package main

import (
	"IsaacCoyote/common/config"
	"IsaacCoyote/common/game"
	"IsaacCoyote/common/isaac"
	"IsaacCoyote/common/logging"
	"IsaacCoyote/pkg/coyote"
	"IsaacCoyote/pkg/coyote/enums"
	"IsaacCoyote/util"
	"fmt"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer fmt.Scanln()

	configM, err := config.NewConfigManager("config.yaml")
	if err != nil {
		zap.L().Error("初始化配置失败", zap.Error(err))
		return
	}
	err = configM.Init()
	if err != nil {
		zap.L().Error("获取配置文件失败", zap.Error(err))
		return
	}

	_, err = logging.ApplyNewLogger(configM.GetConfig().Debug)
	if err != nil {
		zap.L().Error("初始化日志失败", zap.Error(err))
		return
	}

	coyoteConfig := coyote.Config{
		Address: configM.GetConfig().Coyote.Address,
		Port:    configM.GetConfig().Coyote.Port,
	}
	if configM.GetConfig().Coyote.Address == "" {
		localAddressList, err := util.GetLocalIP()
		if err != nil || len(localAddressList) == 0 {
			zap.L().Error("获取IP失败, 请手动填写ip", zap.Error(err))
			return
		}
		if len(localAddressList) > 1 {
			zap.L().Error("或许你有多个地址, 请手动指定")
			zap.L().Info("ip", zap.Any("ips", localAddressList))
			return
		}

		coyoteConfig.Address = localAddressList[0]
	}
	c := coyote.NewCoyote(&coyoteConfig)
	go func() {
		err = c.Run()
		if err != nil {
			zap.L().Panic("Coyote Service Error", zap.Error(err))
			return
		}
	}()

	coyoteSession := c.NewSession()
	_ = util.PrintTerminalQRCode(coyoteSession.GetQRCodeContent())
	err = util.ShowQRCode("qrcode.png", coyoteSession.GetQRCodeContent())
	if err != nil {
		zap.L().Error("获取二维码失败", zap.Error(err))
		return
	}

	zap.L().Info("等待连接...... 请使用使用 DG-LAB app 扫码二维码")
	coyoteSession.RegisterCallback(enums.OnSessionBind, func(session *coyote.Session, callbackData coyote.CallbackData[any]) {
		zap.L().Info("连接成功")
	})
	coyoteSession.WaitForBind()

	isaacListener := isaac.NewGameListener()
	go func() {
		for {
			err = isaacListener.Run()
			if err != nil {
				zap.L().Error("Isaac Service Error", zap.Error(err))
				return
			}
		}
	}()

	coyoteGame := game.NewGame(&configM.GetConfig().Game, coyoteSession, isaacListener)
	err = coyoteGame.Run()
	if err != nil {
		zap.L().Error("Game Service Error", zap.Error(err))
		return
	}
}
