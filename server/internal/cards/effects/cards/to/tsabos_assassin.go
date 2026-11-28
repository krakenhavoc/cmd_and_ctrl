package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tsabo's Assassin — Creature — Phyrexian Zombie Assassin {2}{B}{B}, 1/1
// (Invasion, #5616):
//
//	"{T}: Destroy target creature if it shares a color with the most common
//	 color among all permanents or a color tied for most common. A creature
//	 destroyed this way can't be regenerated."
//
// The condition is read as the ability resolves, through the layered colours
// (SharesAMostCommonColor, #6207): a creature that changed colour in response
// is judged as it is then, and a colourless one shares nothing. The target is
// any creature; the test is not part of the target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d00c44b-8953-4b18-af73-2944321b6cd1",
		Name:         "Tsabo's Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature if it shares a color with the most common color among all permanents or a color tied for most common.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  tsabosAssassinDestroy,
		}},
	})
}

func tsabosAssassinDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok || !effects.SharesAMostCommonColor(g, &c) {
		return nil
	}
	return effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}.Apply(ctx)
}
