package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broken Bond — Sorcery {1}{G} (EDHREC rank 2234):
//
//	"Destroy target artifact or enchantment. You may put a land card
//	 from your hand onto the battlefield."
//
// Naturalize at sorcery speed with a free land drop attached: a
// single target, re-checked at resolution, then the shared
// MayPutALandFromHand offer (#654) — a pick-from-hand prompt whose
// continuation puts the land onto the battlefield through the CR 614
// entry pipeline. Putting a land onto the battlefield is not playing
// one (CR 305.4), so the turn's land drop survives it.
//
// The offer sits INSIDE the legal-target branch. Broken Bond's only
// target is its target: if it has become illegal the spell does not
// resolve at all (CR 608.2b) and none of its effects happen, land
// drop included. A player who kills the artifact in response gets the
// whole spell countered, not a free land.
//
// Until #654 the clause was omitted and declared: the prompt existed
// (#552) but the hand-to-battlefield move did not.
func init() {
	Register(Spec{
		OracleID:     "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3",
		Name:         "Broken Bond",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return MayPutALandFromHand("Broken Bond").Apply(ctx)
		},
	})
}
