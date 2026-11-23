package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saradoc, Master of Buckland — Legendary Creature — Halfling Citizen {3}{W},
// 2/4 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Saradoc or another nontoken creature you control with power 2
//	 or less enters, create a 1/1 white Halfling creature token.
//	 Tap two other untapped Halflings you control: Saradoc gets +2/+0 and
//	 gains lifelink until end of turn."
//
// The watch is Radagast's "this or another nontoken creature" with a power
// test on the other creature, read as it entered (Saradoc himself counts
// whatever his power). The cost is Adaptive Gemguard's TapOthersCost over
// Halflings, with "other" excluding Saradoc: summoning sickness does not
// matter because the cost has no {T} (CR 302.6). The pump applies to
// Saradoc only while he is still the permanent that activated.
//
// No simplification.
func init() {
	const tapTwo = "two other untapped Halflings you control"
	effects.Register(effects.Spec{
		OracleID:     "3c7f7a67-6a71-4f40-8554-2e724e6916a6",
		Name:         "Saradoc, Master of Buckland",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && !effects.IsToken(c) && c.PowerForComparison() <= 2
			}, "Saradoc, Master of Buckland — create a 1/1 white Halfling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Halfling"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two other untapped Halflings you control: Saradoc gets +2/+0 and gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:         2,
				Filter:        effects.TargetPermanent(tapTwo, effects.HasSubtype("Halfling")),
				Label:         tapTwo,
				ExcludeSource: true,
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Saradoc, Master of Buckland — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"lifelink"}, Label: "Saradoc, Master of Buckland — lifelink"}.Apply(ctx)
			},
		}},
	})
}
