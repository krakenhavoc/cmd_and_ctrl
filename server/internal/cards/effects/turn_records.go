package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_records.go — #6217: the card-side builders over the engine's turn
// records (game/turn_records.go).

// CantCastSharingAColorWithTheLastSpell is Mana Maze's "Players can't cast
// spells that share a color with the spell most recently cast this turn":
// a cast restriction on every player, read against the spell most
// recently cast this turn as it was cast (CR 105.2). A colourless spell
// shares a colour with nothing, either way round, and before the turn's
// first spell nothing is refused. The spell being cast is read as it
// would be on the stack (game.SourceCharacteristics).
func CantCastSharingAColorWithTheLastSpell(label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			last, ok := q.Game.LastSpellCastColorsThisTurnForEffect()
			if !ok || len(last) == 0 {
				return false
			}
			card := q.Card
			ch := game.SourceCharacteristics(&card)
			if ch == nil {
				return false
			}
			for _, c := range ch.Colors {
				if HasFold(last, c) {
					return true
				}
			}
			return false
		},
	}
}

// DiscardTheLastCardYouDrewThisTurn is "Discard the last card you drew this
// turn" as an activated ability's cost (Jandor's Ring): the one card it
// takes is the last card the activator drew this turn, while it is still
// in their hand (game.DiscardCost.LastDrawn). Combine it with the rest:
//
//	Plus(ManaCost("{2}"), TapCost(), DiscardTheLastCardYouDrewThisTurn())
func DiscardTheLastCardYouDrewThisTurn() game.AbilityCost {
	c := DiscardN(1, "the last card you drew this turn")
	c.DiscardCards.LastDrawn = true
	return c
}
