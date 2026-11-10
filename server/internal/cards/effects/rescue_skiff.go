package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rescue Skiff — Artifact — Spacecraft {5}{W}, 5/6 (Edge of Eternities,
// #4461):
//
//	"When this Spacecraft enters, return target creature or enchantment
//	 card from your graveyard to the battlefield.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 10+.)
//	 10+ | Flying"
//
// Karmic Guide's reanimation (reanimateSingleTarget) with a target that is a
// creature card or an enchantment card in your graveyard, judged as the
// trigger goes on the stack and again as it resolves. Station, the 10+
// creature threshold, its 5/6 body and flying are the shared constructors
// (Atmospheric Greenhouse). The Skiff is no creature as it enters, so it
// cannot return itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "112aaaf0-3301-4f79-9640-d36c75a0fb30",
		Name:         "Rescue Skiff",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rescue Skiff — return target creature or enchantment card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					reanimateSingleTarget(NewContext(g, item), item.Controller)
					return nil
				}),
				TargetCardInGraveyard("target creature or enchantment card in your graveyard", Or(Creature(), Enchantment()), YouOwn())),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(10, 5, 6),
			ThresholdKeywords(10, "flying"),
		),
	})
}
