package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aura Gnarlid — Creature — Beast {2}{G}, 2/2 (Wilds of Eldraine
// Commander, #5368):
//
//	"Creatures with power less than this creature's power can't block it.
//	 This creature gets +1/+1 for each Aura on the battlefield."
//
// The pump is a layer 7c modify counting every Aura on the battlefield,
// any player's, and an Aura attached to the Gnarlid itself included
// (cmdr2B19AurasOnTheBattlefield). The block restriction compares the
// blocker's power with the Gnarlid's as it is when the block is declared,
// so the pump counts (cmdr2B19LesserPowerThanAttacker); a creature with
// power equal to or greater than the Gnarlid's may still block it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "586aebf2-4653-49fa-8425-13bd1f511bf0",
		Name:         "Aura Gnarlid",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each Aura on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := cmdr2B19AurasOnTheBattlefield(g)
				c.Power += n
				c.Toughness += n
			},
		}},
		BlockRules: []game.BlockRule{
			CantBeBlockedByComparing(OnSelf(), cmdr2B19LesserPowerThanAttacker(),
				"creatures with power less than Aura Gnarlid's"),
		},
	})
}
