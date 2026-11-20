package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// frf_b1_helpers.go — the shape Fate Reforged batch 1 (#5552) prints more
// than once: "as long as you control a <colour> or <colour> permanent", on
// the set's two-colour creatures (Abzan Kin-Guard, Battle Brawler) and the
// five Runemarks. Its own file, as #231 asks of a batch, so a concurrent
// batch never collides here.
//
// Append-only: add a builder, never change what one means.

// frfYouControlEither reports whether `source`'s controller controls a
// permanent of either colour, read after the layers (the source itself
// counts if it is that colour). Each colour is a word of the source's text
// (CR 612.2), so a text-changing effect moves it, as conB1WhileYouControlA
// reads its one colour. `a` and `b` are mana letters ("R", "W").
func frfYouControlEither(a, b string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(_ *game.Card, g *game.Game, source *game.Card) bool {
		return b18ControlsPermanentOfColor(g, source.Controller, source.TextColor(a)) ||
			b18ControlsPermanentOfColor(g, source.Controller, source.TextColor(b))
	}
}

// frfSelfWhileYouControlEither is the AppliesTo of "This creature <has/gets
// …> as long as you control a <colour> or <colour> permanent": the source
// alone, while it is a creature and frfYouControlEither holds.
func frfSelfWhileYouControlEither(a, b string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	either := frfYouControlEither(a, b)
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && either(target, g, source)
	}
}

// frfRunemark is a Runemark's two statics: "Enchanted creature gets +2/+2.
// Enchanted creature has <keyword> as long as you control a <a> or <b>
// permanent." The pump is layer 7c, the keyword layer 6, and the condition
// is re-read on every recompute, so the keyword comes and goes with the
// coloured permanents while the pump stays.
func frfRunemark(keyword, a, b string) []game.StaticAbility {
	return []game.StaticAbility{
		PumpAttached(2, 2),
		GrantToAttachedWhile(frfYouControlEither(a, b), keyword),
	}
}
