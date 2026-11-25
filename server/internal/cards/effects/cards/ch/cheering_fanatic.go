package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cheering Fanatic — Creature — Goblin {1}{R}, 2/2 (Battlebond):
//
//	"Whenever this creature attacks, choose a card name. Spells with the
//	 chosen name cost {1} less to cast this turn."
//
// The name is asked as the trigger resolves (The Clone Saga's prompt, CR
// 201.4). The reduction is a timed record (game.NamedSpellsCostLessMod): it
// applies to every player's casts of a spell with that name on any face, not
// only the Fanatic's controller, it is a CR 601.2f reduction of generic mana,
// and it lasts until end of turn whether or not the Fanatic does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18543bb3-14f9-45a0-8ed3-5f11832527d9",
		Name:         "Cheering Fanatic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Cheering Fanatic — choose a card name; spells with the chosen name cost {1} less to cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					g.QueueCardNameChoiceKeyedForEffect(item.Controller, item.SourceCardID,
						"Cheering Fanatic — choose a card name", effects.NewContext(g, item).ItemTail(cheeringFanaticNamed))
					return nil
				}),
		},
	})
}

// cheeringFanaticNamed is the rest of the trigger once the name (the tail's
// Word) is known. An on-disk key: never renamed, never reused.
var cheeringFanaticNamed = effects.CardTail("then/cheering-fanatic-named", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.Word == "" {
		return nil
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeGame, ctx.Controller(),
		[]game.Mod{game.NamedSpellsCostLessMod(o.Tail.Word, 1)}, ctx.Game.UntilEndOfTurnDuration(),
		"Cheering Fanatic — spells named "+o.Tail.Word+" cost {1} less to cast this turn")
	return nil
})
