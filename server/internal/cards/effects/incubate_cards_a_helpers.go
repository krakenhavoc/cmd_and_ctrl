package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shared bodies for the incubate cards (#3568, ADR 0185), so two cards
// that read the same thing do not carry the same closure twice.

// incubateThatSpellsManaValue is "incubate X, where X is that spell's
// mana value" on a cast trigger (Chrome Host Seedshark, Brimaz, Blight
// of Oreskos): the spell wherever it is now, with its announced X while
// it is still on the stack (triggeringSpellManaValue).
func incubateThatSpellsManaValue(g *game.Game, item *game.StackItem) error {
	return Incubate{N: triggeringSpellManaValue(g, item)}.Apply(NewContext(g, item))
}

// incubateItsPower is "when this creature dies, incubate X, where X is
// its power" (Bloated Processor, Furnace Gremlin). Power is the
// creature's last-known power as it died (CR 603.10a), read through
// SourcePermanent, which serves the stored record once the creature has
// left: counters and pumps count, and it is not clamped.
func incubateItsPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if info, ok := ctx.SourcePermanent(); ok {
		n = info.Power
	}
	return Incubate{N: n}.Apply(ctx)
}
