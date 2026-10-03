package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Nishoba — Creature — Cat Beast Spirit {5}{G}{W}, 0/0:
//
//	"Trample
//	 This creature enters with seven +1/+1 counters on it.
//	 Whenever this creature deals damage, you gain that much life.
//	 If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature."
//
// ADR 0108 §8 (#1906): one of the Phantoms (phantoms.go). "Whenever this
// creature deals damage" is read the way Spirit Link's is, off each
// damage event it deals, with the amount actually dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "e43e06fb-52b7-4f38-8fac-f31973b043f7",
		Name:            "Phantom Nishoba",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    phantomReplacements("Phantom Nishoba", 7),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Phantom Nishoba — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
