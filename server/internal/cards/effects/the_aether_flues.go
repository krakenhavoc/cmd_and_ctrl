package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Aether Flues — Plane — Iquatana (Commander Masters, #4870):
//
//	"When you planeswalk to The Aether Flues and at the beginning of your
//	 upkeep, you may sacrifice a creature. If you do, reveal cards from the
//	 top of your library until you reveal a creature card, put that card
//	 onto the battlefield, then shuffle all other cards revealed this way
//	 into your library.
//	 Whenever chaos ensues, you may put a creature card from your hand onto
//	 the battlefield."
//
// A plane (ADR 0216 §6). The first ability is one ability with two trigger
// conditions; sacrificing is a "you may" asked as it resolves (CR 608.2d),
// and only a creature sacrificed starts the reveal ("if you do"). The
// creature card revealed enters, and the other revealed cards, which never
// left the library, are shuffled in with it. The chaos ability (CR 311.7) is
// a "you may" over the planar controller's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8eb442a9-5e61-4e2e-983d-b49b04be2575",
		Name:         "The Aether Flues",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("The Aether Flues — you may sacrifice a creature; if you do, reveal until a creature card, put it onto the battlefield, then shuffle the rest into your library",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "The Aether Flues — sacrifice a creature?", Creature(), false, aetherFluesSacrificed)
				}),
			WheneverChaosEnsues("The Aether Flues — you may put a creature card from your hand onto the battlefield",
				Do(PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    "The Aether Flues — put a creature card from your hand onto the battlefield",
				})),
		},
	})
}

// aetherFluesSacrificed is "if you do, reveal cards from the top of your
// library until you reveal a creature card, put that card onto the
// battlefield, then shuffle all other cards revealed this way into your
// library". An on-disk key: never renamed, never reused.
var aetherFluesSacrificed = RegisterSacrificedOne("prompt-run/the-aether-flues-reveal", func(ctx *Context, _ uuid.UUID) error {
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, game.Card.IsCreature, "The Aether Flues — revealed until a creature card")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: shuffleLibraryTail,
	}.Apply(ctx)
})

// shuffleLibraryTail is "then shuffle" after a put from the library: the
// controller shuffles their library. An on-disk key.
var shuffleLibraryTail = PutFromLibraryTail("planes/shuffle-library", func(ctx *Context, _ PutFromLibraryResult) error {
	return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
})
