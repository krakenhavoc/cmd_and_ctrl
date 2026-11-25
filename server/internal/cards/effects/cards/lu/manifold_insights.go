package cardslu

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manifold Insights — Sorcery {2}{U} (Commander 2016, #5537):
//
//	"Reveal the top ten cards of your library. Starting with the next opponent
//	 in turn order, each opponent chooses a different nonland card from among
//	 them. Put the chosen cards into your hand and the rest on the bottom of
//	 your library in a random order."
//
// The reveal is the controller's (CR 701.20a); the opponents then choose one
// after another in turn order from the controller's seat (CR 101.4), each from
// the nonland cards no earlier opponent took, so no two choose the same card.
// An opponent with nothing left to choose asks nothing. The prompt is a keyed
// choose_cards pick that carries the revealed cards, then the cards chosen so far,
// as data (Nums: how many were revealed, and the opponent being asked), so a table
// waiting on a choice is a restore point. After the last opponent the chosen cards
// go to the controller's hand and the rest, lands included, to the bottom in a
// random order (Old Stickfingers' PutOnBottomInRandomOrderForEffect).
//
// No simplification.
func init() {
	manifoldInsightsPick = effects.AskedPick("choose-cards/manifold-insights-pick", func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Nums) < 2 || r.Nums[0] > len(r.Carry) {
			return nil
		}
		revealed := slices.Clone(r.Carry[:r.Nums[0]])
		chosen := slices.Clone(r.Carry[r.Nums[0]:])
		if r.Answered {
			chosen = append(chosen, r.Picked...)
		}
		return manifoldInsightsAsk(ctx, revealed, chosen, r.Nums[1]+1)
	})
	effects.Register(effects.Spec{
		OracleID:     "9c793037-6ba8-400a-b24d-6032141e8dae",
		Name:         "Manifold Insights",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 10,
				"Manifold Insights — reveal the top ten cards of your library")
			return manifoldInsightsAsk(ctx, revealed, nil, 0)
		},
	})
}

// manifoldInsightsPick is the keyed continuation of one opponent's choice,
// registered in init (it asks the next opponent with itself).
var manifoldInsightsPick game.ChooseCardsThen

// manifoldInsightsAsk asks the opponent at index `from` (in turn order from the
// controller) or the first after them who has a nonland card left to choose, or
// finishes when none does.
func manifoldInsightsAsk(ctx *effects.Context, revealed, chosen []uuid.UUID, from int) error {
	g := ctx.Game
	opponents := effects.OpponentsInTurnOrder(g, ctx.Controller())
	var left []uuid.UUID
	for _, id := range revealed {
		if slices.Contains(chosen, id) {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
				left = append(left, id)
			}
		}
	}
	if from < len(opponents) && len(left) > 0 {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    opponents[from],
			FromPlayer: ctx.Controller(),
			Source:     ctx.Source(),
			Question:   "Manifold Insights — choose a different nonland card from among the revealed cards",
			Cards:      left,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			ThenKey:    manifoldInsightsPick,
			Carry:      append(slices.Clone(revealed), chosen...),
			CarryNums:  []int{len(revealed), from},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	player := ctx.Controller()
	g.TakeFromLibraryToHandForEffect(player, chosen)
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, effects.CardsStillInALibrary(g, revealed))
}
