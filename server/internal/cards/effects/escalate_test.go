package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate_test.go — CR 702.120a, #2126: "For each mode you choose
// beyond the first as you cast this spell, you pay an additional
// [cost]." The Collective cycle (discard, mana, tap a creature) plus
// the mana-only Borrowed and Alliance cards, through the real cast
// path.

const (
	collectiveBrutalityOracle  = "22a78443-db21-4656-b38a-e3e3186fd94b"
	collectiveResistanceOracle = "bffdfe7b-f17a-41b6-a460-80280fe497d4"
	collectiveDefianceOracle   = "43ddc3e8-936b-4be0-9e64-484bbbd69016"
	collectiveEffortOracle     = "a97e760a-99c8-47ea-a875-451aca913ee7"
	borrowedMalevolenceOracle  = "ee6e0279-9606-4877-a3a0-5d9dd557ff15"
	borrowedHostilityOracle    = "b45fc33b-0658-4fe1-89da-54a07a12ebd4"
	borrowedGraceOracle        = "e858d26f-db45-4c56-b40a-bbaeedbf5549"
	savageAllianceOracle       = "57da1d3f-7469-4efa-9b82-b2d915177d89"
	blessedAllianceOracle      = "48b67778-6616-4563-8456-62f302649768"
)

// castEscalate seeds the spell into the active seat's hand, walks to a
// main phase and casts it with `params`, returning the spell's id and
// the announce error. The hand is NOT cleared: callers set it up.
func castEscalate(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, params)
}

func escPlayerRef(id uuid.UUID, occ int) game.TargetRef {
	return modeRef(game.TargetPlayer, id, occ, 0)
}

func escCardRef(id uuid.UUID, occ int) game.TargetRef {
	return modeRef(game.TargetCard, id, occ, 0)
}

func TestEscalateSpecsAreWired(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		mana         string
		discard, tap int
	}{
		{"Collective Brutality", collectiveBrutalityOracle, "", 1, 0},
		{"Collective Resistance", collectiveResistanceOracle, "{G}", 0, 0},
		{"Collective Defiance", collectiveDefianceOracle, "{1}", 0, 0},
		{"Collective Effort", collectiveEffortOracle, "", 0, 1},
		{"Borrowed Malevolence", borrowedMalevolenceOracle, "{2}", 0, 0},
		{"Borrowed Hostility", borrowedHostilityOracle, "{3}", 0, 0},
		{"Borrowed Grace", borrowedGraceOracle, "{1}{W}", 0, 0},
		{"Savage Alliance", savageAllianceOracle, "{1}", 0, 0},
		{"Blessed Alliance", blessedAllianceOracle, "{2}", 0, 0},
	} {
		ms := game.ModeSpecFor(tc.oracle)
		if ms == nil || ms.Escalate == nil {
			t.Errorf("%s: no escalate cost through the catalog hook", tc.name)
			continue
		}
		e := ms.Escalate
		if e.ManaCost != tc.mana || e.DiscardCards != tc.discard || e.TapCreatures != tc.tap {
			t.Errorf("%s: escalate = {mana %q discard %d tap %d}, want {%q %d %d}",
				tc.name, e.ManaCost, e.DiscardCards, e.TapCreatures, tc.mana, tc.discard, tc.tap)
		}
	}
}

// --- discard-shaped escalate: Collective Brutality ---------------------

func TestCollectiveBrutalityOneModeCostsNothingExtra(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	handBefore := me.Hand.Size()
	oppLife, myLife := opp.Life, me.Life

	id, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{escPlayerRef(opp.ID, 0)}})
	if err != nil {
		t.Fatalf("one mode, no discard: %v", err)
	}
	if me.Hand.Contains(id) || handBefore != 0 || me.Hand.Size() != 0 {
		t.Errorf("one mode should discard nothing: hand %d", me.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("drain: opp %d→%d, me %d→%d, want -2 / +2", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestCollectiveBrutalityOneModeRefusesAStrayDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	fodder := handCard(me, "Fodder", "Sorcery")

	id, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{escPlayerRef(opp.ID, 0)}, DiscardIDs: []uuid.UUID{fodder}})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a discard with one mode: %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(id) || !me.Hand.Contains(fodder) {
		t.Error("a refused cast changed the hand")
	}
}

