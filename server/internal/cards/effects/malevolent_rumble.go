package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malevolent Rumble — Sorcery {1}{G} (Modern Horizons 3):
//
//	"Reveal the top four cards of your library. You may put a permanent
//	 card from among them into your hand. Put the rest into your graveyard.
//	 Create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice
//	 this token: Add {C}.""
//
// Confounding Riddle's take-one-and-bin-the-rest (takeRestGraveyard: the rest
// are put, not milled) over a REVEALED top four and a permanent-card filter,
// with the printed Eldrazi Spawn token after it. The take is a prompt that
// only queues, so the token is made on the line after it: nothing the token
// does can change which cards are in the revealed pile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d30d8968-ec24-4155-97e2-28efd52c8810",
		Name:         "Malevolent Rumble",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			const label = "Malevolent Rumble — you may put a permanent card into your hand; the rest go to your graveyard"
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 4,
				"Malevolent Rumble — revealed from the top of the library")
			if err := (TakeFromLibraryToHand{
				Player:   item.Controller,
				Cards:    revealed,
				Match:    Permanent(),
				Max:      1,
				Optional: true,
				Label:    label,
				Key:      takeRestGraveyard,
			}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Template: EldraziSpawnToken(), N: 1}.Apply(ctx)
		},
	})
}
