package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0a744206-04c4-46cb-b64d-cf192c9324df",
		Name:         "Rhino's Rampage",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		Purpose: ForTargets(BenefitUntilEOT(0, 1, 0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Rhino's Rampage — +1/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			victim, ok := ctx.Game.LookupCardForEffect(theirs.ID)
			if !ok {
				return nil
			}
			lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
			cursor := b25LastEventSeq(ctx.Game)
			if err := b10Fight(ctx, mine.ID, theirs.ID); err != nil {
				return err
			}
			if b27DamageDealtToAfter(ctx.Game, mine.ID, theirs.ID, cursor)-lethal <= 0 {
				return nil
			}
			return ReflexiveTrigger{
				Label: "Rhino's Rampage — destroy up to one target noncreature artifact with mana value 3 or less",
				Body:  rhinosRampageDestroyBody,
			}.Apply(ctx)
		},
	})
}

var rhinosRampageDestroyBody = game.ReflexiveBody("rhinos-rampage/destroy-artifact",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("up to one target noncreature artifact with mana value 3 or less",
			And(Artifact(), Noncreature(), ManaValueLE(3))).WithCount(0, 1)
	}))
