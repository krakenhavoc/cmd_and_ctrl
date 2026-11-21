package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "350f5c0a-563f-41c5-8f7b-10f409bb4d3b",
		Name:         "Primal Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Target player gains 7 life.",
				effects.TargetPlayer("target player"), primalCommandGainLife),
				effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 7})),
			effects.ModeDoing("Put target noncreature permanent on top of its owner's library.",
				effects.TargetPermanent("target noncreature permanent", effects.Noncreature()), effects.TuckTheModesTargetOnTopOfItsOwnersLibrary),
			effects.ModeDoing("Target player shuffles their graveyard into their library.",
				effects.TargetPlayer("target player"), primalCommandShuffle),
			effects.ModeDoing("Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
				nil, primalCommandSearch),
		),
	})
}

func primalCommandGainLife(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.GainLife{Player: t.ID, Amount: 7}.Apply(ctx)
}

func primalCommandShuffle(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.FinaleShuffleGraveyardIntoLibrary(ctx, t.ID)
}

func primalCommandSearch(item *game.StackItem, ctx *effects.Context, _ int) error {
	return effects.B06TutorToHand("Primal Command — a creature card", func(c game.Card) bool { return c.IsCreature() })(item, ctx)
}
