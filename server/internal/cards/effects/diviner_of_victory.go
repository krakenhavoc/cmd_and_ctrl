package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diviner of Victory // Unwind History — Creature — Dwarf Wizard {U},
// 1/1 // Sorcery {1}{U} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 Whenever you scry or surveil, this creature gets +1/+1 until end of
//	 turn."
//
//	Unwind History — "Return target creature an opponent controls with
//	 mana value 3 or less to its owner's hand. Surveil 1."
//
// One ability watching two events (EventScry, EventSurveil), each emitted
// once a scry or surveil is complete, so a scry that looked at nothing
// does not trigger it. The surveil in Unwind History is the bounce's
// continuation, since returning a commander can pause for the command-
// zone question.
//
// No simplification.
func init() {
	const id = "e7b78acd-5288-4528-97a8-a5b43de90bf5"
	Register(Spec{
		OracleID:     id,
		Name:         "Diviner of Victory",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, ByYou,
				"Diviner of Victory — gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if ctx.isNewSourceObject(item.SourceCardID) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
						Label: "Diviner of Victory — +1/+1 until end of turn"}.Apply(ctx)
				}),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Unwind History",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature an opponent controls with mana value 3 or less",
			OpponentControls(), ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return BounceToHand{Target: t.ID, Then: func(c *Context, _ bool) error {
					return Surveil{Player: c.Controller(), N: 1}.Apply(c)
				}}.Apply(ctx)
			}
			return nil
		},
	})
}
