package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oath of Druids — Enchantment {1}{G} (Exodus; Mystery Booster 2, #6100):
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player who controls more creatures than they do and is their
//	 opponent. The first player may reveal cards from the top of their
//	 library until they reveal a creature card. If the first player does,
//	 that player puts that card onto the battlefield and all other cards
//	 revealed this way into their graveyard."
//
// The Oath's controller controls the trigger, but the upkeep player
// chooses its target (OfThatPlayersChoice, The Abyss's seam), and the
// target is judged against the upkeep player: an opponent of theirs who
// controls more creatures than they do (WithMoreCreaturesThanTheChooser).
// With no such player the trigger is removed from the stack as it would
// go on it (CR 603.3d), and if the target no longer qualifies when it
// resolves, the ability does nothing (CR 608.2b). The target is only a
// condition: nothing happens to that player.
//
// On resolution the upkeep player, not the controller, is asked whether
// to reveal. If they do, cards are revealed from the top of their library
// until a creature card; it enters under their control and every other
// card revealed goes to their graveyard. If no creature card is revealed,
// the whole library is revealed and all of it goes to the graveyard.
//
// No simplification.
const oathOfDruidsName = "Oath of Druids"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf0dcb27-f02e-4065-81dc-75fbbb09ea3f",
		Name:         oathOfDruidsName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep(oathOfDruidsName+" — that player may reveal cards until a creature card and put it onto the battlefield", oathOfDruidsAsk),
			effects.WithMoreCreaturesThanTheChooser(effects.OfThatPlayersChoice(
				effects.TargetPlayer("target player who controls more creatures than they do and is their opponent"))),
		)},
	})
}

// oathOfDruidsAsk asks the upkeep player whether to reveal.
func oathOfDruidsAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event.Actor
	if first == uuid.Nil || g.PlayerByIDForEffect(first) == nil {
		return nil
	}
	return effects.MayChoice{
		Player:   first,
		Question: oathOfDruidsName + " — reveal cards from the top of your library until you reveal a creature card?",
		Then:     oathOfDruidsReveal,
	}.Apply(effects.NewContext(g, item))
}

// oathOfDruidsReveal is the upkeep player's "yes". An on-disk key: never
// renamed, never reused.
var oathOfDruidsReveal = effects.MayYes("confirm/oath-of-druids-reveal-until-creature", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	first := r.Chooser
	run, hit := effects.RevealUntil(ctx, first, func(c game.Card) bool { return c.IsCreature() }, oathOfDruidsName)
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit != uuid.Nil {
		if _, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(hit, game.LibraryEntryOptions{Controller: first}); err != nil {
			return err
		}
	}
	return effects.RestIntoGraveyard(ctx.Game, rest)
})
