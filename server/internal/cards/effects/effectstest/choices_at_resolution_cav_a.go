package effectstest

// choices_at_resolution_cav_a_test.go — the #3172 caveat slice "choices
// made at the right time": each card here used to settle a choice when
// its ability went on the stack, or approximate one with a fixed rule, and
// now asks as it resolves. The per-card tests beside each card prove the
// ordinary path; these cover the corners the move opened.

const (
	CavATechnomancerOracle = "4e58ad76-37c7-4531-b207-6890b39a2679"
	CavAGlissaOracle       = "8859a692-d107-4de2-9110-4681fc2761c7"
)
