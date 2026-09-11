package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Konrad, the Grim — Legendary Creature — Human Knight,
// {3}{B}{B}, 5/4:
//
//	"Whenever another creature dies, or a creature card is put into a
//	 graveyard from anywhere other than the battlefield, or a creature
//	 card leaves your graveyard, Syr Konrad deals 1 damage to each
//	 opponent."
//	"{1}{B}: Each player mills a card."
//
// Three trigger conditions and a self-contained engine to feed them:
// the activated ability mills, the mill puts creature cards into
// graveyards, and that is condition two. In a graveyard deck Syr
// Konrad turns every self-mill and every reanimation into table-wide
// reach.
//
// # One event kind covers all three conditions
//
// All three clauses are card motion, and EventZoneMove is the
// engine's catch-all for card motion — draws, mills, discards and
// battlefield exits all emit one. Watching the catch-all rather than
// EventLTB + EventMill + EventDiscardCard is not a shortcut; it is
// what makes the three conditions fire EXACTLY ONCE each. A creature
// dying emits both EventZoneMove and EventLTB, so watching both
// would double the damage — which would be the #259 direction, a
// card stronger than printed.
//
// Read as predicates on (OldZone → NewZone) over a creature card:
//
//	→ graveyard, from battlefield, not Syr Konrad  "another creature dies"
//	→ graveyard, from anywhere else                "put into a graveyard from
//	                                                anywhere other than the
//	                                                battlefield"
//	graveyard →, owned by you                      "leaves your graveyard"
//
// "ANOTHER creature dies" is why the battlefield branch excludes the
// source: Syr Konrad dying does not ping. A creature card leaving
// your graveyard DOES include one going to exile, to hand, to the
// library or to the battlefield — the clause names the departure, not
// the destination.
//
// "YOUR graveyard" is resolved through the card's OWNER rather than
// through a zone-owner field on the event, because a card in a
// graveyard is always in its owner's graveyard: every route to a
// graveyard in this engine sends the card to its owner's. So
// `c.Owner == source.Controller` is the exact reading of "your", and
// a creature card leaving an OPPONENT's graveyard correctly does
// nothing.
//
// The card is looked up AFTER the move, in its new zone, where the
// printed type line is intact. That is the same posture diedCreature
// takes and it carries the same known gap: a permanent that stopped
// being a creature before it left is judged on its printed line.
//
// # The damage
//
// "Syr Konrad deals 1 damage to each opponent" is DAMAGE, sourced
// from Syr Konrad himself — so a prevention shield or a damage
// doubler applies and the source reads correctly. That is
// damageToEachOpponent, whose source is the trigger's source card.
//
// # The mill
//
// "Each player mills a card" is every player INCLUDING the
// controller, which is the point: it is how the controller feeds
// their own graveyard and their own second trigger condition. The
// resulting mills fire Syr Konrad again for every creature card
// milled, controller's and opponents' alike, which is the printed
// behaviour and the reason the card is a mana sink worth building
// around.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "14c3ff84-1e82-4606-a433-869fc52cc382",
		Name:     "Syr Konrad, the Grim",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !c.IsCreature() {
					return false
				}
				if ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard {
					// "another creature dies" / "a creature card is
					// put into a graveyard from anywhere other than
					// the battlefield" — one clause per origin, and
					// Syr Konrad's own death is excluded.
					return ev.OldZone != game.ZoneBattlefield || ev.CardID != source.InstanceID
				}
				// "a creature card leaves your graveyard".
				return ev.OldZone == game.ZoneGraveyard &&
					ev.NewZone != game.ZoneGraveyard &&
					c.Owner == source.Controller
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
				ctx := NewContext(g, item)
				for _, p := range g.Seats {
					if p == nil || p.Eliminated {
						continue
					}
					if err := (MillCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
