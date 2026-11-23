package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feldon of the Third Path — Legendary Creature — Human Artificer {1}{R}{R},
// 2/3:
//
//	"{2}{R}, {T}: Create a token that's a copy of target creature card in
//	 your graveyard, except it's an artifact in addition to its other
//	 types. It gains haste. Sacrifice it at the beginning of the next end
//	 step."
//
// Saheeli, the Sun's Brilliance's shape aimed at a creature card in the
// graveyard: the copy is made from the card's copiable values, with
// Artifact added to its types as part of the copy (CR 707.9b); haste is an
// indefinite grant on the token and the sacrifice is one delayed trigger
// on the tokens made (CR 603.7). The target is checked again as the
// ability resolves, so a card that left the graveyard in response makes
// nothing (CR 608.2b). The {T} makes Feldon subject to summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be094704-5232-4b10-beeb-aa9bbfae063d",
		Name:         "Feldon of the Third Path",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.CreateTokenCopiesTogether{
					Controller: item.Controller,
					Copies:     []uuid.UUID{id},
					Except:     effects.TokenCopyIsAlsoAnArtifact,
					Then:       ctx.ItemTail(feldonRest),
				}.Apply(ctx)
			},
		}},
	})
}

// feldonRest is "It gains haste. Sacrifice it at the beginning of the next
// end step." as a keyed tail. An on-disk key: never renamed, never reused.
var feldonRest = effects.CardTail("feldon-of-the-third-path/haste-and-sacrifice-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.RprB06HasteAndEndStepTail(ctx, o, "Feldon of the Third Path", effects.SacrificeListedCardsBody)
})
