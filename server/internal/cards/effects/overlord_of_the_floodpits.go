package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overlord of the Floodpits — Enchantment Creature — Avatar Horror
// {3}{U}{U}, 5/3 (Duskmourn: House of Horror, #4466):
//
//	"Impending 4—{1}{U}{U} (If you cast this spell for its impending
//	 cost, it enters with four time counters and isn't a creature
//	 until the last is removed. At the beginning of your end step,
//	 remove a time counter from it.)
//	 Flying
//	 Whenever this permanent enters or attacks, draw two cards, then
//	 discard a card."
//
// Impending is the shared constructor (impending.go); the card adds
// flying and the loot, whose discard prompt is built from the hand after
// both draws (b16DrawThenDiscard), so a drawn card may be the discard.
//
// No simplification.
func init() {
	Register(Impending(4, "{1}{U}{U}", Spec{
		OracleID:        "31196f0b-33ee-4981-ac97-24226e8b405f",
		Name:            "Overlord of the Floodpits",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Overlord of the Floodpits — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 2)
				}),
		},
	}))
}
