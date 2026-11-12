package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Shard — Artifact {3} (Mirrodin, #5606):
//
//	"{3}, {T} or {U}, {T}: Return target creature to its owner's hand
//	 unless its controller pays {1}."
//
// One ability with a choice of cost, written as two rows (Tidal
// Control's shape) so the activator picks the cost by picking the row;
// both resolve the same way. As the ability resolves, the target
// creature's controller is asked to pay {1} (CR 118.12a, Combustion Man's
// keyed prompt); if they don't, the creature returns to its owner's
// hand. A creature that is no longer a legal target leaves the ability
// with nothing to do (CR 608.2b), and no one is asked.
//
// No simplification.
func init() {
	const effect = "Return target creature to its owner's hand unless its controller pays {1}."
	Register(Spec{
		OracleID:     "cd8ccb10-a2c8-4a84-aa41-6ce6018f9bb7",
		Name:         "Crystal Shard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: " + effect,
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  crystalShardAsk,
			},
			{
				Label:   "{U}, {T}: " + effect,
				Cost:    Plus(ManaCost("{U}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  crystalShardAsk,
			},
		},
	})
}

// crystalShardAsk asks the target creature's controller for {1}.
func crystalShardAsk(g *game.Game, item *game.StackItem) error {
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
			Question: "Crystal Shard — pay {1}, or " + c.Name + " returns to its owner's hand?",
			Else:     crystalShardBounce,
		}.Apply(ctx)
	}
	return nil
}

// crystalShardBounce is the decline: the creature goes to its owner's
// hand. Keyed, an on-disk identity: a table waiting on the payer is a
// restore point.
var crystalShardBounce = PayUnlessElse("pay-unless/crystal-shard-return-to-hand",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})
