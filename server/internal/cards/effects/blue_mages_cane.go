package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blue Mage's Cane — Artifact — Equipment {2}{U}:
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +0/+2, is a Wizard in addition to its other
//	 types, and has 'Whenever this creature attacks, exile up to one target
//	 instant or sorcery card from defending player's graveyard. If you do,
//	 copy it. You may cast the copy by paying {3} rather than paying its
//	 mana cost.'
//	 Spirit of the Whalaqee — Equip {2}"
//
// The attack trigger is a granted ability (ADR 0093), as Sage's Nouliths'
// is: it belongs to the equipped creature, goes away with the Equipment and
// has the creature as its source, so the defending player is read off the
// creature's own attack. The (up to one) card is targeted as the trigger
// goes on the stack, from the graveyard of the player being attacked.
//
// "If you do" is the card that REALLY reached exile (ExileCardsThenForEffect
// hands the continuation exactly those), so a card a replacement sent
// elsewhere, or one that left the graveyard in response, is not copied. The
// copy is a card copy (CR 707.12): it lives in exile, is offered to the
// trigger's controller for {3} in place of its mana cost (CR 118.9, an
// alternative cost, so an X is 0 and additional costs and cost increases
// still apply), and ceases to exist if it is not cast by the controller's
// next priority pass (CR 704.5e). The original stays exiled either way. The
// cast is taken after the trigger resolves, as every resolution-time cast
// offer is (ADR 0066, CopyAndCast).
//
// No simplification beyond that grant model.
func init() {
	const grantKey = "blue-mages-cane/exile-and-copy"
	attack := withTargetsFrom(
		On(game.EventAttack, Self, "Blue Mage's Cane — exile up to one instant or sorcery card from defending player's graveyard, copy it, and you may cast the copy for {3}",
			blueMagesCaneExileAndCopy),
		targetInstantOrSorceryInDefendersGraveyard)
	Register(Spec{
		OracleID:     "6290233b-b015-4909-8aba-bf8d65cc37b5",
		Name:         "Blue Mage's Cane",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{attack},
			Text:      "Whenever this creature attacks, exile up to one target instant or sorcery card from defending player's graveyard. If you do, copy it. You may cast the copy by paying {3} rather than paying its mana cost.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(0, 2),
			AttachedIsAlso("Wizard"),
			GrantAbilitiesToAttached(grantKey),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Blue Mage's Cane")},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// targetInstantOrSorceryInDefendersGraveyard is "up to one target instant
// or sorcery card from defending player's graveyard", read off the source's
// attack.
func targetInstantOrSorceryInDefendersGraveyard(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "card_in_graveyard", Label: "up to one target instant or sorcery card from defending player's graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return (c.IsInstant() || c.IsSorcery()) && c.Owner == defender
		},
		Min: 0, Max: 1,
	}
}

// blueMagesCaneExileAndCopy exiles the chosen card and offers a copy of
// what reached exile for {3}.
func blueMagesCaneExileAndCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		return CopyAndCast{Cards: exiled, Cost: "{3}"}.Apply(NewContext(g, item))
	})
}
