package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Audacious Reshapers — Creature — Human Artificer {2}{R}, 3/3:
//
//	"{T}, Sacrifice an artifact: Reveal cards from the top of your
//	 library until you reveal an artifact card. Put that card onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order. This creature deals damage to you equal to the number of
//	 cards revealed this way."
//
// The sacrifice is part of the cost, so the artifact is gone before the
// reveal. The reveal is RevealUntilThenPutOntoBattlefield's, unrolled so
// the count of revealed cards (the artifact card included) is in hand for
// the damage: a library with no artifact reveals itself whole, bottoms
// every card and still deals that many. The damage is dealt by the
// Reshapers as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "525c0e33-f073-4bed-be8a-14e436a0bcf1",
		Name:         "Audacious Reshapers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Reveal cards from the top of your library until you reveal an artifact card. Put that card onto the battlefield and the rest on the bottom of your library in a random order. This creature deals damage to you equal to the number of cards revealed this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue}, // a free permanent off the library, no answer
			Cost:    effects.Plus(effects.TapCost(), effects.B10SacrificeAnArtifact()),
			Effect:  audaciousReshapersReveal,
		}},
	})
}

func audaciousReshapersReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsArtifact() },
		"Audacious Reshapers — revealed until an artifact card")
	if err := (effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: effects.PutRestOnBottomInRandomOrderTail,
	}).Apply(ctx); err != nil {
		return err
	}
	dmg := effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: len(run)}
	if ref, ok := ctx.SourceRef(); ok {
		dmg.SourceObject = &ref
	}
	return dmg.Apply(ctx)
}
