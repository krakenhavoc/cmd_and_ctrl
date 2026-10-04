package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scheming Silvertongue // Sign in Blood — Creature — Vampire Warlock
// {1}{B}, 1/1 // Sorcery {B}{B} (preparation card, CR 722):
//
//	"Flying, lifelink
//	 At the beginning of your second main phase, if you gained 2 or
//	 more life this turn, this creature becomes prepared. (While it's
//	 prepared, you may cast a copy of its spell. Doing so unprepares
//	 it.)"
//
//	Sign in Blood — "Target player draws two cards and loses 2 life."
//
// A preparation card (ADR 0090): the creature registers under the bare
// oracle ID and its prepare spell under "<oracle>#1". The prepare spell
// is Sign in Blood word for word, so it shares that card's body.
//
// "Your second main phase" is the postcombat main phase, and the
// "if you gained 2 or more life this turn" clause is an intervening if
// (CR 603.4): asked as the phase begins, so no life gain means no
// trigger, and again as the ability resolves. Lifelink damage counts,
// because it is life gain (CR 702.15b). A Silvertongue that is already
// prepared gains nothing (CR 722.3a).
//
// No simplification.
const schemingSilvertongueOracleID = "44443716-8356-40cd-a879-35264b29108c"

func init() {
	Register(Spec{
		OracleID:        schemingSilvertongueOracleID,
		Name:            "Scheming Silvertongue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepPostcombatMain, true), silvertongueGainedTwo),
				"Scheming Silvertongue — becomes prepared", silvertongueBecomePrepared),
		},
	})
	Register(Spec{
		OracleID:     schemingSilvertongueOracleID + "#1",
		Name:         "Sign in Blood",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    signInBloodOnResolve,
	})
}

// silvertongueGainedTwo is the intervening if as a trigger predicate.
func silvertongueGainedTwo(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b15LifeGainedThisTurn(g, source.Controller) >= 2
}

// silvertongueBecomePrepared re-checks the life gain as the ability
// resolves (CR 603.4), then makes the creature prepared.
func silvertongueBecomePrepared(g *game.Game, item *game.StackItem) error {
	if b15LifeGainedThisTurn(g, item.Controller) < 2 {
		return nil
	}
	return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
