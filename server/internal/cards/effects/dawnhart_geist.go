package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnhart Geist — Creature — Spirit Warlock {1}{W}, 1/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever you cast an enchantment spell, you gain 2 life."
//
// Argothian Enchantress's cast condition over a life gain. It fires on
// casting, so a countered enchantment still gains the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed7a2ee0-289d-4e5f-92e0-1539395baf33",
		Name:         "Dawnhart Geist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, enchantmentSpellCastByYou, "Dawnhart Geist — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
