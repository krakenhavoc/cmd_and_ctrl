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
// The Pest's own "when this token dies, you gain 1 life" ships since
// ADR 0083 (#1248). It is a catalog template (`token:pest`, declared
// in tokens.go and shared with Sedgemoor Witch) and its trigger is
// found through `game.CatalogKey`'s token-key fallback like a printed
// card's, so it fires from the graveyard in the beat before CR 704.5d
// removes the token. It does NOT publish the Pest in the catalogue
// page: a token template is filed in `defs`, never in `registry`,
// which is what `All()` counts. Before that this card shipped as
// `caveats` with a 1/1 body and nothing more.
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
//
// The label is the printed line (#1276's oracle check reads its cost
// prefix), and it is a const because OncePerTurnActivation keys the
// per-object tally on the exact Label string.
const beledrosUntapLabel = "Pay 10 life: Untap all lands you control. Activate only once each turn."

func init() {
	Register(Spec{
		OracleID:        "90194ff1-db61-463f-b5a3-15cd85311d0e",
		Name:            "Beledros Witherbloom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Beledros Witherbloom — create a Pest", Do(CreateToken{Template: PestToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:     beledrosUntapLabel,
			Cost:      PayLife(10),
			Condition: OncePerTurnActivation(beledrosUntapLabel),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return untapAllLandsControlledBy(g, item.Controller, NewContext(g, item))
			},
		}},
	})
}
