package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warriors' Lesson — Instant {G} (Theros):
//
//	"Until end of turn, up to two target creatures you control each gain
//	 "Whenever this creature deals combat damage to a player, draw a card.""
//
// Nimble Trapfinder's granted draw trigger (an ADR 0093 bundle), given to
// each target still legal as the spell resolves until end of turn
// (CR 611.2c); each creature carries its own trigger.
//
// No simplification.
func init() {
	const grant = "warriors-lesson/draw-on-damage"
	effects.Register(effects.Spec{
		OracleID:     "1d4605c0-47c5-4561-ab53-9cff677330aa",
		Name:         "Warriors' Lesson",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Whenever this creature deals combat damage to a player, draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantAbilitiesFor{Target: t.ID, Keys: []string{grant}, Label: "Warriors' Lesson — draw on combat damage"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
