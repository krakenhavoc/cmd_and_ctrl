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
// The engine's mana ability has a ProducedFunc — evaluated to decide
// what lands in the pool — and a Rider that runs immediately after
// (CR 605.3b, all inside one atomic mana-ability resolution). So the
// counter is placed by the Rider and the amount is computed as
// "charge counters + 1", which is the count the printed ability sees
// AFTER its own first instruction. Reading the board first and then
// adding one is the same arithmetic as placing the counter and then
// reading the board, and the two orders are indistinguishable from
// outside the resolution because a mana ability does not use the
// stack and nothing can respond in between.
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
			ProducedFunc: ProducedOneColor(b39AutogeneratorOutput),
			Rider:        b39AutogeneratorCharge,
			Label:        "Put a charge counter on this artifact. Add X mana of any one color, where X is the number of charge counters on it",
		}},
	})
}

// b39AutogeneratorOutput is X: the charge counters already on the
// Autogenerator plus the one this activation is about to add.
func b39AutogeneratorOutput(g *game.Game, _, source uuid.UUID) int {
	return b39ChargeCountersOn(g, source) + 1
}

// b39AutogeneratorCharge is the "put a charge counter on this
// artifact" half, run as the ability's rider so the count above and
// the counter on the card agree once the activation is over.
func b39AutogeneratorCharge(g *game.Game, _, source uuid.UUID) error {
	return g.AddCounterForEffect(source, game.CounterCharge, 1)
}
