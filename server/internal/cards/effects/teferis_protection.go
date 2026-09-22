package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Protection — Instant {2}{W}:
//
//	"Until your next turn, your life total can't change and you gain
//	 protection from everything. All permanents you control phase out.
//	 Exile Teferi's Protection."
//
// The white "I am not here this turn cycle" button, and three clauses
// wide. One of them is the seam this card is in the sprint for; two
// are not, and the split is worth stating precisely because the card
// is famous for the halves that are missing.
//
// WHAT SHIPS. "You gain protection from everything" until your next
// turn — CR 702.16i, the GRANTED half of #1197. It is a real
// protection, read at the same three choke points a permanent's is:
// nothing an opponent (or you) controls can target you, every source
// of damage is prevented (CR 702.16e, combat and noncombat alike),
// and an "enchant player" Curse on you falls off. The duration is the
// CR 611.2b one ADR 0063 Decision 3 built — it sits through all three
// opponents' turns and ends as yours begins, which is the whole
// reason the card is a Fog for a whole rotation rather than for a
// turn.
//
// "Exile Teferi's Protection" also ships, and needs no new machinery:
// the instruction is part of the spell's own text, so it runs inside
// OnResolve, and #489's spellMovedItselfLocked is the check that
// stops the resolution frame putting it in the graveyard afterwards.
// Leaving it out would have been a card that is STRONGER than
// printed — a Regrowth target the printed card never gives you —
// which is the direction #259 forbids.
//
// TWO SIMPLIFICATIONS, and both are named on the card's caveats:
//
//   - **PHASING (CR 702.26) IS NOT IMPLEMENTED.** "All permanents you
//     control phase out" is the larger half of the card and it is a
//     seam of its own: phasing is not a zone change (CR 702.25f), it
//     is a state a permanent is in, and this engine models no such
//     state anywhere (activation_tally.go has said so since S38). So
//     your board is still there, still attackable, still a legal
//     target for a Wrath. What the card DOES give you is the half
//     about you: you cannot be targeted or damaged. Strictly weaker
//     than printed.
//   - **"YOUR LIFE TOTAL CAN'T CHANGE"** is not implemented either.
//     It is a replacement effect with a duration longer than end of
//     turn, and TurnScopedReplacements carries no duration field at
//     all (ADR 0063 Decision 8 states that as deliberate — no card
//     needed one until this one). Most of what it stops is already
//     stopped by the protection: damage is prevented at the source.
//     What gets through is life LOSS that is not damage — a drain, an
//     "each opponent loses 3 life" — and, the other way, life GAIN
//     you would rather not have had. Weaker than printed on the half
//     that matters.
//
// Deferred to whichever sprint teaches the engine phasing; the life
// lock wants a replacement registry with a duration on it, which is
// the smaller of the two.
func init() {
	Register(Spec{
		OracleID:     "0d4ecdb1-ec90-497f-a7a4-1c68092b8757",
		Name:         "Teferi's Protection",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Phasing isn't implemented — \"all permanents you control phase out\" does nothing, so your board is still on the battlefield and can still be attacked, targeted and wrathed.",
			"\"Your life total can't change\" isn't implemented — damage is prevented by the protection, but life loss that isn't damage (a drain) still reaches you, and you can still gain life.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			if err := (GainPlayerKeyword{
				Player:   me,
				Keyword:  ProtectionFromEverything,
				Label:    "Teferi's Protection — protection from everything",
				Duration: DurationUntilYourNextTurn(ctx, me),
			}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Teferi's Protection." The spell moves ITSELF, so
			// the resolution frame's CR 608.2m graveyard route is
			// skipped by spellMovedItselfLocked (#489) — which is
			// exactly the case that check was written for.
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
