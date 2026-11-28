package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB760AttackIfAnotherWithCounterAttacks is Magnetic Web's "If a
// creature with a magnet counter on it attacks, all creatures with magnet
// counters on them attack if able": a static on the Web that gives every
// creature carrying a `kind` counter one CR 508.1d requirement, which exists
// only in a declaration in which another creature with such a counter is
// attacking (game.AttackRequirement.IfAnotherWithCounterAttacks). Like
// Viashino Bey's and Ekundu Cyclops' lines it is a constraint on declaring
// an attack, not a trigger.
func S82OldB760AttackIfAnotherWithCounterAttacks(kind, label string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		Label: label,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature() && target.Counters[kind] > 0
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := requirementFrom(source, uuid.Nil)
			r.IfAnotherWithCounterAttacks = kind
			c.AttackRequirements = append(c.AttackRequirements, r)
		},
	}
}

// S82OldB760EachWithCounterBlocksTheAttacker is Magnetic Web's second
// ability's effect: "all creatures with magnet counters on them block that
// creature this turn if able". "That creature" is the attacker the trigger's
// event names, as it was when it attacked (CR 603.10); each creature with a
// `kind` counter now gets Provoke's "blocks that attacker this turn if able"
// record, which asks nothing of one that is not a defender or cannot block.
func S82OldB760EachWithCounterBlocksTheAttacker(kind, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		attacker := AttackingObjectOf(ctx)
		for _, c := range g.BattlefieldCardsForEffect() {
			if !c.IsCreature() || c.Counters[kind] <= 0 || c.InstanceID == attacker.ID {
				continue
			}
			if err := (BlocksAttackerUntilEOT{Blocker: c.InstanceID, Attacker: attacker, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
