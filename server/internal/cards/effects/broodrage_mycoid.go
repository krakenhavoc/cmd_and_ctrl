package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodrage Mycoid — Creature — Fungus {3}{B}, 4/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"At the beginning of your end step, if you descended this turn,
//	 create a 1/1 black Fungus creature token with "This token can't
//	 block.""
//
// Enterprising Scallywag's descended end-step trigger (CR 700.11, an
// intervening if per CR 603.4, AtYourEndStepIfDescended), making Synapse
// Necromage's Fungus: the catalog's 1/1 black Fungus with its can't-block
// static (FungusCantBlockToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eeba4706-b4f2-48f9-bcb1-e42a5fc68c40",
		Name:         "Broodrage Mycoid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfDescended("Broodrage Mycoid — create a 1/1 black Fungus token that can't block",
				Do(CreateToken{Template: FungusCantBlockToken(), N: 1})),
		},
	})
}
