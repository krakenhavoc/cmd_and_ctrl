package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b09_helpers.go — Streets of New Capenna, S82 batch 9 (#5491).

// sncSecondResolutionThisTurn is "if this is the second time this ability
// has resolved this turn" (Captain Michael Burnham's tally): the count
// includes the resolution in progress, so the second reads exactly 2.
func SncSecondResolutionThisTurn(g *game.Game, item *game.StackItem, label string) bool {
	return g.ResolvedThisTurn(item.SourceCardID, label) == 2
}

// sncCreatureAttacksDuringItsControllersNextCombat is "<creature> attacks
// during its controller's next combat phase if able" (Sizzling Soloist).
// The requirement is pinned to the creature and lasts through its
// controller's next turn (Gideon, Battle-Forged's duration), the first
// turn after this one. When its controller is the active player and combat
// has not begun yet, the next combat phase is this turn's, so it lasts to
// end of turn instead.
func SncCreatureAttacksDuringItsControllersNextCombat(ctx *Context, creature uuid.UUID, label string) error {
	g := ctx.Game
	c, ok := g.LookupCardForEffect(creature)
	if !ok {
		return nil
	}
	var d game.Duration
	active := g.Seats[g.Turn.ActiveSeat]
	switch phase := game.PhaseOf(g.Turn.Step); {
	case active != nil && c.Controller == active.ID && (phase == game.PhaseBeginning || phase == game.PhasePrecombatMain):
		d = DurationUntilEndOfTurn(ctx)
	default:
		var ok bool
		if d, ok = g.UntilEndOfItsControllersNextTurnDuration(creature); !ok {
			return nil
		}
	}
	return ScopedEffectFor{
		Target:   creature,
		Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
		Duration: d,
		Label:    label,
	}.Apply(ctx)
}
