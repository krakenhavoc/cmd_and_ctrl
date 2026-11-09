package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bre of Clan Stoutarm — Legendary Creature — Giant Warrior {2}{R}{W}
// (Lorwyn Eclipsed, #4458):
//
//	"{1}{W}, {T}: Another target creature you control gains flying and
//	 lifelink until end of turn.
//	 At the beginning of your end step, if you gained life this turn, exile
//	 cards from the top of your library until you exile a nonland card. You
//	 may cast that card without paying its mana cost if the spell's mana
//	 value is less than or equal to the amount of life you gained this
//	 turn. Otherwise, put it into your hand."
//
// The trigger checks "you gained life this turn" when the end step begins
// and again as it resolves (CR 603.4). The exile is a keyed mill to exile
// until a nonland card; the free cast happens during the resolution (CR
// 608.2g) with the life gained as the mana value cap, and a card that is
// too expensive, or whose free cast is declined, goes to hand (the same
// offer Breaching Dragonstorm makes). X in the card's cost is 0.
func init() {
	Register(Spec{
		OracleID:     "59384481-3ca0-49dd-b94f-d3680b336f1e",
		Name:         "Bre of Clan Stoutarm",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Another target creature you control gains flying and lifelink until end of turn.",
			Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 0, 0, "flying", "lifelink")), game.AnswerCombatGrant),
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "lifelink"},
					Label: "Bre of Clan Stoutarm — flying and lifelink until end of turn"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && b15LifeGainedThisTurn(g, source.Controller) > 0
			}, "Bre of Clan Stoutarm — exile cards until a nonland card; cast it free if its mana value is at most the life you gained this turn, otherwise put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the intervening "if" is checked again on resolution.
					if b15LifeGainedThisTurn(g, item.Controller) == 0 {
						return nil
					}
					ctx := NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 0, game.ZoneExile, UntilNonland, ctx.ItemTail(breOfClanStoutarmExiled))
				}),
		},
	})
}

// breOfClanStoutarmExiled is the keyed "then" of the exile: the nonland card
// is cast free when its mana value fits the life gained this turn, else it
// goes to hand. An on-disk key: never renamed, never reused.
var breOfClanStoutarmExiled = CardTail("bre-of-clan-stoutarm/after-exile", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		return ctx.Game.CastExiledFreeOrToHandForEffect(ctx.Controller(), ctx.Source(), id,
			b15LifeGainedThisTurn(ctx.Game, ctx.Controller()))
	}
	return nil
})
