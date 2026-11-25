package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b217_helpers.go — S82 OLD batch 217 (Rivals of Ixalan, #5531).

// DoubleCombatDamageThisDealsToPlayers is "If this creature would deal combat
// damage to a player, it deals double that damage to that player instead"
// (Charging Tuskodon, CR 614.1): a replacement on the creature's own combat
// damage to a player, any player, itself included.
func DoubleCombatDamageThisDealsToPlayers(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || !ev.IsCombatDamage || ev.DamageSource != src.InstanceID {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DamageTarget)
			return p != nil && !p.Eliminated
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// WheneverThisDamagesACreatureOfTypeDestroyIt is "Whenever this creature deals
// damage to a <subtype>, destroy that creature" (Dinosaur Hunter): any damage,
// combat or not. It names the damaged creature as the object it was when the
// trigger fired, so one that died to the damage, or died and came back, is not
// destroyed (CR 400.7). The type is read through the layers as the damage was
// dealt, and through the text of the card the word is printed on (CR 612.2).
func WheneverThisDamagesACreatureOfTypeDestroyIt(label, subtype string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.Target)
			if !ok || !c.IsCreature() || !c.HasSubtype(source.TextSubtype(subtype)) {
				return false
			}
			z := g.FindCardZoneForEffect(ev.Target)
			return z != nil && z.Kind == game.ZoneBattlefield
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, ev.Target)
			return item
		},
		Effect: DestroyDamagedCreature,
	}
}

// AtYourEndStepSacrificeThisIfYouDidntAttack is "At the beginning of your end
// step, if you didn't attack with a creature this turn, sacrifice this Aura"
// (See Red): an intervening "if" (CR 603.4), checked as the end step begins
// and again as the trigger resolves.
func AtYourEndStepSacrificeThisIfYouDidntAttack(label string) game.TriggeredAbility {
	return On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && !B18AttackedThisTurn(g, source.Controller)
	}, label, sacrificeThisAuraIfYouStillDidntAttack)
}

// sacrificeThisAuraIfYouStillDidntAttack is the second check of that intervening
// "if" (CR 603.4), made as the trigger resolves.
func sacrificeThisAuraIfYouStillDidntAttack(g *game.Game, item *game.StackItem) error {
	if B18AttackedThisTurn(g, item.Controller) {
		return nil
	}
	return SacrificeThisIfStillOnBattlefield(g, item)
}
