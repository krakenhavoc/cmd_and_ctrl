package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desperate Research — Sorcery {1}{B} (Invasion, S82 older sets batch 662, #5616):
//
//	"Choose a card name other than a basic land card name. Reveal the top
//	 seven cards of your library and put all of them with that name into your
//	 hand. Exile the rest."
//
// Necromentia's name choice as the spell resolves, refusing a basic land card's
// name (CR 201.4a), then the reveal of seven (fewer from a short library) and
// a take of every card with the name; the rest of the revealed cards are
// exiled, face up as revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "308b49ad-ebab-41c4-9e09-44202549bafc",
		Name:         "Desperate Research",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.QueueCardNameChoiceKeyedNotBasicForEffect(ctx.Controller(), ctx.Source(),
				"Desperate Research — choose a card name other than a basic land card name",
				ctx.ItemTail(desperateResearchNamed))
			return nil
		},
	})
}

// desperateResearchNamed reveals the top seven and takes the named cards. An
// on-disk key.
var desperateResearchNamed = effects.CardTail("then/s82-old-b662-desperate-research-named", func(ctx *effects.Context, o game.TailOutcome) error {
	name := o.Tail.Word
	if name == "" {
		return nil
	}
	player := ctx.Controller()
	revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 7,
		"Desperate Research — reveal the top seven cards of your library")
	if len(revealed) == 0 {
		return nil
	}
	return effects.TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.CardNameMatches(c, name) },
		All:    true,
		Label:  "Desperate Research — put all of them with that name into your hand",
		Key:    desperateResearchTake,
	}.Apply(effects.KeyedContext(ctx.Game, player, ctx.Source()))
})

// desperateResearchTake exiles the rest of the revealed cards. An on-disk key.
var desperateResearchTake = effects.RegisterTake("choose-cards/s82-old-b662-desperate-research-take", effects.TakeDef{
	Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if len(res.Rest) == 0 {
			return nil
		}
		return g.ExileCardsKeyedForEffect(res.Rest, game.KeyedTail{})
	},
})
