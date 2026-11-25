package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_returns_as_enchantment.go — #6235: Takklemaggot.
//
//	"When enchanted creature dies, that creature's controller chooses a
//	 creature that this card could enchant. If the player does, return
//	 this card to the battlefield under your control attached to that
//	 creature. If they don't, return this card to the battlefield under
//	 your control as a non-Aura enchantment. It loses "enchant creature"
//	 and gains "At the beginning of that player's upkeep, this
//	 enchantment deals 1 damage to that player.""
//
// The engine half is game/aura_returns_as_enchantment.go. A card wires
// it as:
//
//	Targets:   EnchantCreature(),
//	Grants:    []AbilityGrant{ChosenPlayerUpkeepDamageGrant("<card>/upkeep-damage", 1, "<name>")},
//	Triggered: []game.TriggeredAbility{
//		<its own upkeep trigger>,
//		AuraReturnsOnHostDeath("<name>", "<card>/upkeep-damage"),
//	},

// ChosenPlayerUpkeepDamageGrant is the gained "At the beginning of that
// player's upkeep, this enchantment deals N damage to that player." as a
// grant bundle: "that player" is the permanent's ChosenPlayer, which
// AuraReturnsOnHostDeath names as the card returns.
func ChosenPlayerUpkeepDamageGrant(key string, n int, name string) AbilityGrant {
	text := "At the beginning of that player's upkeep, this enchantment deals 1 damage to that player."
	if n != 1 {
		text = "At the beginning of that player's upkeep, this enchantment deals " + strconv.Itoa(n) + " damage to that player."
	}
	return AbilityGrant{
		Key:  key,
		Text: text,
		Triggered: []game.TriggeredAbility{On(game.EventBeginUpkeep,
			func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ChosenPlayer != uuid.Nil && ev.Actor == source.ChosenPlayer
			},
			name+" — deals "+strconv.Itoa(n)+" damage to that player",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := item.Trigger.Event.Actor
				if player == uuid.Nil {
					return nil
				}
				d := DealDamage{Source: item.SourceCardID, Target: player, Amount: n}
				if ref, ok := ctx.SourceRef(); ok {
					d.SourceObject = &ref
				}
				return d.Apply(ctx)
			})},
	}
}

// AuraReturnsOnHostDeath is the dies trigger for the card named `name`:
// the dead creature's controller (as it last existed, CR 608.2h) chooses
// a creature this card could enchant, and the card returns from its
// owner's graveyard under the trigger's controller's control attached to
// it, or, with no such creature, as a non-Aura enchantment that loses
// enchant and gains the bundle `grantKey` with that player as its
// ChosenPlayer. A card that has left the graveyard doesn't return.
func AuraReturnsOnHostDeath(name, grantKey string) game.TriggeredAbility {
	return WhenEnchantedCreatureDies(name+" — that creature's controller chooses a creature this card could enchant",
		func(g *game.Game, item *game.StackItem) error {
			ev := item.Trigger.Event
			chooser := ev.LastKnownController
			if chooser == uuid.Nil {
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					chooser = c.Controller
				}
			}
			aura := item.SourceCardID
			if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			hosts := g.CreaturesAuraCardCouldEnchantForEffect(aura, item.Controller)
			if len(hosts) == 0 || chooser == uuid.Nil || g.PlayerByIDForEffect(chooser) == nil {
				return returnAsNonAuraEnchantment(g, aura, item.Controller, chooser, grantKey)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:   chooser,
				Source:    aura,
				Question:  name + " — choose a creature it could enchant; it returns attached to that creature",
				Cards:     hosts,
				Min:       1,
				Max:       1,
				ThenKey:   auraReturnsOnHostDeathPick,
				Carry:     []uuid.UUID{aura, item.Controller},
				CarryWord: grantKey,
			})
			return nil
		})
}

// auraReturnsOnHostDeathPick is the chosen creature: the Aura returns
// attached to it, or as a non-Aura enchantment when nothing was picked
// (the creature is gone by the answer).
var auraReturnsOnHostDeathPick = game.RegisterChooseCardsThen("choose-cards/aura-returns-on-host-death", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 {
		return nil
	}
	aura, controller := r.Carry[0], r.Carry[1]
	if z := g.FindCardZoneForEffect(aura); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if len(r.Picked) == 0 || !OnBattlefield(g, r.Picked[0]) {
		return returnAsNonAuraEnchantment(g, aura, controller, r.Chooser, r.Word)
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: controller, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]}},
		game.KeyedTail{})
})

// returnAsNonAuraEnchantment returns the Aura card `aura` from its
// graveyard under `controller`'s control as a non-Aura enchantment that
// has lost enchant and gained the bundle `grantKey`, naming `player`.
func returnAsNonAuraEnchantment(g *game.Game, aura, controller, player uuid.UUID, grantKey string) error {
	as := &game.EntersAs{
		Types:        []string{"Enchantment"},
		LoseSubtypes: []string{"Aura"},
		LoseKeywords: []string{"enchant"},
		ChosenPlayer: player,
		Label:        "returned as a non-Aura enchantment",
	}
	if grantKey != "" {
		as.Grants = []string{grantKey}
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: aura, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: controller, EntersAs: as},
		game.KeyedTail{})
}
