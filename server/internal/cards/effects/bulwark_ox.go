package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bulwark Ox — Creature — Ox Mount {1}{W}:
//
//	"Whenever this creature attacks while saddled, put a +1/+1 counter
//	 on target creature.
//	 Sacrifice this creature: Creatures you control with counters on
//	 them gain hexproof and indestructible until end of turn.
//	 Saddle 1"
//
// The set of creatures that gain the keywords is fixed when the ability
// resolves (CR 611.2c): a creature that gets a counter afterwards in the
// turn does not join it. The Ox is already in the graveyard by then, so
// it is not counted — it was sacrificed as the cost.
//
// No simplification.
func init() {
	// b36CounterOnChosenAnimal is the shared body: a +1/+1 counter on the
	// announced target, if it is still legal.
	attack := AttacksWhileSaddled("Bulwark Ox — put a +1/+1 counter on target creature", b36CounterOnChosenAnimal)
	attack.Targets = TargetCreature("target creature")

	Register(Spec{
		OracleID:     "bcd56fbf-fde3-41c2-83f5-9636e92d2893",
		Name:         "Bulwark Ox",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice this creature: Creatures you control with counters on them gain hexproof and indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    SacrificeThis(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl(), b64HasAnyCounter()),
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Bulwark Ox — hexproof and indestructible until end of turn",
					}.Apply(NewContext(g, item))
				},
			},
			Saddle(1),
		},
		Triggered: []game.TriggeredAbility{attack},
	})
}
