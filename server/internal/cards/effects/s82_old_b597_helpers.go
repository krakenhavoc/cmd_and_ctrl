package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SourceDealsTwoToThatPlayer is the Effect body for "<this> deals 2 damage
// to that player" off a cast trigger (Mindsparker, Ishi-Ishi, Akki
// Crackshot): "that player" is the caster, read off the item's carried
// event, and a source that left in response still deals the damage, named
// as the departed object (DealDamage.SourceObject, CR 608.2h).
func SourceDealsTwoToThatPlayer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	caster := ctx.Trigger().Event.Actor
	ref, ok := ctx.SourceRef()
	if !ok || caster == uuid.Nil {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: caster, Amount: 2}.Apply(ctx)
}
