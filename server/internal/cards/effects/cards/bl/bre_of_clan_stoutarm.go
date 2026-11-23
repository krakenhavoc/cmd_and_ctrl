package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "59384481-3ca0-49dd-b94f-d3680b336f1e",
		Name:         "Bre of Clan Stoutarm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Another target creature you control gains flying and lifelink until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying", "lifelink")), game.AnswerCombatGrant),
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "lifelink"},
					Label: "Bre of Clan Stoutarm — flying and lifelink until end of turn"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && effects.B15LifeGainedThisTurn(g, source.Controller) > 0
			}, "Bre of Clan Stoutarm — exile cards until a nonland card; cast it free if its mana value is at most the life you gained this turn, otherwise put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the intervening "if" is checked again on resolution.
					if effects.B15LifeGainedThisTurn(g, item.Controller) == 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 0, game.ZoneExile, effects.UntilNonland, ctx.ItemTail(breOfClanStoutarmExiled))
				}),
		},
	})
}

// breOfClanStoutarmExiled is the keyed "then" of the exile: the nonland card
// is cast free when its mana value fits the life gained this turn, else it
// goes to hand. An on-disk key: never renamed, never reused.
var breOfClanStoutarmExiled = effects.CardTail("bre-of-clan-stoutarm/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		return ctx.Game.CastExiledFreeOrToHandForEffect(ctx.Controller(), ctx.Source(), id,
			effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller()))
	}
	return nil
})
