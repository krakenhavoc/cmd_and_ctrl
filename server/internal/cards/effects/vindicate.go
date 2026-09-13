package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vindicate — Sorcery {1}{W}{B} (EDHREC rank 1158):
//
//	"Destroy target permanent."
//
// The original answer-anything. Beast Within without the Beast,
// Anguished Unmaking without the exile and the life; any permanent
// at all, lands included. Indestructible is honoured by the destroy
// path (#380).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63c1ac21-e3d8-40c2-8c09-3f31c52992ef",
		Name:         "Vindicate",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
