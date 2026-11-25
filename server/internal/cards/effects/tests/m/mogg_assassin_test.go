package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6171 — Mogg Assassin: "{T}: You choose target creature an opponent
// controls, and that opponent chooses target creature. Flip a coin. If
// you win the flip, destroy the creature you chose. If you lose the
// flip, destroy the creature your opponent chose."

const moggAssassinOracle = "36a6925a-621e-47d6-8cb3-1b2886e1986e"

type moggTable struct {
	g                *game.Game
	me, a, b         *game.Player
	assassin, aBear  uuid.UUID
	bBear, otherBear uuid.UUID
}

func newMoggTable(t *testing.T) moggTable {
	g := effectstest.NewCatalogGame(t)
	m := moggTable{g: g, me: g.Seats[0], a: g.Seats[1], b: g.Seats[2]}
	m.assassin = effectstest.ApaPush(g, m.me.ID, m.me.ID, effectstest.ApaCreature("Mogg Assassin", moggAssassinOracle, 2, 1))
	m.aBear = effectstest.ApaPush(g, m.a.ID, m.a.ID, game.Card{Name: "A's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	m.bBear = effectstest.ApaPush(g, m.b.ID, m.b.ID, game.Card{Name: "B's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	m.otherBear = effectstest.ApaPush(g, g.Seats[3].ID, g.Seats[3].ID, game.Card{Name: "D's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	return m
}

// CR 601.2c (via CR 602.2b) and CR 115.1: both targets are chosen as the
// ability is put on the stack. With four players, the second target is
// chosen by the controller of the first, with no "which opponent"
// question and no prompt for any other opponent. That opponent may choose
// any creature, another player's included, judged for the controller.
func TestMoggAssassinSecondTargetIsChosenByTheFirstTargetsController(t *testing.T) {
	m := newMoggTable(t)
	g := m.g
	if err := g.ActivateCatalogAbility(m.me.ID, m.assassin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: m.aBear, Slot: 0}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c := effectstest.LatestOptionPickFor(g, m.me.ID); c != nil {
		t.Fatalf("the controller was asked to name an opponent: %q", c.Reason)
	}
	for _, p := range []*game.Player{m.b, g.Seats[3]} {
		if effectstest.LatestPickTarget(g, p.ID) != nil {
			t.Fatalf("%s, who does not control the first target, was asked to choose", p.Name)
		}
	}
	pick := effectstest.OtPickFor(t, g, m.a.ID, m.me.ID)
	for _, want := range []uuid.UUID{m.aBear, m.bBear, m.otherBear, m.assassin} {
		if !slices.Contains(pick.PickTargetCards, want) {
			t.Errorf("\"target creature\" does not offer %s", want)
		}
	}
	effectstest.OtAnswer(t, g, m.a.ID, effectstest.OtCard(m.bBear))
	item := effectstest.OtItemFrom(g, m.assassin)
	if item == nil || len(item.Targets) != 2 || len(item.TargetChoosers) != 1 || item.TargetChoosers[0].Player != m.a.ID {
		t.Fatalf("item %+v, want two targets and the first target's controller as the second's chooser", item)
	}
}

// The flip: won, the creature the controller chose is destroyed; lost,
// the creature the opponent chose.
func TestMoggAssassinDestroysTheChosenCreatureByTheFlip(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		m := newMoggTable(t)
		g := m.g
		if err := g.ActivateCatalogAbility(m.me.ID, m.assassin, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: m.aBear, Slot: 0}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.OtAnswer(t, g, m.a.ID, effectstest.OtCard(m.bBear))
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, m.me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) == 0 {
			t.Fatal("no coin was flipped")
		}
		aGone, bGone := !g.Battlefield.Contains(m.aBear), !g.Battlefield.Contains(m.bBear)
		if evs[len(evs)-1].Won {
			won++
			if !aGone || bGone {
				t.Errorf("won flip: A's Bear destroyed %v (want true), B's Bear destroyed %v (want false)", aGone, bGone)
			}
		} else {
			lost++
			if aGone || !bGone {
				t.Errorf("lost flip: A's Bear destroyed %v (want false), B's Bear destroyed %v (want true)", aGone, bGone)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}

// The first clause is "an opponent controls": the controller's own
// creature is refused.
func TestMoggAssassinFirstTargetMustBeAnOpponents(t *testing.T) {
	m := newMoggTable(t)
	mine := effectstest.ApaPush(m.g, m.me.ID, m.me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := m.g.ActivateCatalogAbility(m.me.ID, m.assassin, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}},
	}); err == nil {
		t.Fatal("the controller's own creature was accepted as the first target")
	}
}
