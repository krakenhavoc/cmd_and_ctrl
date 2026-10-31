package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain America, Living Legend — Legendary Creature — Human Soldier Hero
// {1}{W}{U}, 3/4 (Marvel Super Heroes, #4455, seam #4576):
//
//	"Vigilance
//	 Whenever a creature you control becomes tapped during your turn, if
//	 it's the first time that creature has become tapped this turn, untap
//	 it."
//
// "Becomes tapped" is any tap (CR 701.26a), read from two event kinds for
// the reason ThisBecameTapped gives: the engine taps an attacker without an
// EventTapCard, so an EventAttack whose creature is now tapped counts and a
// vigilance attacker does not. "The first time" is the engine's per-object
// became-tapped tally (TurnTally.BecameTapped, Game.TimesBecameTappedThisTurn):
// it is written ahead of the trigger harvester, so the tap that triggers
// this reads 1, a second tap reads 2, a creature that left and came back is
// a new object that reads 0 (CR 400.7), and the tally is flushed with the
// turn.
//
// The "if" is a fact about the tap that triggered, so it is decided when the
// tap happens and not asked again at resolution. The creature is the one the
// event named, read as it is now: one that left the battlefield, or came back
// as a new object, is not untapped.
//
// No simplification.
const captainAmericaLivingLegendLabel = "Captain America, Living Legend — untap it"

func init() {
	Register(Spec{
		OracleID:        "67aa946d-e694-494e-ba0c-44880d5bc03e",
		Name:            "Captain America, Living Legend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b32CreatureYouControlBecameTapped(ev, source, g) &&
						IsYourTurn(g, source.Controller) &&
						g.TimesBecameTappedThisTurn(ev.CardID) == 1
				},
				captainAmericaLivingLegendLabel, untapTheTriggeringCreature),
		},
	})
}

// untapTheTriggeringCreature is "untap it" for the creature a trigger's
// event was about, as it is now (CR 608.2h): nothing if it has left the
// battlefield or is a new object.
func untapTheTriggeringCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	return UntapTarget{Target: ctx.Trigger().Event.CardID}.Apply(ctx)
}
