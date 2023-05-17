package exchange

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/cache/errors"
	"github.com/mitchellh/mapstructure"
)

type FilterType string

type Filter struct {
	MinPrice          string                   `json:"minPrice,omitempty"`
	MaxPrice          string                   `json:"maxPrice,omitempty"`
	FilterType        futures.SymbolFilterType `json:"filterType,omitempty"`
	TickSize          string                   `json:"tickSize,omitempty"`
	StepSize          string                   `json:"stepSize,omitempty"`
	Limit             string                   `json:"limit,omitempty"`
	Notional          string                   `json:"notional,omitempty"`
	MultiplierDown    string                   `json:"multiplierDown,omitempty"`
	MultiplierUp      string                   `json:"multiplierUp,omitempty"`
	MultiplierDecimal string                   `json:"ultiplierDecimal,omitempty"`
}

type Filters []*Filter

func (f *Filters) Parse(data []map[string]interface{}) {
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   f,
		TagName:  "json",
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(data)
}

type Symbol struct {
	Symbol      string   `json:"symbol,omitempty"`
	Pair        string   `json:"pair,omitempty"`
	Filters     *Filters `json:"filters,omitempty"`
	MarginAsset string   `json:"marginAsset,omitempty"`
	BaseAsset   string   `json:"baseAsset,omitempty"`
}

func (s *Symbol) GetPriceFilter() (*Filter, error) {
	for _, filter := range *s.Filters {
		if filter.FilterType == futures.SymbolFilterTypePrice {
			return filter, nil
		}
	}
	return nil, errors.ErrorFilterNotFound
}

func (s *Symbol) GetLotSizeFilter() (*Filter, error) {
	for _, filter := range *s.Filters {
		if filter.FilterType == futures.SymbolFilterTypeLotSize {
			return filter, nil
		}
	}
	return nil, errors.ErrorFilterNotFound
}
