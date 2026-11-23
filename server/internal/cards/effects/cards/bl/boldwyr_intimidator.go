package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boldwyr Intimidator — Creature — Giant Warrior {5}{R}{R}, 5/5
// (Foundations Jumpstart):
//
//	"Cowards can't block Warriors.
//	 {R}: Target creature becomes a Coward until end of turn.
//	 {2}{R}: Target creature becomes a Warrior until end of turn."
//
// "Becomes a Coward / Warrior" ADDS the subtype until end of turn (CR
// 205.1a, 611.2c), as Kargan Intimidator does; the block rule is the same
// one, on the blocker's side of the pair and binding every creature.
//
// No simplification.
func init() {
	becomes := func(subtype, label string) effects.Effect {
		return func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, label, game.AddSubtypesMod(ctx.TextSubtypes(subtype)...))
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "6ad1d8d0-06c0-4804-af5f-60ffb358f371",
		Name:         "Boldwyr Intimidator",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.OfCreatureType("Coward")), effects.OnMatching(effects.OfCreatureType("Warrior")),
				"Cowards can't block Warriors"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: Target creature becomes a Coward until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  becomes("Coward", "Boldwyr Intimidator — becomes a Coward until end of turn"),
			},
			{
				Label:   "{2}{R}: Target creature becomes a Warrior until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{2}{R}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  becomes("Warrior", "Boldwyr Intimidator — becomes a Warrior until end of turn"),
			},
		},
	})
}
