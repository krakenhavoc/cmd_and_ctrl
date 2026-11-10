package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drag the Canal — Instant {U}{B} (Murders at Karlov Manor, #4469):
//
//	"Create a 2/2 white and blue Detective creature token. If a creature
//	 died this turn, you gain 2 life, surveil 2, then investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// The Detective is Museum Nightwatch's token. "A creature died this turn"
// is any creature, anyone's, read off the turn tally as the spell
// resolves (b11CreaturesDiedThisTurn, Gruesome Discovery's count). The
// steps are ResolveSteps, so the Clue waits for the surveil's answer
// (CR 608.2c) and the investigate sees the library the player chose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c0ffb66-94db-4ba9-afde-c5b4fbf09578",
		Name:         "Drag the Canal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1},
			If{Cond: dragTheCanalACreatureDied{}, Then: []Applier{
				GainLife{Amount: 2},
				Surveil{N: 2},
				Investigate{N: 1},
			}},
		),
	})
}

// dragTheCanalACreatureDied is "if a creature died this turn": any
// creature, under anyone's control. The engine's creature-died footprint
// reads only the controller's own creatures, so it declares none.
type dragTheCanalACreatureDied struct{}

func (dragTheCanalACreatureDied) Holds(ctx *Context) bool {
	return b11CreaturesDiedThisTurn(ctx.Game) > 0
}

func (dragTheCanalACreatureDied) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
