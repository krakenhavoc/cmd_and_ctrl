package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b09_helpers.go — shared pieces of Crimson Vow batch 9 (training and
// +1/+1 counters, #5495).

// vowB09TrainingThen is training whose resolving ability goes on to open a
// second ability: "Whenever this creature trains" (CR 702.149c: "when a
// resolving training ability puts one or more +1/+1 counters on this
// creature"). The trainee's counter is the shared trainingCounter's, and the
// "trains" ability is a CR 603.12 reflexive trigger created only when the
// counter really went on (the trainee still the same object on the
// battlefield), built like exploit.go's.
func VowB09TrainingThen(label string, payoff game.BodyRef) game.TriggeredAbility {
	return On(game.EventAttack, AllOf(ThisAttacked, attacksWithGreaterPower), label,
		func(g *game.Game, item *game.StackItem) error {
			if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
				return nil
			}
			if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 1); err != nil {
				return err
			}
			return ReflexiveTrigger{Label: label + ": it trained", Body: payoff}.Apply(NewContext(g, item))
		})
}

// vowB09HasPlusOneCounter is "creature ... with a +1/+1 counter on it".
func VowB09HasPlusOneCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Counters[game.CounterPlusOne] > 0
	}
}

// vowB09YouPutPlusOneThisTurn is "you've put one or more +1/+1 counters on a
// creature this turn", read off this turn's event log. Caller holds g.mu (a
// static's AppliesTo does).
func VowB09YouPutPlusOneThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventCounterPlaced && ev.Label == game.CounterPlusOne &&
			counterPlacedBy(ev, you, g) && wasACreatureWhenCountered(ev, g) && CounterPlacementDelta(ev, g) > 0 {
			return true
		}
	}
	return false
}

// vowB09GrantUntilEOTOnTarget is "target creature gains <keywords> until end
// of turn" for the first legal target.
func VowB09GrantUntilEOTOnTarget(label string, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalCardTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: label}.Apply(ctx)
	}
}

// HumanSoldierTrainingToken is the 1/1 green and white Human Soldier creature
// token with training (Torens, Fist of the Angels).
func HumanSoldierTrainingToken() game.Card { return TokenFromCatalog(printedHumanSoldierTrainingToken) }

func printedHumanSoldierTrainingToken() TokenTemplate {
	return TokenTemplate{
		Slug: "human-soldier-training",
		Card: game.Card{
			Name:      "Human Soldier",
			TypeLine:  "Token Creature — Human Soldier",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G", "W"},
		},
		Triggered: []game.TriggeredAbility{Training("Human Soldier — training: put a +1/+1 counter on it")},
		Text:      "Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)",
	}
}
