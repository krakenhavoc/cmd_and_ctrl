package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Share the Spoils — Enchantment {1}{R} (Forgotten Realms Commander,
// #5904):
//
//	"When this enchantment enters and whenever an opponent loses the
//	 game, exile the top card of each player's library.
//	 During each player's turn, that player may play a land or cast a
//	 spell from among cards exiled with this enchantment, and they may
//	 spend mana as though it were mana of any color to cast that spell.
//	 When they do, exile the top card of their library."
//
// The first ability is one trigger with two conditions (Gorex's shape):
// this object entering, or an opponent of its controller losing the game
// (EventPlayerEliminated, emitted before the loser's objects leave). It
// exiles the top card of the library of each player still in the game,
// each LINKED to this object (CR 607.2a, Game.ExileLinkedForEffect).
//
// The second is a gated STANDING permission over exile, derived off the
// battlefield on every query, granted to EACH player (the #5904 seam,
// CastPermission.EachPlayer), so it ends the moment Share the Spoils
// leaves and a new Share the Spoils has no claim on its earlier self's
// cards (CR 400.7; the ruling):
//
//   - "During each player's turn, that player" is TimingYourTurnOnly,
//     read against the holder; the card's own timing still applies, so a
//     land needs a main phase, an empty stack and an unused land drop
//     (the ruling).
//   - "A land or a spell", once: the follow-up records the use for that
//     player and this object (Game.NoteDoneThisTurn under a per-player
//     label, kept in the turn tally a restore point carries), and the
//     gate withholds the permission from that player for the rest of the
//     turn. One land or one spell, not one of each (the ruling).
//   - "Spend mana as though it were mana of any color" is AnyColor.
//   - "When they do" is a triggered ability of Share the Spoils created
//     by that same follow-up, for a spell cast or a land played through
//     the permission (FollowUpOnLandPlay, the #5904 seam), and put on the
//     stack under its controller (Game.QueueTriggerForPermissionPlayForEffect).
//     It exiles the top card of the playing player's library, linked to
//     this object, so that card can be played this way too.
//
// No simplification.
const (
	shareTheSpoilsName = "Share the Spoils"
	// shareTheSpoilsDoneLabel, followed by the player's ID, is the tally
	// the once-each-turn limit reads for that player.
	shareTheSpoilsDoneLabel = "Share the Spoils — played from among the exiled cards (done) "
)

// shareTheSpoilsPlayed spends the playing player's use this turn and
// creates the "when they do" trigger. An on-disk key.
var shareTheSpoilsPlayed = game.RegisterCastFollowUp("share-the-spoils-played",
	func(g *game.Game, f game.CastFollowUp) error {
		g.NoteDoneThisTurn(f.Source, shareTheSpoilsDoneLabel+f.Player.String())
		g.QueueTriggerForPermissionPlayForEffect(f, game.ReflexiveTrigger{
			Label:  "Share the Spoils — exile the top card of their library",
			Body:   shareTheSpoilsExileTheirTopBody,
			Params: game.EffectParams{Player: f.Player},
		})
		return nil
	})

// shareTheSpoilsExileTheirTopBody exiles the top card of the library of
// the player who played the card (Params.Player), linked to this Share
// the Spoils. An on-disk key.
var shareTheSpoilsExileTheirTopBody = game.ReflexiveBody("share-the-spoils/exile-their-top",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		src, _ := g.SourceObjectForEffect(item)
		return shareTheSpoilsExileTopOf(g, src, p.Player)
	}, nil)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d39af66c-4997-48b8-a704-f21d93247d7d",
		Name:         shareTheSpoilsName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventPlayerEliminated},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller
				},
				"Share the Spoils — exile the top card of each player's library",
				shareTheSpoilsExileEachTop),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: func(g *game.Game, holder, source uuid.UUID) bool {
				return g.ResolvedThisTurn(source, shareTheSpoilsDoneLabel+holder.String()) == 0
			},
			Permission: game.CastPermission{
				Zone:               game.ZoneExile,
				Filter:             game.PermissionFilter{ExiledWithSource: true},
				EachPlayer:         true,
				Timing:             game.TimingYourTurnOnly,
				AnyColor:           true,
				FollowUp:           shareTheSpoilsPlayed,
				FollowUpOnLandPlay: true,
				Label:              "During your turn, play a land or cast a spell from among cards exiled with Share the Spoils, spending mana as though it were mana of any color (once each turn)",
			},
		}},
	})
}

// shareTheSpoilsExileEachTop exiles the top card of each player's
// library, linked to this Share the Spoils. A Share the Spoils that has
// already left still exiles them; they are linked to no object then.
func shareTheSpoilsExileEachTop(g *game.Game, item *game.StackItem) error {
	src, _ := g.SourceObjectForEffect(item)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := shareTheSpoilsExileTopOf(g, src, p.ID); err != nil {
			return err
		}
	}
	return nil
}

// shareTheSpoilsExileTopOf exiles the top card of `player`'s library,
// linked to `src`. An empty library, or a player who has left, exiles
// nothing.
func shareTheSpoilsExileTopOf(g *game.Game, src game.ObjectRef, player uuid.UUID) error {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	return g.ExileLinkedForEffect(src, p.Library.Cards[len(p.Library.Cards)-1].InstanceID)
}
