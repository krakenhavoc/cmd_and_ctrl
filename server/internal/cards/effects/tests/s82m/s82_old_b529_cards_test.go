package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b529_cards_test.go — Shadowmoor, batch 529 (S82, #5591): life gain
// and loss. Each test plays the real catalog entry through the engine.

const (
	oldB529SmolderOracle     = "0300d8a1-ff4e-4e05-a097-4822f80c47e1"
	oldB529TyrannizeOracle   = "27da0af7-85d3-42aa-9426-b86c5e464f81"
	oldB529DawnglowOracle    = "4db813f4-3bd1-4177-9cf7-6c9799a343c5"
	oldB529WoeleecherOracle  = "52e73390-75e4-45a1-8dbe-96e7cceb0a14"
	oldB529MentorOracle      = "5b60a257-e4f8-412e-b146-b7983a384753"
	oldB529BarghestOracle    = "6b08f344-1ff5-4d04-a435-d3a1c5e4211b"
	oldB529ApothecaryOracle  = "a74d80ec-9f42-418d-ab98-897acc23c319"
	oldB529SpiteflameOracle  = "b4ab38e5-c446-4845-b509-4d6f4438949d"
	oldB529BoonOracle        = "b86d3142-568a-4eea-b1ae-5fcfc1453533"
	oldB529ToilOracle        = "c8323e46-c068-45e2-b6ba-7972b4af728a"
	oldB529OracleNectarsOrcl = "f7ea8725-db0e-4ce5-ac77-7dfc2a4e9286"
)

func oldB529Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func oldB529EmptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

func TestOldB529EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB529SmolderOracle: "Smolder Initiate", oldB529TyrannizeOracle: "Tyrannize",
		oldB529DawnglowOracle: "Dawnglow Infusion", oldB529WoeleecherOracle: "Woeleecher",
		oldB529MentorOracle: "Resplendent Mentor", oldB529BarghestOracle: "Hollowborn Barghest",
		oldB529ApothecaryOracle: "Apothecary Initiate", oldB529SpiteflameOracle: "Spiteflame Witch",
		oldB529BoonOracle: "Boon Reflection", oldB529ToilOracle: "Toil to Renown",
		oldB529OracleNectarsOrcl: "Oracle of Nectars",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB529BoonReflectionDoublesOnlyYourGains(t *testing.T) {
	g, me, opp := oldB529Table(t)
	effectstest.B12Push(g, me.ID, "Boon Reflection", "Enchantment", oldB529BoonOracle, 0, 0)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2)
	})
	if me.Life != mine+6-2 {
		t.Errorf("your life %d → %d, want %d (gain 3 doubled, loss 2 untouched)", mine, me.Life, mine+4)
	}
	if opp.Life != theirs+3 {
		t.Errorf("an opponent's gain was doubled: %d → %d", theirs, opp.Life)
	}
}

// oldB529CastSpell has `caster` cast a one-colour sorcery for free from a main phase.
func oldB529CastSpell(t *testing.T, g *game.Game, caster *game.Player, colour string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.DstHand(caster, "Test Spell", "Sorcery", "{"+colour+"}", colour)
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestOldB529ApothecaryInitiateGainsOnlyWhenTheOneIsPaid(t *testing.T) {
	for _, tc := range []struct {
		name         string
		colour       string
		opponentCast bool
		pay          bool
		want         int
	}{
		{"white spell, pay", "W", false, true, 1},
		{"white spell, decline", "W", false, false, 0},
		{"an opponent's white spell, pay", "W", true, true, 1},
		{"black spell", "B", false, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := oldB529Table(t)
			effectstest.B12Push(g, me.ID, "Apothecary Initiate", "Creature — Kithkin Cleric", oldB529ApothecaryOracle, 1, 1)
			caster := me
			if tc.opponentCast {
				caster = opp
				g.WithWriteLock(func() { g.Turn.ActiveSeat = opp.Seat })
			}
			oldB529CastSpell(t, g, caster, tc.colour)
			effectstest.FlashIfFund(g, me, "C")
			before := me.Life
			effectstest.PassPriorityAroundTable(t, g)
			if tc.colour == "W" {
				effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			} else if effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatal("a black spell asked for a payment")
			}
			if got := me.Life - before; got != tc.want {
				t.Errorf("life changed by %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB529ApothecaryInitiateCannotPayWithoutMana(t *testing.T) {
	g, me, _ := oldB529Table(t)
	effectstest.B12Push(g, me.ID, "Apothecary Initiate", "Creature — Kithkin Cleric", oldB529ApothecaryOracle, 1, 1)
	oldB529CastSpell(t, g, me, "W")
	effectstest.FlashIfFund(g, me)
	before := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
	}
	if me.Life != before {
		t.Errorf("life %d → %d with no mana to pay {1}", before, me.Life)
	}
}

func TestOldB529SmolderInitiateDrainsTheChosenPlayerOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g, me, opp := oldB529Table(t)
		effectstest.B12Push(g, me.ID, "Smolder Initiate", "Creature — Elemental Shaman", oldB529SmolderOracle, 1, 1)
		oldB529CastSpell(t, g, me, "B")
		effectstest.FlashIfFund(g, me, "C")
		effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
		before, mine := opp.Life, me.Life
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		want := before
		if pay {
			want = before - 1
		}
		if opp.Life != want || me.Life != mine {
			t.Errorf("pay=%v: opponent %d → %d (want %d), you %d → %d", pay, before, opp.Life, want, mine, me.Life)
		}
	}
}

func TestOldB529SmolderInitiateIgnoresNonBlackSpells(t *testing.T) {
	g, me, _ := oldB529Table(t)
	effectstest.B12Push(g, me.ID, "Smolder Initiate", "Creature — Elemental Shaman", oldB529SmolderOracle, 1, 1)
	oldB529CastSpell(t, g, me, "R")
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Stack.Size() != 1 {
		t.Errorf("a red spell triggered the Initiate (stack %d)", g.Stack.Size())
	}
}

func TestOldB529SpiteflameWitchMakesEveryPlayerLoseOne(t *testing.T) {
	g, me, _ := oldB529Table(t)
	witch := effectstest.B12Push(g, me.ID, "Spiteflame Witch", "Creature — Elemental Shaman", oldB529SpiteflameOracle, 2, 1)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.B10AddMana(me, "B", "R")
	effectstest.B16Activate(t, g, me.ID, witch, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-1 {
			t.Errorf("seat %d life %d → %d, want one less", p.Seat, lives[p.ID], p.Life)
		}
	}
}

func TestOldB529SpiteflameWitchNeedsBothColours(t *testing.T) {
	g, me, _ := oldB529Table(t)
	witch := effectstest.B12Push(g, me.ID, "Spiteflame Witch", "Creature — Elemental Shaman", oldB529SpiteflameOracle, 2, 1)
	effectstest.B10AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("{B}{B} paid for {B}{R}")
	}
}

