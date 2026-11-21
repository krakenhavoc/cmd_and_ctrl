package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Animal Magnetism — Sorcery {4}{G} (Commander Masters, #5948):
//
//	"Reveal the top five cards of your library. An opponent chooses a
//	 creature card from among them. Put that card onto the battlefield and
//	 the rest into your graveyard."
//
// Fact or Fiction's reveal, then Extrapolate the Impossible's two prompts:
// the caster names the opponent who chooses (the engine skips that question
// at two seats), and that opponent picks one of the creature cards just
// revealed (a reveal_pick, whose candidates are public because they were
// revealed). The picked card enters under the caster's control and the rest
// of the five go to the graveyard. With no creature card among them nobody
// is asked and all five go to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1bdf7f1-e3f7-4a13-9cf0-a2394df6cc74",
		Name:         "Animal Magnetism",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 5,
				"Animal Magnetism — reveal the top five cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			if len(animalMagnetismCreatures(ctx.Game, revealed)) == 0 {
				return animalMagnetismSettle(ctx.Game, ctx.Controller(), uuid.Nil, revealed)
			}
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Animal Magnetism — choose an opponent to choose a creature card from among the revealed cards",
				ThenKey:  animalMagnetismOpponent,
				Carry:    revealed,
			}.Apply(ctx)
		},
	})
}

// animalMagnetismCreatures is the creature cards among the revealed ones
// that are still in the library.
func animalMagnetismCreatures(g *game.Game, ids []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			out = append(out, id)
		}
	}
	return out
}

// animalMagnetismOpponent is the chosen opponent's prompt: pick one of the
// revealed creature cards. Keyed: the revealed cards ride the prompt. Nobody
// named (every opponent has left, CR 800.4a) puts every card into the
// graveyard, the only thing left to do with them.
var animalMagnetismOpponent = ChoosePlayerThen("option-pick/animal-magnetism-opponent", func(ctx *Context, chooser uuid.UUID, revealed []uuid.UUID) error {
	if chooser == uuid.Nil {
		return animalMagnetismSettle(ctx.Game, ctx.Controller(), uuid.Nil, revealed)
	}
	_, err := ctx.Game.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  chooser,
		Owner:    ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Animal Magnetism — choose a creature card; it enters under their control and the rest go to their graveyard",
		Cards:    animalMagnetismCreatures(ctx.Game, revealed),
		Min:      1,
		Max:      1,
		ThenKey:  animalMagnetismTake,
		Carry:    revealed,
	})
	return err
})

// animalMagnetismTake is the pick's answer, keyed: the chosen card onto the
// battlefield under the caster's control, the rest into their graveyard.
var animalMagnetismTake = game.RegisterChooseCardsThen("choose-cards/animal-magnetism-take", func(g *game.Game, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	return animalMagnetismSettle(g, r.From, picked, r.Carry)
})

// animalMagnetismSettle moves the cards: `picked` (if any, and still in the
// library) onto the battlefield under `you`, the others into their graveyard.
func animalMagnetismSettle(g *game.Game, you, picked uuid.UUID, revealed []uuid.UUID) error {
	if picked != uuid.Nil {
		if z := g.FindCardZoneForEffect(picked); z != nil && z.Kind == game.ZoneLibrary {
			if _, err := g.PutFromLibraryOntoBattlefieldForEffect(picked, game.LibraryEntryOptions{Controller: you}); err != nil {
				return err
			}
		}
	}
	for _, id := range revealed {
		if id == picked {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		if err := g.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
