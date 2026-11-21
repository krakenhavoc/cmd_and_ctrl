package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ED-E, Lonesome Eyebot — Legendary Artifact Creature — Robot {3}, 2/1
// (Fallout, #5948):
//
//	"Flying
//	 ED-E My Love — Whenever you attack, if the number of attacking
//	 creatures is greater than the number of quest counters on ED-E,
//	 put a quest counter on it.
//	 {2}, Sacrifice ED-E: Draw a card, then draw an additional card for
//	 each quest counter on ED-E."
//
// "Whenever you attack" is once per declaration: the engine emits one
// attack event per creature, so the trigger waits for the settled batch
// (ThatManyEvents). The "if" is intervening (CR 603.4): it counts the
// creatures attacking right now and ED-E's quest counters, as the
// trigger is put on the stack and again as it resolves. The sacrifice is
// a cost, so the quest counters are read from the departure record (the
// counters it had as it last existed, CR 608.2h): one card, plus one for
// each, drawn together.
//
// No simplification.
func init() {
	counter := withAtBatchEnd(On(game.EventAttack,
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller)
		},
		"ED-E, Lonesome Eyebot — put a quest counter on it",
		func(g *game.Game, item *game.StackItem) error {
			return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
		}), ThatManyEvents())
	Register(Spec{
		OracleID:        "1b1bbad2-13b3-4716-85ef-3956268e23ad",
		Name:            "ED-E, Lonesome Eyebot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "the number of attacking creatures is greater than the number of quest counters on ED-E",
				Test: func(ctx *Context) bool {
					src, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
					return ok && len(AttackingCreatures(ctx.Game)) > src.Counters["quest"]
				},
			}, counter),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice ED-E: Draw a card, then draw an additional card for each quest counter on ED-E.",
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := 1
				if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
					n += info.Counters["quest"]
				}
				return DrawCards{N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
