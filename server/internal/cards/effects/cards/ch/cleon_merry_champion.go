package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleon, Merry Champion — Legendary Creature — Human Soldier {2}{R}, 2/2
// (Foundations Jumpstart):
//
//	"Double strike
//	 Heroic — Whenever you cast a spell that targets Cleon, exile the top
//	 card of your library. You may play that card until the end of your
//	 next turn."
//
// Heroic is the cast trigger in shield_families_recipient.go (CR 207.2c);
// the impulse is Alania's Pathmaker's seat-turn window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e0e5060-2be0-476b-8108-c8c7cd76942f",
		Name:            "Cleon, Merry Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Cleon, Merry Champion — exile the top card of your library; until the end of your next turn, you may play it (heroic)", effects.Do(effects.Instruction(effects.BnsB50ImpulseTopUntilYourNextTurnEnds))),
		},
	})
}
