package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ring Goes South — Sorcery {3}{G}:
//
//	"The Ring tempts you. Then reveal cards from the top of your
//	 library until you reveal X land cards, where X is the number of
//	 legendary creatures you control. Put those land cards onto the
//	 battlefield tapped and the rest on the bottom of your library in a
//	 random order."
//
// X is counted after the tempt, so the new Ring-bearer, legendary
// through the Ring, counts. With X = 0 nothing is revealed. A library
// with fewer than X lands is revealed entirely and every land in it
// enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d36f0ab-d333-4a08-ad1b-5b7e1ddb647c",
		Name:         "The Ring Goes South",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TheRingTemptsYou{Then: theRingGoesSouthReveal}.Apply(ctx)
		},
	})
}

// theRingGoesSouthReveal is the sentence after the tempt.
func theRingGoesSouthReveal(ctx *Context, _ uuid.UUID) error {
	player := ctx.Controller()
	x := len(legendaryCreaturesYouControl(ctx.Game, player))
	if x <= 0 {
		return nil
	}
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	var run []uuid.UUID
	lands := map[uuid.UUID]bool{}
	for i := len(p.Library.Cards) - 1; i >= 0 && len(lands) < x; i-- {
		c := p.Library.Cards[i]
		run = append(run, c.InstanceID)
		if !c.IsToken() && c.IsLand() {
			lands[c.InstanceID] = true
		}
	}
	ctx.Game.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: ctx.Source(),
		Reason: "The Ring Goes South — revealed until enough land cards",
		Cards:  run,
	})
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return lands[c.InstanceID] },
		All:    true,
		Tapped: true,
		Then:   PutRestOnBottomInRandomOrder,
	}.Apply(ctx)
}
