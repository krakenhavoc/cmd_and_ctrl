package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Unbinding — Enchantment — Aura Curse {6}{U}:
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, that player reveals
//	 cards from the top of their library until they reveal a creature
//	 card. Put that card onto the battlefield under your control. That
//	 player puts the rest of the revealed cards into their graveyard."
//
// The enchanted player's library is revealed (revealUntil, a token
// never stops the run, CR 108.2); the creature card enters under the
// Curse's controller's control (PutFromLibraryOntoBattlefield with the
// controller as Player and the enchanted player as LibraryOwner, the
// Lonis shape) and stays its owner's (CR 108.3). The rest go into that
// player's graveyard, not milled (a routed move). A library with no
// creature card reveals itself and all of it goes to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e79cb6c-df9a-46d1-b3c6-240297e7d47d",
		Name:         "Curse of Unbinding",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Curse of Unbinding — they reveal until a creature card; you put it onto the battlefield, the rest go to their graveyard",
				Do(Instruction(func(ctx *Context) error {
					victim := ctx.Item.Trigger.Event.Actor
					run, hit := revealUntil(ctx, victim, func(c game.Card) bool { return c.IsCreature() },
						"Curse of Unbinding — reveal until a creature card")
					return PutFromLibraryOntoBattlefield{
						Player:       ctx.Controller(),
						LibraryOwner: victim,
						Cards:        run,
						Match:        func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
						All:          true,
						ThenTail:     PutRestIntoGraveyardTail,
					}.Apply(ctx)
				}))),
		},
	})
}
