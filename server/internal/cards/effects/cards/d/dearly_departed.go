package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dearly Departed — Creature — Spirit {4}{W}{W}, 5/5:
//
//	"Flying
//	 As long as this creature is in your graveyard, each Human creature
//	 you control enters with an additional +1/+1 counter on it."
//
// A replacement effect (CR 614.1c, 614.12) that functions from the
// graveyard (CR 113.6b), the seam dredge and Retriever Phoenix use:
// ReplacementEffect.FromGraveyard, which now also gathers a move onto the
// battlefield from every graveyard. "You" is the owner of the card in the
// graveyard, and only that player's Humans are replaced. Tokens enter
// through the same pipeline and get the counter too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "539be0f3-0863-4d44-87ee-5912dd003475",
		Name:            "Dearly Departed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches:       []game.EventKind{game.EventZoneMove},
			FromGraveyard: true,
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature() && entering.Controller == src.Owner &&
					entering.HasSubtype(src.TextSubtype("Human"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 1)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Owner
			},
			Label: "Dearly Departed: enters with an additional +1/+1 counter",
		}},
	})
}
