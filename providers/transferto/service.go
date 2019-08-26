package transferto

import (
	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/mailroom/goflow"
	"github.com/nyaruka/mailroom/models"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	loginConfig    = "TRANSFERTO_ACCOUNT_LOGIN"
	apiTokenConfig = "TRANSFERTO_AIRTIME_API_TOKEN"
	currencyConfig = "TRANSFERTO_ACCOUNT_CURRENCY"
)

func init() {
	goflow.SetAirtimeService(&transferTo{})
}

type transferTo struct{}

func (s *transferTo) Transfer(session flows.Session, sender urns.URN, recipient urns.URN, amounts map[string]decimal.Decimal) (*flows.AirtimeTransfer, error) {
	t := &flows.AirtimeTransfer{
		Sender:    sender,
		Recipient: recipient,
		Status:    flows.AirtimeTransferStatusFailed,
	}

	org := session.Assets().Source().(*models.OrgAssets).Org()
	login := org.ConfigValue(loginConfig, "")
	token := org.ConfigValue(apiTokenConfig, "")
	currency := org.ConfigValue(currencyConfig, "")

	if login == "" || token == "" {
		return t, errors.New("missing transferto configuration")
	}

	client := NewClient(login, token, session.Engine().HTTPClient())

	info, err := client.MSISDNInfo(recipient.Path(), currency, "1")
	if err != nil {
		return t, err
	}

	t.Currency = info.DestinationCurrency

	// look up the amount to send in this currency
	amount, hasAmount := amounts[t.Currency]
	if !hasAmount {
		return t, errors.Errorf("no amount configured for transfers in %s", t.Currency)
	}

	t.DesiredAmount = amount

	if info.OpenRange {
		// TODO add support for open-range topups once we can find numbers to test this with
		// see https://shop.transferto.com/shop/v3/doc/TransferTo_API_OR.pdf
		return nil, errors.Errorf("transferto account is configured for open-range which is not yet supported")
	}

	// find the product closest to our desired amount
	var useProduct string
	useAmount := decimal.Zero
	for p, product := range info.ProductList {
		price := info.LocalInfoValueList[p]
		if price.GreaterThan(useAmount) && price.LessThanOrEqual(amount) {
			useProduct = product
			useAmount = price
		}
	}
	t.ActualAmount = useAmount

	reservedID, err := client.ReserveID()
	if err != nil {
		return t, err
	}

	topup, err := client.Topup(reservedID, sender.Path(), recipient.Path(), useProduct, "")
	if err != nil {
		return t, err
	}

	t.ActualAmount = topup.ActualProductSent
	t.Status = flows.AirtimeTransferStatusSuccess
	return t, nil
}
