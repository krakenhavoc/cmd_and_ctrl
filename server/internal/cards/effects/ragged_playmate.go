package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ragged Playmate — Artifact Creature — Toy {1}{R}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"{1}, {T}: Target creature with power 2 or less can't be blocked this
//	 turn."
//
// Access Tunnel's ability on a creature, with the power cap at 2: the
// restriction (RestrictUntilEOT, CantBeBlocked) goes on the target if it
// is still legal as the ability resolves (CR 608.2b), and lasts until the
// cleanup step. The {T} in the cost means a Playmate that came under your
// control this turn can't activate it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0732ebf-91ad-4e3c-a09a-f410d20587fc",
		Name:         "Ragged Playmate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Ragged Playmate — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
