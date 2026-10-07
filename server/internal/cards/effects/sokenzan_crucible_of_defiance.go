package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sokenzan, Crucible of Defiance — Legendary Land:
//
//	"{T}: Add {R}.
//	 Channel — {3}{R}, Discard this card: Create two 1/1 colorless
//	 Spirit creature tokens. They gain haste until end of turn. This
//	 ability costs {1} less to activate for each legendary creature you
//	 control."
//
// Boseiju's channel shape (hand-zone activation, `DiscardSelf`, the
// legendary-creature discount) with Mardu Charm's token-then-grant
// resolution: the haste goes only to the two tokens this ability made,
// through the creation's continuation, so a Spirit already in play is
// untouched and a token that was replaced away is never granted
// anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5ee72d5-3a9e-4fe5-8802-3286ee612055",
		Name:         "Sokenzan, Crucible of Defiance",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:         "Channel — {3}{R}, Discard this card: Create two 1/1 colorless Spirit creature tokens. They gain haste until end of turn",
			Cost:          game.AbilityCost{Mana: "{3}{R}", DiscardSelf: true},
			Zones:         []game.ZoneKind{game.ZoneHand},
			CostModifiers: []game.CostModifier{ChannelDiscountPerLegendaryCreature()},
			Effect:        sokenzanChannel,
		}},
	})
}

func sokenzanChannel(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: TokenCard("1/1 colorless Spirit"), Count: 2}},
	}, func(g *game.Game, created []uuid.UUID) error {
		next := NewContext(g, item)
		for _, id := range created {
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Sokenzan, Crucible of Defiance"}).Apply(next); err != nil {
				return err
			}
		}
		return nil
	})
}
