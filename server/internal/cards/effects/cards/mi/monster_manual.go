package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monster Manual // Zoological Study — an adventure card (Commander Legends:
// Battle for Baldur's Gate, S82 GAP #5948):
//
//	Monster Manual {3}{G}, Artifact — Book:
//	"{1}{G}, {T}: You may put a creature card from your hand onto the
//	 battlefield."
//
//	Zoological Study {2}{G}, Sorcery — Adventure:
//	"Mill five cards, then return a creature card milled this way to your
//	 hand."
//
// The Book is Root Elemental's put (the card is put, not cast, and the "may" is
// a real decline) behind an activated ability. The adventure half is the
// "<oracle_id>#1" spec, and the engine exiles the resolved adventure with
// permission to cast the Book (CR 715.3d). Zoological Study mills five, then
// offers the creature cards that reached the graveyard (CR 400.7), Sparring
// Dummy's pick, and the caster takes one if any was milled.
func init() {
	const oracleID = "6d4e7943-1fdf-4085-8af0-658badbd0cb8"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Monster Manual",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}: You may put a creature card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.Creature(),
					Optional: true,
					Label:    "Monster Manual — you may put a creature card from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Zoological Study",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{N: 5, ThenTail: ctx.ItemTail(zoologicalStudyMilled)}.Apply(ctx)
		},
	})
}

// zoologicalStudyMilled offers the creature cards milled this way. An on-disk
// key.
var zoologicalStudyMilled = effects.CardTail("then/zoological-study-milled", func(ctx *effects.Context, o game.TailOutcome) error {
	var creatures []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 || ctx.Item == nil {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Zoological Study — return a creature card milled this way to your hand",
		Cards:      creatures,
		Min:        1,
		Max:        1,
		ThenKey:    zoologicalStudyTake,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// zoologicalStudyTake returns the picked card to hand. An on-disk key.
var zoologicalStudyTake = effects.AskedPick("choose-cards/zoological-study-take", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
