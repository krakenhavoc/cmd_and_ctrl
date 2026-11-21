package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap3_test.go — card-level coverage for the play-rate slice cut from
// tracker #293's two gap lists after the #2342 / #2358 / #2359 slices.

const (
	G3MangaraOracle         = "cbcb6d9a-6ae5-4bcd-8013-2b657553764a"
	G3IcetillOracle         = "109cdefd-e8cc-4ac7-b6ba-2cfdef8d780f"
	G3PuresteelOracle       = "74a62c7b-4753-4af2-b7a1-9a4ae8988801"
	G3MirarisWakeOracle     = "852657c0-18a4-4b28-b9ae-7728acdb5044"
	G3LifeFromLoamOracle    = "ac8fba34-512c-4f24-999a-ab72f1ce4acb"
	G3FlareOracle           = "4ff07b7c-e97a-4b57-bcf1-2f22c37a8bd6"
	G3FloodcallerOracle     = "4879c8f0-8832-4290-bc71-9838940f75cd"
	G3NivParunOracle        = "33666a98-812f-4892-9f8d-33e0cbecc340"
	G3RangerCaptainOracle   = "cada3481-cc2b-4412-b9b5-0436af53aad2"
	G3CoruscationOracle     = "88bb91b5-2ccd-4ce9-8cd4-e54d63c12abf"
	G3CankerbloomOracle     = "d5b80895-621a-40df-bf48-6c7295658f21"
	G3IrrigatedFarmlandOrcl = "406eabe2-df62-49e2-bb39-c0227509d875"
	G3EscapeTunnelOracle    = "0056fc91-4398-471c-b561-7ff99750ac8a"
	G3SunscorchedOracle     = "8d2b2675-19df-4f40-9e8e-196ec097b91c"
	G3TocasiasOracle        = "25c983e0-a8c9-4784-91a4-8fe04c6df882"
)

func HandHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
