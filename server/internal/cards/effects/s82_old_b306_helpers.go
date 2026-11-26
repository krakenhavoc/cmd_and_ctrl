package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b306_helpers.go — shared shapes for the Battle for Zendikar Devoid
// cards of S82 batch 306 (#5547).

// OldB306IfYouControlNoOtherColorless narrows a trigger on the source to the
// intervening "if you control no other colorless creatures" (CR 603.4):
// checked as it triggers and again as it resolves.
func OldB306IfYouControlNoOtherColorless(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && !OldB305ControlsAnotherColorlessCreature(g, source.Controller, source.InstanceID) &&
			(when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "you control no other colorless creatures", Test: func(ctx *Context) bool {
		return !OldB305ControlsAnotherColorlessCreature(ctx.Game, ctx.Controller(), ctx.Source())
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
