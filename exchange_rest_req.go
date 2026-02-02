package myhyperliquidapi

type ExchangeReqCommon[T any] struct {
	Action       T          `json:"action" msgpack:"action"`
	Nonce        *uint64    `json:"nonce" msgpack:"nonce"`
	Signature    *Signature `json:"signature" msgpack:"signature"`
	VaultAddress *string    `json:"vaultAddress,omitempty" msgpack:"vaultAddress,omitempty"`
	ExpiresAfter *uint64    `json:"expiresAfter,omitempty" msgpack:"expiresAfter,omitempty"`
}

type ExchangeOrderAPI struct {
	client *ExchangeRestClient
	req    *ExchangeReqCommon[ExchangeOrderAction]
}

type Order struct {
	Asset         *int       `json:"a" msgpack:"a"`                     // asset 资产索引（通常是一个整数，代表特定的币种，如 0 可能代表 BTC）。
	IsBuy         *bool      `json:"b" msgpack:"b"`                     // isBuy true 为买入，false 为卖出。
	Price         *string    `json:"p" msgpack:"p"`                     // price 价格。注意这里是 String，为了保证高精度，避免浮点数误差。
	Size          *string    `json:"s" msgpack:"s"`                     // size 数量。同样是 String。
	ReduceOnly    *bool      `json:"r" msgpack:"r"`                     // reduceOnly 只减仓。如果为 true，该订单只会平掉现有仓位，不会开新仓。
	OrderType     *OrderType `json:"t" msgpack:"t"`                     // 订单类型
	ClientOrderId *string    `json:"c,omitempty" msgpack:"c,omitempty"` // cloid (client order id)
}

type OrderType struct {
	Limit   *OrderTypeLimit   `json:"limit,omitempty" msgpack:"limit,omitempty"`     // 限价单
	Trigger *OrderTypeTrigger `json:"trigger,omitempty" msgpack:"trigger,omitempty"` // 触发单/止盈止损
}

type OrderTypeLimit struct {
	Tif *string `json:"tif" msgpack:"tif"` // tif, "Alo" | "Ioc" | "Gtc Alo: 只有作为挂单（Maker）时才生效（Post-only）。Ioc: 立即成交否则取消。Gtc: 一直有效直到成交或取消。
}

type OrderTypeTrigger struct {
	TriggerPx *string `json:"triggerPx" msgpack:"triggerPx"` // triggerPx
	IsMarket  *bool   `json:"isMarket" msgpack:"isMarket"`
	Tpsl      *string `json:"tpsl" msgpack:"tpsl"` // tpsl
}

type OrderBuilder struct {
	order *Order
}

func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		order: &Order{},
	}
}

func (ob *OrderBuilder) Build() Order {
	// Manual deep copy to avoid side effects when reusing builder
	dst := Order{}
	src := ob.order

	if src.Asset != nil {
		dst.Asset = GetPointer(*src.Asset)
	}
	if src.IsBuy != nil {
		dst.IsBuy = GetPointer(*src.IsBuy)
	}
	if src.Price != nil {
		dst.Price = GetPointer(*src.Price)
	}
	if src.Size != nil {
		dst.Size = GetPointer(*src.Size)
	}
	if src.ReduceOnly != nil {
		dst.ReduceOnly = GetPointer(*src.ReduceOnly)
	}
	if src.ClientOrderId != nil {
		dst.ClientOrderId = GetPointer(*src.ClientOrderId)
	}

	if src.OrderType != nil {
		dst.OrderType = &OrderType{}
		if src.OrderType.Limit != nil {
			dst.OrderType.Limit = &OrderTypeLimit{}
			if src.OrderType.Limit.Tif != nil {
				dst.OrderType.Limit.Tif = GetPointer(*src.OrderType.Limit.Tif)
			}
		}
		if src.OrderType.Trigger != nil {
			dst.OrderType.Trigger = &OrderTypeTrigger{}
			if src.OrderType.Trigger.TriggerPx != nil {
				dst.OrderType.Trigger.TriggerPx = GetPointer(*src.OrderType.Trigger.TriggerPx)
			}
			if src.OrderType.Trigger.IsMarket != nil {
				dst.OrderType.Trigger.IsMarket = GetPointer(*src.OrderType.Trigger.IsMarket)
			}
			if src.OrderType.Trigger.Tpsl != nil {
				dst.OrderType.Trigger.Tpsl = GetPointer(*src.OrderType.Trigger.Tpsl)
			}
		}
	}
	return dst
}

// asset 资产索引（通常是一个整数，代表特定的币种，如 0 可能代表 BTC）。
func (ob *OrderBuilder) Asset(asset int) *OrderBuilder {
	ob.order.Asset = GetPointer(asset)
	return ob
}

