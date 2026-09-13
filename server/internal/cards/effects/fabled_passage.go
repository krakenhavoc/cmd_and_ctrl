package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fabled Passage — Land (EDHREC rank 46):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	card, put it onto the battlefield tapped, then shuffle. Then if
//	you control four or more lands, untap that land."
//
// Evolving Wilds that stops costing you a turn once the game is
// going, which is why it is played over Wilds and Terramorphic
// Expanse in almost every deck that can spare the slot.
//
// The activation cost is Evolving Wilds' — tap and sacrifice, no
// life — so it reuses nothing from fetchlandCost(), which pays a
// life. The land count is taken AFTER the fetch, so the newly
// arrived land counts toward the four; that is the printed reading
// ("Then if you control four or more lands") and it is why a Passage
// cracked with three lands already out comes in untapped.
//
// S22: "that land" is whatever the SEARCHER chose, delivered by the
// search's Then continuation. This card used to peek at the library
// and re-derive the first-match pick before searching, because the
// primitive returned no handle on what it moved and the pick was
// deterministic. Neither half of that is true any more — the player
// picks, and the pick may arrive several client round-trips later —
// so the untap runs inside the continuation instead.
func init() {
	Register(Spec{
		OracleID:     "0c85b8f7-0bd0-4680-9ec5-d4b110460a54",
		Name:         "Fabled Passage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fabledPassageFetch,
		}},
	})
}

func fabledPassageFetch(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return SearchLibrary{
		Player:        controller,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Fabled Passage — a basic land",
		Then: func(g *game.Game, found []uuid.UUID) error {
			// A whiffed or declined search still cost the land; there
			// is nothing to untap.
			if len(found) == 0 {
				return nil
			}
			lands := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsLand() && c.Controller == controller {
					lands++
				}
			}
			if lands < 4 {
				return nil
			}
			return g.UntapTargetForEffect(found[0])
		},
	}.Apply(NewContext(g, item))
}
