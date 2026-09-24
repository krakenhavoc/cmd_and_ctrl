package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Protection — Instant {2}{W}:
//
//	"Until your next turn, your life total can't change and you gain
//	 protection from everything. All permanents you control phase out.
//	 Exile Teferi's Protection."
//
// The white "I am not here this turn cycle" button, and three clauses
// wide. Each of the three is its own engine seam, and the card is
// complete as of #1200 — the split below is kept because it is the
// clearest account of what each clause actually does.
//
// WHAT SHIPPED FIRST. "You gain protection from everything" until your next
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
// "ALL PERMANENTS YOU CONTROL PHASE OUT" SHIPS TOO, since #1199
// (CR 702.26, ADR 0084). It is the larger half of the card and the
// half it is famous for, and it is one call: every permanent this
// player controls leaves the battlefield SLICE — simultaneously,
// dragging every Aura and Equipment with it (CR 702.26g) — and comes
// back at the start of their next untap step (CR 502.1) with its
// counters, its damage, its tapped state and its attachments
// untouched, because phasing is not a zone change (CR 702.26d).
//
// The board is therefore not attackable, not targetable and not
// wrathable for a whole turn cycle, which is what the card is for.
// Nothing here has to arrange any of it: "treated as though it does
// not exist" is true of every battlefield walk in the engine because
// a phased-out permanent is not in the slice they walk.
//
// The PERMANENTS are taken by CONTROLLER, not by owner, and the list
// is snapshotted before any of it moves — the printed clause is
// "permanents you control" and the rule is simultaneous.
//
// "YOUR LIFE TOTAL CAN'T CHANGE" SHIPS TOO, since #1200 (CR 119.7 /
// CR 119.8, ADR 0085) — and with it the card's last caveat. It is one
// call, for the reason ADR 0084's closing section predicted: the lock
// is a player-scoped statement with a CR 611.2 duration, so it goes on
// Player.Statics beside the protection this same OnResolve grants a
// few lines below, and its consumer is the one CR 614 life window
// #482 built. Nothing else about the card had to learn it.
//
// What that adds on top of the protection is the half the protection
// never covered: life LOSS that is not damage — a drain, an "each
// opponent loses 3 life", a Phyrexian cost somebody makes you pay —
// and, the other way, life GAIN you would rather not have had (your
// own Ad Nauseam, a Sylvan Library cast under the shield). Damage was
// already prevented at the source by CR 702.16e, and now would not
// move the total even if it got through.
//
// THREE PRINTED CLAUSES, THREE SHIPPED, NO SIMPLIFICATIONS. The card
// took three sprints and three ADRs to finish: #1197 / ADR 0072's
// amendment (the protection), #1199 / ADR 0084 (the phasing),
// #1200 / ADR 0085 (the life-total lock).
func init() {
	Register(Spec{
		OracleID:     "0d4ecdb1-ec90-497f-a7a4-1c68092b8757",
		Name:         "Teferi's Protection",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			// "Your life total can't change" (CR 119.7, CR 119.8).
			// First for no reason but the printed order: the two
			// grants are independent entries on one slice and neither
			// reads the other.
			if err := (LockLifeTotal{
				Player:   me,
				Label:    "Teferi's Protection — your life total can't change",
				Duration: DurationUntilYourNextTurn(ctx, me),
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (GainPlayerKeyword{
				Player:   me,
				Keyword:  ProtectionFromEverything,
				Label:    "Teferi's Protection — protection from everything",
				Duration: DurationUntilYourNextTurn(ctx, me),
			}).Apply(ctx); err != nil {
				return err
			}
			// "All permanents you control phase out" (CR 702.26).
			// The set is snapshotted before any of it moves: the
			// phase-out is simultaneous, and a permanent attached to
			// another permanent in the same list must not be dragged
			// out twice (CR 702.26h).
			var mine []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == me {
					mine = append(mine, c.InstanceID)
				}
			}
			if err := (PhaseOut{Targets: mine}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Teferi's Protection." The spell moves ITSELF, so
			// the resolution frame's CR 608.2n graveyard route is
			// skipped by spellMovedItselfLocked (#489) — which is
			// exactly the case that check was written for.
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
