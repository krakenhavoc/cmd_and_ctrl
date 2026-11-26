package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murmurs from Beyond — Instant — Arcane {2}{U} (Commander 2011, #5574):
//
//	"Reveal the top three cards of your library. An opponent chooses one of
//	 them. Put that card into your graveyard and the rest into your hand."
//
// The caster reveals the cards (public), picks which opponent chooses
// ("an opponent" is the controller's choice, as Split the Spoils', and asked
// only when more than one is in the game), and that opponent picks one of
// those cards, the caster looking on. The chosen card goes to the caster's
// graveyard and the others, still in the library, to their hand. A library of
// fewer than three reveals what it has; a single card is simply the one chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e8be8b4-f7be-49cb-b32e-d92015dea7a2",
		Name:         "Murmurs from Beyond",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 3, "Murmurs from Beyond")
			if len(ids) == 0 || len(ctx.Opponents()) == 0 {
				return nil
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Murmurs from Beyond — choose an opponent; they choose one of the revealed cards to go to your graveyard",
				ThenKey:  murmursOpponentChosen,
				Carry:    append([]uuid.UUID(nil), ids...),
			}.Apply(ctx)
		},
	})
}

// murmursOpponentChosen asks the chosen opponent to pick one of the revealed
// cards. An on-disk key.
var murmursOpponentChosen = effects.ChoosePlayerThen("option-pick/murmurs-from-beyond-chosen", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chosen,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Murmurs from Beyond — choose one of the revealed cards: it goes to their graveyard, the rest to their hand",
		Cards:      carry,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    murmursCardChosen,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// murmursCardChosen puts the chosen card into the graveyard and the rest of the
// revealed cards into the hand. An on-disk key.
var murmursCardChosen = effects.AskedPick("choose-cards/murmurs-from-beyond-chosen", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if err := effects.RestIntoGraveyard(ctx.Game, r.Picked[:1]); err != nil {
		return err
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if id == r.Picked[0] {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			rest = append(rest, id)
		}
	}
	ctx.Game.TakeFromLibraryToHandForEffect(ctx.Controller(), rest)
	return nil
})
