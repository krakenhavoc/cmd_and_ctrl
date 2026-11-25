package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b306_helpers.go — shared shapes for the Battle for Zendikar Devoid
// cards of S82 batch 306 (#5547).

// OldB306Ingest is the ingest keyword (CR 702.115a): "whenever this creature
// deals combat damage to a player, that player exiles the top card of their
// library".
func OldB306Ingest(name string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(name+" — ingest: that player exiles the top card of their library",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return MillToZone{Player: item.Trigger.Event.Target, N: 1, To: game.ZoneExile}.Apply(NewContext(g, item))
		})
}

// OldB306OtherColorlessCreatureYouControl is true when `you` control a
// colorless creature other than `self`, read through the layers.
func OldB306OtherColorlessCreatureYouControl(g *game.Game, you, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.InstanceID != self && c.IsCreature() && c.IsColorless() {
			return true
		}
	}
	return false
}

// OldB306IfYouControlNoOtherColorless narrows a trigger on the source to the
// intervening "if you control no other colorless creatures" (CR 603.4):
// checked as it triggers and again as it resolves.
func OldB306IfYouControlNoOtherColorless(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && !OldB306OtherColorlessCreatureYouControl(g, source.Controller, source.InstanceID) &&
			(when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "you control no other colorless creatures", Test: func(ctx *Context) bool {
		return !OldB306OtherColorlessCreatureYouControl(ctx.Game, ctx.Controller(), ctx.Source())
	}}, t)
}

// OldB306ThisGetsPowerWhenYouCastColorless is "Whenever you cast a colorless
// spell, this creature gets +1/+0 until end of turn" (Kozilek's Sentinel).
func OldB306ThisGetsPowerWhenYouCastColorless(name string) game.TriggeredAbility {
	return WheneverYouCast(Colorless(), name+" — this creature gets +1/+0 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			if !OnBattlefield(g, item.SourceCardID) {
				return nil
			}
			return BoostUntilEOT{Target: item.SourceCardID, Power: 1,
				Label: name + " — +1/+0 until end of turn"}.Apply(NewContext(g, item))
		})
}
