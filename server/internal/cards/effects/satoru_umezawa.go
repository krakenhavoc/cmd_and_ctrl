package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satoru Umezawa — Legendary Creature — Human Ninja {1}{U}{B}, 2/4:
//
//	"Whenever you activate a ninjutsu ability, look at the top three cards
//	 of your library. Put one of them into your hand and the rest on the
//	 bottom of your library in any order. This ability triggers only once
//	 each turn.
//	 Each creature card in your hand has ninjutsu {2}{U}{B}."
//
// The second ability is the first card to give an ABILITY to cards that
// are not on the battlefield (#3233, ADR 0093's amendment of 2026-10-13):
// a layer-6 static declared with GrantAbilitiesInZone(ZoneHand, …) names
// a bundle holding one Ninjutsu row, and the engine stamps it on every
// creature card in its controller's hand on each layer pass. The row is
// the card's own ability as far as the activation path, the wire's
// zone_abilities and the bot's legal moves are concerned, and it ends the
// pass after Satoru leaves the battlefield or loses its abilities.
//
// The first ability watches the activation event (game.Event.Label), so
// it triggers off a ninjutsu you activate from ANY card, a printed
// ninjutsu or a granted one, Satoru's own grant included. "Only once each
// turn" is a check before the trigger fires, as Curator of Sun's
// Creation's is (#936); it counts triggers of this object, so a second
// Satoru has its own.
//
// No simplification.
func init() {
	const label = "Satoru Umezawa — look at the top three cards, take one"
	Register(Spec{
		OracleID:     "ed1d21ba-ddfe-4f82-9e37-93f9e2c7ba33",
		Name:         "Satoru Umezawa",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       "satoru-umezawa/ninjutsu",
			Zone:      game.ZoneHand,
			Activated: []ActivatedAbility{Ninjutsu("{2}{U}{B}")},
			Text:      "Ninjutsu {2}{U}{B}",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneHand, CardsYouOwn(creatureCard), "satoru-umezawa/ninjutsu"),
				"Each creature card in your hand has ninjutsu {2}{U}{B}."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, AllOf(ByYou, ANinjutsuAbility, satoruFirstThisTurn(label)), label,
				LookAtTopThenTakeOneRestOnBottom(3, "Satoru Umezawa — put one into your hand")),
		},
	})
}

// creatureCard is the "creature card" of a zone grant: a card whose
// printed types include creature. Off the battlefield Effective() is the
// printed characteristic, so this reads the card as it is in the zone.
func creatureCard(card *game.Card) bool { return card.IsCreature() }

// satoruFirstThisTurn is "this ability triggers only once each turn".
func satoruFirstThisTurn(label string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return source != nil && !b11TriggeredThisTurn(g, source.InstanceID, label)
	}
}
