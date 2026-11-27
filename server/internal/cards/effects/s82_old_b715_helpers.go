package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b715_helpers.go — shared shapes of S82 OLD batch 715 (Urza's
// Legacy, #5623).

// RadiantOtherFlyers counts the creatures with flying on the battlefield other
// than `source` itself, controlled by anyone, after the layers.
func RadiantOtherFlyers(g *game.Game, source *game.Card) int {
	n := 0
	for _, seat := range ApnapPlayers(g) {
		for _, id := range PermanentsControlledByMatching(g, seat, And(Creature(), HasKeyword("flying"))) {
			if id != source.InstanceID {
				n++
			}
		}
	}
	return n
}

// ThisBecomesBlocked is "whenever this creature becomes blocked": the
// attacker the event names is the source. Once per attacker however many
// creatures block it (CR 509.1h).
func ThisBecomesBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
}

// ThisGetsForEachBlocker is "it gets +2/+2 until end of turn for each creature
// blocking it": the blockers are counted as the trigger resolves, so a blocker
// removed from combat in response no longer counts (BlockersOf).
func ThisGetsForEachBlocker(power, toughness int, label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		n := len(BlockersOf(g, item.SourceCardID))
		if n <= 0 {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: power * n, Toughness: toughness * n, Label: label}.Apply(NewContext(g, item))
	}
}
