package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assault Formation — Enchantment {1}{G} (Tarkir: Dragonstorm Commander):
//
//	"Each creature you control assigns combat damage equal to its
//	 toughness rather than its power.
//	 {G}: Target creature with defender can attack this turn as though it
//	 didn't have defender.
//	 {2}{G}: Creatures you control get +0/+1 until end of turn."
//
// The first ability is a layer-6 static that sets
// Characteristic.DamageByToughnessAlways on every creature its controller
// has (CR 510.1c). It is the unconditional form: a 4/1 assigns 1, a 2/3
// assigns 3 (Ghalta's flag only answers with the toughness when it is the
// greater, which would let a 4/1 keep its 4). No power changes, so every
// rule that reads power still sees the real value. The waiver is Hightide
// Hermit's (a defender waiver on the target until end of turn, the keyword
// staying), and the pump is Dawnwing Marshal's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d6d3e6d-f958-4c17-b117-960fb8033601",
		Name:         "Assault Formation",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{G}: Target creature with defender can attack this turn as though it didn't have defender.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{G}"),
				Targets: TargetCreature("target creature with defender", HasKeyword("defender")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return waiveDefenderUntilEOT(ctx, t.ID, "Assault Formation — can attack this turn as though it didn't have defender")
						}
					}
					return nil
				},
			},
			{
				Label:   "{2}{G}: Creatures you control get +0/+1 until end of turn.",
				Cost:    ManaCost("{2}{G}"),
				Purpose: game.Purpose{TeamPump: &game.Pump{Toughness: 1}, Answers: game.AnswerPump},
				Effect: Do(BoostUntilEOT{
					Match: And(Creature(), YouControl()), Toughness: 1,
					Label: "Assault Formation — creatures you control get +0/+1",
				}),
			},
		},
	})
}
