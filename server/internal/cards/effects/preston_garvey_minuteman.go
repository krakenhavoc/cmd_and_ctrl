package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "08a0c661-5ed5-47a6-aa1d-3114bf5db3fb",
		Name:         "Preston Garvey, Minuteman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Preston Garvey — create a Settlement Aura token attached to up to one target land you control",
				prestonSettlement),
				TargetPermanent("up to one target land you control", Land(), YouControl()).WithCount(0, 1)),
			WheneverThisAttacks("Preston Garvey — untap each enchanted permanent you control", prestonUntap),
		},
	})
}

func prestonSettlement(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opts := game.TokenEntryOptions{}
	if host, ok := b16FirstLegalTargetCard(ctx); ok {
		opts.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(GapB63SettlementToken()), 1, opts)
	return err
}

func prestonUntap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, YouControl()) {
		if c.Tapped && b18AurasAttachedTo(g, c.InstanceID) > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
