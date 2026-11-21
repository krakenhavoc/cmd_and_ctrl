package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raul, Trouble Shooter — Legendary Creature — Zombie Mutant Rogue
// {1}{U}{B}, 1/4 (Fallout, #5948):
//
//	"Once during each of your turns, you may cast a spell from among cards
//	 in your graveyard that were milled this turn.
//	 {T}: Each player mills a card. (They each put the top card of their
//	 library into their graveyard.)"
//
// The first line is a standing cast permission over the graveyard (ADR 0066,
// Danitha's shape) with a budget of one cast per turn (PerType permanent,
// which asks no question and spends on any spell), on its controller's own
// turn only. The card's own filter is PermissionFilter.MilledThisTurn: a
// card that arrived in a graveyard this turn as a mill's graveyard leg
// (CR 701.17a), the mark Card.Milled that the mill's route stamps as it lands
// the card. A card a surveil, a discard or a search put there is not one, nor
// is one that has left and come back (CR 400.7), and a milled land is not cast.
// Timing is the card's own (an instant on any turn of its controller's with
// priority, a sorcery in a main phase with an empty stack). Raul does not
// have to be untapped for it. The ability mills the table in turn order, the
// controller first, one card each (CR 701.17a), which is how a Raul's own
// mill finds its spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d247a86d-917d-4833-82c4-876943d54e2c",
		Name:         "Raul, Trouble Shooter",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{CardInZone: true, MilledThisTurn: true},
			PerType:  []string{game.PermissionTypePermanent},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Once during each of your turns, cast a spell from among cards in your graveyard that were milled this turn (Raul, Trouble Shooter)",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Each player mills a card.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B02EachPlayerMills(g, item, 1)
			},
		}},
	})
}