func TestCollectiveBrutalityTwoModesPayOneDiscardAndThreePayTwo(t *testing.T) {
	for _, tc := range []struct {
		name  string
		modes []int
		pays  int
	}{
		{"two modes", []int{1, 2}, 1},
		{"three modes", []int{0, 1, 2}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			me.Hand.Cards = nil
			var fodder []uuid.UUID
			for i := 0; i < 3; i++ {
				fodder = append(fodder, handCard(me, "Fodder", "Sorcery"))
			}
			victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
			targets := []game.TargetRef{}
			for occ, m := range tc.modes {
				switch m {
				case 0, 2:
					targets = append(targets, escPlayerRef(opp.ID, occ))
				case 1:
					targets = append(targets, escCardRef(victim, occ))
				}
			}

			// One discard too few is refused, and changes nothing.
			if _, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
				game.CastSpellParams{Modes: tc.modes, Targets: targets, DiscardIDs: fodder[:tc.pays-1]}); !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("%d discards for %d modes: %v, want ErrInvalidParam", tc.pays-1, len(tc.modes), err)
			}
			// One too many is refused too: the plan is consumed exactly.
			if _, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
				game.CastSpellParams{Modes: tc.modes, Targets: targets, DiscardIDs: fodder[:tc.pays+1]}); !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("%d discards for %d modes: %v, want ErrInvalidParam", tc.pays+1, len(tc.modes), err)
			}
			if me.Graveyard.Size() != 0 {
				t.Fatalf("a refused cast discarded %d cards", me.Graveyard.Size())
			}

			spell, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
				game.CastSpellParams{Modes: tc.modes, Targets: targets, DiscardIDs: fodder[:tc.pays]})
			if err != nil {
				t.Fatalf("%d discards for %d modes: %v", tc.pays, len(tc.modes), err)
			}
			// Paid at announce: the spell is still on the stack.
			if !g.Stack.Contains(spell) {
				t.Error("the spell should still be on the stack")
			}
			for _, id := range fodder[:tc.pays] {
				if !me.Graveyard.Contains(id) {
					t.Errorf("escalate discard %v is not in the graveyard", id)
				}
			}
			for _, id := range fodder[tc.pays:] {
				if !me.Hand.Contains(id) {
					t.Errorf("unpaid card %v left the hand", id)
				}
			}
		})
	}
}

func TestCollectiveBrutalityTooFewCardsToPayIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// The spell is the only card: nothing to discard, so no second mode.
	id, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, Targets: []game.TargetRef{escCardRef(victim, 0), escPlayerRef(opp.ID, 1)}})
	if err == nil {
		t.Fatal("two modes with an empty hand was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
}

func TestCollectiveBrutalityEscalateDiscardTriggersPayoffsWhileOnStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)
	fodder := []uuid.UUID{handCard(me, "Fodder", "Sorcery"), handCard(me, "Fodder", "Sorcery")}
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	spell, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, DiscardIDs: fodder, Targets: []game.TargetRef{
			escPlayerRef(opp.ID, 0), escCardRef(victim, 1), escPlayerRef(opp.ID, 2),
		}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// The discards happened as a cost, so the Mako's triggers are queued
	// or on the stack above the spell, which has not resolved.
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell should still be on the stack")
	}
	if opp.Life != 40 {
		t.Fatalf("the spell resolved before its payoffs: opp life %d", opp.Life)
	}
	passPriorityAroundTable(t, g)
	counters := 0
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == mako {
			counters = c.Counters["+1/+1"]
		}
	}
	if counters != 2 {
		t.Errorf("Marauding Mako counters = %d, want 2 (one per escalate discard)", counters)
	}
}

func TestCollectiveBrutalityModesDoWhatTheyPrint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	fodder := []uuid.UUID{handCard(me, "Fodder", "Sorcery")}
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if _, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, DiscardIDs: fodder, Targets: []game.TargetRef{
			escCardRef(bear, 0), escPlayerRef(opp.ID, 1),
		}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("-2/-2 should have killed the 2/2")
	}
	if opp.Life != 38 || me.Life != 42 {
		t.Errorf("life: opp %d me %d, want 38 / 42", opp.Life, me.Life)
	}
}

