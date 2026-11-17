package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argothian Enchantress — Creature — Human Druid {1}{G}, 0/1 (Urza's
// Saga, #5624):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Whenever you cast an enchantment spell, draw a card."
//
// Satyr Enchanter's cast trigger (enchantmentSpellCastByYou): the spell's
// type is read off the stack and the draw resolves before the enchantment
// does. Shroud is the printed keyword the targeting gate reads
// (CR 702.18a), so not even its controller can target it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11f17f85-ca97-4551-838f-7cb32f0e5f10",
		Name:            "Argothian Enchantress",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, enchantmentSpellCastByYou, "Argothian Enchantress — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
