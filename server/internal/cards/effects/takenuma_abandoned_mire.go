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
// Otawara's own comment declared its "{1} less per legendary
// creature" clause unreachable because activated-ability pricing did
// not exist. It does now (#1184, cost_modifier.go's `Activations`
// partition and `CostQuery.Ability`) — Boom Scholar is the first
// card to use it, and this is the second: a `CostModifier` scoped to
// Takenuma's own channel ability (`q.Card.InstanceID ==
// q.Source.InstanceID`) with a per-legendary-creature `Amount`.
//
// DECLARED SIMPLIFICATION, weaker than printed, and Boom Scholar's
// own: the discount is REAL at payment, but the ability's menu row
// still advertises the printed {3}{B} — the announce-time cost
// preview does not yet read CostModifiers for an activation.
func init() {
	Register(Spec{
		OracleID:     "ac2dd694-d2f1-4025-8400-12332bdc882a",
		Name:         "Takenuma, Abandoned Mire",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The activation discount applies when you pay for it, but the ability's menu row still lists the printed {3}{B}.",
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}",
			},
		},
		CostModifiers: []game.CostModifier{
			{
				Kind:        game.CostReduction,
				Activations: true,
				Label:       "This ability costs {1} less to activate for each legendary creature you control.",
				AppliesTo: func(q game.CostQuery) bool {
					return q.Ability != nil && q.Card.InstanceID == q.Source.InstanceID
				},
				Amount: func(q game.CostQuery) int {
					return takenumaLegendaryCreaturesControlledBy(q.Game, q.Controller)
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "Channel — {3}{B}, Discard this card: Mill three cards, then return a creature or planeswalker card from your graveyard to your hand.",
			Cost:  game.AbilityCost{Mana: "{3}{B}", DiscardSelf: true},
			Zones: []game.ZoneKind{game.ZoneHand},
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

// takenumaLegendaryCreaturesControlledBy counts the legendary
// creatures a player controls — the channel ability's cost
// reduction.
func takenumaLegendaryCreaturesControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.IsLegendary() {
			n++
		}
	}
	return n
}
