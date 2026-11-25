package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b292_helpers.go — shared body for Shadows over Innistrad batch 292.

// PutThisFromYourGraveyardOntoTheBattlefieldTransformed is "put this card
// from your graveyard onto the battlefield transformed": the activated
// ability's whole effect (Startled Awake). A card that left the graveyard
// before the ability resolves stays where it went (CR 400.7).
func PutThisFromYourGraveyardOntoTheBattlefieldTransformed(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{
		Target: item.SourceCardID, Dest: game.ZoneBattlefield, Transformed: true,
	}.Apply(NewContext(g, item))
}
