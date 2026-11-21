package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vesuvan Drifter — Creature — Shapeshifter {2}{U}, 2/4 (March of the
// Machine: The Aftermath, #5476):
//
//	"Flying
//	 You may look at the top card of your library any time.
//	 At the beginning of each combat, you may reveal the top card of your
//	 library. If you reveal a creature card this way, this creature
//	 becomes a copy of that card until end of turn, except it has flying."
//
// The look is private (LibraryTopOwner, Augur of Autumn's shape). The
// trigger fires at the beginning of EACH player's combat. The "you may" is
// the CR 603.5 prompt as it resolves; revealing is public and the card
// stays on top. A creature card revealed is copied by the duration copy
// effect (BecomeCopy, CR 707.2, CR 611.2): the copiable values of a card
// in the library, read as the effect begins and stored, so drawing the
// card afterwards does not change the copy. The exception adds flying to
// the copy's values (CR 707.9b); the copy ends at cleanup. A noncreature
// card reveals but copies nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "727ef984-1643-4a09-8283-a665508e85dd",
		Name:              "Vesuvan Drifter",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtBeginningOfEachCombat("Vesuvan Drifter — reveal the top card of your library; if it's a creature card, this creature becomes a copy of it until end of turn, except it has flying",
				vesuvanDrifterReveal),
				"Vesuvan Drifter — reveal the top card of your library?"),
		},
	})
}

func vesuvanDrifterReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{
		Player:   ctx.Controller(),
		N:        1,
		Reason:   "Vesuvan Drifter — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(revealed[0])
	if !ok || !c.IsCreature() {
		return nil
	}
	return effects.BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       revealed[0],
		Duration: effects.CopyUntilEndOfTurn,
		Except:   func(v *game.PrintedValues) { v.AddKeyword("flying") },
		Label:    "Vesuvan Drifter — becomes a copy of the revealed creature card until end of turn, except it has flying",
	}.Apply(ctx)
}
