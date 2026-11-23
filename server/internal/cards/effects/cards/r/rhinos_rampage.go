package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhino's Rampage — {R/G} Sorcery (Marvel Super Heroes, #4455):
//
//	"Target creature you control gets +1/+0 until end of turn. It fights
//	 target creature an opponent controls. When excess damage is dealt to the
//	 creature an opponent controls this way, destroy up to one target
//	 noncreature artifact with mana value 3 or less."
//
// The fight and the excess are The Last Agni Kai's (b10Fight; the lethal
// amount is read before the fight and the damage dealt to the victim after it,
// CR 120.4a, 701.14), and the pump comes first, so the boosted power is the
// damage. "When excess damage is dealt" is a reflexive trigger (CR 603.12) that
// chooses its "up to one" artifact only when there is excess; a fight that
// deals none, or that never happens because a target is gone, destroys nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a744206-04c4-46cb-b64d-cf192c9324df",
		Name:         "Rhino's Rampage",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Rhino's Rampage — +1/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			victim, ok := ctx.Game.LookupCardForEffect(theirs.ID)
			if !ok {
				return nil
			}
			lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
			cursor := effects.B25LastEventSeq(ctx.Game)
			if err := effects.B10Fight(ctx, mine.ID, theirs.ID); err != nil {
				return err
			}
			if effects.B27DamageDealtToAfter(ctx.Game, mine.ID, theirs.ID, cursor)-lethal <= 0 {
				return nil
			}
			return effects.ReflexiveTrigger{
				Label: "Rhino's Rampage — destroy up to one target noncreature artifact with mana value 3 or less",
				Body:  rhinosRampageDestroyBody,
			}.Apply(ctx)
		},
	})
}

var rhinosRampageDestroyBody = game.ReflexiveBody("rhinos-rampage/destroy-artifact",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetPermanent("up to one target noncreature artifact with mana value 3 or less",
			effects.And(effects.Artifact(), effects.Noncreature(), effects.ManaValueLE(3))).WithCount(0, 1)
	}))
