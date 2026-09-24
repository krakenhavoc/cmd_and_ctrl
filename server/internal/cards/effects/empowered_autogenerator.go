package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Empowered Autogenerator — Artifact {4} (EDHREC rank 4061):
//
//	"This artifact enters tapped.
//	 {T}: Put a charge counter on this artifact. Add X mana of any
//	 one color, where X is the number of charge counters on this
//	 artifact."
//
// A mana rock that starts behind and ends absurd: four mana, enters
// tapped, then adds one, two, three, four … and never stops. It is
// slow enough to be fair in a two-player game and completely unfair
// in Commander, where the turns exist for it to pay off, and it is a
// natural home for every proliferate effect in the deck.
//
// THE COUNTER GOES ON FIRST, AND X COUNTS IT. The printed ability is
// one instruction followed by another in the same resolution, so a
// freshly-cast Autogenerator taps for ONE the first time, not zero.
// That ordering is the whole card and it is the one thing this file
// has to get right.
//
// #1370: it used to be gotten wrong by a guess instead of a read. The
// engine's mana ability computes its ProducedFunc BEFORE running a
// Rider (CR 605.3b, all inside one atomic mana-ability resolution) —
// right for a card whose printed sentence puts the mana clause first
// (a painland's "Add {C}. This land deals 1 damage to you"), backwards
// for this one, whose counter-placement clause comes FIRST. Putting
// the placement in Rider and computing X as "charge counters + 1" is
// only the same arithmetic as "place, then read" when nothing else is
// touching the counter — the moment a real doubler (Doubling Season)
// is on the board, the Rider places 2 (or more) while the guess still
// says +1, and X comes out short.
//
// PreRider is the fix: it runs BEFORE ProducedFunc, in printed order,
// so X reads the count the placement actually landed on rather than
// predicting it. The placement itself goes through
// AddCounterMustSettleNowForEffect, not the ordinary
// AddCounterForEffect / AddCounterThenForEffect: CR 605.3a's mana
// ability resolution has no priority window inside it, so a CR 616
// ordering prompt (Doubling Season next to a Hardened Scales — which
// does not apply here; see below) cannot pause here even though it
// could for an ordinary resolving spell's counter placement. The
// gathered order settles it instead, exactly as
// RepEventProduceMana already does for the mana itself.
//
// Hardened Scales never enters into it: it only replaces PLACEMENT OF
// A +1/+1 COUNTER ON A CREATURE, and this artifact places a CHARGE
// counter on itself, an artifact. Doubling Season doubles ANY counter
// kind on ANY permanent its controller controls, so it is the one
// doubler that matters here.
//
// "ANY ONE COLOR" IS ONE PICK FOR ALL X, not X independent picks —
// ProducedOneColor's shape. It does NOT narrow to the commander's
// colour identity: the printed text has no such clause, so the full
// five-colour choice is offered and the commander's colours are
// merely listed first.
//
// The tapped entry is the real CR 614 self-replacement, so the
// Autogenerator cannot be cast and tapped on the same turn without an
// untapper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6832f9e2-294a-44a2-8af9-eb16ccfdfc36",
		Name:         "Empowered Autogenerator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			PreRider:     b39AutogeneratorCharge,
			ProducedFunc: ProducedOneColor(b39AutogeneratorOutput),
			Label:        "Put a charge counter on this artifact. Add X mana of any one color, where X is the number of charge counters on it",
		}},
	})
}

// b39AutogeneratorOutput is X: the charge counters on the
// Autogenerator, read AFTER b39AutogeneratorCharge's placement has
// landed (it runs first, as the ability's PreRider) — a fact, not a
// guess at what the placement is about to do.
func b39AutogeneratorOutput(g *game.Game, _, source uuid.UUID) int {
	return b39ChargeCountersOn(g, source)
}

// b39AutogeneratorCharge is the "put a charge counter on this
// artifact" half, run as the ability's PreRider so it lands BEFORE
// b39AutogeneratorOutput reads the count — the printed order, and the
// only order that reports what a counter doubler actually placed.
// mustSettleNow because CR 605.3a's mana-ability resolution cannot
// pause for the CR 616 ordering prompt a second doubler would raise.
func b39AutogeneratorCharge(g *game.Game, _, source uuid.UUID) error {
	_, err := g.AddCounterMustSettleNowForEffect(source, game.CounterCharge, 1)
	return err
}
