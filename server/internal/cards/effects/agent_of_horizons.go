package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent of Horizons — Creature — Human Rogue {2}{G}, 3/2 (Theros, #5561):
//
//	"{2}{U}: This creature can't be blocked this turn."
//
// The restriction lasts until end of turn and lands only on the same
// object that activated it (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "668a6a9d-80fc-437b-99b8-6807e92eabca",
		Name:         "Agent of Horizons",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}"),
			Effect:  agentOfHorizonsUnblockable,
		}},
	})
}

// agentOfHorizonsUnblockable is the ability's body: nothing if the
// source has left or come back as a new object.
func agentOfHorizonsUnblockable(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
		Label: "Agent of Horizons — can't be blocked this turn"}.Apply(NewContext(g, item))
}
