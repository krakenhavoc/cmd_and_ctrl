package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eerie_cards_test.go — ADR 0103 PR 4: the Duskmourn eerie creatures.
// Every card is driven through BOTH of its real triggers: a real cast
// of an enchantment spell, and a real unlock of a Room's second door.

const EerieRoomOracle = "eerie-test-room-oracle"

// eerieAnswers are the choices eerieSettle gives any prompt it meets.
type EerieAnswers struct {
	Target game.TargetRef // a pick_target prompt; zero means the first candidate
	Card   uuid.UUID      // a choose_cards prompt; zero means the first candidate
	Option int            // an option_pick prompt
}

// eerieSettle resolves the stack, answering trigger order, trigger
// targets, surveils (keep everything on top), option picks and card
// picks for whoever is asked. A trigger prompt (a "you may") is left
// pending for the test to answer.
func EerieSettle(t *testing.T, g *game.Game, a EerieAnswers) {
	t.Helper()
	for i := 0; i < 96; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerPrompt:
				return
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoicePickTarget:
				ref := a.Target
				if ref.ID == uuid.Nil {
					switch {
					case len(c.PickTargetCards) > 0:
						ref = game.TargetRef{Kind: game.TargetCard, ID: c.PickTargetCards[0]}
					case len(c.PickTargetPlayers) > 0:
						ref = game.TargetRef{Kind: game.TargetPlayer, ID: c.PickTargetPlayers[0]}
					default:
						t.Fatalf("a target prompt with no candidates: %+v", c)
					}
				}
				if err := g.ResolvePickTarget(c.ID, c.Chooser, ref); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				answered = true
			case game.PendingChoiceSurveil:
				if err := g.ResolveSurveil(c.ID, c.Chooser, nil, c.ScryCards); err != nil {
					t.Fatalf("ResolveSurveil: %v", err)
				}
				answered = true
			case game.PendingChoiceOptionPick:
				if err := g.ResolveOptionPick(c.ID, c.Chooser, a.Option); err != nil {
					t.Fatalf("ResolveOptionPick: %v", err)
				}
				answered = true
			case game.PendingChoiceChooseCards:
				pick := append([]uuid.UUID(nil), c.ChooseCards...)
				for _, id := range c.ChooseCards {
					if id == a.Card {
						pick = []uuid.UUID{id}
					}
				}
				n := c.ChooseMax
				if n <= 0 || n > len(pick) {
					n = len(pick)
				}
				if err := g.ResolveChooseCards(c.ID, c.Chooser, pick[:n]); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// eerieEnchantment casts a plain enchantment spell and settles.
func EerieEnchantment(t *testing.T, g *game.Game, a EerieAnswers) {
	t.Helper()
	CastCatalogSpell(t, g, "Test Aura", "Enchantment", "", nil)
	EerieSettle(t, g, a)
}

// eerieUnlock fully unlocks a Room seat 0 controls by taking the unlock
// special action on its second door, and settles.
func eerieUnlock(t *testing.T, g *game.Game, a EerieAnswers) {
	t.Helper()
	me := g.Seats[0]
	id := RoomsDOnBattlefield(g, me.ID, EerieRoomOracle, "Left Door", "Right Door", game.DoorBit(game.DoorLeft))
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight}); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	EerieSettle(t, g, a)
}

// forEachEerieSource runs `run` twice on a fresh table in the main
// phase, once per eerie condition. fire() raises that condition.
func ForEachEerieSource(t *testing.T, run func(t *testing.T, g *game.Game, fire func(EerieAnswers))) {
	t.Helper()
	t.Run("an enchantment enters", func(t *testing.T) {
		g := NewCatalogGame(t)
		AdvanceToMain(t, g)
		run(t, g, func(a EerieAnswers) { EerieEnchantment(t, g, a) })
	})
	t.Run("a Room is fully unlocked", func(t *testing.T) {
		g := NewCatalogGame(t)
		AdvanceToMain(t, g)
		run(t, g, func(a EerieAnswers) { eerieUnlock(t, g, a) })
	})
}

// eerieBody puts a catalog creature on seat 0's battlefield.
func EerieBody(g *game.Game, name, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	me := g.Seats[0]
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: me.ID, Controller: me.ID,
	})
}

func EerieDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

func EerieTokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name && effects.IsToken(c) {
			out = append(out, c)
		}
	}
	return out
}
