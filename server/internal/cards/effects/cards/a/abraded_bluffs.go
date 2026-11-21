package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abraded Bluffs — Land — Desert (EDHREC rank 2949):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {R} or {W}."
//
// The Boros member of Outlaws of Thunder Junction's tapped Desert
// duals. The tapped entry is the real CR 614 self-replacement; the
// enters trigger is targeted ("target opponent") and dealt by the
// land, so it is colourless noncombat damage; the two printed colours
// are not narrowed to the commander's identity (the painland
// posture — the text does not say "in your commander's color
// identity").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca7d093c-0533-493f-9ad3-8af30118fbfc",
		Name:         "Abraded Bluffs",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:       "Abraded Bluffs — 1 damage to target opponent",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
	})
}
