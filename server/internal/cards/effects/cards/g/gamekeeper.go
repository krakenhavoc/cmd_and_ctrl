package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gamekeeper — Creature — Elf {3}{G}, 2/2 (Dominaria Remastered):
//
//	"When this creature dies, you may exile it. If you do, reveal cards
//	 from the top of your library until you reveal a creature card. Put
//	 that card onto the battlefield and put all other cards revealed this
//	 way into your graveyard."
//
// A "you may" dies trigger (CR 603.5). Exiling is attempted only while the
// Gamekeeper is still the card that died in the graveyard (CR 400.7);
// "if you do" is whether it really reached exile (ExileTarget's tail), so
// a Gamekeeper that was moved elsewhere in response reveals nothing. The
// reveal is RevealUntilThenPutOntoBattlefield's, with the rest going to
// the graveyard instead of the library's bottom. With no creature card
// in the library every card is revealed and all of them go to the
// graveyard (the printed outcome).
//
// No simplification.
func init() {
	const label = "Gamekeeper — you may exile it; if you do, reveal until a creature card, put it onto the battlefield and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:     "c6921cfe-48f9-4378-b145-d9b430c2aaf6",
		Name:         "Gamekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				dead := item.Trigger.Event.CardID
				if !effects.DiedCardStillInGraveyard(ctx, dead) {
					return nil
				}
				return effects.ExileTarget{Target: dead, ThenTail: ctx.ItemTail(gamekeeperExiled)}.Apply(ctx)
			}), "Gamekeeper — exile it?"),
		},
	})
}

// gamekeeperExiled is "if you do": the reveal, once the Gamekeeper really
// reached exile. An on-disk key.
var gamekeeperExiled = effects.CardTail("gamekeeper/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	player := ctx.Controller()
	run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		"Gamekeeper — revealed until a creature card")
	return effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: effects.PutRestIntoGraveyardTail,
	}.Apply(ctx)
})
