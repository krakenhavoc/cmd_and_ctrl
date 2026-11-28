package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infectious Rage — Enchantment — Aura {1}{R} (Judgment):
//
//	"Enchant creature
//	 Enchanted creature gets +2/-1.
//	 When enchanted creature dies, choose a creature at random this Aura
//	 can enchant. Return this card to the battlefield attached to that
//	 creature."
//
// Takklemaggot's return-on-host-death shape (game.CreaturesAuraCardCouldEnchant),
// with the creature drawn from the keyed random-pick stream instead of
// chosen by a player. The Aura returns under its owner's control, only if it
// is still in the graveyard, and stays there when no creature is left.
//
// No simplification.
func init() {
	const name = "Infectious Rage"
	effects.Register(effects.Spec{
		OracleID:     "ed0402d7-02cb-4763-a661-82280d6c9af8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(2, -1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies(name+" — return this card attached to a creature chosen at random",
				func(g *game.Game, item *game.StackItem) error {
					aura := item.SourceCardID
					if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					card, ok := g.LookupCardForEffect(aura)
					if !ok {
						return nil
					}
					owner := card.Owner
					hosts := g.CreaturesAuraCardCouldEnchantForEffect(aura, owner)
					pick := g.ChooseAtRandomForEffect(effects.RandomDraw(effects.NewContext(g, item)), hosts, 1)
					if len(pick) == 0 {
						return nil
					}
					return g.PutOntoBattlefieldTogetherKeyedForEffect(
						[]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
						game.ZoneEntryOptions{Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: pick[0]}},
						game.KeyedTail{})
				}),
		},
	})
}
