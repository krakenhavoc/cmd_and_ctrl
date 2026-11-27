package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summoning Trap — Instant — Trap {4}{G}{G} (Zendikar, #5583):
//
//	"If a creature spell you cast this turn was countered by a spell or
//	 ability an opponent controlled, you may pay {0} rather than pay this
//	 spell's mana cost.
//	 Look at the top seven cards of your library. You may put a creature card
//	 from among them onto the battlefield. Put the rest on the bottom of your
//	 library in any order."
//
// The trap cost reads this turn's event log (game.CreatureSpellCounteredBy
// OpponentThisTurn): a creature spell the controller cast and a counter
// whose cause is a spell or ability an opponent controls. The body is
// Turntimber Symbiosis's look-and-put with the player ordering the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2faa8dc1-0370-458c-9aef-44f81a9b5cd5",
		Name:         "Summoning Trap",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "trap",
			Label:    "If a creature spell you cast this turn was countered by a spell or ability an opponent controlled, pay {0}",
			ManaCost: "{0}",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				return g.CreatureSpellCounteredByOpponentThisTurn(controller)
			},
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := item.Controller
			return effects.PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Summoning Trap — you may put a creature card from among them onto the battlefield",
				ThenTail: effects.BnsB06PutRestOnBottomInAnyOrder,
			}.Apply(ctx)
		},
	})
}
