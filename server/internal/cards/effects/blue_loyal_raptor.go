package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blue, Loyal Raptor — Legendary Creature — Dinosaur {2}{G}{U}, 5/4 (Jurassic
// World Collection, #5948):
//
//	"Partner with Owen Grady, Raptor Trainer
//	 For each kind of counter on Blue, Loyal Raptor, each other Dinosaur you
//	 control enters with a counter of that kind on it."
//
// Partner with is the shared entry trigger (PartnerWith). The second ability
// is a CR 614 entry replacement on another Dinosaur entering under Blue's
// controller's control: one counter of every kind on Blue at that moment, in
// a fixed order (the kinds sorted by name); with none it adds nothing.
// Counter doublers apply to the counters, since they ride the entry event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f13ed2e-7790-499f-87bf-07480b567811",
		Name:         "Blue, Loyal Raptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Blue, Loyal Raptor", "Owen Grady, Raptor Trainer"),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if !bnsB13OtherCreatureYouControlEnters(ev, g, src) {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.HasSubtype(src.TextSubtype("Dinosaur"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				kinds := make([]string, 0, len(src.Counters))
				for kind, n := range src.Counters {
					if n > 0 {
						kinds = append(kinds, kind)
					}
				}
				sort.Strings(kinds)
				for _, kind := range kinds {
					ev.AddCounterAtETB(kind, 1)
				}
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Blue, Loyal Raptor: each other Dinosaur you control enters with a counter of each kind on this creature",
		}},
	})
}
