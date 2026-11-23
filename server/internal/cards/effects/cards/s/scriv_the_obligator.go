package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scriv, the Obligator — Legendary Creature — Inkling Bird {2}{W}{B}, 2/3
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying, deathtouch
//	 Whenever Scriv enters or attacks, create a white Aura enchantment
//	 token named Contract attached to target creature an opponent
//	 controls. The token has enchant creature and "Whenever enchanted
//	 creature attacks, it gets +2/+0 until end of turn if it's attacking
//	 one of your opponents. Otherwise, its controller loses 2 life.""
//
// Flying and deathtouch are the printed keywords. "Enters or attacks" is
// ONE ability with two conditions (WhenThisEntersOrAttacks). The target is
// chosen as it goes on the stack and re-checked as it resolves: a target
// that is illegal then means nothing happens and no Contract is created
// (the ruling).
//
// The Contract is the token catalog's printedContractToken (soc_b12_contract_token.go): a white Aura token with its own attack trigger,
// created already attached to the target (TokenEntryOptions.AttachedTo,
// CR 303.4i), and controlled by Scriv's controller, who is the "you" of
// its text. Because that trigger is the Aura's, a creature taken over by
// someone else is still affected, and "its controller" is whoever controls
// the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64c986f0-e786-419c-9b51-8c316c8fae0e",
		Name:            "Scriv, the Obligator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Scriv, the Obligator — create a Contract token attached to target creature an opponent controls",
					scrivContract),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

func scrivContract(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.ContractToken()), 1, game.TokenEntryOptions{
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
	return err
}
