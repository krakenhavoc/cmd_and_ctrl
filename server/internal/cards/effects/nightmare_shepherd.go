package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightmare Shepherd — Enchantment Creature — Demon {2}{B}{B}, 4/4:
//
//	"Flying
//	 Whenever another nontoken creature you control dies, you may exile
//	 it. If you do, create a token that's a copy of that creature,
//	 except it's 1/1 and it's a Nightmare in addition to its other
//	 types."
//
// Hofri Ghostforge's exile-then-copy (cmdr2S47ExileThenCopyTrigger): the
// exile is a "may" asked when the trigger resolves, only while the card is
// still the object that died (CR 400.7), and the copy is the creature's
// copiable values as it last existed (CR 707.2, 608.2h). The exceptions
// are part of the copiable values (CR 707.9a), so a copy of the token is
// a 1/1 Nightmare too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8ee0dfc-f588-43b6-ab96-f3a3c58a8a3d",
		Name:            "Nightmare Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			cmdr2S47ExileThenCopyTrigger(
				"Nightmare Shepherd — exile it; if you do, create a 1/1 Nightmare token copy",
				"Nightmare Shepherd — exile that creature and create a 1/1 token copy of it?",
				"nightmare-shepherd/copy",
				func(t *game.Card) {
					v := game.PrintedValues{TypeLine: t.TypeLine}
					v.AddSubtype("Nightmare")
					t.TypeLine = v.TypeLine
					t.Power, t.Toughness, t.VariableToughness = 1, 1, false
				}),
		},
	})
}
