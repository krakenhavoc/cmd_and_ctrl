package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rofellos's Gift — Sorcery {G} (Urza's Destiny, S82 #5621):
//
//	"Reveal any number of green cards in your hand. Return an enchantment card
//	 from your graveyard to your hand for each card revealed this way."
//
// OldB706RevealAnyNumber, then the controller chooses that many enchantment cards
// in their graveyard (as many as there are, when fewer) and returns them to their
// hand. The enchantments are chosen after the reveal, from the graveyard as it
// is then (CR 608.2c); nothing is asked when there are none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c309759-e607-4bad-b615-f7c673e508ac",
		Name:         "Rofellos's Gift",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB706RevealAnyNumber("Rofellos's Gift", "G", "choose-cards/s82-old-b706-rofellos-gift-reveal", rofellosGiftReturn),
	})
}

// rofellosGiftReturn asks which enchantment cards come back.
func rofellosGiftReturn(ctx *effects.Context, n int) error {
	cards := effects.GraveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsEnchantment() })
	if len(cards) == 0 || n <= 0 {
		return nil
	}
	k := min(n, len(cards))
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Rofellos's Gift — return an enchantment card from your graveyard to your hand for each card revealed",
		Cards:      cards,
		Min:        k,
		Max:        k,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    rofellosGiftTake,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// rofellosGiftTake returns the chosen enchantments to the hand. An on-disk key.
var rofellosGiftTake = effects.AskedPick("choose-cards/s82-old-b706-rofellos-gift-take", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if id == uuid.Nil {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
