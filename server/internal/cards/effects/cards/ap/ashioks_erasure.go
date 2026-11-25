package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok's Erasure — Enchantment {2}{U}{U} (Theros Beyond Death, #5514):
//
//	"Flash
//	 When this enchantment enters, exile target spell.
//	 Your opponents can't cast spells with the same name as the exiled card.
//	 When this enchantment leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// Spell Queller's two triggers over one linked exile: the enters trigger
// exiles the spell with the link (Card.ExiledWith names this OBJECT,
// CR 400.7), without countering it, so a spell that can't be countered is
// exiled all the same; the leave trigger reads the link off the source's
// last-known object and returns the card to its OWNER's hand. The cast
// restriction is a static of the enchantment (Iona's shape), read live from
// the link: each opponent of its controller can't cast a spell whose name is
// the exiled card's, and it lifts the moment the enchantment is gone.
//
// No simplification.
func init() {
	exile := effects.Targeting(
		effects.WhenThisEnters("Ashiok's Erasure — exile target spell", effects.GapExileTargetSpellLinked),
		effects.TargetSpell("target spell"))
	leave := effects.On(game.EventLTB, effects.Self, "Ashiok's Erasure — return the exiled card to its owner's hand",
		effects.B36ReturnCardsExiledWithToOwnersHands)
	effects.Register(effects.Spec{
		OracleID:        "13cf3ef7-d88e-4cc7-a4d8-cedc671f356a",
		Name:            "Ashiok's Erasure",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{exile, leave},
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells with the same name as the exiled card.",
			Forbids: func(q game.CastQuery) bool {
				if q.Game == nil || q.Controller == q.Source.Controller {
					return false
				}
				for _, c := range q.Game.ExiledWithForEffect(game.PermissionCardRef{ID: q.Source.InstanceID, Epoch: q.Source.ObjectEpoch}) {
					if c.Name != "" && (game.PermissionFilter{Named: c.Name}).Matches(q.Card) {
						return true
					}
				}
				return false
			},
		}},
	})
}
