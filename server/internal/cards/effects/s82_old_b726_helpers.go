package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza's Saga activated abilities, batch 726 (#5624): helpers the cards of the
// batch read, and the keys a card needs registered as an on-disk identity.

// S82OldB726TargetGainsKeywordUntilEOT is "Target creature gains <keyword>
// until end of turn" as an activated ability's body (Elvish Herder). A target
// that has left the battlefield gets nothing (CR 608.2b).
func S82OldB726TargetGainsKeywordUntilEOT(name, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return GrantKeywordUntilEOT{
			Target:   target,
			Keywords: []string{keyword},
			Label:    name + " — " + keyword + " until end of turn",
		}.Apply(ctx)
	}
}

// S82OldB726TapEachLegalTarget is "Tap X target lands" as a body (Mishra's
// Helix): each target still on the battlefield is tapped.
func S82OldB726TapEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !OnBattlefield(g, t.ID) {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB726FlipToBounceOrSacrifice is Viashino Sandswimmer's "Flip a coin. If
// you win the flip, return this creature to its owner's hand. If you lose the
// flip, sacrifice this creature." The flip is called by the controller.
func S82OldB726FlipToBounceOrSacrifice(g *game.Game, item *game.StackItem) error {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Viashino Sandswimmer — call the coin flip",
		ThenKey:  s82OldB726SandswimmerFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var s82OldB726SandswimmerFlipped = CoinFlipThen("coin-flip/viashino-sandswimmer", func(ctx *Context, r game.CoinFlipped) error {
	if SourceIsNewObject(ctx.Game, ctx.Item) || !OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	if WonFirstFlip(r) {
		return BounceToHand{Target: ctx.Source()}.Apply(ctx)
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})

// S82OldB726ReturnCreaturesThatDiedThisTurn is No Rest for the Wicked's body:
// return to your hand all creature cards in your graveyard that were put there
// from the battlefield this turn. The set is read once, before any moves.
func S82OldB726ReturnCreaturesThatDiedThisTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
		return c.IsCreature() && PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
	})
	for _, id := range ids {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB726EachPlayerMillsTwo is "Each player mills two cards" (Whetstone):
// every player in turn order, starting with the active player.
func S82OldB726EachPlayerMillsTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return MillEachPlayerInTurn(ctx, ApnapPlayers(g), ConstantMill(2), game.KeyedTail{})
}

// S82OldB726ThisIsAnEnchantment is the intervening "if this permanent is an
// enchantment" of Opal Acrolith, read from the trigger's source as it stands.
func S82OldB726ThisIsAnEnchantment(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(source.InstanceID)
	return ok && c.HasCardType("enchantment")
}

// S82OldB726ThisBecomesAnEnchantment is "This permanent becomes an enchantment"
// (Opal Acrolith's {0}): its types are set to Enchantment alone, so the
// creature type and Soldier are gone, until it is animated again. It has no
// duration, so it lasts as long as this object does (CR 611.2a, 400.7).
func S82OldB726ThisBecomesAnEnchantment(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !OnBattlefield(g, self) || SourceIsNewObject(g, item) {
		return nil
	}
	return ScopedEffectFor{
		Target:   self,
		Mods:     []game.Mod{game.SetTypesMod([]string{"Enchantment"}, nil)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Opal Acrolith — becomes an enchantment",
	}.Apply(NewContext(g, item))
}

// S82OldB726TargetBecomesAnEnchantmentWithNoAbilities is Soul Sculptor's body:
// the target creature becomes an enchantment (it is no longer a creature) and
// loses all abilities until a player casts a creature spell.
func S82OldB726TargetBecomesAnEnchantmentWithNoAbilities(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.SetTypesMod([]string{"Enchantment"}, nil), game.LoseAllAbilitiesMod()},
		Duration: DurationUntilACreatureSpellIsCast(),
		Label:    "Soul Sculptor — becomes an enchantment and loses all abilities until a player casts a creature spell",
	}.Apply(ctx)
}