func TestCollectiveBrutalityFirstModePicksAnInstantOrSorceryFromTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	opp.Hand.Cards = nil
	ids := revealHand(opp, rhForest(), rhBolt(), pbBear(), game.Card{Name: "Divination", TypeLine: "Sorcery"})

	if _, err := castEscalate(t, g, "Collective Brutality", "Sorcery", collectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{escPlayerRef(opp.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pick := openPick(t, g); !sameIDs(pick.DiscardOptions, []uuid.UUID{ids[1], ids[3]}) {
		t.Errorf("options = %v, want the Bolt and the Divination", pick.DiscardOptions)
	}
}

// --- tap-shaped escalate: Collective Effort ----------------------------

func TestCollectiveEffortTapsOneCreaturePerExtraMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	mine2 := pushVanillaCreature(g, me.ID, "Mine Too", 0, 3)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	giant := pushVanillaCreature(g, opp.ID, "Giant", 5, 5)

	// One mode: no tap, and a stray tap is refused.
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{escCardRef(ench, 0)}, TeamworkIDs: []uuid.UUID{mine}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a tap with one mode: %v, want ErrInvalidParam", err)
	}
	// Two modes with no tap named is refused.
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(giant, 0), escCardRef(ench, 1)}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two modes, no tap: %v, want ErrInvalidParam", err)
	}
	// Two modes with two taps is refused (exactly one is owed).
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(giant, 0), escCardRef(ench, 1)}, TeamworkIDs: []uuid.UUID{mine, mine2}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two modes, two taps: %v, want ErrInvalidParam", err)
	}
	for _, id := range []uuid.UUID{mine, mine2} {
		if c := findBattlefieldCardByID(g, id); c == nil || c.Tapped {
			t.Fatal("a refused cast tapped a creature")
		}
	}
	// Two modes, one tap: pays.
	spell, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(giant, 0), escCardRef(ench, 1)}, TeamworkIDs: []uuid.UUID{mine}})
	if err != nil {
		t.Fatalf("two modes, one tap: %v", err)
	}
	if c := findBattlefieldCardByID(g, mine); c == nil || !c.Tapped {
		t.Error("the named creature should be tapped")
	}
	if !g.Stack.Contains(spell) {
		t.Error("the spell should still be on the stack")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(giant) || g.Battlefield.Contains(ench) {
		t.Error("both destroy bullets should have resolved")
	}
}

func TestCollectiveEffortTapCostIgnoresSummoningSicknessAndRefusesTappedOrForeign(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	sick := pushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	spent := pushVanillaCreature(g, me.ID, "Spent", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == spent {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	two := func(tap uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Modes: []int{1, 2}, TeamworkIDs: []uuid.UUID{tap},
			Targets: []game.TargetRef{escCardRef(ench, 0), escPlayerRef(opp.ID, 1)}}
	}
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "", two(spent)); !errors.Is(err, game.ErrAlreadyTapped) {
		t.Errorf("tapping a tapped creature: %v, want ErrAlreadyTapped", err)
	}
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "", two(theirs)); err == nil {
		t.Error("tapping an opponent's creature was accepted")
	}
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "", two(ench)); err == nil {
		t.Error("tapping a non-creature was accepted")
	}
	// Not the {T} symbol: a creature that came in this turn may pay.
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "", two(sick)); err != nil {
		t.Errorf("a summoning-sick creature should be able to pay: %v", err)
	}
}

func TestCollectiveEffortWithNoUntappedCreatureCannotEscalate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	id, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, Targets: []game.TargetRef{escCardRef(ench, 0), escPlayerRef(opp.ID, 1)}})
	if err == nil {
		t.Fatal("escalating with no creature to tap was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
	// A single mode is still castable with an empty board.
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{escCardRef(ench, 0)}}); err != nil {
		t.Fatalf("one mode with no creatures: %v", err)
	}
}

