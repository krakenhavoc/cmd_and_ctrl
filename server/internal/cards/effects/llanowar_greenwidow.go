package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "23391456-5dfd-4402-88fa-3b08f5e23864",
		Name:            "Llanowar Greenwidow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated: []ActivatedAbility{{
			Label:   "Domain — {7}{G}: Return this card from your graveyard to the battlefield tapped. It gains \"If this permanent would leave the battlefield, exile it instead of putting it anywhere else.\" This ability costs {1} less to activate for each basic land type among lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{7}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			CostModifiers: []game.CostModifier{
				DomainCostReduction("This ability costs {1} less to activate for each basic land type among lands you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := returnThisFromGraveyardTapped(g, item); err != nil {
					return err
				}
				if onBattlefield(g, item.SourceCardID) {
					ExileInsteadOfLeavingBattlefield(g, item.SourceCardID, item.Controller,
						"Llanowar Greenwidow: if it would leave the battlefield, exile it instead")
				}
				return nil
			},
		}},
	})
}
