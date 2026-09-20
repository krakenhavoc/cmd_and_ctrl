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
// # Why this is two TriggeredAbility entries
//
// One printed ability, two "up to one" target clauses, and they point
// at different zones: a creature on the battlefield and a creature
// card in a graveyard. game.TriggeredAbility.Targets is a single
// *TargetSpec, so one declaration cannot hold both. The alternative
// considered was raising the graveyard half as a CR 603.12 reflexive
// trigger off the first, and it was rejected twice over: nothing on
// this card says "when you do", so the return is not conditional on
// the counters having happened, and a reflexive goes on the stack
// ABOVE its parent, which would resolve the printed clauses backwards.
//
// Two AtYourEndStep declarations keep each clause's own target rules
// exactly right and keep the printed order available (the controller
// orders their own simultaneous triggers, CR 603.3b). What it costs
// is the one thing a player can see, and it is declared: two stack
// items rather than one, each separately responded to. Nothing about
// the outcome changes — neither clause reads anything the other
// writes.
//
// "Other target creature you control" excludes Betor by name, the
// same way every other "another target creature you control" in the
// catalog does.
//
// Zero is a legal answer to both clauses. With no life gained the
// first ability still targets and places no counters; with no life
// lost the bound is 0, so only a zero-cost creature card is a legal
// target and the clause is usually answered by choosing nothing.
func init() {
	Register(Spec{
		OracleID:     "990b5e12-6e04-4832-9d64-87278f12cbda",
		Name:         "Betor, Ancestor's Voice",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Betor's two end-step clauses arrive as two separate triggers, so each one goes on the stack and is answered on its own.",
		},
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Betor, Ancestor's Voice — +1/+1 counters equal to the life you gained", betorCountersForLifeGained),
				TargetCreature("up to one other target creature you control",
					YouControl(), b03NotNamed("Betor, Ancestor's Voice")).WithCount(0, 1),
			),
			Targeting(
				AtYourEndStep("Betor, Ancestor's Voice — return a creature card within the life you lost", returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("up to one target creature card in your graveyard with mana value at most the life you lost this turn",
					YouOwn(), Creature(), ManaValueAtMostLifeLostThisTurn()).WithCount(0, 1),
			),
		},
	})
}

// betorCountersForLifeGained places the counters. The number is read
// as the ability RESOLVES, so life gained in response counts — the
// printed clause names no fixed amount at announce.
func betorCountersForLifeGained(g *game.Game, item *game.StackItem) error {
	gained := b15LifeGainedThisTurn(g, item.Controller)
	if gained <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: gained}.Apply(ctx)
	}
	return nil
}
