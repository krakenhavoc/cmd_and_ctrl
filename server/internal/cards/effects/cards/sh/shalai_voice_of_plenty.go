package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shalai, Voice of Plenty — Legendary Creature — Angel {3}{W}, 3/4
// (Tarkir: Dragonstorm Commander):
//
//	"Flying
//	 You, planeswalkers you control, and other creatures you control have
//	 hexproof.
//	 {4}{G}{G}: Put a +1/+1 counter on each creature you control."
//
// "You" is the player keyword (Keen-Eared Sentry's); the permanents are
// a layer-6 grant (CR 702.11b) to every other creature and every
// planeswalker its controller has. The activated ability is Gavony
// Township's body.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if target.Controller != source.Controller {
			return false
		}
		if target.IsPlaneswalker() {
			return true
		}
		return target.IsCreature() && target.InstanceID != source.InstanceID
	}, effects.KeywordHexproof)
	grant.Label = "Planeswalkers you control and other creatures you control have hexproof."

	effects.Register(effects.Spec{
		OracleID:        "a0c47ab6-dfb4-46ee-a3f7-9e1521b4bb4b",
		Name:            "Shalai, Voice of Plenty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		PlayerKeywords:  []string{effects.KeywordHexproof},
		Static:          []game.StaticAbility{grant},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{G}: Put a +1/+1 counter on each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B11PutCountersOnEachCreatureYouControl(g, item, 1)
			},
		}},
	})
}
