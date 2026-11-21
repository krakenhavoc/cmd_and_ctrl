package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catti-brie of Mithral Hall — Legendary Creature — Human Archer {G}{W}, 2/2
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Reach, first strike
//	 Whenever Catti-brie attacks, put a +1/+1 counter on it for each Equipment
//	 attached to it.
//	 {1}, Remove all +1/+1 counters from Catti-brie: It deals X damage to
//	 target attacking or blocking creature an opponent controls, where X is
//	 the number of counters removed this way."
//
// The counters are counted as the attack trigger resolves, from the Equipment
// attached to Catti-brie then, whoever controls it. The activated ability
// takes every +1/+1 counter she holds (RemoveAllCountersFromThis): the
// payment is never fewer, and X is the number the paid-cost record names, so
// it is right even when she leaves in response. With no counters it is
// payable and deals 0 damage. She is the source of the damage, from her
// last-known information if she has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f44ffa35-bd88-4996-8b20-46329565df9b",
		Name:            "Catti-brie of Mithral Hall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Catti-brie — put a +1/+1 counter on it for each Equipment attached to it",
				func(g *game.Game, item *game.StackItem) error {
					if _, ok := g.LookupCardForEffect(item.SourceCardID); !ok {
						return nil
					}
					n := effects.S82AfrB08EquipmentAttachedTo(g, item.SourceCardID)
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Remove all +1/+1 counters from Catti-brie: It deals X damage to target attacking or blocking creature an opponent controls, where X is the number of counters removed this way.",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.RemoveAllCountersFromThis(game.CounterPlusOne)),
			Targets: effects.TargetCreature("target attacking or blocking creature an opponent controls",
				effects.AttackingOrBlocking(), effects.OpponentControls()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				x := ctx.CountersRemoved()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
					}
				}
				return nil
			})),
		}},
	})
}
