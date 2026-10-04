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
// "Other" is object identity (effects.RemoveCountersAmongOthers, CR 109.1):
// the counters come off permanents other than Tekuthal himself.
//
// The indestructible counter is CR 122.1b: a permanent with an
// indestructible counter on it has indestructible. The engine reads
// keyword counters itself (ADR 0101), in layer 6 at the counter's own
// timestamp, so the card declares nothing for it — and a Tekuthal that
// lost all its abilities BEFORE the counter arrived is still
// indestructible, which the static it used to carry got wrong.
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
		Replacements: []game.ReplacementEffect{
			ProliferateTwice("Tekuthal, Inquiry Dominus — proliferate twice"),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{U/P}{U/P}, Remove three counters from among other artifacts, creatures, and planeswalkers you control: Put an indestructible counter on Tekuthal, Inquiry Dominus.",
			Cost: Plus(
				ManaCost("{1}{U/P}{U/P}"),
				RemoveCountersAmongOthers("", 3, "other artifacts, creatures, and planeswalkers you control",
					Or(Artifact(), Creature(), Planeswalker())),
			),
			Effect: putIndestructibleCounterOnSource(),
		}},
	})
}
