package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_caps.go — the two clauses that make CR 502.3's "the active
// player determines which permanents they control will untap" a real
// decision (#826, ADR 0070). Its neighbours are untap_step.go (the
// permission that widens the set) and untap_restrictions.go (the
// restriction that narrows it); this file is the ceiling and the
// opt-out, which narrow it *with a choice*.
//
// Both slots are read once per untap step, after layers, through the
// catalog's ability key — so an orb that has lost its abilities caps
// nothing. Neither is a trigger and neither is a static: like the
// permission, they change what a turn-based action does, which has no
// layer to sit in and nothing to put on the stack.

// cantUntapMoreThan builds "players can't untap more than N <kind>
// during their untap steps" — Winter Moon, Damping Field.
//
// `match` nil counts every permanent, which is Static Orb's "more than
// two permanents". The cap is scoped to the active player's own untap
// step by the engine (ADR 0070 Decision 4), so the predicate never has
// to ask whose step it is.
func cantUntapMoreThan(label string, n int, match CardPredicate) game.UntapCap {
	out := game.UntapCap{Label: label, Max: n}
	if match != nil {
		out.Counts = func(g *game.Game, _ *game.Card, target *game.Card) bool {
			return target != nil && match(g, target.Controller, *target)
		}
	}
	return out
}

// cantUntapMoreThanWhileSourceUntapped is the same clause with Winter
// Orb's and Static Orb's "as long as this artifact is untapped"
// condition on the front.
//
// The condition is continuous, so it is read at the instant the step
// asks — the same reason Quest for Renewal's quest counters are read
// in UntapStepPermission.AppliesTo rather than latched. Tapping the orb
// in response is not possible (the untap step grants no priority), but
// tapping it on the previous turn is exactly how the card is played.
func cantUntapMoreThanWhileSourceUntapped(label string, n int, match CardPredicate) game.UntapCap {
	out := cantUntapMoreThan(label, n, match)
	out.Applies = func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
		return source != nil && !source.Tapped
	}
	return out
}

// mayChooseNotToUntapSelf builds "you may choose not to untap this
// <permanent> during your untap step" — Rust Tick, Amber Prison, and
// the other 43 cards in the family, every one of which is self-only.
func mayChooseNotToUntapSelf(label string) game.UntapOptOut {
	return game.UntapOptOut{
		Label: label,
		Optional: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target != nil && source != nil && target.InstanceID == source.InstanceID
		},
	}
}
