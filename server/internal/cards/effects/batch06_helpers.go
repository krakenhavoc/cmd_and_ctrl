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
// The engine keeps no "entered this turn" flag (SummonedThisTurn is
// summoning sickness, which a creature that entered during an
// opponent's turn still carries on yours), so this reads the event
// log: an EventETB for the card that is more recent than the most
// recent EventBeginUpkeep. Every turn passes through its upkeep, and
// nothing can enter during the untap step before it, so "since the
// last upkeep began" is "this turn". A game with no upkeep event yet
// treats every entry as this turn's, which is correct for that
// window.
func b06EnteredThisTurn(g *game.Game, cardID uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		switch ev.Kind {
		case game.EventBeginUpkeep:
			return false
		case game.EventETB:
			if ev.CardID == cardID {
				return true
			}
		}
	}
	return false
}

// b06AnOpponentLostAtLeastThisTurn reports whether some single
// opponent of `controller` has lost at least `n` life this turn —
// Bloodchief Ascension's end-step condition. Same event-log walk as
// b06EnteredThisTurn, summing per player with b04OpponentLostLife,
// which already distinguishes the two ways a life total goes down
// (a negative EventChangeLife, and an EventDealDamage to a player,
// which emits no EventChangeLife of its own).
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
