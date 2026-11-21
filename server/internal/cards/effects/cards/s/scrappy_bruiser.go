package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrappy Bruiser — Creature — Raccoon Warrior {3}{R}, 3/4 (March of the
// Machine, S82 MOM batch 13):
//
//	"Whenever this creature attacks, up to one target attacking creature
//	 gets +2/+0 and gains trample until end of turn. Return that creature
//	 to its owner's hand at end of combat."
//
// The target is an attacking creature chosen as the trigger goes on the
// stack ("up to one", so it may be none). On resolution it gets +2/+0 and
// trample until end of turn (Berserk's pump), then a delayed trigger at
// the end of combat (Arthur's, CR 603.7) returns that creature to its
// owner's hand if it is still the same object on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2976f47b-9626-4934-bcee-7ee7f57944c3",
		Name:         "Scrappy Bruiser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Scrappy Bruiser — up to one target attacking creature gets +2/+0 and gains trample until end of turn; return it to its owner's hand at end of combat",
					scrappyBruiserResolve),
				effects.TargetCreature("up to one target attacking creature", effects.AttackingCreature()).WithCount(0, 1)),
		},
	})
}

func scrappyBruiserResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := effects.UntilEndOfTurn(ctx, id, nil, "Scrappy Bruiser — +2/+0 and trample",
		game.ModifyPTMod(2, 0), game.AddKeywordsMod("trample")); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEndCombat,
		Label:  "Scrappy Bruiser — return that creature to its owner's hand",
		Cards:  []uuid.UUID{id},
		Body:   effects.ZaraReturnToHandBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}
