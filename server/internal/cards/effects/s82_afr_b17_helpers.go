package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_afr_b17_helpers.go — the Adventures in the Forgotten Realms creature
// lands (Den of the Bugbear, Hive of the Eye Tyrant, Lair of the Hydra,
// Cave of the Frost Dragon, Hall of Storm Giants), S82 batch 17.

// afrB17EntersTappedWithTwoOtherLands is "If you control two or more other
// lands, this land enters tapped.": the slowland condition read the other
// way round. EntersTappedUnless taps the land when its condition is false,
// so the condition passed is "fewer than two other lands".
func afrB17EntersTappedWithTwoOtherLands() game.ReplacementEffect {
	return EntersTappedUnless(func(g *game.Game, src *game.Card) bool {
		return countOtherLandsControlled(g, src) < 2
	})
}

// afrB17Manland is the creature-land animation: ONE continuous effect from
// the resolving ability (CR 611.2a, until end of turn), pinned to this
// object (CR 611.2c, CR 400.7), applied in every layer it touches at one
// timestamp (CR 613.6): the creature type and subtype (layer 4), its colour
// (layer 5), keywords and any granted ability bundle (layer 6) and the base
// power and toughness (layer 7b). "It's still a land" (CR 205.1b) is the
// types being ADDED to. A source that left and came back is a new object
// and is not animated (CR 400.7).
//
// grantKey, when set, names an AbilityGrant on the card (a quoted ability
// the land has while it is a creature: ADR 0093).
func afrB17Manland(g *game.Game, item *game.StackItem, name, subtype, color string, power, toughness int, keywords []string, grantKey string) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtypes(subtype)...),
		game.SetColorsMod(ctx.TextColors(color)...),
	}
	mods = append(mods, game.SetBasePTMods(power, toughness)...)
	if len(keywords) > 0 {
		mods = append(mods, game.AddKeywordsMod(keywords...))
	}
	label := name + " — becomes a creature until end of turn"
	if grantKey != "" {
		return GrantAbilitiesFor{
			Target: item.SourceCardID,
			Keys:   []string{grantKey},
			Also:   mods,
			Label:  label,
		}.Apply(ctx)
	}
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     mods,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    label,
	}.Apply(ctx)
}

// afrB17TapForMana is the land's "{T}: Add {X}".
func afrB17TapForMana(symbol string) []ManaAbility {
	return []ManaAbility{{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + symbol + "}",
		Label:    "Add {" + symbol + "}",
	}}
}
