package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Composer of Spring — Creature — Satyr Bard {1}{G}, 1/3 (Commander
// Masters):
//
//	"Constellation — Whenever an enchantment you control enters, you may
//	 put a land card from your hand onto the battlefield tapped. If you
//	 control six or more enchantments, instead you may put a creature or
//	 land card from your hand onto the battlefield tapped."
//
// A put, not a play (CR 305.4), so the land drop is untouched; the card
// arrives tapped through the entry pipeline. The count of six is taken as
// the trigger resolves ("if you control", not an intervening if): the
// enchantment that triggered it is counted, and so is one that has since
// left no longer. The creature that is put is tapped as well, as printed.
// The trigger is "an enchantment you control", so it never fires for
// Composer itself, which is not an enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b5d1b0a-544c-41e6-bc32-c2b0c9bb4576",
		Name:         "Composer of Spring",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, gapB102EnchantmentEnteredUnderYourControl,
				"Composer of Spring — you may put a land card (or, with six enchantments, a creature or land card) from your hand onto the battlefield tapped",
				gapB102ComposerPut),
		},
	})
}

// gapB102EnchantmentEnteredUnderYourControl is "whenever an enchantment you
// control enters".
func gapB102EnchantmentEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsEnchantment()
}

func gapB102ComposerPut(g *game.Game, item *game.StackItem) error {
	put := PutFromHandOntoBattlefield{
		Match:    Land(),
		Optional: true,
		Tapped:   true,
		Label:    "Composer of Spring — you may put a land card from your hand onto the battlefield tapped",
	}
	if gapB102EnchantmentsControlledBy(g, item.Controller) >= 6 {
		put.Match = Or(Creature(), Land())
		put.Label = "Composer of Spring — you may put a creature or land card from your hand onto the battlefield tapped"
	}
	return put.Apply(NewContext(g, item))
}

// gapB102EnchantmentsControlledBy counts the enchantments a player controls,
// read through the layers.
func gapB102EnchantmentsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsEnchantment() {
			n++
		}
	}
	return n
}
