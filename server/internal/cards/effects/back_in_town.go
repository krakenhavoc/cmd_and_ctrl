package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Back in Town — Sorcery {X}{2}{B} (Outlaws of Thunder Junction Commander,
// #5368):
//
//	"Return X target outlaw creature cards from your graveyard to the
//	 battlefield. (Assassins, Mercenaries, Pirates, Rogues, and Warlocks
//	 are outlaws.)"
//
// Return to the Ranks' shape: exactly X targets (CountFromX), chosen as the
// spell is cast, and every one still legal as it resolves comes back
// together under your control (CR 608.2b, CR 603.6a). "Outlaw" is the
// five-subtype batch (b40Outlaw), read off the card in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce723db6-2a80-42a5-a176-cd322b4e1e27",
		Name:         "Back in Town",
		Completeness: CompletenessFull,
		Targets: targetsCountedByX(TargetCardInGraveyard("X target outlaw creature cards from your graveyard",
			YouOwn(), Creature(), b40Outlaw())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets:    legalTargetCardIDs(ctx),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
