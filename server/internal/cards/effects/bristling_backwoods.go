package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristling Backwoods — Land — Desert (EDHREC rank 3590):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {R} or {G}."
//
// The Gruul member of Outlaws of Thunder Junction's tapped Desert
// duals — Abraded Bluffs' shape exactly. The tapped entry is the
// real CR 614 self-replacement; the enters trigger is targeted
// ("target opponent") and dealt by the land, so it is colourless
// noncombat damage; the two printed colours are not narrowed to the
// commander's identity (the painland posture — the text does not
// say "in your commander's color identity").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cbc9f83-8979-42a5-a466-a8d89c8e6de8",
		Name:         "Bristling Backwoods",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bristling Backwoods — 1 damage to target opponent",
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
			Produced:                "{R|G}",
			Label:                   "Add {R} or {G}",
			IgnoreCommanderIdentity: true,
		}},
	})
}
