package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyblade's Boon — Enchantment — Aura {1}{W} (Modern Horizons 2, #5501):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying.
//	 {2}{W}: Return Skyblade's Boon to its owner's hand. Activate only if
//	 Skyblade's Boon is on the battlefield or in your graveyard."
//
// The ability is offered from the battlefield and from the graveyard and
// returns the card from wherever it is as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a03a9ddb-595a-4dff-b7cb-7ad60655f1ee",
		Name:         "Skyblade's Boon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Return Skyblade's Boon to its owner's hand. Activate only if Skyblade's Boon is on the battlefield or in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{W}"),
			Zones:   []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard},
			Effect:  effects.S82OldB38ReturnThisToOwnersHand,
		}},
	})
}
