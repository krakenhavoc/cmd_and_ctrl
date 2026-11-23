package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_colour_lki_test.go — #1689, CR 603.10a: a dies /
// leaves-the-battlefield condition is judged on the permanent's
// COLOUR as it last existed on the battlefield, not the card's colour
// in its new zone. Reader: leftAsColor. Teysa, Orzhov Scion's
// "whenever another black creature you control dies" is the only
// catalog condition that tested a departed permanent's colour.

// colorByEffectOracle is a test oracle whose static paints the
// permanent black — a layer-5 colour change (CR 105.3), the shape
// Darkest Hour gives every creature its controller controls.
const colorByEffectOracle = "test-1689-cards-color-by-effect"

// paintedNonBlackOracle is a test oracle whose static paints the
// permanent white — a printed black creature an effect has made
// another colour, the reverse direction #1689 also has to get right.
const PaintedNonBlackOracle = "test-1689-cards-painted-non-black"

func init() {
	paint := func(color string) func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
		return func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Colors = []string{color}
		}
	}
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	}
	effects.Register(effects.Spec{
		OracleID: colorByEffectOracle,
		Name:     "Test Green Creature Painted Black",
		Static: []game.StaticAbility{{
			Layer: game.Layer5Color, AppliesTo: self, Apply: paint("B"),
		}},
	})
	effects.Register(effects.Spec{
		OracleID: PaintedNonBlackOracle,
		Name:     "Test Black Creature Painted White",
		Static: []game.StaticAbility{{
			Layer: game.Layer5Color, AppliesTo: self, Apply: paint("W"),
		}},
	})
}

// pushCardsColorByEffect seeds a printed-green creature whose own
// static paints it black, and checks the layer took.
func PushCardsColorByEffect(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: "Test Green Beast", TypeLine: "Creature — Beast",
		OracleID: colorByEffectOracle, Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	id := PushBattlefieldCardWithTimestamp(g, c)
	g.ReadSnapshot(func() {})
	if card, ok := BattlefieldCard(g, id); !ok || !card.HasColor("B") {
		t.Fatal("setup: the static did not paint the creature black")
	}
	return id
}
