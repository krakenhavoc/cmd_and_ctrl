package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erratic Portal — Artifact {4} (Mystery Booster 2):
//
//	"{1}, {T}: Return target creature to its owner's hand unless its
//	 controller pays {1}."
//
// Crystal Shard's single-mana shape: the creature's controller is asked as
// the ability resolves, and declining returns it (CR 118.12). The bounce
// branch is Crystal Shard's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a85cd7e-625b-4043-b7f1-999def26eb36",
		Name:         "Erratic Portal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Return target creature to its owner's hand unless its controller pays {1}.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  erraticPortalAsk,
		}},
	})
}

func erraticPortalAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return PayUnless{
			Chooser:  c.Controller,
			Cost:     "{1}",
			Question: "Erratic Portal — pay {1}, or " + c.Name + " returns to its owner's hand?",
			Else:     crystalShardBounce,
		}.Apply(ctx)
	}
	return nil
}
