package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reap Intellect — Sorcery {X}{2}{U}{B} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Target opponent reveals their hand. You choose up to X nonland cards from it
//	 and exile them. For each card exiled this way, search that player's graveyard,
//	 hand, and library for any number of cards with the same name as that card and
//	 exile them. Then that player shuffles."
//
// Agonizing Remorse's reveal and choice of a nonland card, for up to X of them
// (the pick is any number from none to X, S82's UpTo pick), exiled as the pick is
// answered. The searches wait for that pick (CR 608.2c) and are one search over
// the names of every card exiled: Eradicate's extraction (the caster searches the
// graveyard, hand and library in one prompt and may leave matches in the hidden
// zones), whose found set is the union of what each card's own search would
// find. The library is shuffled whatever was taken, and also when nothing was
// exiled, as printed. With X zero the hand is revealed and the library shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa97eb3d-8b8d-418b-8615-839812d60fe0",
		Name:         "Reap Intellect",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			x := ctx.X()
			if x <= 0 {
				ctx.Game.RevealHandForEffect(victim)
				return effects.ShuffleLibrary{Player: victim}.Apply(ctx)
			}
			return effects.ChooseFromRevealedHand{
				Player: victim,
				Filter: effects.Nonland(),
				Label:  "nonland card",
				Count:  x,
				UpTo:   true,
				Exile:  true,
				Then:   reapIntellectSearch,
			}.Apply(ctx)
		},
	})
}

// reapIntellectSearch is the clause after the exile: the caster searches the
// victim's graveyard, hand and library for any number of cards sharing a name with
// a card exiled this way, exiling them, and the victim's library is shuffled.
// An on-disk key: never renamed, never reused.
var reapIntellectSearch = effects.RevealedPickThen("revealed-pick/reap-intellect-search", func(ctx *effects.Context, pick game.RevealedPick) error {
	var names []string
	for _, c := range pick.Chosen {
		names = append(names, effects.ExtractionNames(c)...)
	}
	if len(names) == 0 {
		return effects.ShuffleLibrary{Player: pick.FromPlayer}.Apply(ctx)
	}
	return effects.ExtractionSearch(ctx, "Reap Intellect", pick.FromPlayer, names, false)
})
