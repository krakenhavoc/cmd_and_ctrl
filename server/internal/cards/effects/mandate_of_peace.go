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
// combat phase"), and the last card either #1317 or #1316 needed the
// other to finish: the combat-ending third sentence is
// game.EndCombatPhaseForEffect (#1317, shipped by PR #1343, ADR 0045's 2026-09-23
// amendment), and the cast-ban second sentence is #1316's
// game.CastBanRule{Kind: CastBanOutright} granted to each opponent
// (ADR 0066's 2026-09-23 amendment). Both land here together, so the
// card ships full with neither half caveated.
//
// "CAST THIS SPELL ONLY DURING COMBAT" is CR 307.6's family read off
// the PHASE rather than the step: any of the five combat steps
// satisfies it, which game.PhaseOf(g.Turn.Step) == game.PhaseCombat
// answers without a five-way switch.
//
// "YOUR OPPONENTS CAN'T CAST SPELLS THIS TURN" is
// game.CastBanRule{Kind: CastBanOutright} (no ExceptFromZone — nothing
// here is excepted, unlike Avatar's Wrath) granted to each opponent
// with DurationUntilEndOfTurn — #1316's other named shape, "this
// turn". Granted BEFORE the combat phase ends, in the printed order:
// the third sentence is the last instruction, as the verb requires
// (nothing is left to run once the spell has exiled itself), and
// nothing about the ban depends on combat still being open.
//
// "END THE COMBAT PHASE." game.EndCombatPhaseForEffect exiles the
// whole stack — this spell included — removes everything from combat
// and walks the cursor straight to the postcombat main phase without
// entering the end of combat step (CR 724.2e). It is called LAST:
// #489's spellMovedItselfLocked is what stops the resolution frame
// from routing an already-exiled spell anywhere else afterward.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c309ec42-34a0-4083-a36c-7814643d7960",
		Name:         "Mandate of Peace",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
		},
		CastConditionLabel: "Cast this spell only during combat.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			d := DurationUntilEndOfTurn(ctx)
			for _, opp := range ctx.Opponents() {
				if err := (RestrictCasting{
					Player:   opp,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright},
					Label:    "Mandate of Peace — can't cast spells this turn",
					Duration: d,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			ctx.Game.EndCombatPhaseForEffect()
			return nil
		},
	})
}
