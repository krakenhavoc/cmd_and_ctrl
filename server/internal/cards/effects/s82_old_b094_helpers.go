package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB094PartySizeForCost is the number of creatures in the payer's party
// (CR 700.8) as a cost query reads it: the amount of "this ability costs {1}
// less to activate for each creature in your party" (Seafloor Stalker). With
// no game attached to the query it is 0.
func OldB094PartySizeForCost(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return FinB9PartySize(q.Game, q.Controller)
}
