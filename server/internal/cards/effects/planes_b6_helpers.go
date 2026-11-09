package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b6_helpers.go holds the shapes planes batch 6 (ADR 0216 slice 4,
// #4870) share. Append-only.

// b6ToughnessDamagePair is "Each of those creatures deals damage equal to
// its toughness to the other" over clause slots 0 and 1 of the resolving
// ability (The Great Aerie): both amounts are read before either deals, as
// the damage is simultaneous (CR 120.3), and a missing or illegal creature
// leaves the other dealing nothing to it and receiving nothing.
func b6ToughnessDamagePair(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids [2]uuid.UUID
	var amt [2]int
	var ok [2]bool
	for i := range ids {
		t, legal := ctx.ClauseTarget(i)
		if !legal || t.Kind != game.TargetCard {
			continue
		}
		c, found := g.LookupCardForEffect(t.ID)
		if !found {
			continue
		}
		ids[i], amt[i], ok[i] = t.ID, g.ToughnessAnywhereForEffect(c), true
	}
	for i := range ids {
		j := 1 - i
		if ok[i] && ok[j] {
			if err := (DealDamage{Source: ids[i], Target: ids[j], Amount: amt[i]}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// b6TreasuresPingEverything is The Western Cloud's chaos ability: create
// three tapped Treasure tokens, and each of them deals 1 damage to each
// creature and each planeswalker. The permanents are listed once, as the
// tokens enter (the Treasures are artifacts, so not among them), and each
// Treasure is the source of its own damage.
func b6TreasuresPingEverything(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	tokens, err := g.CreateTokensForEffect(item.Controller, ctx.TextToken(TreasureToken()), 3, game.TokenEntryOptions{Tapped: true})
	if err != nil {
		return err
	}
	var victims []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() || c.IsPlaneswalker() {
			victims = append(victims, c.InstanceID)
		}
	}
	for _, tok := range tokens {
		for _, v := range victims {
			if err := (DealDamage{Source: tok, Target: v, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// b6PlaneswalkedToThisPlaneThisTurn is the intervening "if you planeswalked to
// <this plane> this turn" (Unyaro): the planeswalk to the source plane is in
// the turn's events. It declares no footprint (the planeswalk event is not a
// state the independence check models).
type b6PlaneswalkedToThisPlaneThisTurn struct{}

func (b6PlaneswalkedToThisPlaneThisTurn) Holds(ctx *Context) bool {
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind == game.EventPlaneswalkTo && ev.CardID == ctx.Source() {
			return true
		}
	}
	return false
}

func (b6PlaneswalkedToThisPlaneThisTurn) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// b6UntapThenPhaseOutAllCreaturesUntilPlaneswalk is Unyaro's end-step
// ability: untap all creatures, which then phase out until a player
// planeswalks (the plane that made the effect must still be face up). The
// phase-out takes each creature's attachments with it (CR 702.26g).
func b6UntapThenPhaseOutAllCreaturesUntilPlaneswalk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if err := untapEach(ctx, ids); err != nil {
		return err
	}
	return g.PhaseOutUntilPlaneswalkForEffect(item.SourceCardID, item.SourceCardID, ids...)
}
