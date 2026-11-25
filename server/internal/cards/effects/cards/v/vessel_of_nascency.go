package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vessel of Nascency — Enchantment {G} (Shadows over Innistrad, S82 OLD):
//
//	"{1}{G}, Sacrifice this enchantment: Reveal the top four cards of your
//	 library. You may put an artifact, creature, enchantment, land, or
//	 planeswalker card from among them into your hand. Put the rest into your
//	 graveyard."
//
// Bitter Revelation's take with a one-card cap; the four are revealed to the
// table (CR 701.20a) and the rest go to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e33dec01-c114-436a-befc-dab4e806c10b",
		Name:         "Vessel of Nascency",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this enchantment: Reveal the top four cards of your library. You may put an artifact, creature, enchantment, land, or planeswalker card from among them into your hand. Put the rest into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 4,
					"Vessel of Nascency — reveal the top four cards of your library")
				if len(revealed) == 0 {
					return nil
				}
				return effects.TakeFromLibraryToHand{
					Cards:    revealed,
					Match:    effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment(), effects.Land(), effects.Planeswalker()),
					Max:      1,
					Optional: true,
					Label:    "Vessel of Nascency — you may put an artifact, creature, enchantment, land, or planeswalker card from among them into your hand",
					Key:      vesselOfNascencyTake,
				}.Apply(ctx)
			},
		}},
	})
}

// vesselOfNascencyTake: the rest go into the graveyard. An on-disk key.
var vesselOfNascencyTake = effects.RegisterTake("choose-cards/vessel-of-nascency-take",
	effects.TakeDef{Then: effects.TakeRestIntoGraveyard})
