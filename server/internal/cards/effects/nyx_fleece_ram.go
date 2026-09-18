package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyx-Fleece Ram — Enchantment Creature — Sheep {1}{W}, 0/5 (EDHREC
// rank 4545):
//
//	"At the beginning of your upkeep, you gain 1 life."
//
// Two mana for a wall that blocks almost everything in the early
// game and drips a life a turn. Commander plays it in two seats: a
// lifegain deck that wants a free trigger every upkeep, and an
// enchantress deck that wants a cheap ENCHANTMENT that also happens
// to be a creature — constellation counts it, Eidolon of Blossoms
// draws off it, and the 0/5 body is a bonus.
//
// The upkeep trigger is AtYourUpkeep, so it uses the stack and can be
// responded to, as printed. The enchantment type rides the card's
// type line from Scryfall, which is what lets the batch's own Boon of
// the Spirit Realm see it enter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a20113c-7ea8-4edf-af9f-148ccd326a88",
		Name:         "Nyx-Fleece Ram",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Nyx-Fleece Ram — you gain 1 life", func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
