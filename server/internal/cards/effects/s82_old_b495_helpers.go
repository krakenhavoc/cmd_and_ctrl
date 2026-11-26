package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b495_helpers.go — shared pieces of S82 OLD batch 495 (Zendikar life
// gain and loss, removal, #5583).

// Old495AlliesYouControl counts the Allies `player` controls, read through the
// layered view and the source's text (CR 612.2).
func Old495AlliesYouControl(ctx *Context, player uuid.UUID) int {
	ally := ctx.TextSubtype("Ally")
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(ally) {
			n++
		}
	}
	return n
}

// Old495AnOpponentHasLifeAtMost reports whether some opponent of `you` still in
// the game has n or less life.
func Old495AnOpponentHasLifeAtMost(g *game.Game, you uuid.UUID, n int) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated {
			continue
		}
		if p.Life <= n {
			return true
		}
	}
	return false
}

// Old495AnOpponentGainedLifeThisTurn is Needlebite Trap's alternative-cost
// condition: some opponent of `you` gained life this turn.
func Old495AnOpponentGainedLifeThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated {
			continue
		}
		if B15LifeGainedThisTurn(g, p.ID) > 0 {
			return true
		}
	}
	return false
}

// Old495AnOpponentHadThreeCardsPutIntoTheirGraveyard is Ravenous Trap's
// alternative-cost condition: "an opponent had three or more cards put into
// their graveyard from anywhere this turn". It reads the turn's events, one
// arrival per card (GraveyardArrivals), so a card that has since left the
// graveyard still counts and a token never does (CR 108.2b). One opponent has
// to reach three by themselves.
func Old495AnOpponentHadThreeCardsPutIntoTheirGraveyard(g *game.Game, you uuid.UUID) bool {
	arrivals := GraveyardArrivals{}
	counts := map[uuid.UUID]int{}
	for _, ev := range g.EventsThisTurn() {
		for _, p := range g.Seats {
			if p == nil || p.ID == you || p.Eliminated {
				continue
			}
			if arrivals.arrival(ev, g, p.ID) {
				counts[p.ID]++
			}
		}
	}
	for _, n := range counts {
		if n >= 3 {
			return true
		}
	}
	return false
}

// Old495ExactlyOneCreatureIsAttacking is Pitfall Trap's alternative-cost
// condition: exactly one creature is attacking.
func Old495ExactlyOneCreatureIsAttacking(g *game.Game, _ uuid.UUID) bool {
	return len(AttackingCreatures(g)) == 1
}

// Old495DesecratedEarth is the body of "Destroy target land. Its controller
// discards a card.": the controller is the one as the spell resolves (read
// before the land is destroyed), and discards even if the land survives (an
// indestructible one) — the discard is not conditional on the destruction. A
// target that is no longer legal does nothing (CR 608.2b).
func Old495DesecratedEarth(item *game.StackItem, ctx *Context) error {
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	land, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	owner := land.Controller
	if err := DestroyTheTargetPermanent(item, ctx); err != nil {
		return err
	}
	return DiscardOne(ctx.Game, item, owner)
}

// Old495HagraDiabolistDrain is "you may have target player lose life equal to
// the number of Allies you control": the Allies are counted as the ability
// resolves. The "may" is the trigger's.
func Old495HagraDiabolistDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		n := Old495AlliesYouControl(ctx, item.Controller)
		if n <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
	}
	return nil
}
