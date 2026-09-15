package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beledros Witherbloom — Legendary Creature — Elder Dragon {5}{B}{G},
// 4/4 (EDHREC rank 1458):
//
//	"Flying
//	 At the beginning of each upkeep, create a 1/1 black and green
//	 Pest creature token with "When this token dies, you gain 1
//	 life."
//	 Pay 10 life: Untap all lands you control. Activate only once
//	 each turn."
//
// A Pest every upkeep — every player's, as printed — on a flying
// body. The upkeep trigger has no actor test (Tendershoot Dryad's
// shape).
//
// Two sandbox simplifications, declared, both weaker than printed:
//
//   - The Pest's "when this token dies, you gain 1 life" is not on
//     the token. A token template carries no triggered abilities and
//     a token has no oracle ID for the catalog to key one on, so the
//     Pest is a 1/1 body and nothing more. Registering a synthetic
//     spec for the token would publish it in the public catalogue as
//     a card, which is why it is not done.
//   - The activated ability is not implemented. "Activate only once
//     each turn" is a restriction on ACTIVATION, and an activated
//     ability has no once-per-turn shape (only a loyalty cost
//     carries one). Shipping the untap without the limit would be
//     stronger than printed (#259); gating it at resolution would
//     charge 10 life for a second activation that does nothing,
//     which is a trap rather than a card. So the untap waits on a
//     per-turn activation limit, and the caveat says so.
func init() {
	Register(Spec{
		OracleID:     "90194ff1-db61-463f-b5a3-15cd85311d0e",
		Name:         "Beledros Witherbloom",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Pest tokens don't gain you 1 life when they die.",
			"The \"Pay 10 life: Untap all lands you control\" ability isn't implemented.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Beledros Witherbloom — create a Pest", Do(CreateToken{Template: b13PestToken(), N: 1})),
		},
	})
}
