package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Compost — Enchantment {1}{G} (Secret Lair Drop):
//
//	"Whenever a black card is put into an opponent's graveyard from
//	 anywhere, you may draw a card."
//
// One trigger per card, not per batch. The watch is Robot Domination's over
// every way a card reaches a graveyard (a death, a discard, a mill, a
// countered spell, a spell resolving), narrowed to a black card whose owner
// (its graveyard being its owner's) is an opponent of Compost's controller. A
// token is not a card, and a card already in a graveyard that moves to another
// is not put into one from anywhere.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4681bcc2-8a46-44cd-8b04-6ab140904dea",
		Name:         "Compost",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.OnAny(
				[]game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !CardPutIntoAGraveyard(ev, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Owner != source.Controller && c.HasColorWord("B")
				},
				"Compost — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Compost — draw a card?", 1),
		},
	})
}

// CardPutIntoAGraveyard is "a card is put into a graveyard from anywhere",
// whatever its types: CreatureCardPutIntoAGraveyard without the creature test.
func CardPutIntoAGraveyard(ev game.Event, g *game.Game) bool {
	return effects.OldB720CardPutIntoAGraveyard(ev, g)
}
