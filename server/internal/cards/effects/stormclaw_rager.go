package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormclaw Rager — Creature — Ogre Warrior {1}{B}{R}, 2/2 (March of the
// Machine, S82 MOM batch 13):
//
//	"{1}, Sacrifice another creature or artifact: Put a +1/+1 counter on
//	 this creature and draw a card. Activate only as a sorcery."
//
// Acolyte of Aclazotz's "another creature or artifact" sacrifice cost
// (the Rager can't pay with itself), at sorcery speed. The counter and the
// draw are both on resolution, in that order; the counter is skipped if the
// Rager has left, but the draw still happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1a77708-7ddf-4645-bf1e-1d8147a57f23",
		Name:         "Stormclaw Rager",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}, Sacrifice another creature or artifact: Put a +1/+1 counter on this creature and draw a card. Activate only as a sorcery.",
			Purpose:      game.Purpose{Draws: 1},
			SorcerySpeed: true,
			Cost: game.AbilityCost{
				Mana: "{1}",
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, DrawCards{N: 1}),
		}},
	})
}
