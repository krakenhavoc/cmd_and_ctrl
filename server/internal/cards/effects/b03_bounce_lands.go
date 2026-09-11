package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bounce_lands.go — the Ravnica "bounce lands" (karoos) the
// roadmap's batch 03 (#296) ranks: Dimir Aqueduct, Orzhov Basilica,
// Izzet Boilerworks, Gruul Turf.
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {A}{B}."
//
// Azorius Chancery is already in the catalog as a single file and is
// the template; it is deliberately left untouched (it taps itself in
// OnETB, a workaround from before the entering card's own
// replacements were consulted). These four use the current shape —
// SelfEntersTapped, the real CR 614 replacement — and are otherwise
// the Chancery row for row, including its one simplification:
//
// Sandbox simplification (the Chancery's, inherited): "return a land
// you control" is a CHOICE, not a target — there is no "target" in
// the printed text. It is modelled as a target clause so the
// controller gets the existing board picker, which makes the choice
// targetable in response and locks it in at announce rather than at
// resolution. The trigger can never fizzle for want of a legal
// choice: the land itself is always a land its controller controls,
// and bouncing itself is a legal, sometimes correct, line.
func init() {
	for _, l := range []struct{ oracleID, name, a, b string }{
		{"378a1d57-e2f1-4b84-9692-1564602e9e99", "Dimir Aqueduct", "U", "B"},
		{"aa00ae0b-7c0f-427e-8102-ce0e2a6af5df", "Orzhov Basilica", "W", "B"},
		{"1cb9d94a-3039-4f2e-8fcc-6996f9a45f74", "Izzet Boilerworks", "U", "R"},
		{"657243dd-e479-4f4b-99d2-09b55d833a35", "Gruul Turf", "R", "G"},
	} {
		name := l.name
		Register(Spec{
			OracleID:     l.oracleID,
			Name:         name,
			Completeness: CompletenessCaveats,
			Caveats:      []string{"You pick the land to return when the trigger goes on the stack rather than on resolution, so opponents can respond to the choice."},
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + l.a + "}{" + l.b + "}",
				Label:    "Add {" + l.a + "}{" + l.b + "}",
			}},
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetPermanent("a land you control", And(Land(), YouControl())),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, name+" — return a land you control",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 {
								return nil
							}
							return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
						})
				},
			}},
		})
	}
}
