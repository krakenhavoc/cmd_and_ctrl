package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bone Dragon — Creature — Dragon Skeleton {3}{B}{B}, 5/4 (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Flying
//	 {3}{B}{B}, Exile seven other cards from your graveyard: Return this card
//	 from your graveyard to the battlefield tapped."
//
// Haunted Dead's return-tapped body from the graveyard, with the exile cost
// Grim Lavamancer pays: seven cards named at announce, never the Dragon itself
// (ExileCostOptionsForEffect leaves the source out), so six other cards can't
// pay. Nothing happens if the Dragon has left the graveyard by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51926430-b98a-424d-9347-c36938132825",
		Name:            "Bone Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{B}, Exile seven other cards from your graveyard: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{B}"), effects.ExileFromGraveyard(7, "seven other cards", nil)),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := item.SourceCardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}).Apply(effects.NewContext(g, item))
			},
		}},
	})
}
