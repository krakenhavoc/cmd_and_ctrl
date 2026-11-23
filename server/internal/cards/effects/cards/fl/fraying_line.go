package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fraying Line — Artifact {4} (Commander Legends: Battle for Baldur's Gate,
// #5948):
//
//	"When this artifact enters, put a rope counter on target creature you
//	 control.
//	 At the beginning of each player's upkeep, that player may pay {2}. If
//	 they do, they put a rope counter on a creature they control. Otherwise,
//	 exile this artifact and each creature without a rope counter on it,
//	 then remove all rope counters from all creatures."
//
// The upkeep trigger asks the player whose upkeep it is (not the Line's
// controller) whether to pay {2} (CR 118.12 style keyed prompt, one for each
// upkeep). Paying has them choose a creature they control for the counter (none
// to choose means none is put); declining or being unable to pay exiles the Line
// and every creature without a rope counter in one event, then takes every rope
// counter off every creature, read after the exile so a creature that stayed
// loses its counter too (Oblivion Stone's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b69d2bd6-3e99-4c51-89af-0acedd48739e",
		Name:         "Fraying Line",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Fraying Line — put a rope counter on target creature you control", frayingLineEnters),
				effects.PermanentYouControl("target creature you control", effects.Creature())),
			effects.AtEachUpkeep("Fraying Line — that player may pay {2}; otherwise exile it and each creature without a rope counter", frayingLineUpkeep),
		},
	})
}

const frayingLineRope = "rope"

func frayingLineEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: frayingLineRope, N: 1}.Apply(ctx)
}

func frayingLineUpkeep(g *game.Game, item *game.StackItem) error {
	player := effects.TriggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return effects.PayUnless{
		Chooser:  player,
		Cost:     "{2}",
		Question: "Fraying Line — pay {2}? If you do, put a rope counter on a creature you control; otherwise exile Fraying Line and each creature without a rope counter",
		Else:     frayingLineAnswer,
	}.Apply(effects.NewContext(g, item))
}

// frayingLineAnswer is both branches of the upkeep payment: paid, the payer
// puts a rope counter on a creature they control; declined, the exile and the
// clean-up. An on-disk key: never renamed, never reused.
var frayingLineAnswer = effects.PayUnlessBranches("pay-unless/fraying-line-upkeep", frayingLinePaid, frayingLineDeclined)

var frayingLinePut = effects.RegisterPermanentsPick("prompt-run/fraying-line-rope", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if effects.OnBattlefield(ctx.Game, id) {
				return effects.AddCounter{Target: id, Kind: frayingLineRope, N: 1}.Apply(ctx)
			}
		}
		return nil
	},
})

func frayingLinePaidPick(ctx *effects.Context, payer uuid.UUID) error {
	return effects.ChoosePermanents{
		Player:     payer,
		Question:   "Fraying Line — put a rope counter on a creature you control",
		Candidates: effects.CreaturesOfChooser(1, 1),
		Keyed:      frayingLinePut,
	}.Apply(ctx)
}

func frayingLineDeclined(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	line := ctx.Source()
	if err := (effects.ExileAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.InstanceID == line || (c.IsCreature() && c.Counters[frayingLineRope] <= 0)
		},
	}).Apply(ctx); err != nil {
		return err
	}
	type roped struct {
		id uuid.UUID
		n  int
	}
	var all []roped
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if n := c.Counters[frayingLineRope]; n > 0 {
			all = append(all, roped{c.InstanceID, n})
		}
	}
	for _, r := range all {
		if err := (effects.AddCounter{Target: r.id, Kind: frayingLineRope, N: -r.n}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

func frayingLinePaid(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return frayingLinePaidPick(ctx, r.Chooser)
}
