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
// One sandbox simplification, declared, weaker than printed:
//
//   - The Pest's "when this token dies, you gain 1 life" is not on
//     the token. A token template carries no triggered abilities and
//     a token has no oracle ID for the catalog to key one on, so the
//     Pest is a 1/1 body and nothing more. Registering a synthetic
//     spec for the token would publish it in the public catalogue as
//     a card, which is why it is not done. See the "Triggered and
//     static abilities on non-copy tokens" row in
//     docs/engine-seams.md — still open.
//
// The activated ability IS implemented. "Activate only once each
// turn" reads OncePerTurnActivation, which gates the OFFER (the
// Condition the enumerator and the client both consult before
// announcing) rather than the effect — so a second attempt this turn
// is never offered and never costs the 10 life, not a trap that
// charges for nothing. It reads Game.ResolvedThisTurn, the same
// per-OBJECT "this turn" tally (CR 400.7) every other once-per-turn
// gate in the catalog uses, keyed on this ability's own label.
//
// No tap symbol in the cost, so summoning sickness (CR 302.6) never
// applies — that rule is keyed off the {T} cost component
// (AbilityCost.Tap), which this ability has none of. A freshly cast
// Beledros can pay 10 life and untap lands the same turn it enters.
func init() {
	Register(Spec{
		OracleID:     "90194ff1-db61-463f-b5a3-15cd85311d0e",
		Name:         "Beledros Witherbloom",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Pest tokens don't gain you 1 life when they die.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Beledros Witherbloom — create a Pest", Do(CreateToken{Template: TokenCard("1/1 black and green Pest"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:     "Beledros Witherbloom — untap all lands you control",
			Cost:      PayLife(10),
			Condition: OncePerTurnActivation("Beledros Witherbloom — untap all lands you control"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return untapAllLandsControlledBy(g, item.Controller, NewContext(g, item))
			},
		}},
	})
}
