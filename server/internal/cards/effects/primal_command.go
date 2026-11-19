package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Command — Sorcery {3}{G}{G} (Breaking News, #5929):
//
//	"Choose two —
//	 • Target player gains 7 life.
//	 • Put target noncreature permanent on top of its owner's library.
//	 • Target player shuffles their graveyard into their library.
//	 • Search your library for a creature card, reveal it, put it into your
//	   hand, then shuffle."
//
// The same "choose two" template as Ashling's Command: each chosen bullet
// has its own target, checked again as the spell resolves (CR 608.2b), and
// the bullets run in printed order (CR 608.2c). The tuck is Aura
// Extraction's, to the OWNER's library whoever controls the permanent; the
// graveyard shuffle is Finale of Revelation's, for any player; the search
// is Eladamri's Call's. Choosing the shuffle and the search together runs
// the shuffle first, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "350f5c0a-563f-41c5-8f7b-10f409bb4d3b",
		Name:         "Primal Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeWithPurpose(ModeDoing("Target player gains 7 life.",
				TargetPlayer("target player"), primalCommandGainLife),
				ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 7})),
			ModeDoing("Put target noncreature permanent on top of its owner's library.",
				TargetPermanent("target noncreature permanent", Noncreature()), tuckTheModesTargetOnTopOfItsOwnersLibrary),
			ModeDoing("Target player shuffles their graveyard into their library.",
				TargetPlayer("target player"), primalCommandShuffle),
			ModeDoing("Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
				nil, primalCommandSearch),
		),
	})
}

func primalCommandGainLife(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return GainLife{Player: t.ID, Amount: 7}.Apply(ctx)
}

func primalCommandShuffle(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return finaleShuffleGraveyardIntoLibrary(ctx, t.ID)
}

func primalCommandSearch(item *game.StackItem, ctx *Context, _ int) error {
	return b06TutorToHand("Primal Command — a creature card", func(c game.Card) bool { return c.IsCreature() })(item, ctx)
}
