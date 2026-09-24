package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gonti, Night Minister — Legendary Creature — Aetherborn Rogue
// {2}{B}{B}, 3/4 (Edea steal-and-sac deck, #1565):
//
//	"Whenever a player casts a spell they don't own, that player
//	 creates a Treasure token.
//	 Whenever a creature deals combat damage to one of your opponents,
//	 its controller looks at the top card of that opponent's library
//	 and exiles it face down. They may play that card for as long as
//	 it remains exiled. Mana of any type can be spent to cast a spell
//	 this way."
//
// Both triggers are symmetric in a way that is easy to miss, and both
// are the printed card:
//
//   - The Treasure goes to THE PLAYER WHO CAST the spell, whoever that
//     is, not to Gonti's controller. An opponent casting your card off
//     their own Hostage Taker makes a Treasure.
//   - The damage trigger watches ANY creature, and the card goes to
//     THAT CREATURE'S CONTROLLER. The only condition on Gonti's side
//     is that the damaged player is one of Gonti's controller's
//     opponents. One opponent's creature hitting another opponent
//     hands that creature's controller the top card, not you. The
//     controller is read as the creature had it when the damage was
//     dealt, and fixed when the trigger goes on the stack.
//
// Each combat damage event is one trigger, so three creatures
// connecting are three cards, as printed ("whenever a creature",
// not "one or more").
//
// "Play", not "cast": a land exiled this way can be played as a land
// drop. The window is for as long as the card stays in exile.
//
// DECLARED SIMPLIFICATIONS:
//
//   - The card is exiled FACE UP, so every player sees it. The engine
//     has no face-down exile that only the permission holder may look
//     at. The holder already knows it, so the card is not stronger than
//     printed, but the table sees more than it should.
//   - "Mana of any TYPE" is the engine's "as though it were mana of any
//     colour" permission, which does not cover colourless. A {C} in the
//     exiled card's cost still needs colourless mana.
func init() {
	Register(Spec{
		OracleID:     "e16f79c1-ffbb-4893-b62a-d4e9d15e2b16",
		Name:         "Gonti, Night Minister",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The card is exiled face up, so every player can see it, not only the player who may play it.",
			"A card whose mana cost includes {C} still needs colorless mana for that part; other mana can pay only its colored and generic parts.",
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: gontiCastASpellTheyDontOwn,
				Key:       "Gonti, Night Minister — that player creates a Treasure token",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					caster := ev.Actor
					return game.NewTriggeredItem(source, "Gonti, Night Minister — that player creates a Treasure token",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{Controller: caster, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: gontiCombatDamageToYourOpponent,
				Key:       "Gonti, Night Minister — exile the top card of that opponent's library",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					victim := ev.Target
					grantee := uuid.Nil
					if c, ok := g.LookupCardForEffect(ev.Source); ok {
						grantee = c.Controller
					}
					return game.NewTriggeredItem(source, "Gonti, Night Minister — exile the top card of that opponent's library",
						func(g *game.Game, _ *game.StackItem) error {
							return gontiExileTopForPlay(g, victim, grantee, 1)
						})
				},
			},
		},
	})
}

// gontiCastASpellTheyDontOwn — any player cast a spell whose owner is
// somebody else.
func gontiCastASpellTheyDontOwn(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner != ev.Actor
}

// gontiCombatDamageToYourOpponent — any creature dealt combat damage
// to a player who is an opponent of Gonti's controller.
func gontiCombatDamageToYourOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Target == source.Controller {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Target); p == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature()
}

// gontiExileTopForPlay exiles the top n cards of `from`'s library and
// lets `grantee` play them for as long as they remain exiled, spending
// mana as though it were any colour. Shared with Outrageous Robbery.
func gontiExileTopForPlay(g *game.Game, from, grantee uuid.UUID, n int) error {
	if grantee == uuid.Nil || n <= 0 {
		return nil
	}
	_, err := g.ExileTopWithPermissionForEffect(from, grantee, n, game.CastPermission{
		AnyColor: true,
		Duration: game.WhileInZoneDuration(),
	})
	return err
}
