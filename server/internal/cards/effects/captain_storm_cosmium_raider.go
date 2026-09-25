package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Storm, Cosmium Raider — 2/2 Legendary Creature — Human
// Pirate for {U}{R}:
//
//	"Whenever an artifact you control enters, put a +1/+1 counter on
//	 target Pirate you control."
//
// The deck makes artifacts constantly — every Treasure off the
// commander, Corsair Captain, Ragavan — so this is a two-drop that
// turns the Treasure engine into a clock. First card in the catalog
// to pair the artifact-ETB watcher with a target clause.
//
// The printed trigger is per-artifact ("Whenever AN artifact...
// enters"), not the batched "one or more" wording, so firing once
// per artifact and requiring a fresh legal Pirate each time (CR
// 603.3d — no prompt, no ability, if none exists) matches the card
// exactly. No simplification.
func init() {
	Register(Spec{
		OracleID:     "431e85e3-15e6-471b-ba71-6058394c9a96",
		Name:         "Captain Storm, Cosmium Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			},
			Targets: TargetPermanent("target Pirate you control", YouControl(), IsPirateCard()),
			Key:     "Captain Storm — +1/+1 counter on a Pirate",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: item.Targets[0].ID, Kind: "+1/+1", N: 1}.Apply(ctx)
			},
		}},
	})
}
