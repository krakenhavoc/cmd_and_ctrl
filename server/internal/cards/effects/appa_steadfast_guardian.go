package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Appa, Steadfast Guardian — 3/4 Legendary Creature — Bison Ally
// for {2}{W}{W}:
//
//	"Flash
//	Flying
//	When Appa enters, airbend any number of other target nonland
//	permanents you control. (Exile them. While each one is exiled,
//	its owner may cast it for {2} rather than its mana cost.)
//	Whenever you cast a spell from exile, create a 1/1 white Ally
//	creature token."
//
// The two abilities are the two halves of one plan: airbend your
// own board in response to a wrath, then rebuy it at {2} a piece
// and collect an Ally for each rebuy. The second ability is why
// airbending your OWN permanents is a payoff rather than a cost.
//
// Shape notes:
//
//   - "Any number of OTHER target …" is a multi-target trigger
//     (Min 0, Max unbounded) whose clause is built per trigger by
//     TargetsFrom, which receives the source — so the "other" is a
//     real NotSelf predicate and Appa is never offered to himself
//     (CR 115.1: an illegal target cannot be chosen). A static
//     `Targets` could not say it: a TargetSpec built at init() has no
//     InstanceID to exclude, and that gap is what the old caveat
//     declared.
//   - The chosen permanents leave as ONE exile ("Exile them"), through
//     AirbendAll, and each one's rebuy is granted only once it has
//     actually landed in exile. That matters most for the play this
//     card exists for — airbending your own commander out of a wrath:
//     its owner is asked about the command zone (CR 903.9), and a
//     "no" leaves it in exile castable for {2}.
//   - A trigger with no legal target is dropped before the prompt
//     (CR 603.3d, engine-wide). For a Min-0 clause the printed card
//     would instead put the ability on the stack targeting nothing,
//     which does nothing when it resolves; no card in the catalog can
//     observe the difference.
//   - "Whenever you cast a spell FROM EXILE" reads EventCast's
//     OldZone, stamped at emit time because a card on the stack no
//     longer remembers where it came from (CR 601.2a). It fires once
//     per spell cast from exile, including one Appa's own airbend put
//     there, which is the printed loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03c141ca-11e4-4927-a6bd-980ee1203c73",
		Name:            "Appa, Steadfast Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				TargetsFrom: appaOtherNonlandPermanentsYouControl,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Appa, Steadfast Guardian — airbend your permanents",
						appaAirbendTargets)
				},
			},
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
			}, "Appa, Steadfast Guardian — create a 1/1 Ally", Do(CreateToken{
				Template: TokenCard("1/1 white Ally"),
				N:        1,
			})),
		},
	})
}

// appaOtherNonlandPermanentsYouControl is the ETB's clause, "any
// number of other target nonland permanents you control", with the
// "other" bound to this Appa.
func appaOtherNonlandPermanentsYouControl(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
	return TargetPermanent(
		"any number of other target nonland permanents you control",
		Nonland(), YouControl(), NotSelf(source.InstanceID),
	).WithCount(0, 0)
}

// appaAirbendTargets airbends every target still legal on resolution
// (CR 608.2b per slot: some may have left or changed control while
// the trigger waited; the ability does as much as it can), as one
// simultaneous exile.
func appaAirbendTargets(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == item.SourceCardID {
			continue
		}
		if !g.TargetStillLegalForEffect(item, t) {
			continue
		}
		ids = append(ids, t.ID)
	}
	return AirbendAll{Targets: ids}.Apply(NewContext(g, item))
}
