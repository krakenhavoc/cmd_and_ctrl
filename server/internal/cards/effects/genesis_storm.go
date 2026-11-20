package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Genesis Storm — Sorcery {4}{G}{G} (Commander 2018, #5526):
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game.
//	 Reveal cards from the top of your library until you reveal a nonland
//	 permanent card. You may put that card onto the battlefield. Then put
//	 all cards revealed this way that weren't put onto the battlefield on
//	 the bottom of your library in a random order."
//
// Echo Storm's cast trigger (echoStormCopies) over Hei Bai's reveal: the
// reveal stops at the first nonland permanent card (a token is revealed but
// never the hit), the hit is a "you may put" pick, and everything revealed
// that did not enter, the declined hit included, goes to the bottom in a
// random order. A library with no nonland permanent card reveals itself
// whole and goes to the bottom. Each copy resolves its own reveal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80e72437-0c17-44c0-b0a7-eab075a64569",
		Name:         "Genesis Storm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Genesis Storm — copy it for each time you've cast your commander from the command zone", echoStormCopies),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			run, hit := revealUntil(ctx, player, func(c game.Card) bool { return isPermanentCard(c) && !c.IsLand() }, "Genesis Storm")
			if hit == uuid.Nil {
				return ctx.Game.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, cardsStillInALibrary(ctx.Game, run))
			}
			return PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    run,
				Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == hit },
				Max:      1,
				Optional: true,
				Label:    "Genesis Storm — you may put the nonland permanent card onto the battlefield",
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
