package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daybreak Ranger // Nightfall Predator — {2}{G} Creature — Human Archer
// Ranger Werewolf 2/2 // Creature — Werewolf 4/4 (#4071, ADR 0132
// amendment 2026-10-16):
//
//	Front: "{T}: This creature deals 2 damage to target creature with
//	        flying.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "{R}, {T}: This creature fights target creature.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The ping is the creature's own damage, so it is dealt even if the
// Ranger has left the battlefield by resolution. The fight is
// Polukranos's (b30SourceFightsFirstLegalTarget): a Predator no longer
// on the battlefield fights nothing (CR 701.14b).
//
// No simplification.
func init() {
	const oracle = "280624aa-5f9a-48fd-85ea-815c96c747b3"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Daybreak Ranger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to target creature with flying.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: target, Amount: 2}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Daybreak Ranger"),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Nightfall Predator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: This creature fights target creature.",
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  b30SourceFightsFirstLegalTarget,
		}},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Nightfall Predator"),
		},
	})
}
