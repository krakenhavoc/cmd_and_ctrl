package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blind Obedience — Enchantment {1}{W} (EDHREC top 1,000, #3654):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)
//	 Artifacts and creatures your opponents control enter tapped."
//
// Extort is the shared keyword trigger; the enters-tapped half is
// Kismet's CR 614 replacement narrowed to artifacts and creatures, so
// an opponent's land or enchantment is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d998c09-7d89-4265-ada4-6d80cbf56dae",
		Name:         "Blind Obedience",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Extort("Blind Obedience")},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || entering.Controller == src.Controller {
					return false
				}
				return entering.IsCreature() || entering.IsArtifact()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Blind Obedience: enter tapped",
		}},
	})
}
