package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hexhaven Invigorator — Creature — Chimera Horror {G}{G}{G}{G}, 6/6:
//
//	"Vigilance
//	 Whenever this creature is dealt damage, you may search your library
//	 for up to that many land cards, put them onto the battlefield
//	 tapped, then shuffle."
//
// Ill-Tempered Loner's reading of "is dealt damage": EventDealDamage
// names the damaged permanent in Target and carries the amount, which
// rides the item as Params.Amount. The search is SearchLibrary with
// Limit = the damage, Optional so the player may decline the search (and
// with it the shuffle), and the lands enter tapped.
//
// The engine emits one damage event per SOURCE, so the trigger waits for
// the damage to finish (AtBatchEnd, ThatMuchDamage) and searches once for the
// TOTAL the creature was dealt at the same time, as printed (CR 120.3, CR 510.2).
func init() {
	const label = "Hexhaven Invigorator — search for up to that many land cards, put them onto the battlefield tapped"
	Register(Spec{
		OracleID:        "12d617b1-c95b-4b1c-a586-ccfdf5898b03",
		Name:            "Hexhaven Invigorator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Key:        label,
			AtBatchEnd: ThatMuchDamage(),
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.IsLand() },
					Dest:          game.ZoneBattlefield,
					Limit:         item.Params.Amount,
					Shuffle:       true,
					TappedOnEntry: true,
					Optional:      true,
					Source:        item.SourceCardID,
					Reason:        "Hexhaven Invigorator — up to that many land cards",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
