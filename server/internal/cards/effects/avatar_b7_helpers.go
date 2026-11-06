package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b7_helpers.go — shared pieces of Avatar: The Last Airbender
// batch 7 (#4459). Append-only.

// avB7EnchantedCreatureOnBattlefield is "enchanted creature" in an
// Aura's activated ability as it resolves: the permanent the Aura is
// attached to, or was attached to when the Aura left (CR 608.2h), and
// only while that permanent is still the same object on the battlefield
// (CR 400.7). A creature that left and came back is a new object the
// Aura never enchanted.
func avB7EnchantedCreatureOnBattlefield(ctx *Context) (uuid.UUID, bool) {
	aura, ok := ctx.SourcePermanent()
	if !ok {
		return uuid.Nil, false
	}
	if aura.AttachedObject != nil {
		host := *aura.AttachedObject
		c, ok := ctx.Game.LookupCardForEffect(host.ID)
		if !ok || c.ObjectEpoch != host.Epoch || !onBattlefield(ctx.Game, host.ID) {
			return uuid.Nil, false
		}
		return host.ID, true
	}
	if aura.AttachedTo.Kind != game.TargetCard || !onBattlefield(ctx.Game, aura.AttachedTo.ID) {
		return uuid.Nil, false
	}
	return aura.AttachedTo.ID, true
}

// avB7ShuffleEnchantedIntoLibrary is "Enchanted creature's owner
// shuffles it into their library" (Watery Grasp): the creature goes to
// its owner's library, and that owner shuffles (Living Library's tuck).
func avB7ShuffleEnchantedIntoLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := avB7EnchantedCreatureOnBattlefield(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{host}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: c.Owner})
}

// avB7ExileTargetLinked is "Exile target card from a graveyard" whose
// card another ability of the same source reads back as "exiled with
// this creature" (CR 406.6, 607.2a): the exile stamps the link (Boiling
// Rock Rioter). A source that has left exiles with no link.
func avB7ExileTargetLinked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ExileTarget{Target: t.ID, Linked: true}.Apply(ctx)
		}
	}
	return nil
}
