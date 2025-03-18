package web

import (
	"fmt"
	"github.com/v03413/bepusdt/app/config"
	"github.com/v03413/bepusdt/app/help"
	"github.com/v03413/bepusdt/app/log"
	"github.com/v03413/bepusdt/app/model"
	"github.com/v03413/bepusdt/app/rate"
	"sync"
	"time"
)

func buildOrder(money float64, apiType, payAddress, orderId, tradeType, redirectUrl, notifyUrl, name string) (model.TradeOrders, error) {
	var lock sync.Mutex
	var order model.TradeOrders

	// 暂时先强制使用互斥锁，后续有需求的话再考虑优化
	lock.Lock()
	defer lock.Unlock()

	// 获取兑换汇率
	var calcRate = rate.GetUsdtCalcRate(config.DefaultUsdtCnyRate)
	if tradeType == model.OrderTradeTypeTronTrx {

		calcRate = rate.GetTrxCalcRate(config.DefaultTrxCnyRate)
	}

	// 获取钱包地址
	var wallet = model.GetAvailableAddress(payAddress, tradeType)
	if len(wallet) == 0 {
		log.Error("订单创建失败：还没有配置收款地址")

		return order, fmt.Errorf("还没有配置收款地址")
	}

	// 计算交易金额
	address, amount := model.CalcTradeAmount(wallet, calcRate, money, tradeType)
	tradeId, err := help.GenerateTradeId()
	if err != nil {

		return order, err
	}

	// 创建交易订单
	var expiredAt = time.Now().Add(config.GetExpireTime() * time.Second)
	var tradeOrder = model.TradeOrders{
		OrderId:     orderId,
		TradeId:     tradeId,
		TradeHash:   tradeId, // 这里默认填充一个本地交易ID，等支付成功后再更新为实际交易哈希
		TradeType:   tradeType,
		TradeRate:   fmt.Sprintf("%v", calcRate),
		Amount:      amount,
		Money:       money,
		Address:     address.Address,
		Status:      model.OrderStatusWaiting,
		Name:        name,
		ApiType:     apiType,
		ReturnUrl:   redirectUrl,
		NotifyUrl:   notifyUrl,
		NotifyNum:   0,
		NotifyState: model.OrderNotifyStateFail,
		ExpiredAt:   expiredAt,
	}
	var res = model.DB.Create(&tradeOrder)
	if res.Error != nil {
		log.Error("订单创建失败：", res.Error.Error())

		return order, res.Error
	}

	return tradeOrder, nil
}
