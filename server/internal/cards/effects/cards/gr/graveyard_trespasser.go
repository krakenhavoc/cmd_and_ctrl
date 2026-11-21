package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graveyard Trespasser // Graveyard Glutton — {2}{B} Creature — Human
// Werewolf 3/3 // Creature — Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "Ward—Discard a card.
//	        Whenever this creature enters or attacks, exile up to one
//	        target card from a graveyard. If a creature card was exiled
//	        this way, each opponent loses 1 life and you gain 1 life.
//	        Daybound"
//	Back:  "Ward—Discard a card.
//	        Whenever this creature enters or attacks, exile up to two
//	        target cards from graveyards. For each creature card exiled
//	        this way, each opponent loses 1 life and you gain 1 life.
//	        Nightbound"
//
// The exile and the drain are Soul-Shackled Zombie's: "exiled this way"
// is what actually reached exile, and the type is read before the move.
// The front targets one card in any graveyard, the back up to two cards
// in any graveyards (not "a single graveyard").
//
// Ward—Discard a card is Ward(WardDiscard(1)) (#3172, ward.go): the
// spell's controller is asked to discard a card of their choice or the
// spell is countered, and a caster with an empty hand is countered
// without a prompt. Both faces carry it, as printed.
func init() {
	const oracle = "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Graveyard Trespasser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Graveyard Trespasser — ward, discard a card"),
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Graveyard Trespasser — exile up to one target card from a graveyard",
					graveyardExileAndDrain(false)),
				effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Graveyard Glutton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Graveyard Glutton — ward, discard a card"),
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Graveyard Glutton — exile up to two target cards from graveyards",
					graveyardExileAndDrain(true)),
				effects.TargetCardInGraveyard("up to two target cards from graveyards").WithCount(0, 2)),
		},
	})
}

// graveyardExileAndDrain exiles the still-legal targets and then drains
// each opponent for 1: once if any creature card was exiled (perCard
// false), once per creature card exiled (perCard true).
func graveyardExileAndDrain(perCard bool) effects.Effect {
	then := graveyardExileDrainOnce
	if perCard {
		then = graveyardExileDrainPerCreature
	}
	return func(g *game.Game, item *game.StackItem) error {
		return effects.ExileTargetCardsThen(effects.NewContext(g, item), then)
	}
}

// graveyardExileDrainOnce and graveyardExileDrainPerCreature are the drain
// after the exile (#4253 slice 9): once if any creature card was exiled,
// or once per creature card.
var (
	graveyardExileDrainOnce = effects.GraveyardExileTail("graveyard-exile/drain-once-if-creature", func(ctx *effects.Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		if effects.CreatureCardsAmong(exiled, wasCreature) == 0 {
			return nil
		}
		return effects.B40DrainEachOpponent(1)(ctx.Game, ctx.Item)
	})
	graveyardExileDrainPerCreature = effects.GraveyardExileTail("graveyard-exile/drain-per-creature", func(ctx *effects.Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		n := effects.CreatureCardsAmong(exiled, wasCreature)
		if n == 0 {
			return nil
		}
		return effects.B40DrainEachOpponent(n)(ctx.Game, ctx.Item)
	})
)
