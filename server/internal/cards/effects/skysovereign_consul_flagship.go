package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skysovereign, Consul Flagship — Legendary Artifact — Vehicle, 6/5,
// for {5}:
//
//	"Flying
//	 Whenever Skysovereign enters or attacks, it deals 3 damage to
//	 target creature or planeswalker an opponent controls.
//	 Crew 3"
//
// "Enters or attacks" is ONE printed ability with two trigger
// conditions, so it is one TriggeredAbility watching two event kinds
// — the Sun Titan shape from AGENTS.md §7. Both EventETB and
// EventAttack carry the permanent in ev.CardID, which is what lets
// the single predicate cover both.
//
// The target is chosen as the ability goes on the stack (CR 603.3d)
// and re-checked at resolution (CR 608.2b); with no legal target the
// trigger is removed with no prompt at all.
func init() {
	Register(Spec{
		OracleID:        "50b14338-9318-4327-a1dd-c0ef38903cc4",
		Name:            "Skysovereign, Consul Flagship",
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("Skysovereign, Consul Flagship"),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("target creature or planeswalker an opponent controls",
				Or(Creature(), Planeswalker()), OpponentControls()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Skysovereign — 3 damage",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return DealDamage{
							Target: item.Targets[0].ID,
							Amount: 3,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
