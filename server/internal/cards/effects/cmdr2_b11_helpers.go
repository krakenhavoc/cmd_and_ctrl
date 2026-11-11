package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b11_helpers.go — helpers for the wave-2 Commander batch 11
// (Faeries, #5368). Append-only; every name starts cmdr2B11.

// cmdr2B11FlashByTappingKey is the optional cost's identity on the card.
const cmdr2B11FlashByTappingKey = "flash-by-tapping"

// cmdr2B11FlashByTapping is "You may cast this spell as though it had
// flash by tapping <n> untapped creatures you control with <keyword> in
// addition to paying its other costs" (Tegwyll's Scouring). CR 601.3c: an
// effect that allows the flash only if an additional cost is paid lets
// the caster begin the cast at instant speed; announcing this optional
// cost opens the window for that cast alone, and the taps ride
// cast_spell's teamwork_ids like conspire's. Tapping is not the {T}
// symbol, so a creature that arrived this turn may pay (CR 302.6).
func cmdr2B11FlashByTapping(n int, label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:      true,
		Key:           cmdr2B11FlashByTappingKey,
		TapOthers:     tapOthersPrice(n, label, preds...),
		Label:         "Tap " + label + " to cast it as though it had flash",
		AsThoughFlash: true,
	}
}

// cmdr2B11BecomeRedDragon is Shadow Puppeteers' "it becomes a red Dragon
// with base power and toughness 4/4 in addition to its other colors and
// types until end of turn" for the creature that attacked. The set is
// that one object (CR 611.2c); an attacker that has left, or come back as
// a new object, is not changed (CR 400.7).
func cmdr2B11BecomeRedDragon(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left || item.Trigger == nil {
		return nil
	}
	return ScopedEffectFor{
		Target: item.Trigger.Event.CardID,
		Mods: []game.Mod{
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
			game.AddColorsMod(ctx.TextColors("R")...),
			game.AddSubtypesMod(ctx.TextSubtypes("Dragon")...),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Shadow Puppeteers — red Dragon with base power and toughness 4/4 until end of turn",
	}.Apply(ctx)
}

// cmdr2B11StolenByTheFaeTokens is "You create X 1/1 blue Faerie creature
// tokens with flying" once the bounce has settled. An on-disk key.
var cmdr2B11StolenByTheFaeTokens = CardTail("cmdr2-b11/stolen-by-the-fae-tokens", func(ctx *Context, _ game.TailOutcome) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 blue Faerie with flying"), N: x}.Apply(ctx)
})
