package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crib Swap — Kindred Instant — Shapeshifter {2}{W} (Lorwyn Eclipsed,
// #4458; first printed in Lorwyn):
//
//	"Changeling (This card is every creature type.)
//	 Exile target creature. Its controller creates a 1/1 colorless
//	 Shapeshifter creature token with changeling."
//
// Resculpt's shape: the token goes to the exiled creature's controller,
// read before the exile, and "its controller" is honoured for your own
// creature as for an opponent's. With the target gone by resolution the
// spell does nothing and no token is made. Changeling is printed and
// honoured in every zone (Nameless Inversion); the token carries the
// keyword on its template (Springleaf Parade's 1/1 Shapeshifter).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2987c385-011a-4032-a516-a46d1e9dc9e8",
		Name:            "Crib Swap",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         TargetCreature("target creature"),
		Purpose:         ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: controller,
				Template:   TokenCard("1/1 colorless Shapeshifter with changeling"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
