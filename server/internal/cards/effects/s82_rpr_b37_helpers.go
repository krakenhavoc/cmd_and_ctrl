package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ProducedFromOwnPermanentColors is Meteor Crater's "Choose a color of a
// permanent you control. Add one mana of that color.": the colours of the
// permanents its controller controls, read as they are now (layered view),
// offered as one pipe slot the player picks from. A board with no coloured
// permanent offers nothing, and the ability adds no mana (the card's ruling).
func ProducedFromOwnPermanentColors() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		seen := map[string]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			for _, col := range c.EffectiveColors() {
				seen[strings.ToUpper(col)] = true
			}
		}
		return PipeString(ColorsOnly(OrderedManaSymbols(seen)))
	}
}

// rprBasicLandCard is "basic land card": a land with the basic supertype
// (CR 205.4a), judged on the card where it is (a hand card here).
func RprBasicLandCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() && c.HasSupertype("basic")
	}
}
