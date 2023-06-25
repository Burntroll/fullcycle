package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Order 					[]*Order
	Transactions 		[]*Transaction
	OrdersChan			chan *Order		//input
	OrdersChanOut 	chan *Order		//output
	Wg							*sync.WaitGroup
}

func NewBook(orderChan chan *Order, orderChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Order:					[]*Order{},
		Transactions: 		[]*Transaction{},
		OrdersChan: 		orderChan,
		OrdersChanOut: 	orderChanOut,
		Wg:							wg,
	}
}
/*
func (b *Book) Trade() {
	buyOrders := NewOrderQueue()
	sellOrders := NewOrderQueue()

	heap.Init(buyOrders)
	heap.Init(sellOrders)

	for order := range b.OrdersChan {
		if order.OrderType == "BUY" {
			buyOrders.Push(order)
			if sellOrders.Len() > 0 && sellOrders.Orders[0].Price <= order.Price {
				sellOrder := sellOrders.Pop().(*Order)
				if sellOrder.PendingShares > 0 {
					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
					b.AddTransaction(transaction, b.Wg)
					sellOrder.Transactions = append(sellOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChanOut <- sellOrder
					b.OrdersChanOut <- order
					if sellOrder.PendingShares > 0 {
						sellOrders.Push(sellOrder)
					}
				}
			}
		} else if order.OrderType == "SELL" {
			sellOrders.Push(order)
			if buyOrders.Len() > 0 && buyOrders.Orders[0].Price >= order.Price {
				buyOrder := buyOrders.Pop().(*Order)
				if buyOrder.PendingShares > 0 {
					transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
					b.AddTransaction(transaction, b.Wg)
					buyOrder.Transactions = append(buyOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChanOut <- buyOrder
					b.OrdersChanOut <- order
					if buyOrder.PendingShares > 0 {
						buyOrders.Push(buyOrder)
					}
				}
			}
		}
	}
}
*/

func (b *Book) Trade() {
	buyOrders := NewOrderQueue()
	sellOrders := NewOrderQueue()

	heap.Init(buyOrders)
	heap.Init(sellOrders)

	for order := range b.OrdersChan {
		switch order.OrderType {
		case "BUY":
			b.processOrder(buyOrders, sellOrders, order, true)
		case "SELL":
			b.processOrder(sellOrders, buyOrders, order, false)
		}
	}
}

func (b *Book) processOrder(ownOrders, oppositeOrders *OrderQueue, order *Order, isBuy bool) {
	ownOrders.Push(order)

	if oppositeOrders.Len() > 0 {
		var condition bool
		if isBuy {
			condition = oppositeOrders.Orders[0].Price <= order.Price
		} else {
			condition = oppositeOrders.Orders[0].Price >= order.Price
		}

		if condition {
			oppositeOrder := oppositeOrders.Pop().(*Order)
			if oppositeOrder.PendingShares > 0 {
				price := order.Price
				if isBuy {
					price = oppositeOrder.Price
				}

				transaction := NewTransaction(oppositeOrder, order, order.Shares, price)
				b.AddTransaction(transaction, b.Wg)
				oppositeOrder.Transactions = append(oppositeOrder.Transactions, transaction)
				order.Transactions = append(order.Transactions, transaction)
				b.OrdersChanOut <- oppositeOrder
				b.OrdersChanOut <- order

				if oppositeOrder.PendingShares > 0 {
					oppositeOrders.Push(oppositeOrder)
				}
			}
		}
	}
}

func (b *Book) AddTransaction(transaction *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	minShares := sellingShares
	if buyingShares < minShares {
		minShares = buyingShares
	}

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	transaction.AddBuyOrderPendingShares(-minShares)
	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.BuyingOrder.Asset.ID, minShares)
	transaction.AddSellOrderPendingShares(-minShares)

	transaction.CalculateTotal(transaction.Shares, transaction.BuyingOrder.Price)

	transaction.CloseBuyOrder()
	transaction.CloseSellOrder()

	b.Transactions = append(b.Transactions, transaction)
}