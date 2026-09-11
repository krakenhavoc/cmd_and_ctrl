package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scavenger Grounds — Land — Desert:
//
//	"{T}: Add {C}."
//	"{2}, {T}, Sacrifice a Desert: Exile all graveyards."
//
// The colourless graveyard-hate land: it costs a deck nothing to
// run and answers a whole archetype once. Registered for the usual
// nonbasic-land reason (no BASIC supertype, so nothing derives its
// mana) and for the ability, which is the point of the card.
//
// The cost is three components at once and every one of them is a
// real AbilityCost field today: {2} mana, {T}, and a sacrifice
// matched against a spec. Scavenger Grounds is ITSELF a Desert and
// sacrificing itself is the normal line — AbilityCost.SacrificeOther
// admits the source when the spec does, exactly as Carrion Feeder
// eats itself. The tap and the sacrifice are independent costs, so
// the land must be untapped to activate even though it is about to
// die.
//
// "A Desert" reads the post-layer subtype rather than the card name,
// so a Desert that arrived as a copy of something else, or a
// hypothetical nonland Desert, is still one. Same posture as
// isTreasure.
//
// "Exile all graveyards" is every card in every player's graveyard,
// not just opponents' — the controller's own graveyard goes too, and
// that symmetry is why the card is played in decks that do not care
// about their own yard. The IDs are snapshotted before the first
// exile because ExileCardForEffect mutates the zone slices underneath
// the walk.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "5ece7d03-9ee7-4953-a06e-9d8e41874903",
		Name:     "Scavenger Grounds",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice a Desert: Exile all graveyards.",
			Cost: Plus(
				ManaCost("{2}"),
				TapCost(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("a Desert", isDesert)},
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var doomed []uuid.UUID
				for _, p := range g.Seats {
					if p == nil || p.Graveyard == nil {
						continue
					}
					for _, c := range p.Graveyard.Cards {
						doomed = append(doomed, c.InstanceID)
					}
				}
				for _, id := range doomed {
					if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

// isDesert is the sacrifice-cost predicate for "Sacrifice a Desert".
// Reads the post-layer subtype, not the name, so a Desert by way of
// a copy or a type-adding effect counts.
func isDesert(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return hasSubtype(c, "Desert")
}
