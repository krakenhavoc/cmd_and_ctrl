package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tuktuk Scrapper — Creature — Goblin Artificer Ally {3}{R}, 2/2 (Worldwake, S82 OLD
// b488):
//
//	"Whenever this creature or another Ally you control enters, you may destroy
//	 target artifact. If that artifact is put into a graveyard this way, this
//	 creature deals damage to that artifact's controller equal to the number of
//	 Allies you control."
//
// The Ally trigger (Zendikar's, shared by the Worldwake Allies): the target is chosen as it goes
// on the stack and the "you may" asked as it resolves. The artifact's controller is read before
// the destroy; the damage follows only an artifact that really went to a graveyard (a regenerated,
// indestructible or replaced one deals none), from the Scrapper as it last existed, equal to the
// Allies you control counted when that is settled (the Scrapper included while it is still there).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85cf2403-b419-4364-8ac9-67dd1ceddf9e",
		Name:         "Tuktuk Scrapper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.On(game.EventETB, effects.WwkThisOrAnotherAllyEntered,
					"Tuktuk Scrapper — you may destroy target artifact; if it is put into a graveyard this way, damage to its controller equal to the number of Allies you control",
					tuktukScrapperDestroy),
					"Tuktuk Scrapper — destroy the artifact?"),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}

func tuktukScrapperDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	victim, ok := effects.ControllerOfTarget(ctx, legal[0].ID)
	if !ok {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{legal[0].ID}, effects.WithCarry(ctx.ItemTail(tuktukScrapperDestroyed), victim))
}

// tuktukScrapperDestroyed is the damage, run once the destroy has settled: only when the
// artifact was put into a graveyard. An on-disk key: never renamed, never reused.
var tuktukScrapperDestroyed = effects.CardTail("then/tuktuk-scrapper-destroyed", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	if len(o.IDs) == 0 || victim == uuid.Nil {
		return nil
	}
	n := effects.ZenB2AlliesYouControl(ctx)
	if n <= 0 {
		return nil
	}
	if ref, ok := ctx.SourceRef(); ok {
		return effects.DealDamage{SourceObject: &ref, Target: victim, Amount: n}.Apply(ctx)
	}
	return effects.DealDamage{Source: ctx.Source(), Target: victim, Amount: n}.Apply(ctx)
})
