package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geth, Lord of the Vault — Legendary Creature — Phyrexian Zombie
// {4}{B}{B}, 5/5:
//
//	"Intimidate
//	 {X}{B}: Put target artifact or creature card with mana value X from
//	 an opponent's graveyard onto the battlefield tapped under your
//	 control. Then that player mills X cards."
//
// X is announced with the activation (CR 107.3a) and the target clause
// reads it (WithManaValueEqualsX), checked at announce and again as the
// ability resolves (CR 608.2b), so a card that left the graveyard fizzles
// the whole ability and nothing is milled. "That player" is the card's
// owner, read before it moves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2310103-ec84-49df-bb12-0f0f175ee801",
		Name:            "Geth, Lord of the Vault",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label: "{X}{B}: Put target artifact or creature card with mana value X from an opponent's graveyard onto the battlefield tapped under your control. Then that player mills X cards.",
			Cost:  effects.ManaCost("{X}{B}"),
			Targets: effects.TargetCardInGraveyard("target artifact or creature card with mana value X from an opponent's graveyard",
				effects.Or(effects.Artifact(), effects.Creature()), effects.Not(effects.YouOwn())).WithManaValueEqualsX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return gethLordOfTheVaultReanimate(g, item)
			},
		}},
	})
}

func gethLordOfTheVaultReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		owner := c.Owner
		if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}).Apply(ctx); err != nil {
			return err
		}
		return effects.MillCards{Player: owner, N: ctx.X()}.Apply(ctx)
	}
	return nil
}
