package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tinybones, Bauble Burglar — Legendary Creature — Skeleton Rogue {1}{B}, 1/3:
//
//	"Whenever an opponent discards a card, exile it from their graveyard
//	with a stash counter on it.
//	During your turn, you may play cards you don't own with stash
//	counters on them from exile, and mana of any type can be spent to
//	cast those spells.
//	{3}{B}, {T}: Each opponent discards a card. Activate only as a
//	sorcery."
//
// The marker is an ordinary counter on the exiled card (CR 122.1), and
// it goes where counters on exiled cards always go: MoveCard clears it
// when the card leaves exile, so a stashed card that is cast, returned
// or flickered is no longer stashed.
//
// The second line is a STANDING permission (#2179), declared with
// Spec.CastPermissions and read off the battlefield on every query, never
// stored. Its filter says "with a stash counter" (WithCounter) and "you
// don't own" (NotOwnedByHolder), its timing is "during your turn"
// (TimingYourTurnOnly, which leaves the card's own timing in force, so a
// creature is still a sorcery-speed cast), and AnyType is "mana of any
// type". Because nothing is stored, it covers cards stashed while
// Tinybones was somewhere else, covers cards stashed by ANOTHER
// Tinybones, and ends the moment this one leaves. Lands are playable
// (CastOnly is not set).
//
// The discard trigger exiles the card only if it is still in a graveyard
// when the trigger resolves (a madness card or a card that was returned
// in response is not "it" any more).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd335b2f-e6a2-45ac-949e-a67278b21cb6",
		Name:         "Tinybones, Bauble Burglar",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneExile,
			Filter: game.PermissionFilter{
				NotOwnedByHolder: true,
				WithCounter:      tinybonesStashCounter,
			},
			AnyType: true,
			Timing:  game.TimingYourTurnOnly,
			Label:   "Play a card you don't own with a stash counter from exile (Tinybones, Bauble Burglar)",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b18OpponentDiscarded(ev, source)
			},
			Key:    "Tinybones, Bauble Burglar — exile the discarded card with a stash counter",
			Effect: tinybonesStashDiscarded,
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}, {T}: Each opponent discards a card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{B}"), TapCost()),
			SorcerySpeed: true,
			Effect:       eachOpponentDiscardsOne,
		}},
	})
}

const tinybonesStashCounter = "stash"

// tinybonesStashDiscarded is the trigger's resolution: the discarded card
// leaves the graveyard it is in for exile and carries a stash counter.
// The counter goes on once the card has actually landed in exile (the
// move can pause on a CR 903.9 prompt for a commander).
func tinybonesStashDiscarded(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if id == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{
		Target: id,
		Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			return ctx.Game.AddCounterForEffect(id, tinybonesStashCounter, 1)
		},
	}.Apply(NewContext(g, item))
}
