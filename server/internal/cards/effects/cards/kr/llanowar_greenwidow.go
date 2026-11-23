package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Greenwidow — Creature — Spider {2}{G}, 4/3:
//
//	"Reach, trample
//	 Domain — {7}{G}: Return this card from your graveyard to the
//	 battlefield tapped. It gains "If this permanent would leave the
//	 battlefield, exile it instead of putting it anywhere else." This
//	 ability costs {1} less to activate for each basic land type among
//	 lands you control."
//
// A graveyard ability (CR 113.6, Zones) with Chronosavant's return, then
// the granted replacement as Dreams of the Dead writes it
// (ExileInsteadOfLeavingBattlefield, CR 614): it covers every exit and
// belongs to this object only, so a later return is a new object with
// neither (CR 400.7). The price is the shared domain count
// (DomainCostReduction, #5666), priced from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23391456-5dfd-4402-88fa-3b08f5e23864",
		Name:            "Llanowar Greenwidow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Domain — {7}{G}: Return this card from your graveyard to the battlefield tapped. It gains \"If this permanent would leave the battlefield, exile it instead of putting it anywhere else.\" This ability costs {1} less to activate for each basic land type among lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{7}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			CostModifiers: []game.CostModifier{
				effects.DomainCostReduction("This ability costs {1} less to activate for each basic land type among lands you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ReturnThisFromGraveyardTapped(g, item); err != nil {
					return err
				}
				if effects.OnBattlefield(g, item.SourceCardID) {
					effects.ExileInsteadOfLeavingBattlefield(g, item.SourceCardID, item.Controller,
						"Llanowar Greenwidow: if it would leave the battlefield, exile it instead")
				}
				return nil
			},
		}},
	})
}
