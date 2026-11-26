package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b363_helpers.go — S82 OLD batch 363 (Conspiracy, #5557).

// S82OldB363TargetAttacksThisTurnIfAble is "Target creature attacks this
// turn if able" as an activated ability's body (Basandra, Battle Seraph;
// Heckling Fiends): Imps' Taunt's attack requirement (AddAttackRequirementMod,
// checked as attackers are declared, CR 508.1d) until end of turn. It binds
// only the creature's controller, and only if that creature can attack this
// turn (CR 508.1d); a target that left is skipped (CR 608.2b).
func S82OldB363TargetAttacksThisTurnIfAble(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label + " — attacks this turn if able",
		}.Apply(ctx)
	}
}
