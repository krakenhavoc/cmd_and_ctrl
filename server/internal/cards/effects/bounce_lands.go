package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bounce_lands.go — the Ravnica "bounce lands" / karoos:
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {G}{U}."
//
// Azorius Chancery's shape, as a table: the tempo cost (tapped, and a
// land drop set back) buys one land that taps for two coloured mana.
// Roadmap batch 02 (#295) ranks Simic Growth Chamber and Golgari Rot
// Farm; batch 11 (#304) adds Guildless Commons, the colourless one;
// the rest of the cycle belongs here when it arrives.
//
// One difference from the Chancery file, deliberate:
//
//   - The tapped entry is a real CR 614 self-replacement. The
//     Chancery's comment says a catalog replacement cannot fire on
//     its own entry; that was true when it was written and has not
//     been since the Temple cycle (gatherActiveReplacementsLocked's
//     third block). The Chancery file itself is left to its owner.
//
// The bounce is the Chancery's own shape: a CHOICE, not a target —
// "return a land you control", with no "target" in the oracle text —
// made on resolution (ReturnOneYouControl, the own_permanents
// prompt, #1214). It used to be a target clause picked when the
// trigger went on the stack, a declared simplification that let
// opponents see and answer the choice before the resolution-time
// picker existed. Each land in the cycle is always a candidate while
// it is still on the battlefield, and returning itself is a normal,
// sometimes correct, line.
func init() {
	for _, land := range []struct{ oracleID, name, produced string }{
		{"046f5783-cc7b-416a-8cf6-2bcef9c2cc1a", "Simic Growth Chamber", "{G}{U}"},
		{"1b301478-b14f-4ef8-94e6-9647d582eabe", "Golgari Rot Farm", "{B}{G}"},
		// Roadmap batch 11 (#304): the colourless karoo — same three
		// clauses, "{T}: Add {C}{C}".
		{"ee723c7c-ec9f-4ffb-8f36-cd7637eb1fae", "Guildless Commons", "{C}{C}"},
	} {
		name := land.name
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         land.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: land.produced,
				Label:    "Add " + land.produced,
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
