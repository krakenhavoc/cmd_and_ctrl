package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Battle of Bywater — Sorcery {1}{W}{W} (EDHREC rank 1834):
//
//	"Destroy all creatures with power 3 or greater. Then create a
//	 Food token for each creature you control. (It's an artifact with
//	 "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// The go-wide deck's one-sided wrath: the big things die and every
// small creature you kept becomes a Food. Power is the current power
// (PowerGE reads counters and anthems), and the Food count is taken
// AFTER the destruction, so a creature of yours that survived — or
// was too small to be swept — is what pays.
//
// Not the batched sweep. Each creature is destroyed through the
// single-permanent verb (Blood Money's posture) and the survivors
// are counted afterwards.
//
// The reason for that has lapsed: the mass-destroy path did not
// check indestructible (#446), so through it an indestructible 4/4
// of yours would die and pay nothing, stronger than printed one way
// and weaker the other. S30 (#470) closed it. Converting this card
// to DestroyAllMatching is now a safe follow-up and would retire the
// simultaneity caveat below with it.
//
// Sandbox simplification, declared: the creatures leave one at a
// time rather than as one simultaneous event, so a "whenever another
// creature dies" watcher that is itself in the wipe sees only the
// creatures destroyed before it. Weaker than printed for that
// watcher's controller, never stronger.
func init() {
	Register(Spec{
		OracleID:     "a94c191d-a938-458e-bc1b-2f44fd8873a3",
		Name:         "The Battle of Bywater",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The creatures are destroyed one after another rather than all at once, so a creature with a \"whenever another creature dies\" ability that is itself destroyed may miss some of the deaths."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, c := range MatchingBattlefield(ctx, And(Creature(), PowerGE(3))) {
				if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			return b17FoodPerCreatureYouControl(ctx)
		},
	})
}
