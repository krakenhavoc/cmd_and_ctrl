package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Somberwald Alpha — Creature — Wolf {3}{G}, 3/2 (Magic Origins):
//
//	"Whenever a creature you control becomes blocked, it gets +1/+1 until
//	 end of turn.
//	 {1}{G}: Target creature you control gains trample until end of turn."
//
// Norin's becomes-blocked watch over every creature its controller
// controls (the Alpha itself included), one trigger per blocked
// creature. The pump goes to that creature if it is still the same
// object on the battlefield as it resolves (CR 400.7). The trample
// grant is an ordinary targeted activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11b12954-dd1b-4c50-8433-15d8c1a15e29",
		Name:         "Somberwald Alpha",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller == source.Controller
			}, "Somberwald Alpha — the blocked creature gets +1/+1 until end of turn", somberwaldAlphaPump),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Target creature you control gains trample until end of turn.",
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample"},
					Label:    "Somberwald Alpha — trample",
				}.Apply(ctx)
			},
		}},
	})
}

func somberwaldAlphaPump(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	if o := item.Trigger.Object; o != nil && g.ObjectEpochForEffect(id) != o.Ref().Epoch {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Somberwald Alpha — +1/+1"}.Apply(effects.NewContext(g, item))
}
