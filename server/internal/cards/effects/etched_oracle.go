package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etched Oracle — Artifact Creature — Wizard {4}, 0/0 (EDHREC rank
// 6423):
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for
//	 each color of mana spent to cast it.)
//	 {1}, Remove four +1/+1 counters from this creature: Target
//	 player draws three cards."
//
// Both halves of this PR on one card: sunburst reads the mana that
// paid (#761), and the draw ability spends the counters it made
// (#625's self form, now on a card that can actually make four).
//
// Sunburst (CR 702.44a) is a keyword since ADR 0109 §11 (#1552): the
// engine counts the instances on the resolving spell and seeds the
// counters onto the entry event off the resolving stack item, so the
// counters are part of the ENTRY. They are on the permanent before
// EventETB, a doubler applies, and a "whenever one or more counters are
// put on a permanent you control" payoff sees them. A second instance —
// Lux Artillery's "it gains sunburst" — counts the colours again
// (CR 702.44d).
//
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) no longer leaves this card guessing: it declares ReadsManaSpent,
// so the caster is asked what they paid outside the pool, or whether they
// cast it without paying its mana cost, and the record it reads is exact
// (#3855, ADR 0068 amendment 2026-10-15).
//
// Spec.WantsDistinctColors makes the cast gate spread the payment
// across colours, so an Oracle cast off four different lands really
// is a 4/4. Cast off four colourless mana it is a 0/0 and dies at
// once, which is exactly what the card does in paper.
//
// The draw ability is instant speed and has no {T}, so a freshly
// landed Oracle with four counters can cash them the turn it arrives.
func init() {
	Register(Spec{
		OracleID:            "7ecfa47e-1165-46a6-884c-290b1c14d020",
		Name:                "Etched Oracle",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label: "{1}, Remove four +1/+1 counters from this creature: Target player draws three cards.",
			Cost: Plus(
				ManaCost("{1}"),
				RemoveCountersFromThis(game.CounterPlusOne, 4),
			),
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 3}),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetPlayer {
						return DrawCards{Player: ref.ID, N: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
