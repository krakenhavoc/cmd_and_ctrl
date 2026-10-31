package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Lad, Diverging Destiny — {2}{U} Legendary Artifact Creature — Human
// Hero 2/2 (Marvel Super Heroes, #4455):
//
//	"Flying, vigilance
//	 You may look at the top card of your library any time.
//	 {T}: Reveal the top card of your library. If it's an artifact card,
//	 draw a card."
//
// The peek is the library's top-visible-to-owner flag. The tap ability
// reveals the top card to the table (so it is no secret afterwards), and
// then, when the revealed card is an artifact card, draws a card, which is
// that same card since nothing moves in between. An empty library reveals
// nothing and draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "3e715032-0d21-441b-b2d4-c91084728718",
		Name:              "Iron Lad, Diverging Destiny",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying", "vigilance"},
		LibraryTopVisible: game.LibraryTopOwner,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Reveal the top card of your library. If it's an artifact card, draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  mshA4IronLadReveal,
		}},
	})
}

// mshA4IronLadReveal reveals the top card and draws if it is an artifact.
func mshA4IronLadReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{
		Player:   ctx.Controller(),
		N:        1,
		Reason:   "Iron Lad — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(revealed[0]); !ok || !c.IsArtifact() {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
}
