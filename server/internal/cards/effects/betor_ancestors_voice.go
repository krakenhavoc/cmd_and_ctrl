package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Betor, Ancestor's Voice — Legendary Creature — Spirit Dragon
// {2}{W}{B}{G}, 3/5:
//
//	"Flying, lifelink
//	 At the beginning of your end step, put a number of +1/+1
//	 counters on up to one other target creature you control equal to
//	 the amount of life you gained this turn. Return up to one target
//	 creature card with mana value less than or equal to the amount of
//	 life you lost this turn from your graveyard to the battlefield."
//
// The commander of the #1117 life-swap deck. Both halves read off
// Game.TurnTally — LifeGained for the counters, LifeLost for the
// reanimation bound — which is the per-turn record the engine keeps
// as EventChangeLife fires (#586), not an event-log scan.
//
// # One trigger, two target clauses in two zones
//
// One printed ability with two "up to one" target clauses that point
// at different zones: a creature on the battlefield and a creature
// card in a graveyard. That is the #764 multi-clause shape — Clauses
// builds one statement whose second clause has its own zone and its
// own predicate (Devious Cover-Up mixes the stack and a graveyard the
// same way), and the trigger's target pick asks each clause in turn
// (CR 603.3d). So it is one stack item, answered once, and the two
// sentences resolve in printed order: the counters, then the return.
// Neither reads anything the other writes.
//
// "Other target creature you control" excludes Betor by name, the
// same way every other "another target creature you control" in the
// catalog does.
//
// Zero is a legal answer to both clauses. With no life gained the
// first clause still targets and places no counters; with no life
// lost the bound is 0, so only a zero-cost creature card is a legal
// target and the clause is usually answered by choosing nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "990b5e12-6e04-4832-9d64-87278f12cbda",
		Name:            "Betor, Ancestor's Voice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Betor, Ancestor's Voice — +1/+1 counters equal to the life you gained, then return a creature card within the life you lost",
					betorEndStep),
				Clauses(
					TargetCreature("up to one other target creature you control",
						YouControl(), b03NotNamed("Betor, Ancestor's Voice")).WithCount(0, 1),
					TargetCardInGraveyard("up to one target creature card in your graveyard with mana value at most the life you lost this turn",
						YouOwn(), Creature(), ManaValueAtMostLifeLostThisTurn()).WithCount(0, 1),
				),
			),
		},
	})
}

// betorEndStep resolves both sentences in printed order: the counters
// on clause 0's creature, then the return of clause 1's card. Each
// clause is re-checked on its own (CR 608.2b), so a creature that left
// in response costs the counters and nothing else.
func betorEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		// The number is read as the ability RESOLVES, so life gained
		// in response counts — the printed clause names no fixed
		// amount at announce.
		if gained := b15LifeGainedThisTurn(g, item.Controller); gained > 0 {
			if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: gained}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	return nil
}
