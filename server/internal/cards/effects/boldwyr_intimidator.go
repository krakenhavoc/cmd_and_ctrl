package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	becomes := func(subtype, label string) Effect {
		return func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return untilEndOfTurn(ctx, id, nil, label, game.AddSubtypesMod(ctx.TextSubtypes(subtype)...))
		}
	}
	Register(Spec{
		OracleID:     "6ad1d8d0-06c0-4804-af5f-60ffb358f371",
		Name:         "Boldwyr Intimidator",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(OfCreatureType("Coward")), OnMatching(OfCreatureType("Warrior")),
				"Cowards can't block Warriors"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: Target creature becomes a Coward until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{R}"),
				Targets: TargetCreature("target creature"),
				Effect:  becomes("Coward", "Boldwyr Intimidator — becomes a Coward until end of turn"),
			},
			{
				Label:   "{2}{R}: Target creature becomes a Warrior until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{2}{R}"),
				Targets: TargetCreature("target creature"),
				Effect:  becomes("Warrior", "Boldwyr Intimidator — becomes a Warrior until end of turn"),
			},
		},
	})
}
