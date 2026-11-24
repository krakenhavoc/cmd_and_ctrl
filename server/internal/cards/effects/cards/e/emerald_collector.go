package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emerald Collector — Creature — Human Pirate {1}{U}, 1/2 (Mystery
// Booster Commander, #5948):
//
//	"Whenever this creature deals combat damage to a player, draw a card.
//	 When you draw your third card in a turn, conjure a card named Mox
//	 Emerald into your hand. This ability triggers only once.
//	 {2}{G}: This creature has base power and toughness 4/4 until end of
//	 turn."
//
// "Your third card in a turn" is read as Sneaky Snacker reads it: the
// draw that makes this turn's count three. The conjured Mox Emerald is a
// real card (ADR 0099 amendment 2026-11-24).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d12534c8-6be7-41b0-852b-4461e7f962a1",
		Name:         "Emerald Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Emerald Collector — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.TriggersOnlyOnce(effects.On(game.EventDrawCard, YouDrawYourThirdCardInATurn,
				"Emerald Collector — conjure a card named Mox Emerald into your hand",
				effects.Do(effects.CheckConjure("Emerald Collector", effects.Conjure{Names: []string{"Mox Emerald"}, Zone: game.ZoneHand})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: This creature has base power and toughness 4/4 until end of turn.",
			Cost:    effects.ManaCost("{2}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, ctx.Source(), nil, "Emerald Collector — base power and toughness 4/4", game.SetBasePTMods(4, 4)...)
			},
		}},
	})
}

// YouDrawYourThirdCardInATurn is "when you draw your third card in a
// turn": the draw that makes this turn's count three.
func YouDrawYourThirdCardInATurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 3
}
