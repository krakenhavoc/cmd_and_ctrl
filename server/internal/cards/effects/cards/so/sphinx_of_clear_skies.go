package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sphinx of Clear Skies — Creature — Sphinx {3}{U}{U}, 5/5 (Dominaria
// United):
//
//	"Flying, ward {2}
//	 Domain — Whenever this creature deals combat damage to a player,
//	 reveal the top X cards of your library, where X is the number of
//	 basic land types among lands you control. An opponent separates those
//	 cards into two piles. Put one pile into your hand and the other into
//	 your graveyard. (Piles can be empty.)"
//
// Fact or Fiction's reveal, opponent's split and pile pick, on a combat
// damage trigger with X the shared domain count (DomainCount, #5666)
// counted as the ability resolves. Either pile may be empty. In a game
// with several opponents the controller names the one who separates, as
// Fact or Fiction's controller does. The same settle moves the piles, a
// card that has gone from where the split found it skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ee394a2-84de-42e8-834e-6c1030be0aef",
		Name:            "Sphinx of Clear Skies",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "ward"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Sphinx of Clear Skies — ward {2}"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Sphinx of Clear Skies — reveal the top X cards of your library, where X is the number of basic land types among lands you control; an opponent separates them into two piles; put one pile into your hand and the other into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), effects.DomainCount(g, item.Controller),
						"Sphinx of Clear Skies — reveal the top X cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					return effects.ChoosePlayer{
						Among:    effects.Opponents,
						Question: "Sphinx of Clear Skies — choose an opponent to separate these cards into two piles",
						ThenKey:  sphinxOfClearSkiesSplit,
						Carry:    revealed,
					}.Apply(ctx)
				}),
		},
	})
}

// sphinxOfClearSkiesSplit hands the revealed cards to the chosen
// opponent. An on-disk key: never renamed, never reused.
var sphinxOfClearSkiesSplit = effects.ChoosePlayerThen("option-pick/sphinx-of-clear-skies-splitter", func(ctx *effects.Context, splitter uuid.UUID, revealed []uuid.UUID) error {
	if splitter == uuid.Nil {
		// Nobody left to separate them (CR 800.4a): one pile, taken.
		return effects.FactOrFictionSettle(ctx.Game, revealed, nil)
	}
	return effects.PileSplit{
		Splitter:      splitter,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Sphinx of Clear Skies — separate these cards into two piles",
		PickQuestion:  "Sphinx of Clear Skies — take one pile into your hand; the other goes to your graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(sphinxOfClearSkiesTake),
	}.Apply(ctx)
})

// sphinxOfClearSkiesTake puts the chosen pile into the hand and the
// other into the graveyard. An on-disk key.
var sphinxOfClearSkiesTake = effects.PileTail("then/sphinx-of-clear-skies-take", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return effects.FactOrFictionSettle(ctx.Game, taken, left)
})
