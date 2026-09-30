package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marionette Master — Creature — Human Artificer {4}{B}{B}:
//
//	"Fabricate 3 (When this creature enters, put three +1/+1 counters
//	 on it or create three 1/1 colorless Servo artifact creature
//	 tokens.)
//	 Whenever an artifact you control is put into a graveyard from the
//	 battlefield, target opponent loses life equal to this creature's
//	 power."
//
// The life loss is "this creature's power", which CR 608.2h reads as
// last-known information once the Master is gone: the trigger names
// the Master's object as it goes on the stack, and the effect reads it
// back through PermanentForEffect — live while it is on the
// battlefield, as it last existed once it has left (a wipe that takes
// the Master and its Servos together still drains for the Master's
// real power, counters included). A target that left the game in
// response makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	drain := Targeting(
		On(game.EventLTB, AnArtifactYouControlWasPutIntoAGraveyard, marionetteMasterLabel, marionetteMasterDrain),
		TargetPlayer("target opponent", Opponent()))
	drain.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, marionetteMasterLabel)
		if ref, ok := g.PermanentRefForEffect(source.InstanceID); ok {
			item.Params.Object = ref
		}
		return item
	}
	Register(Spec{
		OracleID:     "dabbf796-3b88-499e-8839-06fa36fe01ac",
		Name:         "Marionette Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Fabricate("Marionette Master", 3),
			drain,
		},
	})
}

const marionetteMasterLabel = "Marionette Master — target opponent loses life equal to Marionette Master's power"

// marionetteMasterDrain makes the chosen opponent lose life equal to
// the Master's power.
func marionetteMasterDrain(g *game.Game, item *game.StackItem) error {
	info, ok := g.PermanentForEffect(item.Params.Object)
	if !ok || info.Power <= 0 {
		return nil
	}
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -info.Power)
	}
	return nil
}
