package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tamiyo's Compleation — Enchantment — Aura {3}{U} (Neon Dynasty, #5493):
//
//	"Flash
//	 Enchant artifact, creature, or planeswalker
//	 When this Aura enters, tap enchanted permanent. If it's an Equipment,
//	 unattach it.
//	 Enchanted permanent loses all abilities and doesn't untap during its
//	 controller's untap step."
//
// The enters trigger reads the permanent the Aura enchants as it resolves,
// or the one it last enchanted (CR 608.2h): it is tapped, and unattached
// when it is an Equipment (a permanent that is an Equipment through the
// layers). "Loses all abilities" is layer 6 removal and "doesn't untap" an
// untap-step restriction; both end with the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "71d8b7f5-a5da-4528-815e-51c65626b681",
		Name:                  "Tamiyo's Compleation",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.TargetPermanent("enchant artifact, creature, or planeswalker", effects.Or(effects.Artifact(), effects.Creature(), effects.Planeswalker())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		Static:                []game.StaticAbility{effects.LoseAllAbilities()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tamiyo's Compleation — tap enchanted permanent; if it's an Equipment, unattach it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok || !effects.OnBattlefield(g, host) {
						return nil
					}
					if err := (effects.TapTarget{Target: host}).Apply(ctx); err != nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(host); ok && c.HasSubtype("Equipment") {
						return g.UnattachForEffect(host)
					}
					return nil
				}),
		},
	})
}
