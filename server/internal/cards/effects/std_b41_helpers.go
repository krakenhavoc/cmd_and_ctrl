package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b41_helpers.go — shared pieces of Standard set batch 41 (#4452).

// AskExcluding is CollectEvidence.Ask for a source that is itself a card
// in the graveyard and may not pay for its own collection (Lamplight
// Phoenix, whose ruling says it can't be exiled to pay): `exclude` is
// never offered, and a graveyard whose other cards total under N asks
// nothing (CR 701.59b).
func (ce CollectEvidence) AskExcluding(ctx *Context, exclude uuid.UUID, question string) error {
	if ce.n <= 0 {
		return fmt.Errorf("%w: collect evidence of %d", game.ErrInvalidParam, ce.n)
	}
	cards := ctx.Game.CollectEvidenceOptionsForEffect(ctx.Controller(), exclude, ce.n)
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Item.SourceCardID,
		Question: question,
		Cards:    cards,
		Min:      0,
		Max:      len(cards),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  ce.pick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// RevealSecretChoice is "Reveal the creature type you chose" as a cost
// component (A Killer Among Us, game.AbilityCost.RevealSecretChoice): the
// word the controller secretly chose for the source is shown to the table
// as the cost is paid. It claims no X and names no card, so it composes
// with any other component; pair it with ActivatedAbility.Condition
// YouSecretlyChoseForThis so the ability is not offered to anyone else.
func RevealSecretChoice() game.AbilityCost { return game.AbilityCost{RevealSecretChoice: true} }

// YouSecretlyChoseForThis is the activation condition of a reveal-the-
// choice ability: a choice was made for the source and the activator is
// the player who made it (the printed ruling).
func YouSecretlyChoseForThis(g *game.Game, controller, source uuid.UUID) bool {
	return g.CanRevealSecretChoiceForEffect(controller, source)
}
