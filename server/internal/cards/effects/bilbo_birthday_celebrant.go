package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bilbo, Birthday Celebrant — Legendary Creature — Halfling Rogue {W}{B}{G},
// 2/3 (Tales of Middle-earth Commander, #5948):
//
//	"If you would gain life, you gain that much life plus 1 instead.
//	 {2}{W}{B}{G}, {T}, Exile Bilbo: Search your library for any number of
//	 creature cards, put them onto the battlefield, then shuffle. Activate
//	 only if you have 111 or more life."
//
// The first line is Angel of Vitality's life-gain replacement (CR 614.1a),
// which reaches every gain of yours, lifelink included. The ability's cost is
// mana, {T} and the exile of Bilbo himself (ExileThis), and the activation
// is refused unless the activating player has 111 or more life as it is
// activated. The search is Ugin, Eye of the Storms' unbounded search (any
// number from zero to every creature card), put onto the battlefield, then
// the library is shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5789520c-e4b2-44b2-b756-3f6b468dc55b",
		Name:         "Bilbo, Birthday Celebrant",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  LifeGainByController,
				Label:  "Bilbo, Birthday Celebrant: gain that much life plus 1",
			}.Build(),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{B}{G}, {T}, Exile Bilbo: Search your library for any number of creature cards, put them onto the battlefield, then shuffle. Activate only if you have 111 or more life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{W}{B}{G}"), TapCost(), ExileThis()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				return p != nil && p.Life >= 111
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() },
					Dest:      game.ZoneBattlefield,
					Unbounded: true,
					Shuffle:   true,
					Reason:    "Bilbo, Birthday Celebrant — any number of creature cards onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
