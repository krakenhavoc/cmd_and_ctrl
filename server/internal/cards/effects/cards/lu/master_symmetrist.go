package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master Symmetrist — Creature — Rhino Druid {2}{G}{G}, 4/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Reach
//	 Whenever a creature you control with power equal to its toughness
//	 attacks, it gains trample until end of turn."
//
// One trigger for each such attacker (CR 508.3a). "Power equal to its
// toughness" is read off the creature as it is declared as an attacker, layers
// and counters included (it is a trigger condition, not an intervening "if");
// the creature that gains trample is the one the event names, and one that has
// left by resolution gains nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ffce416c-90c6-43ed-9019-e5a32accad6d",
		Name:            "Master Symmetrist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.CurrentPower() == c.CurrentToughness()
			}, "Master Symmetrist — the attacker gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !effects.OnBattlefield(g, attacker) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"trample"},
						Label: "Master Symmetrist — trample"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
