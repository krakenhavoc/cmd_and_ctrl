package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boneyard Parley — Sorcery {5}{B}{B} (Commander 2019):
//
//	"Exile up to five target creature cards from graveyards. An opponent
//	 separates those cards into two piles. Put all cards from the pile of your
//	 choice onto the battlefield under your control and the rest into their
//	 owners' graveyards."
//
// Split the Spoils' shape with the roles of Fact or Fiction: the targets still
// legal as the spell resolves go to exile face up (an exile, so the piles are
// public), the caster picks an opponent ("an opponent" is the controller's
// choice, ChoosePlayer) who separates them, and the caster takes one pile.
// The pile taken enters together under the caster's control, the other goes to
// its cards' owners' graveyards, either pile possibly empty; a card that left
// exile in the meantime is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3d048c8-e8df-4b50-a422-fc94a69f44bb",
		Name:         "Boneyard Parley",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("up to five target creature cards from graveyards", effects.Creature()).WithCount(0, 5),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(boneyardParleyExiled))
		},
	})
}

// boneyardParleyExiled asks for the opponent once the cards are in exile. An
// on-disk key: never renamed, never reused.
var boneyardParleyExiled = effects.CardTail("then/boneyard-parley-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	if len(ctx.Opponents()) == 0 {
		return boneyardParleyDeal(ctx, o.IDs, nil)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Boneyard Parley — choose an opponent to separate these cards into two piles",
		ThenKey:  boneyardParleyChosen,
		Carry:    append([]uuid.UUID(nil), o.IDs...),
	}.Apply(ctx)
})

// boneyardParleyChosen starts the split with the chosen opponent as the
// splitter. An on-disk key: never renamed, never reused.
var boneyardParleyChosen = effects.ChoosePlayerThen("option-pick/boneyard-parley-chosen", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return nil
	}
	if chosen == uuid.Nil {
		// Nobody left to separate them (CR 800.4a): one pile, taken.
		return boneyardParleyDeal(ctx, carry, nil)
	}
	return effects.PileSplit{
		Splitter:      chosen,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Boneyard Parley — separate the cards into two piles",
		PickQuestion:  "Boneyard Parley — choose a pile: it enters the battlefield under your control, the other goes to its owners' graveyards",
		Cards:         carry,
		ThenTail:      ctx.ItemTail(boneyardParleyDealt),
	}.Apply(ctx)
})

// boneyardParleyDealt puts the chosen pile onto the battlefield and the other
// into its cards' owners' graveyards. An on-disk key: never renamed, never
// reused.
var boneyardParleyDealt = effects.PileTail("pile/boneyard-parley-dealt", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return boneyardParleyDeal(ctx, taken, left)
})

func boneyardParleyDeal(ctx *effects.Context, taken, left []uuid.UUID) error {
	if len(taken) > 0 {
		if err := (effects.ReturnFromExileTogether{Targets: taken, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range left {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
