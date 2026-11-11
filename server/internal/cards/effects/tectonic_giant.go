package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tectonic Giant — Creature — Elemental Giant {2}{R}{R}, 3/4 (Duskmourn
// Commander):
//
//	"Whenever this creature attacks or becomes the target of a spell an
//	 opponent controls, choose one —
//	 • This creature deals 3 damage to each opponent.
//	 • Exile the top two cards of your library. Choose one of them. Until the
//	   end of your next turn, you may play that card."
//
// One modal ability with two conditions (Appa, Loyal Sky Bison's shape):
// the source attacking, or becoming the target of a spell, not an ability,
// controlled by an opponent (Event.TargeterKind). The first mode is the
// shared damage to each opponent; the second is Strongbox Raider's exile two,
// choose one, play it through your next turn.
//
// No simplification.
func init() {
	trigger := OnAny([]game.EventKind{game.EventAttack, game.EventBecomesTarget}, cmdr2S43TectonicGiantTrigger,
		"Tectonic Giant — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("This creature deals 3 damage to each opponent.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return damageToEachOpponent(ctx.Game, item, 3)
			}),
		ModeDoing("Exile the top two cards of your library. Choose one of them. Until the end of your next turn, you may play that card.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return cmdr2S43TectonicGiantExile(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "0e5e46e3-f9af-47cc-a740-bf808d5cb4b1",
		Name:         "Tectonic Giant",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
