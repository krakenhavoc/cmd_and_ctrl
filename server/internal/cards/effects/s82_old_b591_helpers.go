package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b591_helpers.go — Saviors of Kamigawa, batch 591 (#5601): the
// Sweep cycle's "return any number of <basic land type>s you control to their
// owner's hand", asked as a pick and handed to the card's own continuation.

// OldB591AskSweep is the first half of a Sweep line (the ability word of the
// four Arcane spells): "return any number of <subtype>s you control to their
// owner's hand". The player picks any number, none included, of the
// permanents of that land type they control as the spell resolves (CR 608.2c,
// 701.12 of the chosen set); nothing is targeted. With none to pick the rest
// of the line scales to zero and nothing is asked. `then` is a registered pick
// (AskedPick) that returns the cards (OldB591ReturnSwept) and does the rest;
// it runs against the spell's own item, so its target is read as usual.
func OldB591AskSweep(ctx *Context, subtype, question string, then game.ChooseCardsThen) error {
	var cands []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(YouControl(), HasSubtype(subtype))) {
		cands = append(cands, c.InstanceID)
	}
	if len(cands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   question,
		Cards:      cands,
		Min:        0,
		Max:        len(cands),
		Zone:       game.ZoneBattlefield,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    then,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// OldB591ReturnSwept returns the picked permanents to their owners' hands and
// reports how many actually reached a hand ("for each <land> returned this
// way"): a card that left on its own, or that a replacement kept away, is not
// counted.
func OldB591ReturnSwept(ctx *Context, picked []uuid.UUID) (int, error) {
	n := 0
	for _, id := range picked {
		if !OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return n, err
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			n++
		}
	}
	return n, nil
}
