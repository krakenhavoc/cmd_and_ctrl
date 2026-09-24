package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Takenuma, Abandoned Mire — Legendary Land (EDHREC rank 244):
//
//	"{T}: Add {B}.
//	 Channel — {3}{B}, Discard this card: Mill three cards, then
//	 return a creature or planeswalker card from your graveyard to
//	 your hand. This ability costs {1} less to activate for each
//	 legendary creature you control."
//
// Otawara, Soaring City's shape (Channel — an ActivatedAbility with
// `Zones: {ZoneHand}` and `AbilityCost.DiscardSelf`, #660) with a
// different back half: a mill, then `ReturnPickedToHand`
// (graveyard_to_hand.go), the shared continuation Skullwinder's own
// return uses, over a candidate set filtered to creature-or-
// planeswalker.
//
// The "{1} less per legendary creature" clause is the channel
// ability's OWN cost clause (ActivatedAbility.CostModifiers, #1296),
// shared with Otawara and Boseiju through
// ChannelDiscountPerLegendaryCreature.
//
// It used to be a board modifier (Spec.CostModifiers with the
// #1184 `Activations` bit, scoped to its own ability), and a caveat
// said "the discount applies when you pay for it". It did not: a board
// modifier is gathered from the BATTLEFIELD (CR 113.6), and a channel
// ability is activated from the HAND, so the scan never found it and
// the channel always cost the printed {3}{B}. The ability's own slot
// travels with the ability, which is where CR 113.6 says it works.
// TestChannelDiscountAppliesFromTheHand pins it.
func init() {
	Register(Spec{
		OracleID:     "ac2dd694-d2f1-4025-8400-12332bdc882a",
		Name:         "Takenuma, Abandoned Mire",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}",
			},
		},
		Activated: []ActivatedAbility{{
			Label:         "Channel — {3}{B}, Discard this card: Mill three cards, then return a creature or planeswalker card from your graveyard to your hand.",
			Cost:          game.AbilityCost{Mana: "{3}{B}", DiscardSelf: true},
			Zones:         []game.ZoneKind{game.ZoneHand},
			CostModifiers: []game.CostModifier{ChannelDiscountPerLegendaryCreature()},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
					return err
				}
				candidates := takenumaGraveyardCandidates(g, item.Controller)
				if len(candidates) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    item.Controller,
					FromPlayer: item.Controller,
					Source:     item.SourceCardID,
					Question:   "Takenuma, Abandoned Mire — return a creature or planeswalker card from your graveyard to your hand",
					Cards:      candidates,
					Min:        1,
					Max:        1,
					Zone:       game.ZoneGraveyard,
					Then:       ReturnPickedToHand(item),
				})
				return nil
			},
		}},
	})
}

// takenumaGraveyardCandidates is "a creature or planeswalker card"
// from a player's own graveyard.
func takenumaGraveyardCandidates(g *game.Game, owner uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(owner)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() || c.IsPlaneswalker() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
