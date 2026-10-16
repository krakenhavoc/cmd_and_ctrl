package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Prowler — Creature — Wolf {1}{G}{G}, 3/4:
//
//	"Vigilance
//	 Whenever this creature enters or attacks, exile a card from a
//	 graveyard.
//	 Spells you cast cost {1} less to cast for each card type they share
//	 with cards exiled with this creature."
//
// The exile is not targeted: the controller picks one card from any
// graveyard as the trigger resolves (nothing happens with every graveyard
// empty) and it is exiled LINKED to this Prowler's trigger
// (ExileTarget.Linked, CR 607.2a), so cards another effect exiled, or
// that this Prowler exiled in an earlier life (CR 400.7), are not its
// cards. The cost reduction is one CostsLessEach: the card types the spell
// shares with ANY of the cards still exiled with the Prowler, each type
// counted once however many exiled cards have it ("for each card type",
// not for each card). Generic mana only, never below zero, applied to
// every spell its controller casts, the Prowler's own cast excluded
// because it is not on the battlefield yet.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa53cc50-f06f-44b0-850f-cfa6e0caa4c5",
		Name:            "Cemetery Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks(cemeteryProwlerLabel, cemeteryProwlerExile),
		},
		CostModifiers: []game.CostModifier{
			CostsLessEach(cemeteryProwlerDiscount,
				"Spells you cast cost {1} less to cast for each card type they share with cards exiled with this creature.",
				YourSpell()),
		},
	})
}

// cemeteryProwlerLabel is the trigger's stack label and Key: the catalog
// row the linked exile names.
const cemeteryProwlerLabel = "Cemetery Prowler — exile a card from a graveyard"

// cemeteryProwlerExile queues the pick and exiles it, linked.
func cemeteryProwlerExile(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, p := range g.Seats {
		if p != nil {
			candidates = append(candidates, graveyardCardsOfPlayer(g, p.ID)...)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	resolving := item
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: cemeteryProwlerLabel,
		Cards:    candidates,
		Min:      1,
		Max:      1,
		// No Zone: the candidates span every seat's graveyard.
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ExileTarget{Target: picked[0], Linked: true}.Apply(NewContext(g, resolving))
		},
	})
	return nil
}

// cemeteryProwlerDiscount is the number of card types the spell shares
// with the cards exiled with the modifier's source by its trigger.
func cemeteryProwlerDiscount(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	ref := game.PermissionCardRef{ID: q.Source.InstanceID, Epoch: q.Source.ObjectEpoch}
	exiled := q.Game.ExiledWithRowForEffect(ref, game.AbilitySlotTriggered, cemeteryProwlerLabel)
	n := 0
	for _, t := range semblanceCardTypes {
		if !q.Card.HasCardType(t) {
			continue
		}
		for _, c := range exiled {
			if c.HasCardType(t) {
				n++
				break
			}
		}
	}
	return n
}
