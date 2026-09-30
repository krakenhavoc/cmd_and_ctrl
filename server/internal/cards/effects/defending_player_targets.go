package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TargetCreatureDefendingPlayerControls is the TargetsFrom clause for
// "whenever this creature attacks, ... target creature defending
// player controls" (Goblin Racketeer, Coveted Peacock, #1599): the
// defending player is a fact about THIS attack, not about the
// caster, so it has to be read off the source's own AttackingTarget —
// stamped before the EventAttack that fires this trigger, see
// attack_target.go — rather than a static Targets clause, whose
// predicates only ever see the permanent's controller as "caster".
//
// Returning nil (no defending player — should not happen for a
// creature whose own attack just fired this trigger, but CR 603.3d's
// answer for "targets nothing after all" is to drop the clause, never
// to error) leaves the ability with no target to act on; GoadTarget's
// own "no legal target" check is what makes that a no-op rather than
// a panic.
func TargetCreatureDefendingPlayerControls(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature defending player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == defender
		},
		Min: 1, Max: 1,
	}
}
