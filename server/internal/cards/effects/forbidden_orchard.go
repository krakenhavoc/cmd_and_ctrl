package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forbidden Orchard — Land (EDHREC rank 724):
//
//	"{T}: Add one mana of any color.
//	 Whenever you tap this land for mana, target opponent creates a
//	 1/1 colorless Spirit creature token."
//
// A five-colour land whose drawback is a political gift: every tap
// hands one opponent of your choice a 1/1. The mana ability is City of
// Brass's ("any color", so no commander-identity narrowing); the
// trigger watches EventManaAbilityActivated — which the engine emits
// for a manual activation and for the auto-tapper alike — rather than
// the tap event, because "tap this land FOR MANA" is narrower than
// "becomes tapped" (City of Brass watches the wider one, as printed).
// The trigger targets, so the controller picks the opponent when it
// fires; with no opponent it is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfd60d1f-9832-4408-b84e-0fd3018b015b",
		Name:         "Forbidden Orchard",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Source == source.InstanceID
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Forbidden Orchard — target opponent creates a 1/1 Spirit",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						return CreateToken{
							Controller: item.Targets[0].ID,
							Template:   b06ColorlessSpiritToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
