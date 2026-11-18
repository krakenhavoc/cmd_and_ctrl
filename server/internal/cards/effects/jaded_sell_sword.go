package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaded Sell-Sword — Creature — Dragon Warrior {3}{R}, 4/3 (Adventures in the
// Forgotten Realms):
//
//	"When this creature enters, if mana from a Treasure was spent to cast it,
//	 it gains first strike and haste until end of turn."
//
// Hired Hexblade's intervening "if" (CR 603.4): the spell's payment rides the
// permanent (CR 400.7d, Card.ManaSpentToCast).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f666a990-bd9c-4160-9770-4758fd40ca92",
		Name:           "Jaded Sell-Sword",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ManaSpentToCast().FromTreasure()
			}), "Jaded Sell-Sword — first strike and haste until end of turn", func(g *game.Game, item *game.StackItem) error {
				if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"first strike", "haste"},
					Label:    "Jaded Sell-Sword — first strike and haste until end of turn",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
