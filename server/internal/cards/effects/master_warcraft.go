package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master Warcraft — Instant {2}{R/W}{R/W}:
//
//	"Cast this spell only before attackers are declared.
//	 You choose which creatures attack this turn.
//	 You choose which creatures block this turn and how those creatures
//	 block."
//
// The timing is a CastCondition: the turn's first declare attackers
// step has not begun (the ruling: "If a turn has multiple combat
// phases, this spell can only be cast before the beginning of the
// Declare Attackers Step of the first combat phase in that turn"). The
// turn tally counts steps as they begin, so an additional combat phase
// does not reopen it.
//
// The rest is one DeclarationControl for the turn, attackers and
// blockers (game/declaration_control.go, #4154): in every declare
// attackers step this turn the caster chooses which of the active
// player's creatures attack, and the active player then chooses what
// each attacks and pays its costs (the ruling: "the person who cast
// Master Warcraft first chooses the complete group of creatures that
// are going to attack. Then, for each of those creatures, the active
// player chooses who or what it's going to attack."); in every declare
// blockers step the caster declares every other defending player's
// blocks. Every restriction and requirement still applies ("Your
// choices must be legal within the normal rules for attacking and
// blocking"), and the caster may choose no attackers or no blocks
// ("You can decide that a creature won't block"). The effect takes
// precedence over a player-control effect (ADR 0167's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9bca7af-29b6-46ba-b847-56715d09d1da",
		Name:         "Master Warcraft",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return beforeAttackersDeclaredThisTurn(g)
		},
		CastConditionLabel: "Cast this spell only before attackers are declared.",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.BeginDeclarationControlForEffect(game.DeclarationControl{
				Declarer:   item.Controller,
				Source:     item.SourceCardID,
				SourceName: "Master Warcraft",
				Attackers:  true,
				Blockers:   true,
			})
			return nil
		},
	})
}

// beforeAttackersDeclaredThisTurn reports whether no declare attackers
// step has begun this turn (CR 506.8a, 506.8d; #4062's turn-tally read).
// A turn that reached its postcombat main phase or its end without one
// skipped its combat, and then the window closed with the precombat
// main phase (CR 506.8e).
func beforeAttackersDeclaredThisTurn(g *game.Game) bool {
	if g.DeclareAttackersHasBegunThisTurnForEffect() {
		return false
	}
	switch game.PhaseOf(g.Turn.Step) {
	case game.PhasePostcombatMain, game.PhaseEnding:
		return false
	}
	return true
}
