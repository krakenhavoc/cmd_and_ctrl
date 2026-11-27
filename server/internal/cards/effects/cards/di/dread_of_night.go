package cardsdi

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread of Night — Enchantment {B} (Tempest, S82 OLD b756, #5628):
//
//	"White creatures get -1/-1."
//
// A layer 7c anthem (CR 613.4c) over every white creature at the table, the
// controller's own included. Colour is read through the layers and "white" is a
// word of the card's text (CR 612.2), so a white creature made blue is spared and
// a blue one made white is not. A creature left at 0 toughness dies to the
// state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b431a73b-cf4c-4154-810f-82b7c6099e66",
		Name:         "Dread of Night",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && slices.Contains(target.EffectiveColors(), source.TextColor("W"))
			}, -1, -1),
		},
	})
}
