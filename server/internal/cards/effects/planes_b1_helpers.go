package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b1_helpers.go holds the shapes the first batch of plane cards
// (ADR 0216 slice 4, #4870) settled and every later batch copies. A plane
// is an ordinary catalog card keyed by its oracle ID (docs/adding-cards.md,
// "Plane cards"); its abilities function only while it is face up in the
// command zone, where game/planechase_sources.go finds them (CR 901.7), and
// "you" on it is the planar controller, stamped on the card's Controller
// (CR 901.6, 311.5). Append-only.

// WheneverChaosEnsues is a plane's chaos ability (CR 311.7), "Whenever
// chaos ensues, <effect>":
//
//	"Each plane card has a triggered ability that triggers “Whenever
//	 chaos ensues.” These are called chaos abilities. ... This ability
//	 triggers if the chaos symbol is rolled on the planar die (see rule
//	 901.9b), if a resolving spell or ability says that chaos ensues, or
//	 if a resolving spell or ability states that chaos ensues for a
//	 particular object. ... A chaos ability is controlled by the current
//	 planar controller."
//
// It watches the planar die's EventRollDie showing the chaos face, rolled
// by anyone (an effect can make a player other than the planar controller
// roll it, CR 901.9's note), or by a resolving "chaos ensues" instruction
// (EventChaosEnsues, game.ChaosEnsuesForEffect, Oteclán), and it is
// controlled by the source's controller, the planar controller.
func WheneverChaosEnsues(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventRollDie, game.EventChaosEnsues}, chaosEnsued, label, effect)
}

// chaosEnsued is the chaos face of the planar die (CR 901.9b) or a
// resolving instruction that says chaos ensues (CR 311.7).
func chaosEnsued(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventChaosEnsues || ev.PlanarFace == game.PlanarFaceChaos
}

// WhenYouPlaneswalkToThisAndAtYourUpkeep is "When you planeswalk to ~ and
// at the beginning of your upkeep, <effect>": ONE triggered ability with two
// trigger conditions (OnAny, as "enters or attacks" is), so the label and the
// effect are shared and a restore names one row. The planeswalk half is
// WhenYouPlaneswalkToThis's condition (CR 901.11b; the starting plane does
// not trigger it, CR 901.5); the upkeep half is the planar controller's
// upkeep, and the planar controller is the active player (CR 901.6).
func WhenYouPlaneswalkToThisAndAtYourUpkeep(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventBeginUpkeep}, youPlaneswalkedToThisOrYourUpkeep, label, effect)
}

func youPlaneswalkedToThisOrYourUpkeep(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventPlaneswalkTo:
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller
	case game.EventBeginUpkeep:
		return ByYou(ev, source, lki, g)
	}
	return false
}

// thisIsntACreatureUntilEndOfTurn is "this <permanent> isn't a creature
// until end of turn" as a trigger's whole effect (Weeping Angel, and Aplan
// Mortarium's Alien Angel token): a layer-4 RemoveTypesMod("Creature") on
// the trigger's source while it is still that object. The layer pass drops
// the creature types with it (CR 205.1a), so it stays an artifact and stops
// being an Alien Angel, and one in combat is removed from it (CR 506.4).
func thisIsntACreatureUntilEndOfTurn(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
			return nil
		}
		return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label, game.RemoveTypesMod("Creature"))
	}
}

// TimeTravel is "time travel" (CR 701.56a):
//
//	"To time travel means to choose any number of permanents you control
//	 with one or more time counters on them and/or suspended cards you own
//	 in exile with one or more time counters on them and, for each of
//	 those objects, put a time counter on it or remove a time counter
//	 from it."
//
// Each such object is asked about in turn, a keyed option pick (so a table
// waiting on it is a restore point): put a counter on it, remove one, or
// leave it. The objects are listed as the instruction begins and each is
// judged again as it is reached (CR 608.2b). Removing a suspended card's
// last time counter emits the counter event suspend's third ability
// watches (CR 702.62a), so the free cast is offered. "Player" zero is the
// effect's controller.
type TimeTravel struct {
	Player uuid.UUID
}

func (t TimeTravel) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return askTimeTravel(ctx, player, timeTravelObjects(ctx.Game, player))
}

// timeTravelObjects lists what `player` may time travel: the permanents they
// control with a time counter, then the suspended cards they own in exile
// (CR 702.62b: in exile, with suspend, with a time counter).
func timeTravelObjects(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.Counters[game.CounterTime] > 0 {
			out = append(out, c.InstanceID)
		}
	}
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Owner == player && game.CardIsSuspended(c) && game.CardHasSuspend(c) {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}

// timeTravelStillEligible re-judges one listed object as it is reached.
func timeTravelStillEligible(g *game.Game, player, id uuid.UUID) (game.Card, bool) {
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterTime] <= 0 {
		return game.Card{}, false
	}
	z := g.FindCardZoneForEffect(id)
	switch {
	case z == nil:
		return game.Card{}, false
	case z.Kind == game.ZoneBattlefield:
		return c, c.Controller == player
	case z.Kind == game.ZoneExile:
		return c, c.Owner == player && game.CardHasSuspend(c)
	}
	return game.Card{}, false
}

// askTimeTravel asks about the first still-eligible object of `ids`.
func askTimeTravel(ctx *Context, player uuid.UUID, ids []uuid.UUID) error {
	for len(ids) > 0 {
		c, ok := timeTravelStillEligible(ctx.Game, player, ids[0])
		if !ok {
			ids = ids[1:]
			continue
		}
		name := c.Name
		if c.FaceDownIsPermanent() {
			name = "a face-down permanent"
		}
		return PickOption{
			Player:   player,
			Question: fmt.Sprintf("Time travel — %s has %d time counters", name, c.Counters[game.CounterTime]),
			Options: []game.ChoiceOption{
				{Label: "Put a time counter on " + name},
				{Label: "Remove a time counter from " + name},
				{Label: "Leave " + name + " as it is"},
			},
			ThenKey:  timeTravelAnswered,
			Carry:    append([]uuid.UUID(nil), ids...),
			WithItem: true,
		}.Apply(ctx)
	}
	return nil
}

// timeTravelAnswered applies one answer and asks about the next object. An
// on-disk key: never renamed, never reused. Registered in init: the answer
// asks the next question, which names this key.
var timeTravelAnswered game.OptionPickThen

func init() { timeTravelAnswered = OptionPickThen("option-pick/time-travel", timeTravelAnswer) }

func timeTravelAnswer(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Option == nil {
		return nil
	}
	id, player := r.Carry[0], r.Chooser
	if _, ok := timeTravelStillEligible(ctx.Game, player, id); ok {
		switch r.Index {
		case 0:
			if err := ctx.Game.AddCounterByForEffect(player, id, game.CounterTime, 1); err != nil {
				return err
			}
		case 1:
			if err := ctx.Game.AddCounterByForEffect(player, id, game.CounterTime, -1); err != nil {
				return err
			}
		}
	}
	return askTimeTravel(ctx, player, r.Carry[1:])
}
