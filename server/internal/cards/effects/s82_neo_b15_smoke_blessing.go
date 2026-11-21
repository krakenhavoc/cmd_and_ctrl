package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_neo_b15_smoke_blessing.go — Smoke Spirits' Aid's Smoke Blessing token
// (Kamigawa: Neon Dynasty Commander, S82 #5493): a red Aura enchantment
// token with enchant creature and "When enchanted creature dies, it deals 1
// damage to its controller and you create a Treasure token."

// SmokeBlessingToken is the Smoke Blessing token's template.
func SmokeBlessingToken() game.Card { return TokenFromCatalog(printedSmokeBlessingToken) }

// printedSmokeBlessingToken is the Smoke Blessing as printed. The trigger is
// the Aura's own, so "you" is the token's controller (CR 109.5), however the
// creature is controlled; "its controller" is whoever controlled the
// creature as it died (last-known information, CR 608.2h), and the creature
// is the damage source.
func printedSmokeBlessingToken() TokenTemplate {
	return TokenTemplate{
		Slug: "smoke-blessing",
		Card: game.Card{
			Name:     "Smoke Blessing",
			TypeLine: "Token Enchantment — Aura",
			Colors:   []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Smoke Blessing — it deals 1 damage to its controller and you create a Treasure token",
				smokeBlessingDies),
		},
		Text: "Enchant creature\nWhen enchanted creature dies, it deals 1 damage to its controller and you create a Treasure token.",
	}
}

func smokeBlessingDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if obj := ctx.Trigger().Object; obj != nil {
		if creature, ok := ctx.TriggeringPermanent(); ok {
			ref := obj.Ref()
			if p := g.PlayerByIDForEffect(creature.Controller); p != nil && !p.Eliminated {
				if err := (DealDamage{SourceObject: &ref, Target: creature.Controller, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
			}
		}
	}
	return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
}
