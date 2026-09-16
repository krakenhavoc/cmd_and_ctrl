package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strix Serenade — Instant {U} (EDHREC rank 1048):
//
//	"Counter target artifact, creature, or planeswalker spell. Its
//	 controller creates a 2/2 blue Bird creature token with flying."
//
// Swan Song's other half: the three spell types Swan Song does not
// cover, the same consolation Bird. The countered spell's controller
// is read off the stack BEFORE the counter runs — CounterTarget
// deletes the StackMeta entry — and the Bird lands under them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ca3aa03-e62d-464a-9111-e754abe17f76",
		Name:         "Strix Serenade",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target artifact, creature, or planeswalker spell", Or(Artifact(), Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			stackID := item.Targets[0].ID
			tokenOwner := ctx.Controller()
			if victim := ctx.Game.StackItemForEffect(stackID); victim != nil {
				tokenOwner = victim.Controller
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: tokenOwner, Template: TokenCard("2/2 blue Bird with flying"), N: 1}.Apply(ctx)
		},
	})
}
