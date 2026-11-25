package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// tap_others_lands_test.go — #6204, Piracy: "Until end of turn, you may
// tap lands you don't control for mana. Spend this mana only to cast
// spells." (CR 602.2: the effect that "specifically says otherwise";
// CR 605.3a.)

func mountainShape() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: "{R}", Label: "Add {R}"}}
}

// piracyBoard is seat 0 in its main phase, seat 1 controlling a Mountain.
func piracyBoard(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g, me := restrictedBoard(t)
	mountain := pushIntrinsicPermanent(g, g.Seats[1], "Mountain", "Basic Land — Mountain", mountainShape(), nil)
	return g, me, mountain
}

func grantPiracy(t *testing.T, g *Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.TapLandsYouDontControlThisTurnForEffect(uuid.Nil, player, "Piracy") {
			t.Fatal("no record was registered")
		}
	})
}

// Without Piracy another player's land is theirs to tap (CR 602.2).
func TestTappingAnotherPlayersLandIsRefusedWithoutPiracy(t *testing.T) {
	g, me, mountain := piracyBoard(t)
	if err := g.ActivateManaAbility(me.ID, mountain, 0, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("tapping an opponent's Mountain = %v, want ErrCardCallerMismatch", err)
	}
}

// With Piracy the {R} goes to the activator's pool, restricted to casting
// spells: it can pay for a spell and not for an activated ability.
func TestPiracyTapsAnOpponentsMountainForSpellOnlyMana(t *testing.T) {
	g, me, mountain := piracyBoard(t)
	grantPiracy(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, mountain, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the opponent's Mountain under Piracy: %v", err)
	}
	if !tappedOnBattlefield(t, g, mountain) {
		t.Fatal("the Mountain is untapped")
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("activator's pool = %d, want the {R}", got)
	}
	if got := g.Seats[1].ManaPool.Total(); got != 0 {
		t.Fatalf("the land's controller got %d mana, want none (CR 106.4: the activator's pool)", got)
	}
	cost := mustParseCost(t, "{R}")
	if !me.ManaPool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeCast}) {
		t.Error("the {R} can't pay for a spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("the {R} paid for an activated ability: Piracy's mana is only to cast spells")
	}
}

// The auto-tapper takes the opponent's land for a spell and casts it.
func TestPiracyAutoTapCastsASpellWithAnOpponentsLand(t *testing.T) {
	g, me, mountain := piracyBoard(t)
	grantPiracy(t, g, me.ID)
	if _, err := tryAutoCast(g, me, "Shock", "Instant", "{R}"); err != nil {
		t.Fatalf("auto-paid cast with the opponent's Mountain: %v", err)
	}
	if !tappedOnBattlefield(t, g, mountain) {
		t.Error("the opponent's Mountain was not tapped for the cast")
	}
}

// Piracy reaches lands only: a nonland permanent stays its controller's.
func TestPiracyDoesNotReachNonlandPermanents(t *testing.T) {
	g, me, _ := piracyBoard(t)
	elf := pushIntrinsicPermanent(g, g.Seats[1], "Elf", "Creature — Elf Druid", []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}, nil)
	grantPiracy(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, elf, 0, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("tapping an opponent's mana creature under Piracy = %v, want ErrCardCallerMismatch", err)
	}
}

// "Until end of turn": the next turn the land is its controller's alone.
func TestPiracyEndsAtEndOfTurn(t *testing.T) {
	g, me, mountain := piracyBoard(t)
	grantPiracy(t, g, me.ID)
	advanceIntoStep(t, g, StepEnd)
	advanceIntoStep(t, g, StepUpkeep)
	if err := g.ActivateManaAbility(me.ID, mountain, 0, ManaAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("tapping the Mountain the next turn = %v, want ErrCardCallerMismatch", err)
	}
}
