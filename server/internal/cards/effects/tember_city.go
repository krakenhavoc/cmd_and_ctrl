package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tember City — Plane — Kinshala (Planechase Anthology, #4870):
//
//	"Whenever a player taps a land for mana, Tember City deals 1 damage to
//	 that player.
//	 Whenever chaos ensues, each other player sacrifices a nonland
//	 permanent of their choice."
//
// A plane (ADR 0216 §6). The first ability is Manabarbs' (the tapped land's
// mana ability, the damage dealt to its player by the plane). The chaos
// ability is an edict on every player but the planar controller (Grave
// Pact's shape), each choosing their own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbb1d13c-a3ce-4151-883a-cce9bbd584d0",
		Name:         "Tember City",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			temberCityDamage(),
			WheneverChaosEnsues("Tember City — each other player sacrifices a nonland permanent",
				Do(EachPlayerSacrifices{ExceptController: true, Match: Nonland(), Label: "a nonland permanent"})),
		},
	})
}

func temberCityDamage() game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventManaAbilityActivated},
		AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			return b20LandTappedForMana(ev, g)
		},
		Key: "Tember City — deal 1 damage to that player",
		Effect: func(g *game.Game, item *game.StackItem) error {
			return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 1}.Apply(NewContext(g, item))
		},
	}
}
