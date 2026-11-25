package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unraveling Mummy — {1}{W}{B} Creature — Zombie 2/3 (Hour of Devastation):
//
//	"{1}{W}: Target attacking Zombie gains lifelink until end of turn.
//	 {1}{B}: Target attacking Zombie gains deathtouch until end of turn."
//
// Each ability names its own target, judged on the stack as well (CR 608.2b).
func init() {
	grant := func(cost, keyword string) effects.ActivatedAbility {
		return effects.ActivatedAbility{
			Label:   cost + ": Target attacking Zombie gains " + keyword + " until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, keyword)),
			Cost:    effects.ManaCost(cost),
			Targets: effects.TargetCreature("target attacking Zombie", effects.AttackingCreature(), effects.Subtype("Zombie")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: "Unraveling Mummy — " + keyword}.Apply(ctx)
			},
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "ea30f395-86fc-444f-8c9e-b9876287084e",
		Name:         "Unraveling Mummy",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{grant("{1}{W}", "lifelink"), grant("{1}{B}", "deathtouch")},
	})
}
