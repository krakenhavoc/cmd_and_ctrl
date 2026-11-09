package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackdaw Savior — Creature — Bird Cleric {2}{W}, 3/1 (Duskmourn: House of
// Horror, #4466):
//
//	"Flying
//	 Whenever this creature or another creature you control with flying
//	 dies, return another target creature card with lesser mana value from
//	 your graveyard to the battlefield."
//
// Scrap Trawler's shape (TargetsFrom, #1223): "lesser" is lesser than the
// mana value of the creature that just died, read off the trigger's
// CR 603.10 last-known snapshot rather than the card now in the graveyard
// (the ruling; X is 0 there). The trigger is one per creature that died, so
// Jackdaw and another flyer dying together trigger twice (the ruling). The
// other creature must have had flying as it last existed and been under your
// control as it died (WasKeyword, leftUnderControlOf); Jackdaw itself counts
// whether or not it still had flying. "Another" is the target not being the
// creature that died: with a lesser mana value it never can be. When nothing
// in the graveyard is cheaper the ability is not put on the stack at all
// (CR 603.3d). The return is to the battlefield under your control, and the
// target is checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e31ef54-b2f2-4f3f-9779-151b74470ce4",
		Name:            "Jackdaw Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventLTB},
			AppliesTo:   jackdawSaviorFlyerDied,
			TargetsFrom: jackdawSaviorLesserCreatureClause,
			Key:         "Jackdaw Savior — return another creature card with lesser mana value from your graveyard to the battlefield",
			Effect:      moseoInfusion,
		}},
	})
}

// jackdawSaviorFlyerDied is the trigger condition: the source itself died,
// or another creature its controller controlled, with flying as it last
// existed, died.
func jackdawSaviorFlyerDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, c) != source.Controller {
		return false
	}
	had, known := ev.WasKeyword("flying")
	return known && had
}

// jackdawSaviorLesserCreatureClause builds "another target creature card
// with lesser mana value from your graveyard" from the creature that died.
func jackdawSaviorLesserCreatureClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	ceiling := 0
	if tc.Object != nil {
		ceiling = tc.Object.ManaValue
	}
	return TargetCardInGraveyard(
		"another target creature card with lesser mana value from your graveyard",
		YouOwn(), Creature(), ManaValueLE(ceiling-1),
	)
}
