package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ca7d093c-0533-493f-9ad3-8af30118fbfc",
		Name:         "Abraded Bluffs",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Abraded Bluffs — 1 damage to target opponent",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{R|W}",
			Label:                   "Add {R} or {W}",
			IgnoreCommanderIdentity: true,
		}},
	})
}
