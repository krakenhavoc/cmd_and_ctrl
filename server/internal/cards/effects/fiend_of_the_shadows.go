package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiend of the Shadows — Creature — Vampire Wizard {3}{B}{B}, 3/3:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles a card from their hand. You may play that card for as long as it
//	 remains exiled.
//	 Sacrifice a Human: Regenerate this creature."
//
// The damaged player chooses the card (a floor of one; an empty hand exiles
// nothing, CR 609.3). The Fiend's controller, who is not the card's owner,
// is granted a play permission (a land counts, as does a spell: it is cast
// for its own cost, with the mana it normally takes)
// that lasts while the card stays in exile. The controller is carried on the
// prompt, so the answer needs no state beyond it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3891a3b6-d8ac-4cd3-b2f6-d6d786e89a67",
		Name:            "Fiend of the Shadows",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			fiendOfTheShadowsTrigger("Fiend of the Shadows — that player exiles a card from their hand; you may play it for as long as it remains exiled"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Human: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Human", HasSubtype("Human"))},
			Effect:  regenerateThisSource,
		}},
	})
}

// fiendOfTheShadowsTrigger is Fiend of the Shadows'
// trigger: the damaged player exiles a card from their hand and the
// trigger's controller may play it while it stays exiled.
func fiendOfTheShadowsTrigger(label string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		victim := item.Trigger.Event.Target
		p := g.PlayerByIDForEffect(victim)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    victim,
			FromPlayer: victim,
			Source:     item.SourceCardID,
			Question:   "Fiend of the Shadows — exile a card from your hand",
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    fiendOfTheShadowsExile,
			Carry:      []uuid.UUID{item.Controller},
			LiveZone:   game.ZoneHand,
		})
		return nil
	})
}

// fiendOfTheShadowsExile is the keyed answer: exile the pick, then let the
// Fiend's controller play it for as long as it remains exiled. An on-disk
// key: never renamed, never reused.
var fiendOfTheShadowsExile = game.RegisterChooseCardsThen("choose-cards/fiend-of-the-shadows-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	for _, id := range r.Picked {
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     r.Carry[0],
			Zone:       game.ZoneExile,
			Duration:   game.WhileInZoneDuration(),
			Source:     r.Source,
			SourceName: "Fiend of the Shadows",
			Label:      "You may play it for as long as it remains exiled",
		}, []game.Card{c})
	}
	return nil
})
