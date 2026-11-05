package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b28_helpers.go — shared pieces for Star Trek batch 28 (#4973).

// registerGainLand registers one member of the "gain land" cycle:
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {A} or {B}."
//
// The body gain_lands.go's table registers for every row, extracted so
// a member printed in a later set (the Star Trek Planets, which add only
// the Planet subtype the type line already carries) can sit in its own
// file without a second copy of the same Spec.
func registerGainLand(oracleID, name, a, b string) {
	Register(Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility(a, b)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Key:       name + " — you gain 1 life",
			Effect:    Do(GainLife{Amount: 1}),
		}},
	})
}

// trekB28FirstChargeCountersOnThisThisTurn is U.S.S. Enterprise-D's
// condition: the event PUT one or more charge counters on the source (a
// removal is not a placement), and no earlier event this turn put any
// on this same object. An object that left the battlefield and came
// back is a new object (CR 400.7), so a placement on its previous
// existence does not count; the walk forgets everything before the
// source's last departure, the boundary counterCountBeforeLocked uses.
func trekB28FirstChargeCountersOnThisThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Target != source.InstanceID || b33CountersPlacedDelta(ev, game.CounterCharge, g) <= 0 {
		return false
	}
	placedBefore := false
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventCounterPlaced && prev.Target == source.InstanceID &&
			b33CountersPlacedDelta(prev, game.CounterCharge, g) > 0 {
			placedBefore = true
			continue
		}
		if prev.CardID != source.InstanceID {
			continue
		}
		switch prev.Kind {
		case game.EventLTB:
			placedBefore = false
		case game.EventZoneMove, game.EventCast:
			if prev.NewZone != game.ZoneBattlefield {
				placedBefore = false
			}
		}
	}
	return !placedBefore
}

// trekB28TwoChargeCountersOnTheTarget is Warp Core's activation: two
// charge counters on the targeted Spacecraft, if it is still legal.
func trekB28TwoChargeCountersOnTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 2}.Apply(ctx)
		}
	}
	return nil
}

// trekB28TappedNewPlanet is the Anomalies' last sentence: "Create a
// tapped Planet land token named New Planet with '{T}: Add one mana of
// any color.'" — the printed New Planet (ttrk 27) Genesis Effect makes.
func trekB28TappedNewPlanet() Applier {
	return CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1}
}