// isBuy true 为买入，false 为卖出。
func (ob *OrderBuilder) IsBuy(isBuy bool) *OrderBuilder {
	ob.order.IsBuy = GetPointer(isBuy)
	return ob
}

// price 价格。注意这里是 String，为了保证高精度，避免浮点数误差。
func (ob *OrderBuilder) Price(price string) *OrderBuilder {
	ob.order.Price = GetPointer(price)
	return ob
}

// size 数量。同样是 String。
func (ob *OrderBuilder) Size(size string) *OrderBuilder {
	ob.order.Size = GetPointer(size)
	return ob
}

// reduceOnly 只减仓。如果为 true，该订单只会平掉现有仓位，不会开新仓。
func (ob *OrderBuilder) ReduceOnly(reduceOnly bool) *OrderBuilder {
	ob.order.ReduceOnly = GetPointer(reduceOnly)
	return ob
}

// cloid (client order id)
func (ob *OrderBuilder) ClientOrderId(cloid string) *OrderBuilder {
	ob.order.ClientOrderId = GetPointer(cloid)
	return ob
}

// ensureOrderType 确保 OrderType 已初始化
func (ob *OrderBuilder) ensureOrderType() {
	if ob.order.OrderType == nil {
		ob.order.OrderType = &OrderType{}
	}
}

// ensureLimit 确保 Limit 已初始化，并清除 Trigger 以防冲突
func (ob *OrderBuilder) ensureLimit() {
	ob.ensureOrderType()
	if ob.order.OrderType.Limit == nil {
		ob.order.OrderType.Limit = &OrderTypeLimit{}
	}
	// 互斥处理：如果是 Limit 单，不应该有 Trigger 属性
	ob.order.OrderType.Trigger = nil
}

// ensureTrigger 确保 Trigger 已初始化，并清除 Limit 以防冲突
func (ob *OrderBuilder) ensureTrigger() {
	ob.ensureOrderType()
	if ob.order.OrderType.Trigger == nil {
		ob.order.OrderType.Trigger = &OrderTypeTrigger{}
	}
	// 互斥处理：如果是 Trigger 单，不应该有 Limit 属性
	ob.order.OrderType.Limit = nil
}

// tif, "Alo" | "Ioc" | "Gtc Alo: 只有作为挂单（Maker）时才生效（Post-only）。Ioc: 立即成交否则取消。Gtc: 一直有效直到成交或取消。
func (ob *OrderBuilder) LimitTif(tif string) *OrderBuilder {
	ob.ensureLimit()
	ob.order.OrderType.Limit.Tif = GetPointer(tif)
	return ob
}

// triggerPx isMarket tpsl
func (ob *OrderBuilder) Trigger(isMarket bool, triggerPx string, tpsl string) *OrderBuilder {
	ob.ensureTrigger()
	ob.order.OrderType.Trigger.IsMarket = GetPointer(isMarket)
	ob.order.OrderType.Trigger.TriggerPx = GetPointer(triggerPx)
	ob.order.OrderType.Trigger.Tpsl = GetPointer(tpsl)
	return ob
}

type ExchangeOrderAction struct {
	Type     *string                     `json:"type" msgpack:"type"`
	Orders   *[]Order                    `json:"orders" msgpack:"orders"`
	Grouping *string                     `json:"grouping,omitempty" msgpack:"grouping,omitempty"` // "na" | "normalTpsl" | "positionTpsl", 订单分组策略，用于处理止盈止损与主仓位的关联逻辑。
	Builder  *ExchangeOrderActionBuilder `json:"builder,omitempty" msgpack:"builder,omitempty"`   // Optional
}

type ExchangeOrderActionBuilder struct {
	B *string `json:"b" msgpack:"b"` // the address the should receive the additional fee 接收返佣/费用的地址。
	F *int    `json:"f" msgpack:"f"` // the size of the fee in tenths of a basis point e.g. if f is 10, 1bp of the order notional  will be charged to the user and sent to the builder 费用比例（单位是 0.1 基点）。例如填 10 代表 1 个基点（0.01%）。
}

func (api *ExchangeOrderAPI) Type(t string) *ExchangeOrderAPI {
	api.req.Action.Type = GetPointer(t)
	return api
}

func (api *ExchangeOrderAPI) Grouping(grouping string) *ExchangeOrderAPI {
	api.req.Action.Grouping = GetPointer(grouping)
	return api
}

func (api *ExchangeOrderAPI) Builder(b string, f int) *ExchangeOrderAPI {
	api.req.Action.Builder = &ExchangeOrderActionBuilder{
		B: GetPointer(b),
		F: GetPointer(f),
	}
	return api
}

func (api *ExchangeOrderAPI) AddOrders(orders ...Order) *ExchangeOrderAPI {
	if api.req == nil {
		api.req = &ExchangeReqCommon[ExchangeOrderAction]{}
	}
	if api.req.Action.Orders == nil {
		api.req.Action.Orders = &[]Order{}
	}
	*api.req.Action.Orders = append(*api.req.Action.Orders, orders...)
	return api
}

