package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Painter's Studio // Defaced Gallery — Enchantment — Room (ADR 0103):
//
//	Painter's Studio {2}{R}: "When you unlock this door, exile the top
//	  two cards of your library. You may play them until the end of your
//	  next turn."
//	Defaced Gallery {1}{R}: "Whenever you attack, attacking creatures
//	  you control get +1/+0 until end of turn."
//
// Painter's Studio is Reckless Impulse's exile (ExileTopNUntilYourNextTurn,
// ADR 0063's duration). Defaced Gallery is one trigger per attack
// declaration (OncePerBatch, Adeline's shape) and pumps the creatures
// attacking when it resolves (BoostUntilEOT reads the set once,
// CR 611.2c).
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "2af01e09-f233-4675-89cd-470816471c62",
		Name:         "Painter's Studio // Defaced Gallery",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Painter's Studio — exile the top two cards; play them until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextTurn(NewContext(g, item), 2)
				}),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Defaced Gallery — attacking creatures you control get +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() && c.Controller == item.Controller && c.AttackingTarget != uuid.Nil
						},
						Power: 1,
						Label: "Defaced Gallery — +1/+0",
					}.Apply(NewContext(g, item))
				})),
		}},
	}))
}
