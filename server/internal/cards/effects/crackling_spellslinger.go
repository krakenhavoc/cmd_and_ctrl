package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crackling Spellslinger — Creature — Human Wizard {3}{R}{R}, 2/2:
//
//	"Flash
//	 When this creature enters, if you cast it, the next instant or sorcery
//	 spell you cast this turn has storm. (When you cast that spell, copy it
//	 for each spell cast before it this turn. You may choose new targets for
//	 the copies.)"
//
// "If you cast it" is an intervening if (CR 603.4): the entry came from the
// stack and its caster controls it (b16EnteredFromStack), checked as the
// trigger fires and again as it resolves, so a reanimated or blinked
// Spellslinger does nothing. The grant is Storm, Force of Nature's watch
// (CR 603.7b, WhenYouNextCast with the storm body): it fires on the next
// instant or sorcery only, a creature spell leaves it waiting, and it ends
// at cleanup. Storm's copies come from stormCopies off the recorded cast
// order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ed7ba558-5341-4c7b-a9c5-b382dee88a13",
		Name:            "Crackling Spellslinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b16EnteredFromStack(g, source.InstanceID)
			}, "Crackling Spellslinger — the next instant or sorcery spell you cast this turn has storm",
				func(g *game.Game, item *game.StackItem) error {
					if !b16EnteredFromStack(g, item.SourceCardID) {
						return nil
					}
					return WhenYouNextCast(
						"Crackling Spellslinger — storm: copy that spell for each spell cast before it",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						stormTheSpellBody,
					).Apply(NewContext(g, item))
				}),
		},
	})
}
