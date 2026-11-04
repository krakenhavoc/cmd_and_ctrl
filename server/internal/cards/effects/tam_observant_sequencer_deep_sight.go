package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tam, Observant Sequencer // Deep Sight — Legendary Creature — Gorgon
// Wizard {2}{G}{U}, 4/3, with the prepare spell Deep Sight, Sorcery {G}{U}
// (Secrets of Strixhaven, #4456; preparation card, CR 722):
//
//	"Landfall — Whenever a land you control enters, Tam becomes prepared.
//	 (While it's prepared, you may cast a copy of its spell. Doing so
//	 unprepares it.)"
//
//	Deep Sight: "You draw a card and gain 1 life."
//
// A preparation card: the permanent under the bare oracle ID and its prepare
// spell under "#1". Landfall is an ability word; the trigger is one
// "whenever a land you control enters" and makes Tam prepared (BecomePrepared,
// CR 722.3a), which does nothing to a Tam that is already prepared or that has
// left the battlefield. The engine makes the copy of the spell in exile as Tam
// becomes prepared and unprepares it as the copy is cast.
//
// Deep Sight draws a card and then gains 1 life.
//
// No simplification.
func init() {
	const oracle = "702e871d-90d8-4468-8f69-5ae42af2c9d3"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Tam, Observant Sequencer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Tam, Observant Sequencer — Landfall: Tam becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Deep Sight",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: 1}.Apply(ctx)
		},
	})
}
