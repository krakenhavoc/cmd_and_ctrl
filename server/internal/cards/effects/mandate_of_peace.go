package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mandate of Peace — {1}{W} Instant:
//
//	"Cast this spell only during combat.
//	 Your opponents can't cast spells this turn.
//	 End the combat phase. (Remove all attackers and blockers from
//	 combat. Exile all spells and abilities from the stack, including
//	 this spell.)"
//
// The one card CR 724.2 names ("One card (Mandate of Peace) ends the
// combat phase"), and the proof card for #1317: the third sentence is
// game.EndCombatPhaseForEffect, which exiles the whole stack — this
// spell included — removes everything from combat and walks the cursor
// straight to the postcombat main phase without entering the end of
// combat step (CR 724.2e). ADR 0045's 2026-09-23 amendment has the
// step-by-step mapping. It is the last instruction, as the verb
// requires: nothing is left to run once the spell has exiled itself.
//
// "Cast this spell only during combat" is the CR 307.6-family
// CastCondition (ADR 0073 §7): the combat phase, from beginning of
// combat through end of combat — PhaseOf, not a step list, so a step
// added to the phase later is covered without touching this file.
//
// The second sentence is NOT here, and it is the caveat. "Your
// opponents can't cast spells this turn" is a cast restriction
// CREATED by a resolving spell, with a duration — the engine's cast
// restrictions come from permanents on the battlefield
// (cast_restriction.go), and a spell in exile contributes nothing.
// That shape is #1316, which a separate PR takes; when it lands this
// card gains the sentence and loses the caveat. Shipping without it is
// weaker than printed, never stronger — every opponent keeps the
// spells the card would have taken from them, which is the direction
// #259 permits.
func init() {
	Register(Spec{
		OracleID:     "c309ec42-34a0-4083-a36c-7814643d7960",
		Name:         "Mandate of Peace",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Your opponents can still cast spells for the rest of the turn — only the combat-ending half works.",
		},
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
		},
		CastConditionLabel: "Cast this spell only during combat.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.EndCombatPhaseForEffect()
			return nil
		},
	})
}
