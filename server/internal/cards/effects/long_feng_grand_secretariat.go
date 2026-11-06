package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long Feng, Grand Secretariat — Legendary Creature — Human Advisor
// {1}{B/G}{B/G}, 2/3 (Avatar: The Last Airbender, #4459):
//
//	"Whenever another creature you control or a land you control is put
//	 into a graveyard from the battlefield, put a +1/+1 counter on target
//	 creature you control."
//
// One ability with one event: an EventLTB to a graveyard of a permanent
// that, as it last existed on the battlefield (CR 603.10a), was under
// Long Feng's controller and was a creature other than Long Feng or a
// land. A land creature is one trigger, not two. The target is chosen as
// the trigger goes on the stack and is checked again on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9bf9fe83-b0df-4ef8-ac54-c0d223d4e410",
		Name:         "Long Feng, Grand Secretariat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventLTB, longFengYourCreatureOrLandWentToTheGraveyard,
					"Long Feng, Grand Secretariat — put a +1/+1 counter on target creature you control",
					PlusOneCounterOnChosenTarget),
				TargetCreature("target creature you control", YouControl()),
			),
		},
	})
}

// longFengYourCreatureOrLandWentToTheGraveyard is the trigger condition:
// another creature or a land, under the source's controller as it left,
// put into a graveyard from the battlefield.
func longFengYourCreatureOrLandWentToTheGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || leftUnderControlOf(ev, c) != source.Controller {
		return false
	}
	if leftAsType(ev, c, "land") {
		return true
	}
	return ev.CardID != source.InstanceID && leftAsType(ev, c, "creature")
}
