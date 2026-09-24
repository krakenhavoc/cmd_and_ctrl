package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perpetual Timepiece — Artifact {2} (EDHREC rank 2025):
//
//	"{T}: Mill two cards. (Put the top two cards of your library
//	 into your graveyard.)
//	 {2}, Exile this artifact: Shuffle any number of target cards
//	 from your graveyard into your library."
//
// The self-mill rock that doubles as graveyard insurance. The mill is
// a tap ability on the stack, two cards, and every mill watcher sees
// them.
//
// The shuffle waited on #1404. "Exile this artifact" is ExileThis()
// paid from the BATTLEFIELD: the Timepiece leaves the battlefield at
// announce (CR 602.2b), before anyone can respond, so it cannot be
// destroyed in response to save the cost and it is gone before the
// ability resolves. It is not a sacrifice (no "whenever you sacrifice"
// payoff sees it) and it does not die (exile is not a graveyard), so
// only leaves-the-battlefield watchers fire — what the printed card
// does, and the reason it could not be shipped as SacrificeThis().
//
// "Any number of target cards" is a zero-to-unbounded graveyard clause,
// so the activator may name nothing and still shuffle. Each target
// still in the graveyard at resolution goes into the library (CR
// 608.2b skips one that left in response), then the library is
// shuffled once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17b4778b-82b1-4845-ad08-00f3ff66877b",
		Name:         "Perpetual Timepiece",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{T}: Mill two cards.",
				Cost:  TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MillCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{2}, Exile this artifact: Shuffle any number of target cards from your graveyard into your library.",
				Cost:    Plus(ManaCost("{2}"), ExileThis()),
				Targets: TargetCardInGraveyard("any number of target cards from your graveyard", YouOwn()).WithCount(0, 0),
				Effect:  shuffleTargetGraveyardCardsIntoLibrary,
			},
		},
	})
}

// shuffleTargetGraveyardCardsIntoLibrary is "Shuffle any number of
// target cards from your graveyard into your library": every target
// still legal goes into its owner's library, then the controller's
// library is shuffled — once, and even when nothing was named, because
// the instruction is the shuffle.
func shuffleTargetGraveyardCardsIntoLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	return ShuffleLibrary{Player: item.Controller}.Apply(ctx)
}