func TestCollectiveEffortCountersLandOnEveryCreatureTheTargetControls(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	other := pushVanillaCreature(g, g.Seats[1].ID, "Other", 2, 2)

	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{escPlayerRef(me.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	counters := func(id uuid.UUID) int {
		if c := findBattlefieldCardByID(g, id); c != nil {
			return c.Counters["+1/+1"]
		}
		return -1
	}
	if counters(a) != 1 || counters(b) != 1 || counters(other) != 0 {
		t.Errorf("counters: a=%d b=%d other=%d, want 1 1 0", counters(a), counters(b), counters(other))
	}
}

func TestCollectiveEffortDestroyRespectsPowerFloor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	small := pushVanillaCreature(g, opp.ID, "Small", 3, 3)
	if _, err := castEscalate(t, g, "Collective Effort", "Sorcery", collectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{escCardRef(small, 0)}}); err == nil {
		t.Fatal("power 3 is not a legal target for \"power 4 or greater\"")
	}
}

// --- mana-shaped escalate ----------------------------------------------

func priceWithModes(t *testing.T, g *game.Game, oracle, cost string, modes []int) (generic, coloured int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{InstanceID: uuid.New(), Name: "Escalate Probe", TypeLine: "Instant", OracleID: oracle,
		ManaCost: cost, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{Modes: modes})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	return price.Total.Generic, len(price.Total.Required)
}

func TestEscalateManaIsOwedOncePerModeBeyondTheFirst(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost  string
		modes               []int
		wantGeneric, wantCo int
	}{
		// Collective Resistance {1}{G}, escalate {G}.
		{"Resistance 1 mode", collectiveResistanceOracle, "{1}{G}", []int{0}, 1, 1},
		{"Resistance 2 modes", collectiveResistanceOracle, "{1}{G}", []int{0, 1}, 1, 2},
		{"Resistance 3 modes", collectiveResistanceOracle, "{1}{G}", []int{0, 1, 2}, 1, 3},
		// Collective Defiance {1}{R}{R}, escalate {1}.
		{"Defiance 1 mode", collectiveDefianceOracle, "{1}{R}{R}", []int{1}, 1, 2},
		{"Defiance 2 modes", collectiveDefianceOracle, "{1}{R}{R}", []int{1, 2}, 2, 2},
		{"Defiance 3 modes", collectiveDefianceOracle, "{1}{R}{R}", []int{0, 1, 2}, 3, 2},
		// Borrowed Hostility {R}, escalate {3}.
		{"Hostility 1 mode", borrowedHostilityOracle, "{R}", []int{0}, 0, 1},
		{"Hostility 2 modes", borrowedHostilityOracle, "{R}", []int{0, 1}, 3, 1},
		// Borrowed Grace {2}{W}, escalate {1}{W}.
		{"Grace 2 modes", borrowedGraceOracle, "{2}{W}", []int{0, 1}, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			gen, co := priceWithModes(t, g, tc.oracle, tc.cost, tc.modes)
			if gen != tc.wantGeneric || co != tc.wantCo {
				t.Errorf("price = generic %d + %d coloured, want %d + %d", gen, co, tc.wantGeneric, tc.wantCo)
			}
		})
	}
}

func TestEscalateManaIsRefusedWhenTheCasterCannotPayIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	art := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	advanceToMain(t, g)
	me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"})

	// {1}{G} pays for one mode, not two ({1}{G}{G}).
	id, err := castEscalate(t, g, "Collective Resistance", "Instant", collectiveResistanceOracle, "{1}{G}",
		game.CastSpellParams{Strict: true, Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(art, 0), escCardRef(ench, 1)}})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("two modes on two mana: %v, want ErrInsufficientMana", err)
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
	if _, err := castEscalate(t, g, "Collective Resistance", "Instant", collectiveResistanceOracle, "{1}{G}",
		game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{escCardRef(art, 0)}}); err != nil {
		t.Fatalf("one mode on two mana: %v", err)
	}
}

// --- every card's modes ------------------------------------------------

