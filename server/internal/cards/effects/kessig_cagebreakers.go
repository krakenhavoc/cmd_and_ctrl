package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kessigCagebreakersLabel = "Kessig Cagebreakers — a tapped and attacking Wolf for each creature card in your graveyard"

// Kessig Cagebreakers — Creature — Human Rogue {4}{G}, 3/4:
//
//	"Whenever this creature attacks, create a 2/2 green Wolf creature
//	 token that's tapped and attacking for each creature card in your
//	 graveyard."
//
// Hanweir Garrison's shape: the tokens attack what the Cagebreakers
// attacked (read off the attack event; CR 506.4), were never declared
// as attackers (CR 508.4), and the graveyard is counted as the trigger
// resolves. A defender gone by then gets no tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7407920-7596-497a-a135-b9343909a0cb",
		Name:         "Kessig Cagebreakers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: ThisAttacked,
			Key:       kessigCagebreakersLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, kessigCagebreakersLabel)
				item.Params.Player = b17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil {
					return nil
				}
				n := creatureCardsInGraveyard(g, item.Controller)
				if n == 0 {
					return nil
				}
				tmpl := TokenCard("2/2 green Wolf")
				tmpl.Tapped = true
				tmpl.AttackingTarget = defender
				return CreateToken{Controller: item.Controller, Template: tmpl, N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
