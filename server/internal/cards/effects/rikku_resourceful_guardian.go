package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rikku, Resourceful Guardian — Legendary Creature — Human Artificer
// {2}{U}, 2/3:
//
//	"Whenever you put one or more counters on a creature, until end of
//	 turn, that creature can't be blocked by creatures your opponents
//	 control.
//	 Steal — {1}, {T}: Move a counter from target creature an opponent
//	 controls onto target creature you control. Activate only as a
//	 sorcery."
//
// The trigger is once per creature per placement
// (WheneverYouPutOneOrMoreCountersOnACreature), any creature, yours or
// not. "Creatures your opponents control" are Rikku's controller's
// opponents, so the rule is The Black Gate's per-player block bar, once
// for each opponent, pinned to that creature until end of turn: an
// opponent's creature you countered can still be blocked by yours. A
// creature that has left by the time the trigger resolves gets nothing.
//
// Steal is Nesting Grounds' move over two clauses: one counter, of the
// kind the activator picks when the first creature has several, put on
// your creature and then taken off theirs only if it landed (CR 122.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81ba75af-eddc-4344-befb-419ebfb855a3",
		Name:         "Rikku, Resourceful Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouPutOneOrMoreCountersOnACreature(nil,
				"Rikku, Resourceful Guardian — that creature can't be blocked by creatures your opponents control this turn",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.Target
					if !onBattlefield(g, id) {
						return nil
					}
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := (CantBeBlockedThisTurnByPlayer{Target: id, Player: opp,
							Label: "Rikku, Resourceful Guardian"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "Steal — {1}, {T}: Move a counter from target creature an opponent controls onto target creature you control. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Targets: Clauses(
				TargetCreature("target creature an opponent controls", OpponentControls()),
				TargetCreature("target creature you control", YouControl()),
			),
			SorcerySpeed: true,
			Effect:       nestingGroundsMove,
		}},
	})
}
