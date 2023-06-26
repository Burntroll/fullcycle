package entity

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuyAsset(t *testing.T) {
	asset1 := NewAsset("asset1", "Asset 1", 100)

	investor := NewInvestor("1")
	investor2 := NewInvestor("2")

	investorAssetPosition := NewInvestorAssetPosition("asset1", 10)
	investor.AddAssetPosition(investorAssetPosition)

	wg := sync.WaitGroup{}
	orderChan := make(chan *Order)
	orderChanOut := make(chan *Order)

	book := NewBook(orderChan, orderChanOut, &wg)
	go book.Trade()

	//Add buy order
	wg.Add(1)
	order := NewOrder("1", investor, asset1, 5, 5, "SELL")
	orderChan <- order

	//Add sell order

	order2 := NewOrder("2", investor2, asset1, 5, 5, "BUY")
	orderChan <- order2
	wg.Wait()

	assert := assert.New(t)
	assert.Equal("CLOSED", order.Status, "Order 1 should be closed")
	assert.Equal(0, order.PendingShares, "Order 1 should have 0 PendingShares")
	assert.Equal("CLOSED", order2.Status, "Order 2 should be closed")
	assert.Equal(0, order2.PendingShares, "Order 2 should have 0 PendingShares")

	assert.Equal(5, investorAssetPosition.Shares, "Investor 1 should have 5 shares of asset 1")
	assert.Equal(5, investor2.GetAssetPosition("asset1").Shares, "Investor 2 should have 5 shares of asset 1")
}

func TestBuyAssetWithDifferentAssents(t *testing.T) {
	asset1 := NewAsset("asset1", "Asset 1", 100)
	asset2 := NewAsset("asset2", "Asset 2", 100)

	investor := NewInvestor("1")
	investor2 := NewInvestor("2")

	investorAssetPosition := NewInvestorAssetPosition("asset1", 10)
	investor.AddAssetPosition(investorAssetPosition)

	investorAssetPosition2 := NewInvestorAssetPosition("asset2", 10)
	investor2.AddAssetPosition(investorAssetPosition2)

	wg := sync.WaitGroup{}
	orderChan := make(chan *Order)
	orderChanOut := make(chan *Order)

	book := NewBook(orderChan, orderChanOut, &wg)
	go book.Trade()

	order := NewOrder("1", investor, asset1, 5, 5, "SELL")
	orderChan <- order

	order2 := NewOrder("2", investor2, asset2, 5, 5, "BUY")
	orderChan <- order2

	assert := assert.New(t)
	assert.Equal("OPEN", order.Status, "Order 1 should be closed")
	assert.Equal(5, order.PendingShares, "Order 1 should have 5 PendingShares")
	assert.Equal("OPEN", order2.Status, "Order 2 should be closed")
	assert.Equal(5, order2.PendingShares, "Order 2 should have 5 PendingShares")
}