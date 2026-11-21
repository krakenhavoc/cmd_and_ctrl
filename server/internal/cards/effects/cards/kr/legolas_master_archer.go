package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legolas, Master Archer — Legendary Creature — Elf Archer {1}{G}{G}, 1/4
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Reach
//	 Whenever you cast a spell that targets Legolas, put a +1/+1 counter on
//	 Legolas.
//	 Whenever you cast a spell that targets a creature you don't control,
//	 Legolas deals damage equal to its power to up to one target creature."
//
// The first trigger is heroic (Wingsteed Rider's Heroic), read off the spell's
// targets as it is cast (CR 601.2c). The second reads the same targets for a
// creature on the battlefield its caster doesn't control, and its own target
// clause is "up to one target creature", chosen as the trigger goes on the
// stack above the spell (CR 603.3d); Legolas's power is read as the ability
// resolves, and nothing is dealt if he has left or the target is no longer
// legal (CR 608.2b).
//
// No simplification.
func init() {
	dmg := effects.On(game.EventCast, youCastASpellThatTargetsACreatureYouDontControl,
		"Legolas, Master Archer — Legolas deals damage equal to its power to up to one target creature",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard || !effects.OnBattlefield(g, item.SourceCardID) {
				return nil
			}
			src, found := g.LookupCardForEffect(item.SourceCardID)
			if !found || src.CurrentPower() <= 0 {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: src.CurrentPower()}.Apply(ctx)
		})
	dmg.Targets = effects.TargetCreature("up to one target creature").WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:        "5e99497f-7c73-49ae-88b3-78d5d74ba9ea",
		Name:            "Legolas, Master Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Legolas, Master Archer — put a +1/+1 counter on Legolas", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			dmg,
		},
	})
}

// youCastASpellThatTargetsACreatureYouDontControl is the When of the second
// trigger: a spell its controller cast targets a creature on the battlefield
// that controller doesn't control.
func youCastASpellThatTargetsACreatureYouDontControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && c.Controller != source.Controller && effects.OnBattlefield(g, t.ID) {
			return true
		}
	}
	return false
}
