package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deserter's Disciple — Creature — Human Rebel Ally {1}{R}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"{T}: Another target creature you control with power 2 or less can't
//	 be blocked this turn."
//
// The power is a targeting restriction, so it is checked again as the
// ability resolves: a target pumped to 3 in response makes it fizzle,
// one pumped after it resolved stays unblockable, and one already
// blocked stays blocked (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73ca3132-d849-4cf9-beba-75ba09789332",
		Name:         "Deserter's Disciple",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Another target creature you control with power 2 or less can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapCost(),
			Targets: Another(TargetCreature("another target creature you control with power 2 or less", YouControl(), PowerLE(2))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Deserter's Disciple — can't be blocked",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
