package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aragorn, the Uniter — Legendary Creature — Human Noble {R}{G}{W}{U}, 5/5:
//
//	"Whenever you cast a white spell, create a 1/1 white Human Soldier
//	 creature token.
//	 Whenever you cast a blue spell, scry 2.
//	 Whenever you cast a red spell, Aragorn deals 3 damage to target
//	 opponent.
//	 Whenever you cast a green spell, target creature gets +4/+4 until
//	 end of turn."
//
// Four independent cast triggers (K'rrik's WheneverYouCast over OfColor).
// A spell of two of the colours triggers both abilities, once each, and a
// colourless one triggers none. The colour is read off the spell on the
// stack, through the source's text (YouCast). The two targeted triggers
// choose their target as they go on the stack; the damage is Aragorn's, so
// it is noncombat damage from a four-colour source.
//
// No simplification.
func init() {
	damage := effects.WheneverYouCast(effects.OfColor("R"), "Aragorn, the Uniter — 3 damage to target opponent",
		func(g *game.Game, item *game.StackItem) error {
			for _, t := range item.Targets {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(effects.NewContext(g, item))
			}
			return nil
		})
	damage.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	pump := effects.WheneverYouCast(effects.OfColor("G"), "Aragorn, the Uniter — target creature gets +4/+4 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Aragorn, the Uniter — +4/+4"}.Apply(ctx)
			}
			return nil
		})
	pump.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "3f31c423-d45c-4c3e-b7d4-ca125be0f92c",
		Name:         "Aragorn, the Uniter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.OfColor("W"), "Aragorn, the Uniter — create a 1/1 white Human Soldier",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
			effects.WheneverYouCast(effects.OfColor("U"), "Aragorn, the Uniter — scry 2", effects.Do(effects.Scry{N: 2})),
			damage,
			pump,
		},
	})
}
