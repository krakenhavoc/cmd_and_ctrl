package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torrent of Lava — Sorcery {X}{R}{R} (Mirage, #5633):
//
//	"Torrent of Lava deals X damage to each creature without flying.
//	 As long as Torrent of Lava is on the stack, each creature has "{T}: Prevent
//	 the next 1 damage that would be dealt to this creature by Torrent of Lava
//	 this turn.""
//
// The second paragraph is a stack-zone static (CR 113.6b, #6222) granting a
// bundle to every creature; each creature's ability names the spell through its
// grantor (StackItem.GrantedBy) and shields that creature from it by 1.
//
// No simplification.
func init() {
	const grant = "torrent-of-lava/prevent-one"
	effects.Register(effects.Spec{
		OracleID:     "f467ec85-ffe2-4963-82e0-509a7873fd63",
		Name:         "Torrent of Lava",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to this creature by Torrent of Lava this turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.TapCost(),
				Effect:  effects.OldB795PreventOneFromTheGrantingSpell,
			}},
			Text: "{T}: Prevent the next 1 damage that would be dealt to this creature by Torrent of Lava this turn.",
		}},
		Static: []game.StaticAbility{func() game.StaticAbility {
			s := effects.GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() }, grant)
			s.Label = "As long as Torrent of Lava is on the stack, each creature has \"{T}: Prevent the next 1 damage that would be dealt to this creature by Torrent of Lava this turn.\""
			s.Zones = []game.ZoneKind{game.ZoneStack}
			return s
		}()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), ctx.X())
		},
	})
}
