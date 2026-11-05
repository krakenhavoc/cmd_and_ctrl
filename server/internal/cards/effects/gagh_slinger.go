package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gagh Slinger — Creature — Klingon Warrior {2}{R}, 3/2 (Star Trek
// Commander, #4974):
//
//	"Klingon spells and Warrior spells you cast cost {1} less to cast.
//	 Whenever another Klingon or Warrior you control enters, this
//	 creature deals 1 damage to each opponent."
//
// The discount is one cost modifier (CR 601.2f) whose predicate takes a
// spell with either type, so a Klingon Warrior spell is {1} cheaper, not
// {2}; it spends against generic mana only. The trigger is Agate
// Instigator's "another … you control enters" narrowed to the two types,
// read off the permanent that entered: a 2/1 Klingon Warrior token
// triggers it once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5697e53-eb2f-482b-8b4a-1dad165ac16f",
		Name:         "Gagh Slinger",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Klingon spells and Warrior spells you cast cost {1} less to cast.",
				YourSpell(), trekB37KlingonOrWarriorSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, trekB37AnotherKlingonOrWarriorEntered,
				"Gagh Slinger — 1 damage to each opponent", Do(DamageEachOpponent{N: 1})),
		},
	})
}
