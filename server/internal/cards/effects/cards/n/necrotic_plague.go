package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrotic Plague — Enchantment — Aura {2}{B}{B} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature has "At the beginning of your upkeep, sacrifice this
//	 creature."
//	 When enchanted creature dies, its controller chooses target creature one
//	 of their opponents controls. Return this card from its owner's graveyard
//	 to the battlefield attached to that creature."
//
// The granted ability is a bundle on the enchanted creature, whose "your" is
// that creature's controller. The dies trigger is the Aura's, harvested while
// it is still attached (WhenEnchantedCreatureDies, Gift of Immortality's
// read): the creature's controller as it last was chooses the target
// (ChosenByTriggerActor, with the chooser read off the leaving event's
// last-known controller, CR 603.10a), and the creature must be controlled by
// someone else than that chooser ("one of their opponents"), legality
// judged for the Aura's controller. The Aura comes back from its owner's
// graveyard attached to it (CR 303.4f), and only if it is still in a
// graveyard when the trigger resolves.
//
// No simplification.
const necroticPlagueGrant = "necrotic-plague/upkeep-sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d637892-26d7-46f1-bd7d-89bda523f649",
		Name:         "Necrotic Plague",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: necroticPlagueGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Necrotic Plague — sacrifice this creature", effects.SacrificeThisIfStillOnBattlefield),
			},
			Text: "At the beginning of your upkeep, sacrifice this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(necroticPlagueGrant)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenEnchantedCreatureDies("Necrotic Plague — return this Aura to the battlefield attached to target creature one of its controller's opponents controls",
					S82OldB476PlagueSpreads),
				effects.AnotherPlayerControls(effects.OfThatPlayersChoice(effects.TargetCreature("target creature one of their opponents controls")))),
		},
	})
}

// S82OldB476PlagueSpreads returns the Aura from its owner's graveyard to the
// battlefield attached to the chosen creature, while the Aura is still in
// the graveyard and the creature is still legal.
func S82OldB476PlagueSpreads(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: item.SourceCardID, From: z.Kind}},
		game.ZoneEntryOptions{AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}},
		game.KeyedTail{})
}
