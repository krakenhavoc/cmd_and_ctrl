package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_alt_cost_cards_test.go — ADR 0129 PR 4 (#1995): energy as an
// alternative cost (Nissa, Worldsoul Speaker; Primal Prayers; Amped
// Raptor), replicate (Reiterating Bolt) and a keyword's cost (Inventor's
// Axe's equip, Salvation Colossus's unearth), each through the real
// catalog and the real cast or activation path.

const (
	NissaWorldsoulOracle    = "00037840-6089-42ec-8c5c-281f9f474504"
	PrimalPrayersOracle     = "8d087fe0-d554-4d7c-ba22-32db2cf71887"
	AmpedRaptorOracle       = "3ed14ea7-ee62-4604-af3b-deafb00108c3"
	ReiteratingBoltOracle   = "371da0d8-3845-458e-9b0a-6477017b890e"
	InventorsAxeOracle      = "364fc1a7-b91d-470b-80fe-275828f36cfc"
	SalvationColossusOracle = "8b5c6754-36bd-44ac-b0c0-5a3cd2d9f2c3"
)

// eacHandCard puts a plain card in p's hand.
func EacHandCard(g *game.Game, p *game.Player, name, typeLine, mana string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: mana,
			Owner: p.ID, Controller: p.ID})
	})
	return id
}
