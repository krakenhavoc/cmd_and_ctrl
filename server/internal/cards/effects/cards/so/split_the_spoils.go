package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Split the Spoils — Sorcery {2}{G} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Exile up to five target permanent cards from your graveyard and separate
//	 them into two piles. An opponent chooses one of those piles. Put that pile
//	 into your hand and the other into your graveyard. (Piles can be empty.)"
//
// The targets still legal as the spell resolves go to exile face up (an exile,
// so the piles are public), then the caster picks an opponent (ChoosePlayer:
// "an opponent" is the controller's choice) who is the chooser of a pile split the caster
// makes (Fact or Fiction's, with the roles swapped: the caster separates, the
// opponent takes). The pile the opponent picks goes to the caster's hand and the
// other to their graveyard (FactOrFictionSettle), either pile possibly empty. A
// card that left exile in the meantime is skipped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48ad5bd6-9568-4813-be29-99aef8b9327a",
		Name:         "Split the Spoils",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("up to five target permanent cards from your graveyard", effects.YouOwn(),
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.IsPermanentCard(c) }).WithCount(0, 5),
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
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(splitTheSpoilsExiled))
		},
	})
}

// splitTheSpoilsExiled asks for the opponent once the cards are in exile. An
// on-disk key.
var splitTheSpoilsExiled = effects.CardTail("then/split-the-spoils-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(ctx.Opponents()) == 0 {
		return nil
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Split the Spoils — choose an opponent; they choose one of the two piles you separate",
		ThenKey:  splitTheSpoilsChosen,
		Carry:    append([]uuid.UUID(nil), o.IDs...),
	}.Apply(ctx)
})

// splitTheSpoilsChosen starts the split with the chosen opponent as the chooser.
// An on-disk key.
var splitTheSpoilsChosen = effects.ChoosePlayerThen("option-pick/split-the-spoils-chosen", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) == 0 {
		return nil
	}
	you := ctx.Controller()
	return effects.PileSplit{
		Splitter:      you,
		Chooser:       chosen,
		Owner:         you,
		SplitQuestion: "Split the Spoils — separate the cards into two piles",
		PickQuestion:  "Split the Spoils — choose a pile: it goes to their hand, the other to their graveyard",
		Cards:         carry,
		ThenTail:      ctx.ItemTail(splitTheSpoilsSettled),
	}.Apply(ctx)
})

// splitTheSpoilsSettled puts the chosen pile into the caster's hand and the other
// into their graveyard. An on-disk key.
var splitTheSpoilsSettled = effects.PileTail("pile/split-the-spoils-settled", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return effects.FactOrFictionSettle(ctx.Game, taken, left)
})
