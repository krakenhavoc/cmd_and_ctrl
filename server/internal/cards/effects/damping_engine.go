package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damping Engine — Artifact {4}:
//
//	"A player who controls more permanents than each other player can't
//	 play lands or cast artifact, creature, or enchantment spells. That
//	 player may sacrifice a permanent of their choice for that player to
//	 ignore this effect until end of turn."
//
// THE RESTRICTION is two statics, both live at every query (nothing is
// stored), over one rule: a CastRestriction for the artifact, creature
// and enchantment spells and a LandPlayRestriction (ADR 0109 §4) for the
// land play, a land play not being a cast (CR 305.1). Both ask
// Game.ControlsMoreThanEachOtherPlayerForEffect, so a tie with the
// richest opponent restricts nobody, tokens count, and the Engine's own
// controller is restricted as much as anyone (CR 101.2: "a player").
//
// THE ESCAPE is CR 116.2d, a special action with the player who is
// restricted as its actor, and the cost of it a sacrifice of a permanent
// of THEIR choice (IgnorableEffect; game/ignore_effect.go). Both
// restrictions are marked Ignorable, so the one record the action writes
// (a ModIgnoreEffect naming the player and THIS Engine, until end of
// turn) switches both off at once, and another Engine on the table is
// still in force until it is paid for too.
//
// "Sacrifice a permanent": the Engine itself is among the permanents a
// player may give up. That ends the effect for everybody rather than
// ignoring it, and it is what the words allow.
//
// No simplification.
func init() {
	const (
		castLabel = "A player who controls more permanents than each other player can't cast artifact, creature, or enchantment spells."
		landLabel = "A player who controls more permanents than each other player can't play lands."
	)
	Register(Spec{
		OracleID:     "3584a055-5220-4721-89f8-de33f779122f",
		Name:         "Damping Engine",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: castLabel,
			Forbids: func(q game.CastQuery) bool {
				return q.Game.ControlsMoreThanEachOtherPlayerForEffect(q.Controller) &&
					matchCastCard(q.Game, Or(Artifact(), Creature(), Enchantment()), q.Controller, q.Card)
			},
			Ignorable: true,
		}},
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: landLabel,
			Forbids: func(q game.LandPlayQuery) bool {
				return q.Game.ControlsMoreThanEachOtherPlayerForEffect(q.Player)
			},
			Ignorable: true,
		}},
		IgnorableEffect: &game.IgnorableEffect{
			Label: "Sacrifice a permanent: ignore Damping Engine's effect until end of turn",
			Applies: func(g *game.Game, _ *game.Card, player uuid.UUID) bool {
				return g.ControlsMoreThanEachOtherPlayerForEffect(player)
			},
		},
	})
}
