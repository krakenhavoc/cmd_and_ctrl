package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orah, Skyclave Hierophant — Legendary Creature — Kor Cleric {2}{W}{B},
// 3/3 (Zendikar Rising, #5508):
//
//	"Lifelink
//	 Whenever Orah or another Cleric you control dies, return target Cleric
//	 card with lesser mana value from your graveyard to the battlefield."
//
// The trigger fires for Orah herself or any Cleric the controller had as it
// died, judged on its last-known subtypes and controller (CR 603.10a). The
// target is chosen as the trigger goes on the stack, and "lesser" is read
// off the dead creature's last-known mana value (TargetsFrom, Riveteers
// Ascendancy's shape), so with nothing cheaper in the graveyard the trigger
// is not put on the stack (CR 603.3d). The card returns under its owner's
// control, which is the controller's (it is "your graveyard").
//
// No simplification.
func init() {
	const label = "Orah, Skyclave Hierophant — return target Cleric card with lesser mana value from your graveyard to the battlefield"
	trigger := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := effects.DiedCreature(ev, g)
		return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
			(ev.CardID == source.InstanceID || effects.LeftAsSubtype(ev, dead, source.TextSubtype("Cleric")))
	}, label, func(g *game.Game, item *game.StackItem) error {
		return effects.ReturnFirstLegalGraveyardTarget(g, item, game.ZoneBattlefield)
	})
	trigger.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		ceiling := 0
		if tc.Object != nil {
			ceiling = tc.Object.ManaValue
		}
		return effects.TargetCardInGraveyard("target Cleric card with lesser mana value from your graveyard",
			effects.YouOwn(), effects.Subtype("Cleric"), effects.ManaValueLE(ceiling-1))
	}
	effects.Register(effects.Spec{
		OracleID:        "fe348f31-4530-4c59-9ee2-6839eab973da",
		Name:            "Orah, Skyclave Hierophant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
