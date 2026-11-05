package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vulcan Mind-Meld — Sorcery {U} (Star Trek):
//
//	Target opponent looks at the top card of your library, then exiles
//	it face down. You look at the top two cards of their library, then
//	exile them face down. For each card exiled this way, the player who
//	exiled it may play it for as long as it remains exiled. Mana of any
//	type can be spent to cast spells this way.
//
// Two of Gonti, Night Minister's exiles (gontiExileTopForPlay): the top
// card of the caster's library goes face down to the target opponent, who
// alone may look at it and play it, and the top two of that opponent's
// library go face down to the caster. Face down (CR 406.3), so only the
// player who may play a card can examine it. Each permission lasts while
// the card stays in exile and lets mana of any type pay for it. "Play", so
// a land can be played with the holder's land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6aa204a-0be9-4a38-b54c-82538adfe7f1",
		Name:         "Vulcan Mind-Meld",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			opp, you := item.Targets[0].ID, ctx.Controller()
			if err := gontiExileTopForPlay(ctx, you, opp, 1); err != nil {
				return err
			}
			return gontiExileTopForPlay(ctx, opp, you, 2)
		},
	})
}
