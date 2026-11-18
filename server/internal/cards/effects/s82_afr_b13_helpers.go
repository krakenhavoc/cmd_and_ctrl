package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b13_helpers.go — shared helpers for AFR batch 13 (Warlock Class).

// CondAnyCreatureDied is "if a creature died this turn" as an intervening-if
// Condition, for InterveningIf and If: any player's creature, a token
// included, read off the game-wide turn tally as the item resolves. The tally
// only rises within a turn, so it declares a footprint condition the
// trigger-independence check can argue (game.FootprintCondAnyCreatureDied).
type CondAnyCreatureDied struct{}

func (CondAnyCreatureDied) Holds(ctx *Context) bool {
	return b11CreaturesDiedThisTurn(ctx.Game) > 0
}

func (CondAnyCreatureDied) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondAnyCreatureDied, 0), true
}

// AtYourEndStepIfAnyCreatureDied is "At the beginning of your end step, if a
// creature died this turn, …": the intervening "if" is checked as the end
// step begins and again as the trigger resolves (CR 603.4).
func AtYourEndStepIfAnyCreatureDied(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondAnyCreatureDied{},
		On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			return b11CreaturesDiedThisTurn(g) > 0
		}), label, effect))
}

// BlocksAttackerThisCombat is "<Blocker> blocks <Attacker> this combat if
// able" (Fighter Class): BlocksAttackerUntilEOT's record, lasting until the
// end of the combat phase it was made in (CR 511.3), so an additional combat
// this turn is a different combat. Outside a combat phase it registers
// nothing (CR 611.2b), and a zero Blocker or Attacker registers nothing.
type BlocksAttackerThisCombat struct {
	Blocker  uuid.UUID
	Attacker game.ObjectRef
	Label    string
}

// Apply registers the requirement.
func (b BlocksAttackerThisCombat) Apply(ctx *Context) error {
	if b.Blocker == uuid.Nil || b.Attacker.ID == uuid.Nil {
		return nil
	}
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   b.Blocker,
		Mods:     []game.Mod{game.BlocksAttackerMod(b.Attacker)},
		Duration: d,
		Label:    b.Label,
	}.Apply(ctx)
}
