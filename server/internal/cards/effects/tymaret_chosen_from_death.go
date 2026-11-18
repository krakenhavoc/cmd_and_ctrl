package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tymaret, Chosen from Death — {B}{B} Legendary Enchantment Creature —
// Demigod, 2/*:
//
//	"Tymaret's toughness is equal to your devotion to black.
//	 {1}{B}: Exile up to two target cards from graveyards. You gain 1 life
//	 for each creature card exiled this way."
//
// The toughness is a layer 7a CDA over devotionTo (Daxos's shape); her own
// {B}{B} counts, so she is at least a 2/2 on the battlefield. The ability
// is Rooftop Percher's "up to two target cards from graveyards" with
// Graveyard Trespasser's per-creature-card tail: the life counts the
// creature cards that actually reached exile, read as they were in the
// graveyard (CR 608.2b skips a target that left in response), and no
// creature card means no life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64b4e91f-83a4-4b3f-87d4-8c48907e628d",
		Name:         "Tymaret, Chosen from Death",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Toughness is equal to your devotion to black.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Toughness = devotionTo(g, source.Controller, source.TextColor("B"))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Exile up to two target cards from graveyards. You gain 1 life for each creature card exiled this way.",
			Cost:    ManaCost("{1}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Targets: TargetCardInGraveyard("up to two target cards from graveyards").WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetCardsThen(NewContext(g, item), tymaretGainPerCreature)
			},
		}},
	})
}

// tymaretGainPerCreature is "you gain 1 life for each creature card exiled
// this way". An on-disk key.
var tymaretGainPerCreature = GraveyardExileTail("tymaret-chosen-from-death/gain-per-creature", func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
	n := creatureCardsAmong(exiled, wasCreature)
	if n == 0 {
		return nil
	}
	return GainLife{Amount: n}.Apply(ctx)
})
