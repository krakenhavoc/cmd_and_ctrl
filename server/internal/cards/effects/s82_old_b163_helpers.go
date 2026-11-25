package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b163_helpers.go — shared pieces of Modern Horizons batch 163
// (#5519, S82).

// PutAPlusOneCounterOnThisIfStillHere is "put a +1/+1 counter on this
// creature" as an ability's whole effect (Feral Hydra, Farmstead Gleaner): it
// does nothing once the source has left or come back as a new object
// (CR 400.7).
func PutAPlusOneCounterOnThisIfStillHere(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
