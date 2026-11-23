package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prowling Pangolin — Creature — Pangolin Beast {3}{B}{B}, 6/5:
//
//	"When this creature enters, any player may sacrifice two creatures
//	 of their choice. If a player does, sacrifice this creature."
//
// ADR 0150 (#1903): the sacrifice is offered to every player in APNAP
// order from the active player, the Pangolin's controller included,
// until one makes it. The payer picks the two creatures, and a player
// with fewer than two can't pay (CR 118.3). A Pangolin that has already
// left, or come back as a new object, is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92042cad-09f5-464f-85ad-baad27e6d1eb",
		Name:         "Prowling Pangolin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prowling Pangolin — any player may sacrifice two creatures to make you sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AnyPlayerPays{
						Pay:      effects.SacrificePayment(2, "creature", "creatures", effects.QueryType("creature")),
						Question: "Prowling Pangolin — sacrifice two creatures to make its controller sacrifice it?",
						Then:     prowlingPangolinSacrificed,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// prowlingPangolinSacrificed is "if a player does", keyed (#4264).
var prowlingPangolinSacrificed = effects.AnyPlayerConsequences("any-player/prowling-pangolin-sacrifice",
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error { return effects.SacrificeThisIfStillHere(ctx) },
	nil,
	nil)
