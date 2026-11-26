package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vein Drinker — Creature — Vampire {4}{B}{B}, 4/4 (Commander 2017, #5533):
//
//	"Flying
//	 {R}, {T}: This creature deals damage equal to its power to target
//	 creature. That creature deals damage equal to its power to this
//	 creature.
//	 Whenever a creature dealt damage by this creature this turn dies, put a
//	 +1/+1 counter on this creature."
//
// Garruk Relentless's pair of blows, with the Drinker's own power as the
// first. They are separate instructions in printed order (not a fight, CR
// 701.14 does not apply) inside one resolution, and no state-based action is
// checked between them (CR 704.3): a creature the first blow kills still
// deals its damage, equal to its power as it now is, to the Drinker. If the
// target has gone by then (it left the battlefield) it deals nothing. The counter trigger is Vampiric Dragon's: any creature the
// Drinker damaged this turn, by this ability or in combat, that dies gives it
// a counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "705d89f6-13a8-4f40-a01a-3d5d0ce4f34e",
		Name:            "Vein Drinker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Vein Drinker — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals damage equal to its power to target creature. That creature deals damage equal to its power to this creature.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				ref, ok := g.PermanentRefForEffect(victim)
				if !ok {
					return nil
				}
				drinker, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				if drinker.Power > 0 {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: victim, Amount: drinker.Power}).Apply(ctx); err != nil {
						return err
					}
				}
				info, ok := g.PermanentForEffect(ref)
				if !ok || info.Power <= 0 || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.DealDamage{SourceObject: &ref, Target: item.SourceCardID, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
