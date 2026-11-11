package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tasigur, the Golden Fang — Legendary Creature — Human Shaman {5}{B},
// 4/5 (Tarkir: Dragonstorm Commander):
//
//	"Delve
//	 {2}{G/U}{G/U}: Mill two cards, then return a nonland card of an
//	 opponent's choice from your graveyard to your hand."
//
// Delve is `Delve: true` (CR 702.66, ADR 0100). The ability targets
// nothing (the printed ruling): the mill happens first, then the
// controller names an opponent (only asked when there are several, Plargg
// and Nassari's way) and THAT player picks a nonland card from your whole
// graveyard, not just the milled ones. With no nonland card there, nobody
// is asked. The card returns to your hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "837e3fde-241f-4826-ad04-f146fdcf2278",
		Name:         "Tasigur, the Golden Fang",
		Completeness: CompletenessFull,
		Delve:        true,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G/U}{G/U}: Mill two cards, then return a nonland card of an opponent's choice from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{G/U}{G/U}"),
			Effect: Do(
				MillCards{N: 2},
				Instruction(tasigurAskAnOpponent),
			),
		}},
	})
}

// tasigurAskAnOpponent is "return a nonland card of an opponent's
// choice": the choosing opponent is named by the controller when there is
// more than one.
func tasigurAskAnOpponent(ctx *Context) error {
	if !tasigurHasNonlandCard(ctx) {
		return nil
	}
	opps := ctx.Opponents()
	switch len(opps) {
	case 0:
		return nil
	case 1:
		return tasigurAsk(ctx, opps[0])
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Tasigur, the Golden Fang — choose an opponent to choose a nonland card in your graveyard",
		ThenKey:  tasigurOpponentChosen,
	}.Apply(ctx)
}

func tasigurNonlandCards(ctx *Context) []uuid.UUID {
	return graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return !c.IsLand() })
}

func tasigurHasNonlandCard(ctx *Context) bool { return len(tasigurNonlandCards(ctx)) > 0 }

// tasigurOpponentChosen runs once the controller has named the opponent.
// An on-disk key: never renamed, never reused.
var tasigurOpponentChosen = ChoosePlayerThen("option-pick/tasigur-the-golden-fang-opponent",
	func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
		if chosen == uuid.Nil {
			return nil
		}
		return tasigurAsk(ctx, chosen)
	})

// tasigurAsk puts the "choose a nonland card" prompt to the opponent: the
// controller's graveyard is the pool, the opponent the chooser.
func tasigurAsk(ctx *Context, chooser uuid.UUID) error {
	cards := tasigurNonlandCards(ctx)
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Tasigur, the Golden Fang — choose a nonland card in their graveyard to return to their hand",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    tasigurPicked,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// tasigurPicked returns the chosen card to its owner's hand. An on-disk
// key: never renamed, never reused.
var tasigurPicked = askedPickRule("choose-cards/tasigur-the-golden-fang-return", nil,
	func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
