package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bösium Strip — Artifact {3} (Weatherlight):
//
//	"{3}, {T}: Until end of turn, you may cast instant and sorcery spells from
//	 the top of your graveyard. If a spell cast this way would be put into a
//	 graveyard, exile it instead."
//
// A standing permission over the controller's graveyard for instants and
// sorceries, open only to the card currently on top (CR 404.2, #6194) and
// re-read every time the pile changes, so a card below the top opens when the
// card above it is cast and gone. Casting pays the card's own cost and timing
// (CR 601.2); a spell cast this way goes to exile instead of a graveyard however
// it leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fa6f1b1-00af-433f-b8e1-36db99cd9bba",
		Name:         "Bösium Strip",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Until end of turn, you may cast instant and sorcery spells from the top of your graveyard. If a spell cast this way would be put into a graveyard, exile it instead.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				perm := game.CastPermission{
					Player:                  item.Controller,
					Zone:                    game.ZoneGraveyard,
					Scope:                   game.ScopeStanding,
					Duration:                g.UntilEndOfTurnDuration(),
					Source:                  item.SourceCardID,
					SourceName:              "Bösium Strip",
					Label:                   "Bösium Strip — cast an instant or sorcery from the top of your graveyard",
					AltCostKey:              game.AltCostKeyGrantedCost,
					CastOnly:                true,
					TopOfGraveyardOnly:      true,
					ExileInsteadOfGraveyard: true,
				}
				perm.Filter.InstantOrSorceryOnly = true
				g.GrantCastPermissionForEffect(perm)
				return nil
			},
		}},
	})
}
