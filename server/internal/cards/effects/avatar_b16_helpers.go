package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b16_helpers.go — shared helpers for Avatar: The Last Airbender,
// batches 16 and 17 (#4459).

// avB16CondLessonInYourGraveyard is the intervening "if there's a Lesson
// card in your graveyard" (Walltop Sentries, Dragonfly Swarm), re-checked
// as the ability resolves (CR 603.4). It reads the item's controller's
// graveyard (game.FootprintCondGraveyard). The When half is
// avB16DiedWithALessonInYourGraveyard.
type avB16CondLessonInYourGraveyard struct{}

func (avB16CondLessonInYourGraveyard) Holds(ctx *Context) bool {
	return avB4LessonsInGraveyard(ctx.Game, ctx.Controller()) > 0
}

func (avB16CondLessonInYourGraveyard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 0), true
}

// avB16DiedWithALessonInYourGraveyard is "When this creature dies, if
// there's a Lesson card in your graveyard": the creature itself died and,
// as the ability triggers, its controller's graveyard holds a Lesson card
// (the creature is in it by now, and is never a Lesson). If there is none,
// the ability is not put on the stack at all (the card's ruling).
func avB16DiedWithALessonInYourGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisDied(ev, source, lki, g) && avB4LessonsInGraveyard(g, source.Controller) > 0
}

// avB16WhenThisDiesIfLesson is the whole row for "When this creature
// dies, if there's a Lesson card in your graveyard, <steps>".
func avB16WhenThisDiesIfLesson(label string, steps ...Applier) game.TriggeredAbility {
	return InterveningIf(avB16CondLessonInYourGraveyard{},
		On(game.EventLTB, avB16DiedWithALessonInYourGraveyard, label, Do(steps...)))
}

// creaturesYouControlWithFlyingGetPlusOne is "Creatures you control with
// flying get +1/+1." (Favorable Winds, Air Nomad Legacy): a layer 7c
// modify over the controller's creatures that have flying, read through
// the layers on every pass.
func creaturesYouControlWithFlyingGetPlusOne() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Creatures you control with flying get +1/+1.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: b30CreaturesYouControlWithFlying,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power++
			c.Toughness++
		},
	}
}

// avB16SelfEntersWithCountersIfRaid is "Raid — This creature enters with
// N <kind> counters on it if you attacked this turn" (Cruel
// Administrator). The condition is read as the creature enters (CR
// 614.1c): a creature that entered before its controller attacked this
// turn gets nothing later.
func avB16SelfEntersWithCountersIfRaid(kind string, n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  "Raid — Enters with " + counterCountPhrase(n, kind) + " on it if you attacked this turn.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				b18AttackedThisTurn(g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
	}
}

// avB16PlusOneCountersOnLandsYouControl is the number of +1/+1 counters
// on lands `you` controls (Toph, the Blind Bandit), the land type read
// through the layers.
func avB16PlusOneCountersOnLandsYouControl(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == you && c.IsLand() {
			if k := c.Counters[game.CounterPlusOne]; k > 0 {
				n += k
			}
		}
	}
	return n
}

// avB16SoldierFirebendingToken is the 2/2 red Soldier with firebending 1
// (Cruel Administrator): the catalog's `token:soldier-firebending`, the
// same printed token Firebender Ascension makes.
func avB16SoldierFirebendingToken() game.Card {
	return tokenFromCatalog(printedFirebendingSoldierToken)
}

// avB16Earthbend is a triggered ability's whole effect when it is
// "earthbend N" and nothing else (Toph, the Blind Bandit's enters trigger,
// Earth Village Ruffians' dies trigger): the row's target is the keyword's
// own clause (EarthbendTargets).
func avB16Earthbend(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return EarthbendFirstTarget(EarthbendCount(n))(item, NewContext(g, item))
	}
}
