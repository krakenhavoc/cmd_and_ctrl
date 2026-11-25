package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lifecraft Awakening — {X}{G} Instant (Aether Revolt, #5536):
//
//	"Put X +1/+1 counters on target artifact you control. If it isn't a
//	 creature or Vehicle, it becomes a 0/0 Construct artifact creature."
//
// The counters go on first and the check follows, as printed: a Vehicle
// that is not crewed stays a Vehicle, and a creature keeps its own size. The
// change has no end (CR 611.2a), pinned to the permanent so a new object
// does not inherit it (CR 400.7). Bring to Life's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8919c8dd-d3b3-406f-a387-44cc1b6d92f3",
		Name:         "Lifecraft Awakening",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.IsCreature() || c.HasSubtypeWord("Vehicle") {
				return nil
			}
			mods := append([]game.Mod{game.AddTypesMod("Creature"), game.AddSubtypesMod(ctx.TextSubtypes("Construct")...)},
				game.SetBasePTMods(0, 0)...)
			return effects.ScopedEffectFor{
				Target:   id,
				Mods:     mods,
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Lifecraft Awakening — becomes a 0/0 Construct artifact creature",
			}.Apply(ctx)
		},
	})
}
