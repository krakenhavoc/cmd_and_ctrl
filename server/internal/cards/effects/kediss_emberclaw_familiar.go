package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kediss, Emberclaw Familiar — Legendary Creature — Elemental Lizard
// {1}{R}, 1/1 (EDHREC rank 1162):
//
//	"Whenever a commander you control deals combat damage to an
//	 opponent, it deals that much damage to each other opponent.
//	 Partner (You can have two commanders if both have partner.)"
//
// The partner that turns one commander hit into a table hit. The
// trigger is the combat-damage-to-a-player condition with the
// dealing creature narrowed to a commander (the deck importer's
// IsCommander flag, the field Bastion Protector reads); the amount
// and the struck opponent are read at resolution off the item's
// carried trigger context (item.Trigger.Event, #1223), and the
// commander itself is the damage source for each other opponent, so a
// damage doubler sees the commander. That damage is NOT combat
// damage, so it does not add to the per-commander tally (#380) — the
// printed ruling. Two partner commanders connecting are two triggers,
// one each, as printed.
//
// Partner is a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9c87cc2-943e-49b6-becc-748857549617",
		Name:         "Kediss, Emberclaw Familiar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10CommanderYouControlDealtCombatDamageToOpponent(ev, source, g)
			},
			Key: "Kediss — that much damage to each other opponent",
			Effect: func(g *game.Game, item *game.StackItem) error {
				commander, struck, amount := item.Trigger.Event.Source, item.Trigger.Event.Target, item.Trigger.Event.Amount
				ctx := NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if opp == struck {
						continue
					}
					if err := (DealDamage{Source: commander, Target: opp, Amount: amount}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
