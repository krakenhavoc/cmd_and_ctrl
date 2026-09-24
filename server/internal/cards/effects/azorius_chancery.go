package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Chancery — Land:
//
//	"This land enters tapped. When this land enters, return a land
//	you control to its owner's hand. {T}: Add {W}{U}."
//
// A "bounce land": the tempo cost is real (it enters tapped AND
// sets you back a land drop), and the payoff is that one land taps
// for two coloured mana.
//
// Enters-tapped is a real CR 614 self-replacement (SelfEntersTapped),
// the same one the karoo table in bounce_lands.go uses: the land is
// never untapped on the battlefield and no tap event is emitted. This
// file used to say a catalog replacement could not fire on its own
// source's entry and tapped the land from a hook a beat later; that
// stopped being true with the Temple cycle's entering-card block in
// gatherActiveReplacementsLocked (#360, #578).
//
// The bounce is a CHOICE, not a target — "return a land you control",
// with no "target" in the oracle text — so it is made on resolution
// (ReturnOneYouControl, the own_permanents prompt). It used to be a
// target clause picked when the trigger went on the stack, a declared
// simplification that let opponents see and answer the choice; the
// resolution-time pick over one's own permanents (#1214) retired it.
// The Chancery itself is always a candidate while it is still on the
// battlefield, and returning itself is a normal, sometimes correct,
// line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "189fc8f4-17ac-4f1d-82c8-8401445bdaf4",
		Name:         "Azorius Chancery",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}{U}",
			Label:    "Add {W}{U}",
		}},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Azorius Chancery — return a land you control", Do(ReturnOneYouControl{
				Match:    MatchLand,
				Question: "Azorius Chancery — return a land you control to its owner's hand",
			})),
		},
	})
}
