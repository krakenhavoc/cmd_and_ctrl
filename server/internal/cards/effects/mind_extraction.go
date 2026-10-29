package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Extraction — Sorcery {2}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target player reveals their hand and discards all cards of each of
//	 the sacrificed creature's colors."
//
// The colours are the sacrificed creature's as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1). A
// card that shares at least one of them is discarded; a colourless
// sacrifice reveals the hand and discards nothing (the 2004-10-04
// rulings).
//
// "Discards all" names the cards, so there is nothing to choose: the
// discard prompt asks for exactly the matching cards and accepts only
// that set.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "0077740a-528b-4ee3-b331-fac321b95302",
		Name:           "Mind Extraction",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			player := item.Targets[0].ID
			p := ctx.Game.PlayerByIDForEffect(player)
			if p == nil {
				return nil
			}
			ctx.Game.RevealHandForEffect(player)
			info, ok := ctx.SacrificedPermanent()
			if !ok || len(info.Characteristic.Colors) == 0 {
				return nil
			}
			letters := mindExtractionLetters(info.Characteristic.Colors)
			then, ok := mindExtractionDiscards[letters]
			if !ok {
				return nil
			}
			n := 0
			for _, c := range p.Hand.Cards {
				if mindExtractionMatches(letters, c) {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   player,
				Source:   item.SourceCardID,
				N:        n,
				Question: "Mind Extraction — discard every card that shares a color with the sacrificed creature",
			}, then, nil)
		},
	})
}

// mindExtractionColorOrder is the five colors in the order a key spells them.
const mindExtractionColorOrder = "WUBRG"

// mindExtractionLetters spells a set of colors in WUBRG order: "" for none.
func mindExtractionLetters(colors []string) string {
	var out []byte
	for i := 0; i < len(mindExtractionColorOrder); i++ {
		letter := mindExtractionColorOrder[i : i+1]
		for _, col := range colors {
			if col == letter {
				out = append(out, letter[0])
				break
			}
		}
	}
	return string(out)
}

// mindExtractionMatches reports whether a card shares a color with the set
// spelled by letters.
func mindExtractionMatches(letters string, c game.Card) bool {
	for i := 0; i < len(letters); i++ {
		if c.HasColor(letters[i : i+1]) {
			return true
		}
	}
	return false
}

// mindExtractionDiscards is the keyed discard for each set of colors a
// sacrificed creature can have (#4253): "discards all cards of each of the
// sacrificed creature's colors" is a rule about the SET picked, and a set
// rule is a registered function of the cards alone, so there is one for each
// of the 31 non-empty color sets, under "choose-cards/mind-extraction-discard-"
// plus its letters in wubrg order, lowercase. On-disk keys: never renamed, never reused.
// The rule is the whole effect: the discard has nothing after it.
var mindExtractionDiscards = map[string]game.DiscardThen{}

func init() {
	for mask := 1; mask < 1<<len(mindExtractionColorOrder); mask++ {
		var letters string
		for i := 0; i < len(mindExtractionColorOrder); i++ {
			if mask&(1<<i) != 0 {
				letters += mindExtractionColorOrder[i : i+1]
			}
		}
		set := letters
		mindExtractionDiscards[set] = game.RegisterDiscardThenRule("choose-cards/mind-extraction-discard-"+strings.ToLower(set),
			func(picked []game.Card) bool {
				for _, c := range picked {
					if !mindExtractionMatches(set, c) {
						return false
					}
				}
				return true
			},
			func(*game.Game, game.DiscardsDone) error { return nil })
	}
}
