package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jason Bright, Glowing Prophet — Legendary Creature — Zombie Mutant Advisor
// {2}{U}, 2/3 (Fallout):
//
//	"Whenever a Zombie or Mutant you control dies, if its power was different
//	 from its base power, draw a card.
//	 Come Fly With Me — {2}, Sacrifice a creature: Put a +1/+1 counter on
//	 target creature you control. It gains flying until end of turn."
//
// The dies trigger reads the creature as it last existed on the battlefield
// (CR 603.10a): its types, its controller, and its power against its base
// power (CR 208.4b), both stored on its departure record. Jason is a Zombie
// Mutant himself, so his own death counts. The ability's cost may sacrifice
// Jason, and then the target must be another creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c4f1f94-9916-49aa-956a-c1c6b299b24f",
		Name:         "Jason Bright, Glowing Prophet",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, gapZombieOrMutantDiedWithOddPower,
				"Jason Bright, Glowing Prophet — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Come Fly With Me — {2}, Sacrifice a creature: Put a +1/+1 counter on target creature you control. It gains flying until end of turn.",
			Cost:    Plus(ManaCost("{2}"), SacrificeACreature()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"flying"},
					Label:    "Jason Bright, Glowing Prophet — flying until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
