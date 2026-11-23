package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosnakht, Heir of Rohgahh — Legendary Creature — Kobold Warrior {R}, 0/1
// (Dominaria United Commander, S82):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 Heroic — Whenever you cast a spell that targets Rosnakht, create a 0/1
//	 red Kobold creature token named Kobolds of Kher Keep."
//
// Battle cry is the shared keyword trigger. Heroic reads the spell's
// targets as it is cast (youCastASpellThatTargetsThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86f7d513-1f5f-4551-83e6-56d24da5ab39",
		Name:         "Rosnakht, Heir of Rohgahh",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.BattleCry("Rosnakht, Heir of Rohgahh"),
			effects.On(game.EventCast, effects.YouCastASpellThatTargetsThis,
				"Rosnakht, Heir of Rohgahh — create a 0/1 Kobolds of Kher Keep (heroic)",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 red Kobolds of Kher Keep"), N: 1})),
		},
	})
}
