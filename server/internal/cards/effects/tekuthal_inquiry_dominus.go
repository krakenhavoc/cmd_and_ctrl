package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tekuthal, Inquiry Dominus — Legendary Creature — Phyrexian Horror
// {2}{U}{U}, 3/5:
//
//	"Flying
//	 If you would proliferate, proliferate twice instead.
//	 {1}{U/P}{U/P}, Remove three counters from among other artifacts,
//	 creatures, and planeswalkers you control: Put an indestructible
//	 counter on Tekuthal, Inquiry Dominus."
//
// The card the counter-cost seam was waiting on (#943). Its activated
// ability is the LAST printed shape of AbilityCost.RemoveCounters: a
// removal of any kind SPLIT across permanents. Every other among-cost
// names its kind (Iron Spider's +1/+1, Hopeful Initiate's +1/+1), so
// the payment is one kind and a count per permanent; Tekuthal takes
// three counters of whatever kinds happen to be there, so a loyalty
// counter off a planeswalker and two +1/+1 counters off a creature is
// one legal payment. It is RemoveCountersAmong with an EMPTY kind —
// the same component and the same constructor, with the kind question
// asked once per permanent instead of once per payment.
//
// "Other" is b03NotNamed, the posture every "sacrifice another" cost
// in the catalog takes: a cost clause is built at registration, before
// any instance exists, so the source is excluded by name. Tekuthal is
// legendary, so in a Commander game the name IS the instance — the
// usual "a token copy would be excluded too" gap cannot bite here,
// because the legend rule would not let the copy live under the same
// controller (CR 704.5j).
//
// The indestructible counter is CR 122.1b: a permanent with an
// indestructible counter on it has indestructible. The engine reads no
// keyword counters of its own, so Tekuthal carries the rule for the
// counters it places, exactly as Vraska Joins Up carries it for its
// deathtouch counters (b24KeywordCounterGrant). Vraska's declared gap
// — the grant stops when the source leaves — cannot bite here: the
// counter and the static are on the same permanent, so when Tekuthal
// leaves, its counters leave with it.
//
// The {U/P} symbols are payable either way since #971 put the
// activation-time Phyrexian announcement on ActivateAbilityParams
// (CR 107.4f): blue mana, or 2 life each.
//
// "If you would proliferate, proliferate twice instead" is a
// replacement of a keyword ACTION (CR 701.34), which the engine had no
// seam for when #943 shipped the card — ProliferateForEffect applied a
// choice and never entered the CR 614 pipeline, so the clause was a
// declared caveat and Tekuthal's proliferates happened once. #976
// built the seam: a proliferate is now RepEventKeywordAction, opened
// once per instruction with a count of TIMES, and the clause is
// ProliferateTwice — one line, no per-card machinery.
//
// Two Tekuthals are four proliferates and nobody is asked to order
// them: two objects contributing one declared effect is #792's
// identical-window skip, and ×2 then ×2 is ×4 either way.
func init() {
	Register(Spec{
		OracleID:        "4716ab91-30e6-4c63-8389-a9db8f9414d8",
		Name:            "Tekuthal, Inquiry Dominus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{b24KeywordCounterGrant("indestructible")},
		Replacements: []game.ReplacementEffect{
			ProliferateTwice("Tekuthal, Inquiry Dominus — proliferate twice"),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{U/P}{U/P}, Remove three counters from among other artifacts, creatures, and planeswalkers you control: Put an indestructible counter on Tekuthal, Inquiry Dominus.",
			Cost: Plus(
				ManaCost("{1}{U/P}{U/P}"),
				RemoveCountersAmong("", 3, "other artifacts, creatures, and planeswalkers you control",
					Or(Artifact(), Creature(), Planeswalker()),
					b03NotNamed("Tekuthal, Inquiry Dominus")),
			),
			Effect: putCounterOnSourceWhileOnBattlefield("indestructible", 1),
		}},
	})
}