func TestCollectiveResistanceModesDoWhatTheyPrint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	art := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := pushVanillaCreature(g, me.ID, "Mine", 2, 2)

	if _, err := castEscalate(t, g, "Collective Resistance", "Instant", collectiveResistanceOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{escCardRef(art, 0), escCardRef(ench, 1), escCardRef(bear, 2)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(ench) {
		t.Error("the artifact and enchantment should be destroyed")
	}
	abilities := effectiveAbilities(t, g, bear)
	for _, want := range []string{"hexproof", "indestructible"} {
		found := false
		for _, a := range abilities {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("target creature lacks %s: %v", want, abilities)
		}
	}
}

func TestCollectiveDefianceModesDoWhatTheyPrint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 4, 4)
	handCard(opp, "A", "Sorcery")
	handCard(opp, "B", "Sorcery")
	oppHand := opp.Hand.Size()

	if _, err := castEscalate(t, g, "Collective Defiance", "Sorcery", collectiveDefianceOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{
			escPlayerRef(opp.ID, 0), escCardRef(bear, 1), escPlayerRef(opp.ID, 2),
		}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("4 damage should kill a 4/4")
	}
	if opp.Life != 37 {
		t.Errorf("opp life = %d, want 37 after 3 damage", opp.Life)
	}
	if opp.Hand.Size() != oppHand {
		t.Errorf("wheel: hand %d, want %d (discard all, draw that many)", opp.Hand.Size(), oppHand)
	}
	if opp.Graveyard.Size() < oppHand {
		t.Errorf("the discarded cards should be in the graveyard: %d", opp.Graveyard.Size())
	}
}

func TestBorrowedAndAllianceCardsResolveEachMode(t *testing.T) {
	t.Run("Borrowed Malevolence", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
		theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
		if _, err := castEscalate(t, g, "Borrowed Malevolence", "Instant", borrowedMalevolenceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(mine, 0), escCardRef(theirs, 1)}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p := effectivePower(t, g, mine); p != 3 {
			t.Errorf("+1/+1 on a 2/2: power %d, want 3", p)
		}
		if g.Battlefield.Contains(theirs) {
			t.Error("-1/-1 should kill a 1/1")
		}
	})
	t.Run("Borrowed Hostility", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.Hand.Cards = nil
		mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
		if _, err := castEscalate(t, g, "Borrowed Hostility", "Instant", borrowedHostilityOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(mine, 0), escCardRef(mine, 1)}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p := effectivePower(t, g, mine); p != 5 {
			t.Errorf("+3/+0 on a 2/2: power %d, want 5", p)
		}
		found := false
		for _, a := range effectiveAbilities(t, g, mine) {
			found = found || a == "first strike"
		}
		if !found {
			t.Error("first strike was not granted")
		}
	})
	t.Run("Borrowed Grace", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.Hand.Cards = nil
		a := pushVanillaCreature(g, me.ID, "A", 1, 1)
		if _, err := castEscalate(t, g, "Borrowed Grace", "Instant", borrowedGraceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if effectivePower(t, g, a) != 3 || effectiveToughness(t, g, a) != 3 {
			t.Errorf("a 1/1 under +2/+0 and +0/+2 should be 3/3, got %d/%d", effectivePower(t, g, a), effectiveToughness(t, g, a))
		}
	})
	t.Run("Savage Alliance", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
		x := pushVanillaCreature(g, opp.ID, "X", 3, 3)
		y := pushVanillaCreature(g, opp.ID, "Y", 1, 1)
		if _, err := castEscalate(t, g, "Savage Alliance", "Instant", savageAllianceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{
				escPlayerRef(me.ID, 0), escCardRef(x, 1), escPlayerRef(opp.ID, 2),
			}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
			t.Error("2 damage then 1 to each should have killed both of the opponent's creatures")
		}
		found := false
		for _, a := range effectiveAbilities(t, g, mine) {
			found = found || a == "trample"
		}
		if !found {
			t.Error("trample was not granted to the target player's creatures")
		}
	})
	t.Run("Blessed Alliance", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		tapped := pushVanillaCreature(g, me.ID, "Tapped", 2, 2)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		if _, err := castEscalate(t, g, "Blessed Alliance", "Instant", blessedAllianceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
				escPlayerRef(me.ID, 0), escCardRef(tapped, 1),
			}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if me.Life != 44 {
			t.Errorf("life = %d, want 44", me.Life)
		}
		if c := findBattlefieldCardByID(g, tapped); c == nil || c.Tapped {
			t.Error("the creature should be untapped")
		}
		_ = opp
	})
}
