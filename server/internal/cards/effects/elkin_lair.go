package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elkin Lair — World Enchantment {3}{R}:
//
//	"At the beginning of each player's upkeep, that player exiles a card
//	 at random from their hand. The player may play that card this turn.
//	 At the beginning of the next end step, if the player hasn't played
//	 the card, they put it into their graveyard."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. The
// upkeep player's card is chosen at random from their hand and exiled
// with a "play" permission for that player until end of turn, so a land
// can be played from exile too (on their own turn, with a land play
// left). Once it has landed in exile, a delayed trigger is set up for
// the next end step, which is this turn's (CR 603.7b), naming the card
// as the exiled OBJECT.
//
// "If the player hasn't played the card" is read off where the card is:
// played, it has left exile, and a card that left and came back is a new
// object (CR 400.7, CR 603.7c), so the delayed body puts the card into
// its owner's graveyard only while it is still that object in exile.
// The delayed ability is controlled by Elkin Lair's controller (CR
// 603.7d), and a card still in exile goes to the graveyard of the player
// whose card it was.
//
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters. The delayed
// trigger does not depend on Elkin Lair staying.
//
// No simplifications.
var elkinLairToGraveyardBody = game.DelayedBody("elkin-lair/unplayed-card-to-graveyard", elkinLairToGraveyard)

func init() {
	Register(Spec{
		OracleID:     "b3488947-7142-4e6c-9c92-2728469b3ad9",
		Name:         "Elkin Lair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Elkin Lair — that player exiles a card at random from their hand and may play it this turn", elkinLairUpkeep),
		},
	})
}

func elkinLairUpkeep(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	ctx := NewContext(g, item)
	pick := g.ChooseAtRandomForEffect(game.RandomDraw{Player: player, Source: ctx.Source()}, handCardIDs(g, player), 1)
	if len(pick) == 0 {
		return nil
	}
	card := pick[0]
	return g.ExileCardKeyedForEffect(card, game.KeyedTail{Ref: elkinLairExiled, Player: player, Source: ctx.Source(), Asked: game.NewAskedItem(item)})
}

// elkinLairToGraveyard puts the exiled card into its owner's graveyard
// if it is still the object Elkin Lair exiled: not played, not moved.
func elkinLairToGraveyard(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	return g.PutIntoGraveyardForEffect(id)
}

// elkinLairExiled lets the player play the card that reached exile this turn (#4253 slice 9).
var elkinLairExiled = CardTail("elkin-lair/may-play-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, item, player := ctx.Game, ctx.Item, o.Tail.Player
	exiled := len(o.IDs) == 1
	card := uuid.Nil
	if exiled {
		card = o.IDs[0]
	}

	if !exiled {
		return nil
	}
	c, ok := g.LookupCardForEffect(card)
	if !ok {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(card, game.CastPermission{
		Player:   player,
		Zone:     game.ZoneExile,
		Duration: g.UntilEndOfTurnDuration(),
	})
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Elkin Lair — if " + c.Name + " wasn't played, it goes to the graveyard",
		Cards:  []uuid.UUID{card},
		Body:   elkinLairToGraveyardBody,
		Params: game.EffectParams{Player: player, Object: game.ObjectRef{ID: card, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
})
