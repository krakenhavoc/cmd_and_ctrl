package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curzon Dax, Bound by Blood — Legendary Creature — Trill Warrior
// {2}{R}{W}, 4/3 (Star Trek Commander, #4974):
//
//	"Whenever another Klingon or Warrior you control dies, Curzon Dax deals
//	 1 damage to each opponent.
//	 When Curzon Dax dies, you may put a creature card with mana value 3 or
//	 less from your hand onto the battlefield."
//
// The first trigger reads the dead permanent's types as it last existed on
// the battlefield (CR 603.10a), so a creature that was a Klingon or a
// Warrior only through an effect counts, and it fires once per such
// creature. Curzon himself is a Warrior but the text says "another". The
// damage is Curzon's.
//
// The second is a dies trigger: the put is optional, never a cast, and
// offers only creature cards with mana value 3 or less.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "035bee4c-6f79-4123-9b6c-feba190052e8",
		Name:         "Curzon Dax, Bound by Blood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, trekB39AnotherKlingonOrWarriorYouControlDied,
				"Curzon Dax, Bound by Blood — 1 damage to each opponent",
				Do(DamageEachOpponent{N: 1})),
			WhenThisDies("Curzon Dax, Bound by Blood — you may put a creature card with mana value 3 or less from your hand onto the battlefield",
				Do(PutFromHandOntoBattlefield{
					Match:    And(Creature(), ManaValueLE(3)),
					Optional: true,
					Label:    "Curzon Dax, Bound by Blood — you may put a creature card with mana value 3 or less from your hand onto the battlefield",
				})),
		},
	})
}
