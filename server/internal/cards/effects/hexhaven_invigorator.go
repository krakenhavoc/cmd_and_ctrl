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
// DECLARED SIMPLIFICATION, weaker than printed, the same one Ill-Tempered
// Loner carries: the engine emits one damage event per SOURCE, so a
// creature damaged by two blockers at once searches twice, once per
// blocker's damage, where the printed card searches once for the total.
func init() {
	const label = "Hexhaven Invigorator — search for up to that many land cards, put them onto the battlefield tapped"
	Register(Spec{
		OracleID:        "12d617b1-c95b-4b1c-a586-ccfdf5898b03",
		Name:            "Hexhaven Invigorator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If two or more sources damage it at the same time, it searches once for each source's damage instead of once for the total."},
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     label,
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
