package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rise and Shine — Sorcery {1}{U}:
//
//	"Target noncreature artifact you control becomes a 0/0 artifact
//	 creature. Put four +1/+1 counters on each artifact that became a
//	 creature this way.
//	 Overload {4}{U}{U} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// Haunted Screen's type change: the permanent becomes a creature in
// addition to being an artifact, with base power and toughness 0/0, and
// the effect has no stated duration so it lasts as long as the permanent
// does (CR 611.2a, pinned to this object, CR 400.7). The counters go on
// only the artifacts that became creatures. Overloaded (Cyclonic Rift's
// shape), every noncreature artifact its controller controls as the
// spell resolves is animated and gets them.
//
// No simplification.
func init() {
	noncreatureArtifactYouControl := []CardPredicate{Artifact(), Not(Creature()), YouControl()}
	Register(Spec{
		OracleID:     "59925461-99f5-4217-8343-49c9b1ce24c1",
		Name:         "Rise and Shine",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		Targets:      TargetPermanent("target noncreature artifact you control", noncreatureArtifactYouControl...),
		AlternativeCosts: []game.AlternativeCost{
			// Overload changes the spell, so it declares for itself.
			CostWithPurpose(Overload("{4}{U}{U}"), game.Purpose{Answers: game.AnswerAnimate}),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			if ctx.PaidAltCost("overload") {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsArtifact() && !c.IsCreature() {
						ids = append(ids, c.InstanceID)
					}
				}
			} else if id := FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
				ids = append(ids, id)
			}
			for _, id := range ids {
				if err := (ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.SetBasePowerMod(0),
						game.SetBaseToughnessMod(0),
					},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Rise and Shine — becomes a 0/0 artifact creature",
				}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 4}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
