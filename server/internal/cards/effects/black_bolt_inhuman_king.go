package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Bolt, Inhuman King — Legendary Creature — Inhuman Noble Hero
// {2}{W}{U}, 3/3 (Marvel Super Heroes Commander, #4707):
//
//	"Flying
//	 Whenever you cast a noncreature spell, Black Bolt gets +2/+2 until end
//	 of turn.
//	 Lethal Voice — Whenever Black Bolt becomes the target of a spell or
//	 ability an opponent controls, destroy target nonland permanent that
//	 player controls."
//
// "That player" is the controller of the targeting spell or ability: the
// triggering event's Actor on EventBecomesTarget. ControlledByThatEventActor
// reads it in the one shared target walk (ADR 0019's 2026-10-30 amendment), so
// the picker offers only that opponent's nonland permanents, the announce
// check and the CR 608.2b re-check agree, and a permanent that changes hands
// before the ability resolves is no longer a legal target. The trigger fires
// at announce (CR 601.2c) and goes on the stack above the targeting spell, so
// it resolves first. It does not fire for a spell or ability Black Bolt's own
// controller controls: ev.Actor is that player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a30f30e3-dbbf-4cb2-b3d0-76f57dffb2dd",
		Name:            "Black Bolt, Inhuman King",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Black Bolt, Inhuman King — +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2,
						Label: "Black Bolt, Inhuman King — +2/+2 until end of turn"}.Apply(NewContext(g, item))
				}),
			Targeting(
				On(game.EventBecomesTarget, blackBoltTargetedByAnOpponent,
					"Black Bolt, Inhuman King — Lethal Voice: destroy target nonland permanent that player controls",
					destroyFirstLegalCardTarget),
				ControlledByThatEventActor(TargetPermanent("target nonland permanent that player controls", Nonland()))),
		},
	})
}

// blackBoltTargetedByAnOpponent is "Black Bolt becomes the target of a spell
// or ability an opponent controls": the event names the source and its Actor,
// the controller of the targeting spell or ability, is someone else.
func blackBoltTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID &&
		ev.Actor != uuid.Nil && ev.Actor != source.Controller
}
