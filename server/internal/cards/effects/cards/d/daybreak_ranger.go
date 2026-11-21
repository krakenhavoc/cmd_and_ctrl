package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Daybreak Ranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to target creature with flying.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: 2}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Daybreak Ranger"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Nightfall Predator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: This creature fights target creature.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.B30SourceFightsFirstLegalTarget,
		}},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Nightfall Predator"),
		},
	})
}
