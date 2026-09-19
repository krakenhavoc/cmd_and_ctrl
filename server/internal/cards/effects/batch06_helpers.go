package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch06_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 06 (#299, `edhrec_rank` 694–797). Own file, every
// package-level name prefixed b06, per the convention the parallel
// batches settled on after the batch 02 collision.

// b06EnteredThisTurn reports whether the permanent entered the
// battlefield during the current turn — Oran-Rief's "each green
// creature that entered this turn".
//
// The per-turn tally's per-object cell (#1009). It used to walk the
// event log back to the most recent EventBeginUpkeep on the argument
// that "nothing can enter during the untap step", which is not true:
// an untap-step trigger or an untap-step choice can put a permanent
// onto the battlefield (#70, ADR 0070), and the walk stopped short of
// it and answered "no" for a permanent that really did enter this
// turn. The tally is reset as the turn BEGINS, before the untap step,
// so it answers from the real boundary.
//
// Not summoning sickness: a creature that entered during an
// opponent's turn still carries SummonedThisTurn on yours (CR 302.6)
// and did not enter this turn.
func b06EnteredThisTurn(g *game.Game, cardID uuid.UUID) bool {
	return g.EnteredThisTurn(cardID)
}

// b06AnOpponentLostAtLeastThisTurn reports whether some single
// opponent of `controller` has lost at least `n` life this turn —
// Bloodchief Ascension's end-step condition. PlayerTurnTally.LifeLost
// already distinguishes the two ways a life total goes down (a
// negative EventChangeLife, and an EventDealDamage to a player, which
// emits no EventChangeLife of its own), so this is one cell per
// opponent.
func b06AnOpponentLostAtLeastThisTurn(g *game.Game, controller uuid.UUID, n int) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		if g.TurnTallyFor(p.ID).LifeLost >= n {
			return true
		}
	}
	return false
}

// b06ExileListedCards is the delayed-trigger body Whip of Erebos
// schedules: exile every card the item carries that is still on the
// battlefield. Package-level so the delayed trigger captures nothing.
func b06ExileListedCards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b06EntersTappedUnlessLandType is the Castle condition — "unless you
// control a Plains" — one land type where the checklands name two.
// youControlLandTyped with the same lowercase needle twice is exactly
// that.
func b06EntersTappedUnlessLandType(subtype string) game.ReplacementEffect {
	return SelfEntersTappedUnless(youControlLandTyped(subtype, subtype))
}

// b06SelfETB is the AppliesTo every "when this permanent enters"
// trigger in this batch shares.
func b06SelfETB(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID
}

// isLegendary reads the post-layer supertypes.
func isLegendary(c *game.Card) bool {
	for _, s := range c.Effective().Supertypes {
		if s == "Legendary" {
			return true
		}
	}
	return false
}

// b06TutorToHand is the shared body of the "search your library for a
// <kind> card, reveal it, put it into your hand, then shuffle" spells:
// Idyllic Tutor (enchantment), Eladamri's Call (creature). Grim Tutor
// is not one — it neither reveals nor filters, and it taxes life after
// the search.
func b06TutorToHand(reason string, pred func(game.Card) bool) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return SearchLibrary{
			Player:    ctx.Controller(),
			Predicate: pred,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Reason:    reason,
		}.Apply(ctx)
	}
}