func TestOldB529HollowbornBarghestDrainsOnEmptyHands(t *testing.T) {
	for _, empty := range []bool{true, false} {
		g, me, opp := oldB529Table(t)
		effectstest.B12Push(g, me.ID, "Hollowborn Barghest", "Creature — Demon Dog", oldB529BarghestOracle, 7, 6)
		if empty {
			oldB529EmptyHand(g, me)
		}
		lives := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			lives[p.ID] = p.Life
		}
		effectstest.BeginUpkeep(t, g, me.ID)
		for _, p := range g.Seats {
			want := lives[p.ID]
			if empty && p.ID != me.ID {
				want -= 2
			}
			if p.Life != want {
				t.Errorf("empty=%v your upkeep: seat %d life %d → %d, want %d", empty, p.Seat, lives[p.ID], p.Life, want)
			}
		}
		_ = opp
	}
}

func TestOldB529HollowbornBarghestPunishesAnOpponentUpkeepWithNoCards(t *testing.T) {
	for _, empty := range []bool{true, false} {
		g, me, opp := oldB529Table(t)
		effectstest.B12Push(g, me.ID, "Hollowborn Barghest", "Creature — Demon Dog", oldB529BarghestOracle, 7, 6)
		other := g.Seats[(opp.Seat+1)%len(g.Seats)]
		if other.ID == me.ID {
			other = g.Seats[(opp.Seat+2)%len(g.Seats)]
		}
		if empty {
			oldB529EmptyHand(g, opp)
		}
		oldB529EmptyHand(g, me) // its own upkeep trigger is not under test
		mine, theirs, third := me.Life, opp.Life, other.Life
		effectstest.BeginUpkeep(t, g, opp.ID)
		want := theirs
		if empty {
			want -= 2
		}
		if opp.Life != want {
			t.Errorf("empty=%v: opponent life %d → %d, want %d", empty, theirs, opp.Life, want)
		}
		if me.Life != mine || other.Life != third {
			t.Errorf("empty=%v: only the upkeep's player loses life (you %d → %d, other %d → %d)", empty, mine, me.Life, third, other.Life)
		}
	}
}

func TestOldB529TyrannizeDiscardsUnlessSevenLifeIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g, me, opp := oldB529Table(t)
		for i := 0; i < 3; i++ {
			effectstest.DstHand(opp, "Card", "Sorcery", "{1}")
		}
		hand, life := opp.Hand.Size(), opp.Life
		effectstest.CastCatalogSpell(t, g, "Tyrannize", "Sorcery", oldB529TyrannizeOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		if pay && (opp.Hand.Size() != hand || opp.Life != life-7) {
			t.Errorf("paid: hand %d → %d, life %d → %d; want the hand kept and 7 life paid", hand, opp.Hand.Size(), life, opp.Life)
		}
		if !pay && (opp.Hand.Size() != 0 || opp.Life != life) {
			t.Errorf("declined: hand %d → %d, life %d → %d; want an empty hand and no life paid", hand, opp.Hand.Size(), life, opp.Life)
		}
		_ = me
	}
}

