package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristlebane Outrider — Creature — Kithkin Knight {3}{G}, 3/5 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 As long as another creature entered the battlefield under your
//	 control this turn, this creature gets +2/+0."
//
// The block restriction is Cold-Blooded Crew's, judged as blockers are
// declared (CR 509.1b) on each would-be blocker's current power. The pump
// is a layer 7c static read from the turn's entry tally (Kinbinding's
// EnteredWithCardTypeThisTurn), so a creature that entered and has since
// left still counts, and one that entered under an opponent's control
// does not. "Another" is Galadriel, Elven-Queen's reading: when the
// Outrider itself entered this turn it needs a second entry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48f7217e-ae11-44bd-ac42-7c8d9dd5dc36",
		Name:         "Bristlebane Outrider",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Static: []game.StaticAbility{{
			Label:    "As long as another creature entered the battlefield under your control this turn, this creature gets +2/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && bristlebaneOutriderAnotherCreatureEntered(g, source)
			},
			Apply: pumpBy(2, 0),
		}},
	})
}

// bristlebaneOutriderAnotherCreatureEntered is "another creature entered
// the battlefield under your control this turn": the turn's creature entry
// tally for the source's controller, needing one more entry when the
// source itself entered this turn.
func bristlebaneOutriderAnotherCreatureEntered(g *game.Game, source *game.Card) bool {
	need := 1
	if g.EnteredThisTurn(source.InstanceID) {
		need = 2
	}
	return g.EnteredWithCardTypeThisTurn(source.Controller, "creature") >= need
}
