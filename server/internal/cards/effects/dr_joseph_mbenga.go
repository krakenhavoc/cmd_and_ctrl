package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dr. Joseph M'Benga — Legendary Creature — Human Doctor {2}{G}{G}, 3/3
// (Star Trek, #4973):
//
//	"Other creatures you control with counters on them have hexproof.
//	 If you would put one or more counters on a creature you control, put
//	 twice that many of each of those kinds of counters on that creature
//	 instead."
//
// The static is a layer-6 grant read off the permanent's counters, any
// kind, M'Benga himself excluded. The doubling is a CR 614.1a replacement
// over the counter window with Lt. Cmdr. Geordi La Forge's "if YOU would
// put" test: the placer must be M'Benga's controller, so counters an
// opponent puts on your creature are not doubled, and the creature is
// read through the layers as the counters arrive, M'Benga included (he
// is a creature you control). Counters on a player are never doubled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "046c4052-4c3c-4207-b44e-e9a98a8c017f",
		Name:         "Dr. Joseph M'Benga",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(trekB20OtherCreatureYouControlWithACounter, "hexproof"), "Other creatures you control with counters on them have hexproof."),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if !counterPlacementBy(ev, g, src.Controller, true) || ev.CounterPlayer != uuid.Nil {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.Controller == src.Controller && target.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Dr. Joseph M'Benga: twice that many counters",
		}},
	})
}
