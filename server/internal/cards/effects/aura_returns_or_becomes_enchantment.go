package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_returns_or_becomes_enchantment.go — #6235, Takklemaggot:
//
//	"When enchanted creature dies, that creature's controller chooses a
//	 creature that this card could enchant. If the player does, return
//	 this card to the battlefield under your control attached to that
//	 creature. If they don't, return this card to the battlefield under
//	 your control as a non-Aura enchantment. It loses 'enchant creature'
//	 and gains 'At the beginning of that player's upkeep, this
//	 enchantment deals 1 damage to that player.'"
//
// The choice is the dead creature's controller's (its last-known
// controller, CR 603.10a), among the creatures the card could enchant as
// it enters under your control (game.CreaturesAuraCouldEnchantForEffect:
// its enchant clause, protection and "can't be enchanted", no targeting
// gate, CR 303.4f). A creature chosen: the card returns attached to it,
// the host named by the effect (CR 303.4f's "doesn't specify" is not
// met); one gone or no longer legal by then leaves the card where it is
// (CR 303.4i). No creature chosen, or none to choose: the card returns as
// a non-Aura enchantment — ADR 0210's entry overwrite with the Aura
// subtype removed (layer 4), so it is never an Aura on the battlefield,
// is not asked what it enchants and is not held to an enchant clause
// (CR 704.5m); its enchant ability goes with being an Aura. It gains the
// grant bundle `grantKey` (layer 6, ADR 0093), whose ability reads "that
// player" off the permanent (Card.ChosenPlayer, set as it lands). Either
// way it is a new object (CR 400.7).
//
// Nothing happens if the card is no longer in a graveyard when the
// trigger resolves (CR 608.2b's spirit for "this card").

// ReturnToNewHostOrAsEnchantment is the effect of the dies trigger
// (WhenEnchantedCreatureDies). grantKey names the card's AbilityGrant
// (ThatPlayersUpkeepDealsDamageGrant).
func ReturnToNewHostOrAsEnchantment(name, grantKey string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		aura := item.SourceCardID
		if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		chooser := uuid.Nil
		if item.Trigger != nil {
			chooser = item.Trigger.Event.LastKnownController
		}
		var candidates []uuid.UUID
		if p := g.PlayerByIDForEffect(chooser); p != nil && !p.Eliminated {
			candidates = g.CreaturesAuraCouldEnchantForEffect(aura, item.Controller)
		}
		if len(candidates) == 0 {
			return returnAsNonAuraEnchantment(g, aura, item.Controller, chooser, grantKey, name)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   chooser,
			Source:    aura,
			Question:  name + " — choose a creature for it to enchant, or none to have it return as an enchantment that damages you each upkeep",
			Cards:     candidates,
			Min:       0,
			Max:       1,
			ThenKey:   returnToNewHostPick,
			LiveZone:  game.ZoneBattlefield,
			Carry:     []uuid.UUID{aura, item.Controller},
			CarryWord: grantKey,
		})
		return nil
	}
}

// returnToNewHostPick is the answer: Carry is [the card, "your"
// controller], Word the grant key.
var returnToNewHostPick = game.RegisterChooseCardsThen("choose-cards/aura-returns-to-new-host-or-as-enchantment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) != 2 {
		return nil
	}
	aura, controller := r.Carry[0], r.Carry[1]
	if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if len(r.Picked) == 1 {
		return g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
			game.ZoneEntryOptions{Controller: controller, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]}},
			game.KeyedTail{})
	}
	name := ""
	if c, ok := g.LookupCardForEffect(aura); ok {
		name = c.Name
	}
	return returnAsNonAuraEnchantment(g, aura, controller, r.Chooser, r.Word, name)
})

func returnAsNonAuraEnchantment(g *game.Game, aura, controller, player uuid.UUID, grantKey, name string) error {
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: controller, EntersAs: &game.EntersAs{
			RemoveSubtypes: []string{"Aura"},
			Grants:         []string{grantKey},
			ChosenPlayer:   player,
			Label:          name + " — a non-Aura enchantment",
		}},
		game.KeyedTail{})
}

// ThatPlayersUpkeepDealsDamageGrant is the gained ability as a grant
// bundle: "At the beginning of that player's upkeep, this enchantment
// deals 1 damage to that player." "That player" is Card.ChosenPlayer on
// the permanent; a player who has left takes no upkeep.
func ThatPlayersUpkeepDealsDamageGrant(key string) AbilityGrant {
	text := "At the beginning of that player's upkeep, this enchantment deals 1 damage to that player."
	return AbilityGrant{
		Key: key,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				p := ChosenPlayerOf(g, source.InstanceID)
				return p != uuid.Nil && ev.Actor == p
			}, text, func(g *game.Game, item *game.StackItem) error {
				p := item.Trigger.Event.Actor
				return g.DealDamageToPlayerForEffect(item.SourceCardID, p, 1)
			}),
		},
		Text: text,
	}
}
