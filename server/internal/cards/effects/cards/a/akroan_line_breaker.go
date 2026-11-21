package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroan Line Breaker — Creature — Human Warrior {2}{R}, 2/1 (Journey into Nyx, #5558):
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 this creature gets +2/+0 and gains intimidate until end of turn.
//	 (It can't be blocked except by artifact creatures and/or creatures
//	 that share a color with it.)"
//
// Heroic, with the boost and the keyword on the creature itself through
// thisCreatureUntilEOT (nothing if it has left). Intimidate is the
// engine's blocking rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01443d75-3f86-45c0-81fa-fbf1b3175d4f",
		Name:         "Akroan Line Breaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Akroan Line Breaker — +2/+0 and intimidate until end of turn (heroic)",
				effects.ThisCreatureUntilEOT("Akroan Line Breaker — +2/+0 and intimidate until end of turn", 2, 0, "intimidate")),
		},
	})
}
