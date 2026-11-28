package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b844_helpers.go — Ice Age instants and sorceries (S82 OLD batch 844, #5636).

// S82OldB844DeathWard is "Regenerate target creature." as a spell's resolution.
func S82OldB844DeathWard(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return Regenerate{Target: id}.Apply(ctx)
}

// S82OldB844WordOfUndoing is "Return target creature and all white Auras you own attached to it to their
// owners' hands." The creature and the Auras go back together as one return (CR 400.7): a white Aura is read
// through the layers and must be owned by the caster, whoever controls it.
func S82OldB844WordOfUndoing(_ *game.StackItem, ctx *Context) error {
	t, ok := FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	ids := []uuid.UUID{t.ID}
	white := ctx.TextColor("W")
	for _, a := range ctx.Game.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(t.ID) && a.IsAura() && a.Owner == ctx.Controller() && a.HasColor(white) {
			ids = append(ids, a.InstanceID)
		}
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(ids, game.KeyedTail{})
}

// S82OldB844MindWarp is "Look at target player's hand and choose X cards from it. That player discards those
// cards." With X = 0 the hand is only looked at; with a hand smaller than X, all of it is chosen.
func S82OldB844MindWarp(_ *game.StackItem, ctx *Context) error {
	victim := TargetedPlayer(ctx)
	if victim == uuid.Nil {
		return nil
	}
	if ctx.X() <= 0 {
		ctx.Game.LookAtHandForEffect(ctx.Controller(), victim)
		return nil
	}
	return ChooseFromRevealedHand{Player: victim, Label: "card", Look: true, Count: ctx.X()}.Apply(ctx)
}

// S82OldB844Trailblazer is "Target creature can't be blocked this turn."
func S82OldB844Trailblazer(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Trailblazer — can't be blocked"}.Apply(ctx)
}

// S82OldB844AvalancheTargets is "X target snow lands": exactly X, announced with the spell (CR 601.2c).
func S82OldB844AvalancheTargets() *game.TargetSpec {
	spec := TargetPermanent("X target snow lands", And(Land(), Snow()))
	spec.CountFromX = true
	return spec
}

// S82OldB844StuntedGrowth is "Target player chooses three cards from their hand and puts them on top of their
// library in any order." A player holding fewer than three puts back what they have.
func S82OldB844StuntedGrowth(_ *game.StackItem, ctx *Context) error {
	who := TargetedPlayer(ctx)
	if who == uuid.Nil {
		return nil
	}
	return PutFromHandOnTopInAnyOrder{Player: who, N: 3, Label: "Stunted Growth — choose three cards from your hand"}.Apply(ctx)
}
