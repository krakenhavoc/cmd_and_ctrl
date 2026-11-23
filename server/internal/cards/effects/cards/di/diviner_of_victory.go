package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Diviner of Victory",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, effects.ByYou,
				"Diviner of Victory — gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.IsNewSourceObject(item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
						Label: "Diviner of Victory — +1/+1 until end of turn"}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Unwind History",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature an opponent controls with mana value 3 or less",
			effects.OpponentControls(), effects.ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.BounceToHand{Target: t.ID, ThenTail: ctx.Tail(unwindHistorySurveil)}.Apply(ctx)
			}
			return nil
		},
	})
}

// unwindHistorySurveil is "surveil 1" after the bounce, whether or not the
// creature reached a hand, keyed (#4253): a table waiting on the bounce's
// prompt is a restore point. An on-disk key: never renamed, never reused.
var unwindHistorySurveil = effects.CardTail("then/unwind-history-surveil", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
