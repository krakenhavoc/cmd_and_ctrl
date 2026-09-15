package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodchief Ascension — Enchantment {B} (EDHREC rank 754):
//
//	"At the beginning of each end step, if an opponent lost 2 or more
//	 life this turn, you may put a quest counter on this enchantment.
//	 (Damage causes loss of life.)
//	 Whenever a card is put into an opponent's graveyard from
//	 anywhere, if this enchantment has three or more quest counters
//	 on it, you may have that player lose 2 life. If you do, you gain
//	 2 life."
//
// The one-mana enchantment that turns a mill deck or a wheel into a
// kill. Two triggers:
//
//   - The end-step one is EventBeginEndStep on EVERY player's turn
//     ("each end step"), with the intervening-if read off the event
//     log: b04OpponentLostLife already knows the two shapes a life
//     loss takes (a negative EventChangeLife, or an EventDealDamage
//     to a player, which writes the life total directly and emits no
//     EventChangeLife), and b06AnOpponentLostAtLeastThisTurn sums them per
//     player back to the current turn's upkeep. The counter goes on
//     through AddCounter, so Doubling Season applies.
//   - The graveyard one watches the four event kinds that put a card
//     into a graveyard — EventZoneMove (a permanent dying, a spell
//     resolving or fizzling, a tutor to the yard), EventDiscardCard,
//     EventMill and EventCounterSpell (a countered spell emits only
//     that) — and then asks the one question that unifies them: is
//     the card, right now, in a graveyard whose owner is an
//     opponent? Every path emits exactly one of the four for one
//     card, so nothing fires twice. "That player" is the graveyard's
//     owner, captured by value when the trigger is built.
//
// Both intervening-ifs are checked when the trigger would fire, the
// posture every intervening-if card in the catalog takes. For the
// first that is exact — life lost this turn cannot be un-lost. For
// the second the counters could be removed in response, which no
// catalog card can do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5ea905b-4bb6-48d0-9082-c472703db550",
		Name:         "Bloodchief Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b06AnOpponentLostAtLeastThisTurn(g, source.Controller, 2)
			}, "Bloodchief Ascension — put a quest counter on it", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
			}), "Bloodchief Ascension — put a quest counter on it?"),
			{
				Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if source.Counters["quest"] < 3 {
						return false
					}
					_, ok := b06CardInOpponentsGraveyard(ev, source.Controller, g)
					return ok
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Bloodchief Ascension — have that player lose 2 life? (You gain 2 life.)"},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					victim, _ := b06CardInOpponentsGraveyard(ev, source.Controller, g)
					return game.NewTriggeredItem(source, "Bloodchief Ascension — that player loses 2 life and you gain 2 life",
						func(g *game.Game, item *game.StackItem) error {
							p := g.PlayerByIDForEffect(victim)
							if p == nil || p.Eliminated {
								return nil
							}
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -2); err != nil {
								return err
							}
							return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}

// b06CardInOpponentsGraveyard reports whether ev names a card that is
// now in the graveyard of an opponent of `controller`, and whose
// graveyard it is. The zone is read at event time, and every
// graveyard-entry path emits its event AFTER the move, so the answer
// is current.
func b06CardInOpponentsGraveyard(ev game.Event, controller uuid.UUID, g *game.Game) (uuid.UUID, bool) {
	if ev.CardID == uuid.Nil {
		return uuid.Nil, false
	}
	if ev.Kind == game.EventZoneMove && ev.NewZone != game.ZoneGraveyard {
		return uuid.Nil, false
	}
	z := g.FindCardZoneForEffect(ev.CardID)
	if z == nil || z.Kind != game.ZoneGraveyard || z.Owner == uuid.Nil || z.Owner == controller {
		return uuid.Nil, false
	}
	if g.PlayerByIDForEffect(z.Owner) == nil {
		return uuid.Nil, false
	}
	return z.Owner, true
}
