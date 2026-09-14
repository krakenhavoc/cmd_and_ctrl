package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vilis, Broker of Blood — Legendary Creature — Demon 8/8 for
// {5}{B}{B}{B} (EDHREC rank 1117):
//
//	"Flying
//	 {B}, Pay 2 life: Target creature gets -1/-1 until end of turn.
//	 Whenever you lose life, draw that many cards. (Damage causes
//	 loss of life.)"
//
// The payoff half of the life-for-cards engines, and the card that
// makes Necropotence's and Griselbrand's costs read as upside. Its
// trigger is what the reminder text says it is: EVERY life loss
// counts, not just the ones an effect calls "lose life". A Lightning
// Bolt to the face, unblocked combat damage, the 7 life Griselbrand
// charges, the 1 life Necropotence charges, a painland — all of
// them.
//
// That breadth is why the trigger watches two event kinds. The
// engine's damage paths write the life total directly and emit only
// EventDealDamage; effects that say "loses N life" emit
// EventChangeLife with a negative amount. Neither is emitted for the
// other's loss, so watching both counts each loss exactly once. The
// same two-kind read is why b04OpponentLostLife exists one seat
// over.
//
// Two orderings worth knowing, both of which fall out of the engine
// rather than needing card code:
//
//   - Vilis's own ability pays 2 life at ANNOUNCE (CR 601.2h), so
//     the draw-2 trigger goes on the stack ABOVE the -1/-1 and
//     resolves first. That is paper behaviour and it is observable:
//     the cards are in hand before the creature shrinks.
//   - The trigger fires once per loss, not once per point. Losing 7
//     to Griselbrand is one trigger that draws seven, not seven
//     triggers.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "7c209753-121b-4859-944b-4e33f885e777",
		Name:            "Vilis, Broker of Blood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Pay 2 life: Target creature gets -1/-1 until end of turn.",
			Cost:    Plus(ManaCost("{B}"), PayLife(2)),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{
					Target:    item.Targets[0].ID,
					Power:     -1,
					Toughness: -1,
					Label:     "Vilis, Broker of Blood — -1/-1",
				}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := s22PlayerLostLife(ev, source.Controller, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := s22PlayerLostLife(ev, source.Controller, g)
				return game.NewTriggeredItem(source, "Vilis, Broker of Blood — draw that many cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

// s22PlayerLostLife reports whether `ev` is `playerID` losing life,
// and how much.
//
// b04OpponentLostLife with the opponent check turned inside out, and
// kept as its own function rather than a refactor of that one: the
// batch-04 helper is depended on by eight cards whose text says
// "opponent", and widening it to take a predicate would put the
// wrong answer one argument slip away for all of them.
//
// The two event kinds are the two shapes a life loss takes in this
// engine. Damage writes the life total directly and emits only
// EventDealDamage; a "lose N life" effect, a drain, or a life
// payment emits EventChangeLife with a negative Amount. Nothing
// emits both for one loss, so nothing is double-counted.
//
// Life GAIN is an EventChangeLife with a positive amount and is
// explicitly not a loss — the sign test is the whole filter.
func s22PlayerLostLife(ev game.Event, playerID uuid.UUID, g *game.Game) (int, bool) {
	if playerID == uuid.Nil || ev.Target != playerID || g.PlayerByIDForEffect(playerID) == nil {
		return 0, false
	}
	switch ev.Kind {
	case game.EventChangeLife:
		if ev.Amount < 0 {
			return -ev.Amount, true
		}
	case game.EventDealDamage:
		if ev.Amount > 0 {
			return ev.Amount, true
		}
	}
	return 0, false
}
