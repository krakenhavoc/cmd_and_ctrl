package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glowspore Shaman — Creature — Elf Shaman {B}{G}, 3/1 (Ravnica Remastered,
// #5948):
//
//	"When this creature enters, mill three cards. You may put a land card
//	 from your graveyard on top of your library."
//
// The mill is Eccentric Farmer's, and the pick is made after it, so a land
// just milled is a legal choice (CR 608.2c). The pick is optional and
// offered only when the graveyard holds a land card; the chosen card goes on
// top of the library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac7a9d73-dbdf-4e97-b382-67cd30ef5649",
		Name:         "Glowspore Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glowspore Shaman — mill three cards, then you may put a land card from your graveyard on top of your library",
				Do(MillCards{N: 3}, Instruction(glowsporeShamanPick))),
		},
	})
}

func glowsporeShamanPick(ctx *Context) error {
	player := ctx.Controller()
	lands := graveyardCardIDs(ctx, player, game.Card.IsLand)
	if len(lands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Glowspore Shaman — you may put a land card from your graveyard on top of your library",
		Cards:      lands,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    glowsporeShamanPicked,
	})
	return nil
}

var glowsporeShamanPicked = game.RegisterChooseCardsThen("choose-cards/glowspore-shaman-land-on-top", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	id := r.Picked[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.TuckToLibraryForEffect(id, false)
})
