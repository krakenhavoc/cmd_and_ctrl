package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Snowman — Creature — Goblin {3}{R}, 1/1:
//
//	"Whenever this creature blocks, prevent all combat damage that would be dealt to and dealt by it this turn.
//	 {T}: This creature deals 1 damage to target creature it's blocking."
//
// #1863: "target creature it's blocking" is BlockedBySource, judged at
// announce and again at resolution (CR 608.2b), so a Snowman removed
// from combat or an attacker that left leaves the ability with nothing
// to hit. The block trigger is one trigger however many attackers it
// blocks (CR 509.3a) and the prevention is one to-and-by record
// (ADR 0108 §7). If the Snowman has gone by resolution the damage is
// still dealt by it, as it last existed (CR 608.2h), and the target is
// still judged by what it was blocking then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38e9ad30-4bbf-4b58-8bb2-47520ce351a3",
		Name:         "Goblin Snowman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Goblin Snowman — prevent all combat damage dealt to and by it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return shieldThis(g, item, toAndByShield(ShieldTarget{}, true))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature it's blocking.",
			Cost:    TapCost(),
			Targets: BlockedBySource(TargetCreature("target creature it's blocking")),
			Effect:  goblinSnowmanPings,
		}},
	})
}

// goblinSnowmanPings deals 1 damage from the Snowman (or its last-known
// self) to the first still-legal target.
func goblinSnowmanPings(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		return DealDamage{SourceObject: &ref, Target: t.ID, Amount: 1}.Apply(ctx)
	}
	return nil
}
