package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent-Blade Oni — Creature — Demon Ninja {3}{U}{U}{B}{B}, 6/5 (Outlaws of
// Thunder Junction Commander):
//
//	"Ninjutsu {4}{U}{B}
//	 Whenever this creature deals combat damage to a player, look at that
//	 player's hand. You may cast a spell from among those cards without
//	 paying its mana cost."
//
// The looker becomes a knower of every card in the damaged player's hand
// (CR 701.20; "look at" is not "reveal"), then the free cast is the
// resolution_cast offer every "cast a spell from among them without paying
// its mana cost" shares (Aetherflux Conduit's grant over a hand, here over
// the other player's): free, any timing during the resolution (CR 608.2g),
// pass-closed, one spell, the declined cards staying where they are. The
// spell is cast by you, so you pay any additional cost and choose its
// targets and modes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f515250-0f1f-4fb7-aa75-b703b81ccd4a",
		Name:         "Silent-Blade Oni",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{4}{U}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Silent-Blade Oni — look at that player's hand; you may cast a spell from among those cards without paying its mana cost",
				cmdr2S45SilentBladeOniLook),
		},
	})
}

// cmdr2S45SilentBladeOni is the free-cast offer; its key is on disk: never
// renamed, never reused.
var cmdr2S45SilentBladeOni = registerFreeHandCastOffer("choose-cards/cmdr2-s45-silent-blade-oni", "Silent-Blade Oni")

func cmdr2S45SilentBladeOniLook(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := g.PlayerByIDForEffect(item.Trigger.Event.Target)
	if victim == nil || victim.Hand == nil {
		return nil
	}
	cards := make([]uuid.UUID, 0, len(victim.Hand.Cards))
	for i := range victim.Hand.Cards {
		victim.Hand.Cards[i].AddKnower(item.Controller)
		cards = append(cards, victim.Hand.Cards[i].InstanceID)
	}
	offerFreeHandCastOver(g, item.Controller, item.SourceCardID, cmdr2S45SilentBladeOni,
		"Silent-Blade Oni — cast a spell from among those cards without paying its mana cost?", anySpell, cards)
	return nil
}
