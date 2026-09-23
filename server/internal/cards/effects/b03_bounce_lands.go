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
// the template. These four use the same shape: SelfEntersTapped, the
// real CR 614 replacement, and ReturnOneYouControl for the bounce.
//
// The bounce is the Chancery's own shape: a CHOICE, not a target —
// "return a land you control", with no "target" in the printed text
// — made on resolution (ReturnOneYouControl, the own_permanents
// prompt, #1214). It used to be a target clause picked when the
// trigger went on the stack, a declared simplification that let
// opponents see and answer the choice before the resolution-time
// picker existed. Each land here is always a candidate while it is
// still on the battlefield, and returning itself is a normal,
// sometimes correct, line.
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
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + l.a + "}{" + l.b + "}",
				Label:    "Add {" + l.a + "}{" + l.b + "}",
			}},
			Triggered: []game.TriggeredAbility{
				WhenThisEnters(name+" — return a land you control", Do(ReturnOneYouControl{
					Match:    MatchLand,
					Question: name + " — return a land you control to its owner's hand",
				})),
			},
		})
	}
}
