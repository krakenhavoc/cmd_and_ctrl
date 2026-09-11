package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warren Soultrader — Creature — Zombie Goblin Wizard {2}{B}, 3/3
// (EDHREC rank 416):
//
//	"Pay 1 life, Sacrifice another creature: Create a Treasure token."
//
// A free sacrifice outlet that pays you back in mana — the
// aristocrats deck's Phyrexian Altar on a body. Two cost components,
// life and a sacrifice, both paid at announce (so the sacrificed
// creature's dies-triggers resolve above and before the Treasure).
//
// Sandbox simplification: "ANOTHER creature" is enforced by NAME —
// the sacrifice clause excludes any creature named Warren Soultrader
// — because TargetSpec.CardOK never sees the source and so cannot
// exclude it by instance (#350 lists the "another/other" gap). In a
// singleton format that is the same creature; the one divergence is
// that a token COPY of the Soultrader could not be fed to it either,
// which is weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID: "ace86e56-efde-4eb7-8815-71456a4c3abe",
		Name:     "Warren Soultrader",
		Activated: []ActivatedAbility{{
			Label: "Pay 1 life, Sacrifice another creature: Create a Treasure token.",
			Cost: Plus(PayLife(1), game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature", Creature(), b03NotNamed("Warren Soultrader")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   TreasureToken(),
					N:          1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
