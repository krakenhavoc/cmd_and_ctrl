package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrogen Rotpriest — Creature — Phyrexian Zombie Cleric {2}{B}{G}, 1/5
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 2
//	 Whenever a creature you control with toxic deals combat damage to a
//	 player, that player gets an additional poison counter.
//	 {1}{B}{G}: Target creature you control with toxic gains deathtouch
//	 until end of turn."
//
// The extra poison counter is its own trigger, one per creature that
// dealt the damage (toxic's own counters are part of the damage, CR
// 702.164c). Rotpriest itself qualifies. "With toxic" is read off the
// creature as it dealt the damage.
//
// No simplification.
func init() {
	const activateLabel = "{1}{B}{G}: Target creature you control with toxic gains deathtouch until end of turn."
	Register(Spec{
		OracleID:        "7d9a424e-cab1-410f-9c04-ee0025d04169",
		Name:            "Necrogen Rotpriest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g) && s82OneB05HasToxic(g, ev.Source)
			}, "Necrogen Rotpriest — that player gets an additional poison counter",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					p := item.Trigger.Event.Target
					if g.PlayerByIDForEffect(p) == nil {
						return nil
					}
					return g.AddPlayerCounterByForEffect(item.Controller, p, game.CounterPoison, 1)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   activateLabel,
			Cost:    ManaCost("{1}{B}{G}"),
			Targets: TargetCreature("target creature you control with toxic", YouControl(), WithToxic()),
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"},
					Label: "Necrogen Rotpriest — deathtouch until end of turn"}.Apply(ctx)
			},
		}},
	})
}
