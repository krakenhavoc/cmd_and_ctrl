package effects

// Undying Malice — Instant {B} (Foundations, #4465):
//
//	"Until end of turn, target creature gains "When this creature dies,
//	 return it to the battlefield tapped under its owner's control with a
//	 +1/+1 counter on it.""
//
// Feign Death's body, in registerReturnWithCounterSpell: a duration grant
// (ADR 0093) of a dies trigger that fires from last-known information.
//
// No simplification.
const undyingMaliceReturn = "undying-malice/return"

func init() {
	registerReturnWithCounterSpell("8b061702-1bc3-45aa-a758-446b25034801", "Undying Malice", undyingMaliceReturn)
}
