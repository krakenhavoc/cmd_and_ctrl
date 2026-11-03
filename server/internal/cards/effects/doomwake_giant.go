package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomwake Giant — Enchantment Creature — Giant {4}{B}, 4/6 (Secrets of
// Strixhaven Commander reprint, #4834):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, creatures your opponents control get -1/-1 until
//	 end of turn."
//
// Constellation is Grim Guardian's condition (the Giant's own entry, or
// another enchantment entering under your control). The creatures are
// the ones your opponents control as the trigger resolves, locked then
// (CR 611.2c): one that enters later in the turn is not shrunk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba069664-9f92-47c7-b556-9e5143a22579",
		Name:         "Doomwake Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
			}, "Doomwake Giant — creatures your opponents control get -1/-1 until end of turn (constellation)",
				Do(BoostUntilEOT{
					Match:     And(Creature(), OpponentControls()),
					Power:     -1,
					Toughness: -1,
					Label:     "Doomwake Giant — -1/-1 until end of turn",
				})),
		},
	})
}
