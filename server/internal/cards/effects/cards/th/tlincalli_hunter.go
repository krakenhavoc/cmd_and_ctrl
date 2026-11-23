package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tlincalli Hunter // Retrieve Prey — an adventure card (Commander Legends:
// Battle for Baldur's Gate):
//
//	Tlincalli Hunter {5}{G}{G}, Creature — Scorpion Scout 7/7:
//	"Trample
//	 Once each turn, you may pay {0} rather than pay the mana cost for a creature
//	 spell you cast from exile."
//
//	Retrieve Prey {1}{G}, Sorcery — Adventure:
//	"Exile target creature card from your graveyard. Until the end of your next
//	 turn, you may cast that card. (Then exile this card. You may cast the
//	 creature later from exile.)"
//
// The adventure layout is Brazen Borrower's (CR 715.3d). The static is a granted
// alternative cost (CR 118.9, ADR 0118) like Warped Space's, narrowed to creature
// spells and spent once per turn per Hunter. Retrieve Prey is Practiced
// Scrollsmith's exile-and-permission: the card goes to exile with a cast-only
// permission lasting until the end of its controller's next turn, so a creature
// cast under it from exile can use the Hunter's {0} offer.
//
// No simplification.
func init() {
	const oracleID = "79b69b38-c01d-499a-a865-8c1cecf9fae0"

	effects.Register(effects.Spec{
		OracleID:                oracleID,
		Name:                    "Tlincalli Hunter",
		Completeness:            effects.CompletenessFull,
		PrintedKeywords:         []string{"trample"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.GapB217PayZeroForACreatureSpellFromExileOnceEachTurn()},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Retrieve Prey",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature card from your graveyard",
			effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.ExileCardWithPermissionForEffect(id, game.CastPermission{
				Player:     item.Controller,
				Duration:   ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller),
				CastOnly:   true,
				Source:     item.SourceCardID,
				SourceName: "Retrieve Prey",
				Label:      "Retrieve Prey — cast this card",
			})
		},
	})
}
