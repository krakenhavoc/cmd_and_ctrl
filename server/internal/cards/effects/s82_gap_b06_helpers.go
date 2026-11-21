package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b06_helpers.go — shared pieces of S82 GAP batch 6 (the Mystery
// Booster Commander planeswalkers, #5948): their tokens and a few bodies.

// --- tokens ---------------------------------------------------------

// gapB06FerozsBanToken is Feroz's Ban, the artifact Feroz, Ulgrotha's Warden
// makes: "Creature spells cost {2} more to cast." The cost modifier rides the
// template like Tamiyo's Notebook's.
func gapB06FerozsBanToken() game.Card { return tokenFromCatalog(printedGapB06FerozsBanToken) }

// printedGapB06FerozsBanToken is that token as PRINTED. Its slug is an on-disk
// identity.
func printedGapB06FerozsBanToken() tokenTemplate {
	return tokenTemplate{
		Slug: "ferozs-ban",
		Card: game.Card{
			Name:     "Feroz's Ban",
			TypeLine: "Token Artifact",
		},
		CostModifiers: []game.CostModifier{
			CostsMore(2, "Creature spells cost {2} more to cast.", CreatureSpell()),
		},
		Text: "Creature spells cost {2} more to cast.",
	}
}

// gapB06GiantBadgerToken is Greensleeves's Giant Badger: a 2/2 green Badger
// with "Whenever this creature blocks, it gets +2/+2 until end of turn."
func gapB06GiantBadgerToken() game.Card { return tokenFromCatalog(printedGapB06GiantBadgerToken) }

// printedGapB06GiantBadgerToken is that token as PRINTED. Its slug is an
// on-disk identity.
func printedGapB06GiantBadgerToken() tokenTemplate {
	return tokenTemplate{
		Slug: "giant-badger",
		Card: game.Card{
			Name:      "Giant Badger",
			TypeLine:  "Token Creature — Badger",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Giant Badger — it gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Giant Badger — +2/+2 until end of turn")),
		},
		Text: "Whenever this creature blocks, it gets +2/+2 until end of turn.",
	}
}

// gapB06LoyaltyCatToken is Worzel's Cat: a 1/1 white Cat with "{T}: Put a
// loyalty counter on each planeswalker you control."
func gapB06LoyaltyCatToken() game.Card { return tokenFromCatalog(printedGapB06LoyaltyCatToken) }

// printedGapB06LoyaltyCatToken is that token as PRINTED. Its slug is an
// on-disk identity.
func printedGapB06LoyaltyCatToken() tokenTemplate {
	return tokenTemplate{
		Slug: "cat-loyalty",
		Card: game.Card{
			Name:      "Cat",
			TypeLine:  "Token Creature — Cat",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Text: "{T}: Put a loyalty counter on each planeswalker you control.",
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{T}: Put a loyalty counter on each planeswalker you control.",
			Cost:    game.AbilityCost{Tap: true},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  gapB06LoyaltyOnEachWalker,
		}},
	}
}

// gapB06LoyaltyOnEachWalker puts a loyalty counter on each planeswalker the
// resolving ability's controller controls, as it resolves.
func gapB06LoyaltyOnEachWalker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsPlaneswalker() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
