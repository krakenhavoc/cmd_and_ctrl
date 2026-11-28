package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b797_helpers.go — S82 OLD batch 797 (Mirage, #5633).

var (
	// BarrelingAttackBlockedCondition is "when that creature becomes
	// blocked this turn" (CR 509.3c): the EventBecomesBlocked of the one
	// object CondParams.Object names, id and CR 400.7 epoch, so a creature
	// that left and came back is a new object and never fires it
	// (CR 603.7c).
	BarrelingAttackBlockedCondition = game.DelayedEachCondition("blocked/s82-old-b797-that-creature-becomes-blocked", barrelingAttackBlocked)

	// BarrelingAttackBoostBody is "it gets +1/+1 until end of turn for each
	// creature blocking it", counted as the trigger resolves.
	BarrelingAttackBoostBody = game.SimpleDelayedBody("s82-old-b797/barreling-attack-boost-per-blocker", barrelingAttackBoost)
)

func barrelingAttackBlocked(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBecomesBlocked || p.Object.ID == uuid.Nil || ev.CardID != p.Object.ID {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ev.CardID)
	if !ok || ref.Epoch != p.Object.Epoch {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

func barrelingAttackBoost(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := ctx.PayloadCards()
	if len(cards) == 0 || !OnBattlefield(g, cards[0]) {
		return nil
	}
	n := len(BlockersOf(g, cards[0]))
	if n <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: cards[0], Power: n, Toughness: n, Label: "Barreling Attack — +1/+1 for each creature blocking it"}.Apply(ctx)
}
