package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutual Destruction — Sorcery {B} (Ikoria: Lair of Behemoths, #5512):
//
//	"This spell has flash as long as you control a permanent with flash.
//	 As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy target creature."
//
// Bone Splinters' body (the sacrifice is paid while casting, CR 601.2h) with
// Crashing Tide's flash permission (Spec.FlashIf, CR 601.3), read as the cast
// begins over the keywords the layered view gives each permanent you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b01cd710-3906-486b-8bb5-8c4efa46e4d9",
		Name:           "Mutual Destruction",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		FlashIf: effects.FlashIfYou("as long as you control a permanent with flash",
			func(g *game.Game, you uuid.UUID) bool {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == you && game.HasKeyword(&c, "flash") {
						return true
					}
				}
				return false
			}),
		Targets:   effects.TargetPermanent("target creature", effects.Creature()),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
