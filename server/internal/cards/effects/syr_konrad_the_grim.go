package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Konrad, the Grim — Legendary Creature — Human Knight {3}{B}{B},
// 5/4 (EDHREC rank 259):
//
//	"Whenever another creature dies, or a creature card is put into a
//	 graveyard from anywhere other than the battlefield, or a creature
//	 card leaves your graveyard, Syr Konrad deals 1 damage to each
//	 opponent.
//	 {1}{B}: Each player mills a card."
//
// One triggered ability with three conditions (CR 603.2c — one
// trigger per event, never more), so one TriggeredAbility watching
// four event kinds:
//
//   - "another creature dies" — EventLTB into a graveyard, any
//     controller, the Knight itself excluded.
//   - "a creature card is put into a graveyard from anywhere other
//     than the battlefield" — a discard (EventDiscardCard), a mill
//     (EventMill), or a creature spell countered or otherwise
//     leaving the stack for a graveyard (EventZoneMove). Each path
//     emits exactly one of those, so nothing is counted twice; a
//     death's own ZoneMove is FROM the battlefield and excluded.
//   - "a creature card leaves your graveyard" — EventZoneMove out of
//     a graveyard, for a card its controller OWNS. Reanimation,
//     Regrowth, a Bojuka Bog on your own pile: all count, per card.
//
// The damage is dealt by the Knight, so it is noncombat damage from
// a black source. The mill ability feeds all three clauses at once
// — every player mills, and every creature card milled pings the
// table — which is the card's whole engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "14c3ff84-1e82-4606-a433-869fc52cc382",
		Name:     "Syr Konrad, the Grim",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB, game.EventDiscardCard, game.EventMill, game.EventZoneMove},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventLTB {
					if ev.CardID == source.InstanceID {
						return false // "another"
					}
					_, died := diedCreature(ev, g)
					return died
				}
				return b02CreatureCardEnteredGraveyardNotFromBattlefield(ev, g) ||
					b02CreatureCardLeftYourGraveyard(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Syr Konrad, the Grim — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}: Each player mills a card.",
			Cost:  ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b02EachPlayerMills(g, item, 1)
			},
		}},
	})
}
