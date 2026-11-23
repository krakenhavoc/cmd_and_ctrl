package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relentless Pursuit — Sorcery {2}{G} (Theros Beyond Death, #5514):
//
//	"Reveal the top four cards of your library. You may put a
//	 creature card and/or a land card from among them into your hand.
//	 Put the rest into your graveyard."
//
// In the Presence of Ages' take: the top four are revealed, the caster
// may keep a creature card and a land card (one card that is both fills
// one slot), and the rest go to the graveyard.
//
// No simplification.
func init() {
	const label = "Relentless Pursuit — you may put a creature card and/or a land card into your hand and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:     "d39c378b-f0a2-400f-af40-a632e24b6d72",
		Name:         "Relentless Pursuit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, "Relentless Pursuit — reveal the top four cards"),
				Optional: true,
				Label:    label,
				Key:      relentlessPursuitTake,
			}.Apply(ctx)
		},
	})
}

// relentlessPursuitTake is an on-disk key: never renamed, never reused.
var relentlessPursuitTake = effects.RegisterTake("choose-cards/relentless-pursuit-take",
	effects.TakeDef{
		Slots: []effects.TakeSlot{
			{Label: "a creature card", Match: effects.Creature(), Max: 1},
			{Label: "a land card", Match: effects.Land(), Max: 1},
		},
		Then: effects.TakeRestIntoGraveyard,
	})
