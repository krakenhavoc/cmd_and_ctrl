package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b680_helpers.go — shared bodies for Nemesis batch 680 (#5618).

// OldB680YouControlAForest is "if you control a Forest" (Skyshroud Cutter's
// gate): a permanent you control with the Forest subtype, a land made a
// Forest included (CR 205.3i).
func OldB680YouControlAForest(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtypeWord("Forest") {
			return true
		}
	}
	return false
}

// OldB680MostCreatures is the player who controls more creatures than each
// other player still in the game (Wild Mammoth), or uuid.Nil when no player
// does: a tie for the most, or no creatures at all, names nobody.
func OldB680MostCreatures(g *game.Game) uuid.UUID {
	counts := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			counts[c.Controller]++
		}
	}
	best, bestN, tied := uuid.Nil, 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		switch n := counts[p.ID]; {
		case n > bestN:
			best, bestN, tied = p.ID, n, false
		case n == bestN:
			tied = true
		}
	}
	if tied || bestN == 0 {
		return uuid.Nil
	}
	return best
}

// OldB680HasFading is "permanent with fading" (Parallax Inhibitor): the
// permanent's catalog row carries the fading upkeep trigger (CR 702.32a).
func OldB680HasFading(c game.Card) bool {
	spec, ok := Lookup(c.OracleID)
	if !ok {
		return false
	}
	for _, t := range spec.Triggered {
		if t.Keyword == KeywordFading {
			return true
		}
	}
	return false
}

// OldB680FadeCounterOnEachFadingPermanentYouControl is "Put a fade counter on
// each permanent with fading you control".
func OldB680FadeCounterOnEachFadingPermanentYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && OldB680HasFading(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterFade, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
