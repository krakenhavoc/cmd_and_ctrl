package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Entropic Specter — Creature — Specter Spirit {3}{B}{B}, */*:
//
//	"Flying
//	 As this creature enters, choose an opponent.
//	 Entropic Specter's power and toughness are each equal to the
//	 number of cards in the chosen player's hand.
//	 Whenever this creature deals damage to a player, that player
//	 discards a card."
//
// Three printed pieces, all on existing machinery. The opponent is an
// as-enters choice stored on the permanent (ChoosePlayerAsEnters). The
// size is a layer 7a characteristic-defining ability reading that
// stored seat's hand live, with Psychosis Crawler's DependsOnHandSize
// hint so the layer cache is dropped when a hand moves. And the
// discard is a trigger on ANY damage to a player, combat or not, the
// Specter's controller included; the player who was dealt the damage
// chooses what to discard.
//
// With no one chosen yet the ability defines nothing and the Specter
// is the importer's `*`; the choice is the first thing that happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9c44443-85c7-4d0a-82ab-36137151e4e5",
		Name:            "Entropic Specter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        ChoosePlayerAsEnters("Entropic Specter", Opponents),
		Static: []game.StaticAbility{{
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7A_CDA,
			DependsOnHandSize: true,
			// Until the as-enters answer arrives the ability applies to
			// nothing: the Specter keeps its printed `*` (which the
			// state-based actions skip, #690) rather than reading a
			// 0 toughness off nobody's hand and dying before its
			// controller can choose.
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && ChosenPlayerOf(g, source.InstanceID) != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := cardsInHand(g, ChosenPlayerOf(g, source.InstanceID))
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtDamageToAPlayer,
				"Entropic Specter — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					victim := NewContext(g, item).Trigger().Event.Target
					return discardOne(g, item, victim)
				}),
		},
	})
}

// discardOne makes `player` discard a card of their choice — a
// hellbent player discards nothing.
func discardOne(g *game.Game, item *game.StackItem, player uuid.UUID) error {
	if cardsInHand(g, player) == 0 {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   item.SourceCardID,
		N:        1,
		Question: "Discard a card",
	})
	return nil
}
