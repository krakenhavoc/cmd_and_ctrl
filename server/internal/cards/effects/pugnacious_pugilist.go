package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pugnacious Pugilist — Creature — Ogre Warrior {3}{R}{R}, 4/4:
//
//	"Whenever this creature attacks, create a tapped and attacking 1/1
//	 red Devil creature token with "When this token dies, it deals 1
//	 damage to any target."
//	 Blitz {3}{R}"
//
// Blitz (CR 702.152, #3772). The Devil is Spiked Corridor's catalog
// token, entering tapped and attacking the player the Pugilist is
// attacking (Skyknight Vanguard's shape). It was never declared as an
// attacker, so "whenever a creature attacks" does not see it (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7c22e6ba-5570-460b-962e-b7c0fbb3f394",
		Name:             "Pugnacious Pugilist",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pugnacious Pugilist — a tapped and attacking Devil",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					tmpl := devilToken()
					tmpl.Tapped = true
					tmpl.AttackingTarget = skyknightVanguardDefender(g, item.SourceCardID)
					return CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(ctx)
				}),
		},
	})
}
