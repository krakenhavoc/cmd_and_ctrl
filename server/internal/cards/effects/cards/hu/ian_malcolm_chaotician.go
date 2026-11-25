package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ian Malcolm, Chaotician — {1}{U}{R} Legendary Creature — Human Scientist 2/2
//
//	"Whenever a player draws their second card each turn, that player
//	 exiles the top card of their library.
//	 During each player's turn, that player may cast a spell from among the
//	 cards they don't own exiled with Ian Malcolm, and mana of any type can
//	 be spent to cast it."
//
// The trigger fires on the draw that takes that player's turn tally to
// exactly two (the shape OpponentDrewTheirSecondCardThisTurn reads), for
// any player. It exiles the top card of THAT player's library LINKED to
// this Ian Malcolm object (CR 607.2a, Game.ExileLinkedForEffect); an Ian
// that has already left still exiles it, linked to no object then.
//
// The second line is Share the Spoils' shape (#5904): a gated STANDING
// permission over exile, derived off the battlefield on every query and
// granted to EACH player (CastPermission.EachPlayer), so it ends the
// moment Ian leaves and a new Ian has no claim on its earlier self's cards
// (CR 400.7):
//
//   - "During each player's turn, that player" is TimingYourTurnOnly, read
//     against the holder; the card's own timing still applies.
//   - "The cards they don't own" is NotOwnedByHolder, read against the
//     holder, so each player casts only other players' cards.
//   - "Cast a spell" is a cast (CastOnly: a land is never played this
//     way), once each turn: the follow-up records the use for that player
//     and this object, and the gate withholds the permission from that
//     player for the rest of the turn.
//   - "Mana of any type can be spent" is AnyType.
//
// No simplification.
const ianMalcolmDoneLabel = "Ian Malcolm, Chaotician — cast from among the exiled cards (done) "

// ianMalcolmCast spends the casting player's use this turn. An on-disk key.
var ianMalcolmCast = game.RegisterCastFollowUp("ian-malcolm-chaotician-cast",
	func(g *game.Game, f game.CastFollowUp) error {
		g.NoteDoneThisTurn(f.Source, ianMalcolmDoneLabel+f.Player.String())
		return nil
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "57ff1665-0a7b-4d98-bd8f-cc07cf04f701",
		Name:         "Ian Malcolm, Chaotician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && g.TurnTallyFor(ev.Actor).CardsDrawn == 2
				},
				"Ian Malcolm, Chaotician — that player exiles the top card of their library",
				ianMalcolmExileTheirTop),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: func(g *game.Game, holder, source uuid.UUID) bool {
				return g.ResolvedThisTurn(source, ianMalcolmDoneLabel+holder.String()) == 0
			},
			Permission: game.CastPermission{
				Zone:       game.ZoneExile,
				Filter:     game.PermissionFilter{ExiledWithSource: true, NotOwnedByHolder: true},
				EachPlayer: true,
				CastOnly:   true,
				Timing:     game.TimingYourTurnOnly,
				AnyType:    true,
				FollowUp:   ianMalcolmCast,
				Label:      "During your turn, cast a spell from among the cards you don't own exiled with Ian Malcolm, and mana of any type can be spent to cast it (once each turn)",
			},
		}},
	})
}

// ianMalcolmExileTheirTop exiles the top card of the drawing player's
// library, linked to this Ian Malcolm.
func ianMalcolmExileTheirTop(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Trigger.Event.Actor)
	if p == nil || p.Eliminated || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	src, _ := g.SourceObjectForEffect(item)
	return g.ExileLinkedForEffect(src, p.Library.Cards[len(p.Library.Cards)-1].InstanceID)
}
