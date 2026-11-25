package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB240YouControlADesert is "you control a Desert" (Hour of Devastation's
// Desert payoffs): a permanent with the Desert subtype under `you`.
func OldB240YouControlADesert(g *game.Game, you uuid.UUID) bool {
	return len(PermanentsControlledByMatching(g, you, Subtype("Desert"))) > 0
}

// OldB240DesertCardInYourGraveyard is "there is a Desert card in your
// graveyard".
func OldB240DesertCardInYourGraveyard(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype("Desert") {
			return true
		}
	}
	return false
}

// OldB240YouHaveADesert is "you control a Desert or there is a Desert card in
// your graveyard". It reads only board state, so a static ability, an
// activation condition and an intervening "if" may all call it.
func OldB240YouHaveADesert(g *game.Game, you uuid.UUID) bool {
	return OldB240YouControlADesert(g, you) || OldB240DesertCardInYourGraveyard(g, you)
}

// OldB240SelfWhile is the AppliesTo of a static that affects its own source
// alone "as long as" `test` holds for the source's controller.
func OldB240SelfWhile(test func(g *game.Game, you uuid.UUID) bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && test(g, source.Controller)
	}
}

// OldB240FactOrFictionReveal is Unesh, Criosphinx Sovereign's "reveal the top
// four cards of your library. An opponent separates those cards into two
// piles. Put one pile into your hand and the other into your graveyard", with
// the printed card's name on its prompts: the controller picks the opponent
// who separates, who splits, and the controller takes a pile.
func OldB240FactOrFictionReveal(ctx *Context, card string, n int) error {
	revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), n,
		card+" — reveal the top cards of your library")
	if len(revealed) == 0 {
		return nil
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: card + " — choose an opponent to separate these cards into two piles",
		ThenKey:  oldB240Split,
		Carry:    revealed,
	}.Apply(ctx)
}

// oldB240Split hands the revealed cards to the chosen opponent. An on-disk
// key: never renamed, never reused.
var oldB240Split = ChoosePlayerThen("option-pick/old-b240-unesh-splitter", func(ctx *Context, splitter uuid.UUID, revealed []uuid.UUID) error {
	if splitter == uuid.Nil {
		// Nobody left to separate them (CR 800.4a): one pile, taken.
		return FactOrFictionSettle(ctx.Game, revealed, nil)
	}
	return PileSplit{
		Splitter:      splitter,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Unesh, Criosphinx Sovereign — separate these cards into two piles",
		PickQuestion:  "Unesh, Criosphinx Sovereign — take one pile into your hand; the other goes to your graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(oldB240Take),
	}.Apply(ctx)
})

// oldB240Take puts the chosen pile into the hand and the other into the
// graveyard. An on-disk key.
var oldB240Take = PileTail("then/old-b240-unesh-take", func(ctx *Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return FactOrFictionSettle(ctx.Game, taken, left)
})