func TestOldB529TyrannizeDiscardsWithoutAskingBelowSevenLife(t *testing.T) {
	g, _, opp := oldB529Table(t)
	effectstest.DstHand(opp, "Card", "Sorcery", "{1}")
	g.WithWriteLock(func() { opp.Life = 6 })
	effectstest.CastCatalogSpell(t, g, "Tyrannize", "Sorcery", oldB529TyrannizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("asked a player with 6 life to pay 7")
	}
	if opp.Hand.Size() != 0 || opp.Life != 6 {
		t.Errorf("hand %d, life %d; want an empty hand and 6 life", opp.Hand.Size(), opp.Life)
	}
}

func TestOldB529DawnglowInfusionReadsTheColoursSpent(t *testing.T) {
	for _, tc := range []struct {
		name string
		pool []string
		want int
	}{
		{"green only", []string{"G", "G", "G"}, 2},
		{"white only", []string{"W", "W", "W"}, 2},
		{"both", []string{"G", "W", "G"}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := oldB529Table(t)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.DstHand(me, "Dawnglow Infusion", "Sorcery", "{X}{G/W}", "G", "W")
			g.WithWriteLock(func() {
				for i := range me.Hand.Cards {
					if me.Hand.Cards[i].InstanceID == id {
						me.Hand.Cards[i].OracleID = oldB529DawnglowOracle
					}
				}
			})
			effectstest.FlashIfFund(g, me, tc.pool...)
			before := me.Life
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Life - before; got != tc.want {
				t.Errorf("life changed by %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB529OracleOfNectarsGainsX(t *testing.T) {
	g, me, _ := oldB529Table(t)
	oracle := effectstest.B12Push(g, me.ID, "Oracle of Nectars", "Creature — Elf Cleric", oldB529OracleNectarsOrcl, 2, 2)
	effectstest.B10AddMana(me, "C", "C", "C")
	before := me.Life
	effectstest.B16Activate(t, g, me.ID, oracle, 0, game.ActivateAbilityParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d → %d, want +3", before, me.Life)
	}
	if !effectstest.B16Tapped(t, g, oracle) {
		t.Error("the Oracle did not tap")
	}
}

func TestOldB529ResplendentMentorGivesWhiteCreaturesALifeGainTap(t *testing.T) {
	g, me, opp := oldB529Table(t)
	effectstest.B12Push(g, me.ID, "Resplendent Mentor", "Creature — Kithkin Cleric", oldB529MentorOracle, 2, 2)
	white := effectstest.B16Creature(g, me.ID, "White Bear", "Creature — Bear", 2, 2, "W")
	green := effectstest.B16Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their White Bear", "Creature — Bear", 2, 2, "W")
	for name, id := range map[string]uuid.UUID{"green": green, "opponent's white": theirs} {
		if err := g.ActivateCatalogAbility(map[string]uuid.UUID{"green": me.ID, "opponent's white": opp.ID}[name], id, 0, game.ActivateAbilityParams{}); err == nil {
			t.Errorf("%s creature has the granted ability", name)
		}
	}
	before := me.Life
	effectstest.B16Activate(t, g, me.ID, white, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestOldB529WoeleecherRemovesACounterAndGainsOnlyIfItDid(t *testing.T) {
	for _, counters := range []int{2, 0} {
		g, me, _ := oldB529Table(t)
		w := effectstest.B12Push(g, me.ID, "Woeleecher", "Creature — Elemental", oldB529WoeleecherOracle, 3, 5)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
		if counters > 0 {
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterMinusOne, counters) })
		}
		effectstest.B10AddMana(me, "W")
		before := me.Life
		effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
		effectstest.PassPriorityAroundTable(t, g)
		wantLife, wantCounters := before, 0
		if counters > 0 {
			wantLife, wantCounters = before+2, counters-1
		}
		if me.Life != wantLife || effectstest.CountersOn(g, bear, game.CounterMinusOne) != wantCounters {
			t.Errorf("%d counters: life %d → %d (want %d), counters left %d (want %d)",
				counters, before, me.Life, wantLife, effectstest.CountersOn(g, bear, game.CounterMinusOne), wantCounters)
		}
	}
}

func TestOldB529ToilToRenownCountsTappedArtifactsCreaturesAndLands(t *testing.T) {
	g, me, opp := oldB529Table(t)
	tap := func(id uuid.UUID) { g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) }) }
	tap(effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	tap(effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"))
	tap(effectstest.B12Permanent(g, me.ID, "Mox", "Artifact Creature — Construct")) // counted once
	tap(effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest"))
	tap(effectstest.B12Permanent(g, me.ID, "Pact", "Enchantment")) // not counted
	effectstest.B12Permanent(g, me.ID, "Untapped Forest", "Basic Land — Forest")
	tap(effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Toil to Renown", "Sorcery", oldB529ToilOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("life %d → %d, want +4 (creature, artifact, artifact creature, land)", before, me.Life)
	}
}
