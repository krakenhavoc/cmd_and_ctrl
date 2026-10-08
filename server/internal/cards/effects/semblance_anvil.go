package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Semblance Anvil — Artifact {3}:
//
//	"Imprint — When this artifact enters, you may exile a nonland card
//	 from your hand.
//	 Spells you cast that share a card type with the exiled card cost {2}
//	 less to cast."
//
// Chrome Mox's imprint (an ETB trigger with a pick-from-hand prompt,
// the exile finished by chromeMoxExileThePick) with a Semblance-specific
// candidate list: ANY nonland card, artifacts included. "The exiled
// card" is read back off the event log (b27ExiledWith), so a flickered
// Anvil forgets it (CR 400.7) and an Anvil that imprinted nothing
// discounts nothing — an unlinked Anvil must never discount every spell.
//
// "Share a card type" compares card types, not subtypes: a Creature
// imprint discounts every creature spell, an artifact creature imprint
// both artifact and creature spells. Cost reduction takes generic
// mana only, as every {N} less does, and never below zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbd0c406-c995-46e0-898d-375fdbede203",
		Name:         "Semblance Anvil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(semblanceAnvilImprintLabel, semblanceAnvilImprint),
		},
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Spells you cast that share a card type with the exiled card cost {2} less to cast.",
				YourSpell(), spellSharesTypeWithImprint),
		},
	})
}

// semblanceAnvilImprintLabel keys the "exiled with this artifact"
// record; it must match the label the trigger is built with.
const semblanceAnvilImprintLabel = "Semblance Anvil — exile a nonland card from your hand"

var semblanceCardTypes = []string{"artifact", "creature", "enchantment", "instant", "sorcery", "planeswalker", "battle", "kindred", "land"}

func semblanceAnvilImprint(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if !c.IsLand() {
				candidates = append(candidates, c.InstanceID)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: semblanceAnvilImprintLabel,
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then:     chromeMoxExileThePick,
	})
	return nil
}

// spellSharesTypeWithImprint passes when the spell shares a card type
// with a card imprinted on the modifier's source.
func spellSharesTypeWithImprint(q game.CostQuery) bool {
	for _, id := range b27ExiledWith(q.Game, q.Source.InstanceID, semblanceAnvilImprintLabel) {
		exiled, ok := q.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, t := range semblanceCardTypes {
			if exiled.HasCardType(t) && q.Card.HasCardType(t) {
				return true
			}
		}
	}
	return false
}
