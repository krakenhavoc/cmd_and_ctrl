package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b19_helpers.go — shared bodies for Commander wave 2, batches 19
// and 20 (#5368).

// cmdr2B19AttachedCantAttackYou is "Enchanted creature … can't attack you
// or planeswalkers you control" (Vow of Duty, Vow of Lightning, CR 508.1c):
// a layer 6 restriction on the attached creature naming the Aura's
// controller, read when the layers run, so a stolen Aura protects its new
// controller. A battle that player protects is not covered.
func cmdr2B19AttachedCantAttackYou() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature can't attack you or planeswalkers you control.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:     source.InstanceID,
				SourceName: source.Name,
				NotPlayer:  source.Controller,
			})
		},
	}
}

// cmdr2B19HumanMonkToken is Careful Cultivation's 1/1 green Human Monk
// creature token with "{T}: Add {G}."
func cmdr2B19HumanMonkToken() game.Card { return tokenFromCatalog(printedCmdr2B19HumanMonkToken) }

// printedCmdr2B19HumanMonkToken is that token as PRINTED (Neon Dynasty's
// token card, tneo 10). Its slug is an on-disk identity.
func printedCmdr2B19HumanMonkToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-monk-tap-g",
		Card: game.Card{
			Name:      "Human Monk",
			TypeLine:  "Token Creature — Human Monk",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Text: "{T}: Add {G}.",
	}
}

// cmdr2B19AttachedIsCreature is the condition of Careful Cultivation's
// "As long as enchanted permanent is a creature": the host as the earlier
// layers have left it this pass.
func cmdr2B19AttachedIsCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return AttachedToSource(target, g, source) && target.IsCreature()
}
