package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// con_b1_helpers.go — the shapes Conflux batch 1 (#5587) prints more
// than once: the Shards' "as long as you control a <colour> permanent"
// static and the "<cost>: This creature gains <keyword> until end of
// turn." row. Its own file, as #231 asks of a batch, so a concurrent
// batch never collides here.
//
// Append-only: add a builder, never change what one means.

// conB1WhileYouControlA is the AppliesTo of "This creature <has/gets …>
// as long as you control a <colour> permanent" (Toxic Iguanar, Ember
// Weaver, Cliffrunner Behemoth): the source alone, while its controller
// controls any permanent of that colour, read after the layers (the
// source itself counts if something has made it that colour). The
// colour is a word of the source's text (CR 612.2), so a text-changing
// effect moves it. `letter` is the colour's mana letter ("R", "W").
func ConB1WhileYouControlA(letter string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return SelfCreature(target, g, source) && B18ControlsPermanentOfColor(g, source.Controller, source.TextColor(letter))
	}
}

// conB1DomainBoostTheTarget is the OnResolve of "Domain — Target
// creature gets +1/+1 (sign 1) or -1/-1 (sign -1) until end of turn for
// each basic land type among lands you control" (Might of Alara, Drag
// Down). The count is read once, as the spell resolves (CR 608.2h), and
// the boost ends in the cleanup step (CR 514.2). A target that is no
// longer legal gets nothing (CR 608.2b).
func ConB1DomainBoostTheTarget(sign int, label string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		n := sign * DomainCount(ctx.Game, item.Controller)
		return BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: label}.Apply(ctx)
	}
}

// conB1ThisGainsRow is "<cost>: This creature gains <keyword> until end
// of turn." (Pestilent Kathari's first strike, Thornling's haste,
// trample and indestructible): a layer-6 grant pinned to the ability's
// source that ends in the cleanup step (CR 514.2). With no {T} in the
// cost it can be activated any number of times, the turn the creature
// arrives included.
func ConB1ThisGainsRow(cost, keyword, name string, answers game.Answers) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": This creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: answers, Pump: &game.Pump{Keywords: []string{keyword}}},
		Cost:    ManaCost(cost),
		Effect:  ThisGainsKeywordUntilEndOfTurn(keyword, name+" — "+keyword+" until end of turn"),
	}
}
