package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// volvers.go — the five Volvers of Apocalypse and Planeshift (#2360).
//
// Each prints "Kicker [A] and/or [B]" and two linked clauses:
//
//	"If this creature was kicked with its [A] kicker, it enters with two
//	 +1/+1 counters on it and with <ability>.
//	 If it was kicked with its [B] kicker, it enters with a +1/+1
//	 counter on it and with <ability>."
//
// THE SHAPES. The counters are CR 614.1c entry clauses declared with
// CountersIfKickedWith, one per kicker, so a Volver kicked with both
// enters with three, rides the CR 614 window (Doubling Season doubles
// each) and reads the kicker record off the cast itself —
// game.CastCounts.KickedWith. A reanimated Volver, a flickered one or
// a Clone of one was never cast, so it enters with none and, because
// the record on a permanent is cleared on the way out (CR 400.7d), has
// none of the abilities either.
//
// The abilities are NOT counters and NOT replacements: "enters with
// <ability>" gives the permanent the ability for as long as it is on
// the battlefield, which is a layer 6 grant keyed on the permanent's
// own kicker record (game.CardKickedWith):
//
//   - a keyword (trample, flying, first strike) is a KeywordGrant that
//     applies to the Volver itself when it was kicked with that cost;
//   - "Whenever this creature deals damage, you gain that much life"
//     is a triggered ability whose AppliesTo also checks the record;
//   - "Pay 3 life: Regenerate this creature" is an activated ability
//     whose Condition checks the record. Condition greys an ability the
//     permanent HAS, where the printed text says the permanent does not
//     have it unless kicked. The player-visible result is the same
//     (an unkicked Volver cannot activate it, and an Activated entry
//     costs nothing to look at), and nothing in the catalog reads "has
//     an activated ability" off a Volver.
//
// No simplification.

type volverAbility int

const (
	volverTrample volverAbility = iota
	volverFlying
	volverFirstStrike
	volverLifegain
	volverRegenerate
)

type volverKicker struct {
	cost    string
	counter int
	ability volverAbility
}

func volverKickedWith(cost string) func(g *game.Game, source uuid.UUID) bool {
	return func(g *game.Game, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && game.CardKickedWith(c, cost)
	}
}

func volverSpec(oracleID, name string, first, second volverKicker) Spec {
	spec := Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers(first.cost, second.cost),
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfKickedWith(game.CounterPlusOne, first.cost, first.counter),
			CountersIfKickedWith(game.CounterPlusOne, second.cost, second.counter),
		},
	}
	for _, k := range []volverKicker{first, second} {
		k := k
		switch k.ability {
		case volverTrample, volverFlying, volverFirstStrike:
			keyword := map[volverAbility]string{
				volverTrample: "trample", volverFlying: "flying", volverFirstStrike: "first strike",
			}[k.ability]
			spec.Static = append(spec.Static, KeywordGrant(
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && game.CardKickedWith(*target, k.cost)
				}, keyword))
		case volverLifegain:
			spec.Triggered = append(spec.Triggered, game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && ev.Source == source.InstanceID && game.CardKickedWith(*source, k.cost)
				},
				Key: name + " — you gain that much life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				},
			})
		case volverRegenerate:
			kicked := volverKickedWith(k.cost)
			spec.Activated = append(spec.Activated, ActivatedAbility{
				Label:   "Pay 3 life: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    PayLife(3),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					return kicked(g, source)
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			})
		}
	}
	return spec
}

func init() {
	Register(volverSpec("639d16ed-daa4-4c8d-8550-5c9a29f4d9fe", "Necravolver",
		volverKicker{"{1}{G}", 2, volverTrample}, volverKicker{"{W}", 1, volverLifegain}))
	Register(volverSpec("9431c81f-7b71-4c5d-a63c-1b5aba03601b", "Degavolver",
		volverKicker{"{1}{B}", 2, volverRegenerate}, volverKicker{"{R}", 1, volverFirstStrike}))
	Register(volverSpec("c00adb98-81dc-490f-9e0f-7f8546ef22ce", "Rakavolver",
		volverKicker{"{1}{W}", 2, volverLifegain}, volverKicker{"{U}", 1, volverFlying}))
	Register(volverSpec("03a4cbe2-cf35-4c2c-b035-666f994ffd09", "Anavolver",
		volverKicker{"{1}{U}", 2, volverFlying}, volverKicker{"{B}", 1, volverRegenerate}))
	Register(volverSpec("3ae86510-f828-48f8-adab-759df4b5544e", "Cetavolver",
		volverKicker{"{1}{R}", 2, volverFirstStrike}, volverKicker{"{G}", 1, volverTrample}))
}
