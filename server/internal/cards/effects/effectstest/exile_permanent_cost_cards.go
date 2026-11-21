package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_permanent_cost_cards_test.go — #1600's proof cards for "Exile a
// creature you control" as a cost (ADR 0020's 2026-10-03 amendment):
// The Soul Stone, Food Chain, City of Shadows and Altar of Bhaal. The
// engine half is game/exile_permanent_cost_test.go.
//
// The trigger behaviour is pinned here, with real cards, because it is
// what makes exiling different from sacrificing: Circuit Mender's
// "When this creature leaves the battlefield, draw a card" fires,
// Zulaport Cutthroat's "Whenever this creature or another creature you
// control dies" does not, and Juri's "Whenever you sacrifice a
// permanent" does not.

const (
	TheSoulStoneOracle   = "92cfba68-12f6-4f97-9187-0f6a39656a0f"
	FoodChainOracle      = "5c8e5092-962e-49ef-ab82-8434e475e4e7"
	CityOfShadowsOracle  = "043192d8-6077-46c3-b43f-b7caf6762869"
	circuitMenderOracle  = "1665ca9f-176d-40f1-a4e9-42da4f1236e9"
	JuriRevueOracle      = "3364bad6-6d0a-4141-b410-86e3d9e1916e"
	AltarOfBhaalOracleTC = "0a364b66-95df-480b-a733-e90f6d5c4d2b"
)

// pushCircuitMender seats the LTB-trigger creature with its printed
// mana value ({3}), which Food Chain reads.
func PushCircuitMender(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Circuit Mender", OracleID: circuitMenderOracle,
		TypeLine: "Artifact Creature — Construct", ManaCost: "{3}", Power: 2, Toughness: 3,
		Owner: owner, Controller: owner,
	})
}
