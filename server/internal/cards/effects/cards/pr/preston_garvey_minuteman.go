package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Preston Garvey, Minuteman — Legendary Creature — Human Soldier
// {2}{R}{G}{W}, 4/4 (Fallout, #5948):
//
//	"At the beginning of combat on your turn, create a green Aura
//	 enchantment token named Settlement attached to up to one target land
//	 you control. It has enchant land and "Enchanted land has '{T}: Add one
//	 mana of any color.'"
//	 Whenever Preston Garvey attacks, untap each enchanted permanent you
//	 control."
//
// The Settlement is the printed token (GapB63SettlementToken): its static
// gives the enchanted land Cryptolith Rite's mana ability. The target is
// "up to one": chosen as the trigger goes on the stack, and with none (or if
// the land is no longer a legal target as it resolves) the Aura is created
// attached to nothing and is put into the graveyard by the next state check
// (CR 704.5m, 303.4i). "Enchanted" is any Aura attached to the permanent,
// whoever controls the Aura; the set is chosen as the attack trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08a0c661-5ed5-47a6-aa1d-3114bf5db3fb",
		Name:         "Preston Garvey, Minuteman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Preston Garvey — create a Settlement Aura token attached to up to one target land you control",
				prestonSettlement),
				effects.TargetPermanent("up to one target land you control", effects.Land(), effects.YouControl()).WithCount(0, 1)),
			effects.WheneverThisAttacks("Preston Garvey — untap each enchanted permanent you control", effects.OldB006UntapEachEnchantedPermanentYouControl),
		},
	})
}

func prestonSettlement(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opts := game.TokenEntryOptions{}
	if host, ok := effects.B16FirstLegalTargetCard(ctx); ok {
		opts.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.GapB63SettlementToken()), 1, opts)
	return err
}
