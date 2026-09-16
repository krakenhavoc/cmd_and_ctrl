package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch05_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 05 (#298, `edhrec_rank` 578–693).
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files. Every name
// is b05-prefixed for the same reason — batches 03, 04, 06 and 07 are
// in flight beside this one.

// b05HasSubtype is the CardPredicate form of Card.HasSubtype, for
// "target Forest" (Arbor Elf) and any other subtype-gated clause. It
// reads the post-layer subtype list (#354), so an Urborg-made Swamp or
// a type-changed permanent qualifies exactly as it does in the rules.
func b05HasSubtype(subtype string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtype(subtype) }
}

// b05Legendary is "target legendary permanent" (Minamo, School at
// Water's Edge). Supertypes come from the effective characteristic so
// a layer effect that adds or removes Legendary composes.
func b05Legendary() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return isLegendary(&c) }
}

// b05ControlsSubtype reports whether `controller` controls a
// battlefield permanent with the given subtype — Ophiomancer's "if
// you control no Snakes".
func b05ControlsSubtype(g *game.Game, controller uuid.UUID, subtype string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// b05ControlsCreatureOrPlaneswalker is Plaguecrafter's "each player who
// can't" test: a player with no creature and no planeswalker cannot
// sacrifice one and discards instead.
func b05ControlsCreatureOrPlaneswalker(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && (c.IsCreature() || c.IsPlaneswalker()) {
			return true
		}
	}
	return false
}

// b05OpponentControlsMoreLands is Knight of the White Orchid's
// intervening-if: "if an opponent controls more lands than you".
// Counts post-layer lands, so an animated or type-changed permanent
// counts exactly as CR 305 would have it.
func b05OpponentControlsMoreLands(g *game.Game, controller uuid.UUID) bool {
	counts := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() {
			counts[c.Controller]++
		}
	}
	mine := counts[controller]
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if counts[p.ID] > mine {
			return true
		}
	}
	return false
}

// b05EachPlayerDraws is "each player draws N cards" (Scrawling
// Crawler's upkeep), APNAP from the active seat, eliminated seats
// skipped. Each draw is an ordinary draw, so "whenever an opponent
// draws" watchers — the Crawler's own second ability included — see
// every one of them.
func b05EachPlayerDraws(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < numSeats; i++ {
		p := g.Seats[(start+i)%numSeats]
		if p == nil || p.Eliminated {
			continue
		}
		if err := (DrawCards{Player: p.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
