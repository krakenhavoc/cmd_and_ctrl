package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// effects_this_until_end_of_turn.go — whole activated-ability bodies
// that change the ability's own source until end of turn: "This
// creature gets +N/+M until end of turn" and "This creature loses
// <keyword> until end of turn". First written for the any-player cards
// (ADR 0106 §1, #1793): the Flailing creatures, Fan Favorite, Oona's
// Prowler, Zerapa Minotaur, Ribbon Snake and Vintara Elephant.
//
// "This creature" is the ability's source (CR 113.7a), never "you": on
// an any-player row the activator may not control it, and the effect
// lands on the permanent all the same. Both bodies go through the
// until-end-of-turn primitives, which refuse a source that left and
// came back as a new object (CR 400.7, source_object_guard.go), and
// which do nothing to a source that is no longer on the battlefield.
//
// Append-only: add a builder, never change what one means.

// thisGetsUntilEndOfTurn is "This creature gets +power/+toughness until
// end of turn" (CR 611.2c, ending at CR 514.2). Negative values are the
// shrink ("gets -1/-1", "gets -2/-0").
func thisGetsUntilEndOfTurn(power, toughness int, label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return BoostUntilEOT{
			Target:    item.SourceCardID,
			Power:     power,
			Toughness: toughness,
			Label:     label,
		}.Apply(NewContext(g, item))
	}
}

// thisLosesKeywordUntilEndOfTurn is "This creature loses <keyword>
// until end of turn": a layer-6 ability-removing effect with its own
// timestamp (CR 613.1f), so a grant of the keyword with a LATER
// timestamp puts it back, and an earlier one (or the printed keyword)
// does not.
func thisLosesKeywordUntilEndOfTurn(keyword, label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label,
			game.RemoveKeywordsMod(keyword))
	}
}

// thisGainsKeywordUntilEndOfTurn is "This creature gains <keyword>
// until end of turn": a layer-6 grant pinned to the source (Endling's
// three {B} abilities). A cumulative keyword (undying, prowess) granted
// twice is two instances.
func thisGainsKeywordUntilEndOfTurn(keyword, label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label,
			game.AddKeywordsMod(keyword))
	}
}

// thisLosesOwnReplacementUntilEndOfTurn is "This creature loses '<one of
// its own replacement effects>' until end of turn": row `row` of the
// card's Spec.Replacements switched off by a layer-6 loseOwnAbility
// record with its own timestamp (CR 613.1f, #1859). A grant of the same
// text with a later timestamp is a different row and still applies.
func thisLosesOwnReplacementUntilEndOfTurn(row int, label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label,
			game.LoseOwnAbilityMod(game.AbilitySlotReplacement, row))
	}
}