type ExchangeCancelAPI struct {
	client *ExchangeRestClient
	req    *ExchangeReqCommon[ExchangeCancelAction]
}

type ExchangeCancelAction struct {
	Type    *string   `json:"type" msgpack:"type"`
	Cancels *[]Cancel `json:"cancels" msgpack:"cancels"`
}

type Cancel struct {
	Asset *int `json:"a" msgpack:"a"` // a is asset
	Oid   *int `json:"o" msgpack:"o"` // o is oid (order id)
}

func (api *ExchangeCancelAPI) AddCancelOrder(cancel Cancel) *ExchangeCancelAPI {
	if api.req == nil {
		api.req = &ExchangeReqCommon[ExchangeCancelAction]{}
	}
	if api.req.Action.Cancels == nil {
		api.req.Action.Cancels = &[]Cancel{}
	}
	*api.req.Action.Cancels = append(*api.req.Action.Cancels, cancel)
	return api
}

type ExchangeCancelByCloidAPI struct {
	client *ExchangeRestClient
	req    *ExchangeReqCommon[ExchangeCancelByCloidAction]
}

type CancelByCloid struct {
	Asset *int    `json:"asset" msgpack:"asset"`
	Cloid *string `json:"cloid" msgpack:"cloid"`
}

type ExchangeCancelByCloidAction struct {
	Type    *string          `json:"type" msgpack:"type"`
	Cancels *[]CancelByCloid `json:"cancels" msgpack:"cancels"`
}

func (api *ExchangeCancelByCloidAPI) AddCancelByCloid(cancel CancelByCloid) *ExchangeCancelByCloidAPI {
	if api.req == nil {
		api.req = &ExchangeReqCommon[ExchangeCancelByCloidAction]{}
	}
	if api.req.Action.Cancels == nil {
		api.req.Action.Cancels = &[]CancelByCloid{}
	}
	*api.req.Action.Cancels = append(*api.req.Action.Cancels, cancel)
	return api
}

type ExchangeBatchModifyAPI struct {
	client *ExchangeRestClient
	req    *ExchangeReqCommon[ExchangeBatchModifyAction]
}

type ExchangeBatchModifyAction struct {
	Type     *string   `json:"type" msgpack:"type"`
	Modifies *[]Modify `json:"modifies" msgpack:"modifies"`
}

type Modify struct {
	Oid   any    `json:"oid" msgpack:"oid"` // number, cloid(string, hex)
	Order *Order `json:"order" msgpack:"order"`
}

type ModifyOption func(*Modify)

func (api *ExchangeBatchModifyAPI) OidOption(oid int, order Order) ModifyOption {
	return func(m *Modify) {
		m.Oid = oid
		m.Order = GetPointer(order)
	}
}

func (api *ExchangeBatchModifyAPI) CloidOption(cloid string, order Order) ModifyOption {
	return func(m *Modify) {
		m.Oid = cloid
		m.Order = GetPointer(order)
	}
}

func (api *ExchangeBatchModifyAPI) AddModify(options ...ModifyOption) *ExchangeBatchModifyAPI {
	if api.req == nil {
		api.req = &ExchangeReqCommon[ExchangeBatchModifyAction]{}
	}
	if api.req.Action.Modifies == nil {
		api.req.Action.Modifies = &[]Modify{}
	}

	modify := Modify{}
	for _, opt := range options {
		opt(&modify)
	}

	*api.req.Action.Modifies = append(*api.req.Action.Modifies, modify)
	return api
}

type ExchangeUpdateLeverageAPI struct {
	client *ExchangeRestClient
	req    *ExchangeReqCommon[ExchangeUpdateLeverageAction]
}

type ExchangeUpdateLeverageAction struct {
	Type     *string `json:"type" msgpack:"type"`
	Asset    *int    `json:"asset" msgpack:"asset"`       // index of coin
	IsCross  *bool   `json:"isCross" msgpack:"isCross"`   // true or false if updating cross-leverage
	Leverage *int    `json:"leverage" msgpack:"leverage"` // integer representing new leverage, subject to leverage constraints on that coin
}

func (api *ExchangeUpdateLeverageAPI) Asset(asset int) *ExchangeUpdateLeverageAPI {
	api.req.Action.Asset = GetPointer(asset)
	return api
}

func (api *ExchangeUpdateLeverageAPI) IsCross(isCross bool) *ExchangeUpdateLeverageAPI {
	api.req.Action.IsCross = GetPointer(isCross)
	return api
}

func (api *ExchangeUpdateLeverageAPI) Leverage(leverage int) *ExchangeUpdateLeverageAPI {
	api.req.Action.Leverage = GetPointer(leverage)
	return api
}
