package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epiphany at the Drownyard — Instant {X}{U} (Shadows over Innistrad, S82
// OLD):
//
//	"Reveal the top X plus one cards of your library and separate them into
//	 two piles. An opponent chooses one of those piles. Put that pile into
//	 your hand and the other into your graveyard."
//
// Intrude on the Mind's pile split over X + 1 cards, with no Thopter: you
// separate the revealed cards, the opponent you name chooses (the ruling: you
// decide which opponent chooses) and a pile may be empty. With no opponent
// left to choose (CR 800.4a) everything goes to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33befbd6-56ac-45a7-a4d0-448167162707",
		Name:         "Epiphany at the Drownyard",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), ctx.X()+1,
				"Epiphany at the Drownyard — reveal the top X plus one cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Epiphany at the Drownyard — choose an opponent to choose one of the two piles",
				ThenKey:  epiphanyChooser,
				Carry:    revealed,
			}.Apply(ctx)
		},
	})
}

// epiphanyChooser opens the split with the chosen opponent as the chooser. An
// on-disk key: never renamed, never reused.
var epiphanyChooser = effects.ChoosePlayerThen("option-pick/epiphany-at-the-drownyard-chooser", func(ctx *effects.Context, chooser uuid.UUID, revealed []uuid.UUID) error {
	if chooser == uuid.Nil {
		return effects.OldB288IntoHandAndGraveyard(ctx, epiphanyRest, nil, revealed)
	}
	return effects.PileSplit{
		Splitter:      ctx.Controller(),
		Chooser:       chooser,
		Owner:         ctx.Controller(),
		SplitQuestion: "Epiphany at the Drownyard — separate these cards into two piles",
		PickQuestion:  "Epiphany at the Drownyard — choose a pile to put into their hand; the other goes to their graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(epiphanyTake),
	}.Apply(ctx)
})

var epiphanyTake = effects.PileTail("then/epiphany-at-the-drownyard-take", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return effects.OldB288IntoHandAndGraveyard(ctx, epiphanyRest, taken, left)
})

// epiphanyRest is "the other into your graveyard".
var epiphanyRest = effects.CardTail("tail/epiphany-at-the-drownyard-rest", effects.OldB288BinCarried)
