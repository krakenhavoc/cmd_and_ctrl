package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mysterious Pathlighter — Creature — Faerie {2}{W}, 2/2 (Throne of
// Eldraine, #5516):
//
//	"Flying
//	 Each creature you control that has an Adventure enters with an
//	 additional +1/+1 counter on it. (It doesn't need to have gone on the
//	 Adventure first.)"
//
// A CR 614.1c replacement over every creature entering under the
// Pathlighter's controller's control whose card has an Adventure (Bioengineered
// Future's shape). The Pathlighter has no Adventure itself, so it does not
// boost itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe87e290-8e9d-4269-ad1f-bb31e6eeb7da",
		Name:            "Mysterious Pathlighter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && entering.Controller == src.Controller && effects.CardHasAdventure(entering)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 1)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Mysterious Pathlighter: each creature you control that has an Adventure enters with an additional +1/+1 counter on it",
		}},
	})
}
