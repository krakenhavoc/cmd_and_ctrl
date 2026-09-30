package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metallic Mimic — Artifact Creature — Shapeshifter {2}, 2/1
// (slice 296-m):
//
//	"As this creature enters, choose a creature type.
//	 This creature is the chosen type in addition to its other types.
//	 Each other creature you control of the chosen type enters with an
//	 additional +1/+1 counter on it."
//
// The first two lines are Adaptive Automaton's own construction
// (ChooseCreatureTypeAsEnters + IsAlsoTheChosenType, CR 614.12's
// named-tribe prompt and the layer-4 type add).
//
// The third line is a CR 614 self-replacement watching an OTHER
// creature's entry, Kismet's shape: RepEventMove / ZoneBattlefield,
// the entering card looked up with LookupCardForEffect (pre-push, so
// printed characteristics), matched against TribeFilter{Chosen: true,
// Others: true, YoursOnly: true} — "each other creature you control
// of the chosen type" is exactly that filter's own reading of
// NamedTribe. The bonus counter rides ev.AddCounterAtETB, so
// Doubling Season still doubles it and a Hardened Scales still adds
// to it, the same as any other entry-counters clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f91297a-ec2b-4ea8-9198-aa1daac20ff8",
		Name:         "Metallic Mimic",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Metallic Mimic"),
		Static: []game.StaticAbility{
			IsAlsoTheChosenType(),
		},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
						return false
					}
					entering, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return false
					}
					return (TribeFilter{Chosen: true, Others: true, YoursOnly: true}).Matches(&entering, g, src)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.AddCounterAtETB(game.CounterPlusOne, 1)
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Metallic Mimic: an additional +1/+1 counter",
			},
		},
	})
}
