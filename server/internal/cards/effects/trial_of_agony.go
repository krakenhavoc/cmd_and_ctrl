package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trial of Agony — Sorcery {R} (Duskmourn: House of Horror, #4466):
//
//	"Choose two target creatures controlled by the same opponent. That
//	 player chooses one of those creatures. Trial of Agony deals 5 damage to
//	 that creature, and the other can't block this turn."
//
// The two targets are Breaking of the Fellowship's pair: the first is any
// creature an opponent controls and the second is a DIFFERENT creature that
// player controls (ControlledByTargetsController, judged as the spell is cast
// and again as it resolves; a first target that has left leaves its last-known
// controller as "that player", CR 608.2h). As the spell resolves the creatures
// still legal are put to that player, who chooses one of them (a choose-cards
// prompt addressed to the opponent, not the caster); with only one legal target
// left it is the one chosen and nothing is asked (the ruling). The chosen
// creature is dealt 5 damage by the spell, and the other creature, if there is
// one, can't block this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ef07e5a-3dc3-41dd-a3a9-e8812b4d40d3",
		Name:         "Trial of Agony",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature an opponent controls", OpponentControls()),
			Distinct(ControlledByTargetsController(TargetCreature("another target creature controlled by that player"), 0)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var candidates []uuid.UUID
			var chooser uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				candidates = append(candidates, t.ID)
				chooser = c.Controller
			}
			switch len(candidates) {
			case 0:
				return nil
			case 1:
				return trialOfAgonyApply(ctx, candidates[0], uuid.Nil)
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  chooser,
				Source:   ctx.Source(),
				Question: "Trial of Agony — choose one of those creatures; it is dealt 5 damage and the other can't block this turn",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				ThenKey:  trialOfAgonyChosen,
				LiveZone: game.ZoneBattlefield,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// trialOfAgonyApply deals the 5 damage to `chosen` and stops `other` (when
// there is one) from blocking this turn.
func trialOfAgonyApply(ctx *Context, chosen, other uuid.UUID) error {
	if err := (DealDamage{Source: ctx.Source(), Target: chosen, Amount: 5}).Apply(ctx); err != nil {
		return err
	}
	if other == uuid.Nil || !onBattlefield(ctx.Game, other) {
		return nil
	}
	return RestrictUntilEOT{Target: other, Restrictions: game.CantBlock, Label: "Trial of Agony — can't block"}.Apply(ctx)
}

// trialOfAgonyChosen is the opponent's answer, keyed (#4253): the asking item
// rides the prompt as data, and the creature they did not choose is the other
// candidate. An on-disk key: never renamed, never reused.
var trialOfAgonyChosen = askedPick("choose-cards/trial-of-agony-chosen", func(ctx *Context, r game.CardsChosen) error {
	// Unanswered, the opponent has left the game, and their creatures with them.
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	chosen := r.Picked[0]
	other := uuid.Nil
	for _, id := range r.Candidates {
		if id != chosen {
			other = id
		}
	}
	return trialOfAgonyApply(ctx, chosen, other)
})
