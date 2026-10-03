package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zealot il-Vec — Creature — Human Rebel {2}{W}, 1/1:
//
//	"Shadow (This creature can block or be blocked by only creatures with shadow.)
//	 Whenever this creature attacks and isn't blocked, you may have it deal 1 damage to target creature. If you do, prevent all combat damage this creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the trigger targets a creature
// (it must, even if you won't use it; its ruling), and "you may" is asked
// as it resolves. On yes the Zealot deals 1 damage to the target, then
// its own combat damage this turn is prevented — even if that 1 damage
// was itself prevented or replaced (its ruling). The shield's source is
// the Zealot while it is still the same object (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4107d0e3-608b-4024-b433-40c5d8b63549",
		Name:            "Zealot il-Vec",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenAttacksAndIsNotBlockedEffect("Zealot il-Vec — you may have it deal 1 damage to target creature",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					ctx := NewContext(g, item)
					var target uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							target = t.ID
						}
					}
					if target == uuid.Nil {
						return nil
					}
					return MayChoice{
						Question: "Zealot il-Vec — deal 1 damage to the target creature? (if you do, its combat damage this turn is prevented)",
						OnYes: func(ctx *Context) error {
							if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: 1}).Apply(ctx); err != nil {
								return err
							}
							return shieldAgainstThisCombatDamage(ctx)
						},
					}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}
