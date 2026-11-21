package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEscalateSpecsAreWired(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		mana         string
		discard, tap int
	}{
		{"Collective Brutality", effectstest.CollectiveBrutalityOracle, "", 1, 0},
		{"Collective Resistance", effectstest.CollectiveResistanceOracle, "{G}", 0, 0},
		{"Collective Defiance", effectstest.CollectiveDefianceOracle, "{1}", 0, 0},
		{"Collective Effort", effectstest.CollectiveEffortOracle, "", 0, 1},
		{"Borrowed Malevolence", effectstest.BorrowedMalevolenceOracle, "{2}", 0, 0},
		{"Borrowed Hostility", effectstest.BorrowedHostilityOracle, "{3}", 0, 0},
		{"Borrowed Grace", effectstest.BorrowedGraceOracle, "{1}{W}", 0, 0},
		{"Savage Alliance", effectstest.SavageAllianceOracle, "{1}", 0, 0},
		{"Blessed Alliance", effectstest.BlessedAllianceOracle, "{2}", 0, 0},
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	handBefore := me.Hand.Size()
	oppLife, myLife := opp.Life, me.Life

	id, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{effectstest.EscPlayerRef(opp.ID, 0)}})
	if err != nil {
		t.Fatalf("one mode, no discard: %v", err)
	}
	if me.Hand.Contains(id) || handBefore != 0 || me.Hand.Size() != 0 {
		t.Errorf("one mode should discard nothing: hand %d", me.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("drain: opp %d→%d, me %d→%d, want -2 / +2", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestCollectiveBrutalityOneModeRefusesAStrayDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	fodder := effectstest.HandCard(me, "Fodder", "Sorcery")

	id, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{effectstest.EscPlayerRef(opp.ID, 0)}, DiscardIDs: []uuid.UUID{fodder}})
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			me.Hand.Cards = nil
			var fodder []uuid.UUID
			for i := 0; i < 3; i++ {
				fodder = append(fodder, effectstest.HandCard(me, "Fodder", "Sorcery"))
			}
			victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
			targets := []game.TargetRef{}
			for occ, m := range tc.modes {
				switch m {
				case 0, 2:
					targets = append(targets, effectstest.EscPlayerRef(opp.ID, occ))
				case 1:
					targets = append(targets, effectstest.EscCardRef(victim, occ))
				}
			}

			// One discard too few is refused, and changes nothing.
			if _, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
				game.CastSpellParams{Modes: tc.modes, Targets: targets, DiscardIDs: fodder[:tc.pays-1]}); !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("%d discards for %d modes: %v, want ErrInvalidParam", tc.pays-1, len(tc.modes), err)
			}
			// One too many is refused too: the plan is consumed exactly.
			if _, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
				game.CastSpellParams{Modes: tc.modes, Targets: targets, DiscardIDs: fodder[:tc.pays+1]}); !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("%d discards for %d modes: %v, want ErrInvalidParam", tc.pays+1, len(tc.modes), err)
			}
			if me.Graveyard.Size() != 0 {
				t.Fatalf("a refused cast discarded %d cards", me.Graveyard.Size())
			}

			spell, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// The spell is the only card: nothing to discard, so no second mode.
	id, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, Targets: []game.TargetRef{effectstest.EscCardRef(victim, 0), effectstest.EscPlayerRef(opp.ID, 1)}})
	if err == nil {
		t.Fatal("two modes with an empty hand was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
}

func TestCollectiveBrutalityEscalateDiscardTriggersPayoffsWhileOnStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	mako := effectstest.PushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", effectstest.MaraudingMakoOracle, false)
	fodder := []uuid.UUID{effectstest.HandCard(me, "Fodder", "Sorcery"), effectstest.HandCard(me, "Fodder", "Sorcery")}
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	spell, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, DiscardIDs: fodder, Targets: []game.TargetRef{
			effectstest.EscPlayerRef(opp.ID, 0), effectstest.EscCardRef(victim, 1), effectstest.EscPlayerRef(opp.ID, 2),
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	fodder := []uuid.UUID{effectstest.HandCard(me, "Fodder", "Sorcery")}
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if _, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, DiscardIDs: fodder, Targets: []game.TargetRef{
			effectstest.EscCardRef(bear, 0), effectstest.EscPlayerRef(opp.ID, 1),
		}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("-2/-2 should have killed the 2/2")
	}
	if opp.Life != 38 || me.Life != 42 {
		t.Errorf("life: opp %d me %d, want 38 / 42", opp.Life, me.Life)
	}
}

func TestCollectiveBrutalityFirstModePicksAnInstantOrSorceryFromTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	opp.Hand.Cards = nil
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), game.Card{Name: "Divination", TypeLine: "Sorcery"})

	if _, err := effectstest.CastEscalate(t, g, "Collective Brutality", "Sorcery", effectstest.CollectiveBrutalityOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.EscPlayerRef(opp.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.OpenPick(t, g); !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{ids[1], ids[3]}) {
		t.Errorf("options = %v, want the Bolt and the Divination", pick.DiscardOptions)
	}
}

// --- tap-shaped escalate: Collective Effort ----------------------------

func TestCollectiveEffortTapsOneCreaturePerExtraMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	mine2 := effectstest.PushVanillaCreature(g, me.ID, "Mine Too", 0, 3)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	giant := effectstest.PushVanillaCreature(g, opp.ID, "Giant", 5, 5)

	// One mode: no tap, and a stray tap is refused.
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{effectstest.EscCardRef(ench, 0)}, TeamworkIDs: []uuid.UUID{mine}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a tap with one mode: %v, want ErrInvalidParam", err)
	}
	// Two modes with no tap named is refused.
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(giant, 0), effectstest.EscCardRef(ench, 1)}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two modes, no tap: %v, want ErrInvalidParam", err)
	}
	// Two modes with two taps is refused (exactly one is owed).
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(giant, 0), effectstest.EscCardRef(ench, 1)}, TeamworkIDs: []uuid.UUID{mine, mine2}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two modes, two taps: %v, want ErrInvalidParam", err)
	}
	for _, id := range []uuid.UUID{mine, mine2} {
		if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.Tapped {
			t.Fatal("a refused cast tapped a creature")
		}
	}
	// Two modes, one tap: pays.
	spell, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(giant, 0), effectstest.EscCardRef(ench, 1)}, TeamworkIDs: []uuid.UUID{mine}})
	if err != nil {
		t.Fatalf("two modes, one tap: %v", err)
	}
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || !c.Tapped {
		t.Error("the named creature should be tapped")
	}
	if !g.Stack.Contains(spell) {
		t.Error("the spell should still be on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(giant) || g.Battlefield.Contains(ench) {
		t.Error("both destroy bullets should have resolved")
	}
}

func TestCollectiveEffortTapCostIgnoresSummoningSicknessAndRefusesTappedOrForeign(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	spent := effectstest.PushVanillaCreature(g, me.ID, "Spent", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == spent {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	two := func(tap uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Modes: []int{1, 2}, TeamworkIDs: []uuid.UUID{tap},
			Targets: []game.TargetRef{effectstest.EscCardRef(ench, 0), effectstest.EscPlayerRef(opp.ID, 1)}}
	}
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "", two(spent)); !errors.Is(err, game.ErrAlreadyTapped) {
		t.Errorf("tapping a tapped creature: %v, want ErrAlreadyTapped", err)
	}
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "", two(theirs)); err == nil {
		t.Error("tapping an opponent's creature was accepted")
	}
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "", two(ench)); err == nil {
		t.Error("tapping a non-creature was accepted")
	}
	// Not the {T} symbol: a creature that came in this turn may pay.
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "", two(sick)); err != nil {
		t.Errorf("a summoning-sick creature should be able to pay: %v", err)
	}
}

func TestCollectiveEffortWithNoUntappedCreatureCannotEscalate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	id, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1, 2}, Targets: []game.TargetRef{effectstest.EscCardRef(ench, 0), effectstest.EscPlayerRef(opp.ID, 1)}})
	if err == nil {
		t.Fatal("escalating with no creature to tap was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
	// A single mode is still castable with an empty board.
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{effectstest.EscCardRef(ench, 0)}}); err != nil {
		t.Fatalf("one mode with no creatures: %v", err)
	}
}

func TestCollectiveEffortCountersLandOnEveryCreatureTheTargetControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	other := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Other", 2, 2)

	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{effectstest.EscPlayerRef(me.ID, 0)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	counters := func(id uuid.UUID) int {
		if c := effectstest.FindBattlefieldCardByID(g, id); c != nil {
			return c.Counters["+1/+1"]
		}
		return -1
	}
	if counters(a) != 1 || counters(b) != 1 || counters(other) != 0 {
		t.Errorf("counters: a=%d b=%d other=%d, want 1 1 0", counters(a), counters(b), counters(other))
	}
}

func TestCollectiveEffortDestroyRespectsPowerFloor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 3, 3)
	if _, err := effectstest.CastEscalate(t, g, "Collective Effort", "Sorcery", effectstest.CollectiveEffortOracle, "",
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{effectstest.EscCardRef(small, 0)}}); err == nil {
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
		{"Resistance 1 mode", effectstest.CollectiveResistanceOracle, "{1}{G}", []int{0}, 1, 1},
		{"Resistance 2 modes", effectstest.CollectiveResistanceOracle, "{1}{G}", []int{0, 1}, 1, 2},
		{"Resistance 3 modes", effectstest.CollectiveResistanceOracle, "{1}{G}", []int{0, 1, 2}, 1, 3},
		// Collective Defiance {1}{R}{R}, escalate {1}.
		{"Defiance 1 mode", effectstest.CollectiveDefianceOracle, "{1}{R}{R}", []int{1}, 1, 2},
		{"Defiance 2 modes", effectstest.CollectiveDefianceOracle, "{1}{R}{R}", []int{1, 2}, 2, 2},
		{"Defiance 3 modes", effectstest.CollectiveDefianceOracle, "{1}{R}{R}", []int{0, 1, 2}, 3, 2},
		// Borrowed Hostility {R}, escalate {3}.
		{"Hostility 1 mode", effectstest.BorrowedHostilityOracle, "{R}", []int{0}, 0, 1},
		{"Hostility 2 modes", effectstest.BorrowedHostilityOracle, "{R}", []int{0, 1}, 3, 1},
		// Borrowed Grace {2}{W}, escalate {1}{W}.
		{"Grace 2 modes", effectstest.BorrowedGraceOracle, "{2}{W}", []int{0, 1}, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			gen, co := priceWithModes(t, g, tc.oracle, tc.cost, tc.modes)
			if gen != tc.wantGeneric || co != tc.wantCo {
				t.Errorf("price = generic %d + %d coloured, want %d + %d", gen, co, tc.wantGeneric, tc.wantCo)
			}
		})
	}
}

func TestEscalateManaIsRefusedWhenTheCasterCannotPayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	art := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.AdvanceToMain(t, g)
	me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"})

	// {1}{G} pays for one mode, not two ({1}{G}{G}).
	id, err := effectstest.CastEscalate(t, g, "Collective Resistance", "Instant", effectstest.CollectiveResistanceOracle, "{1}{G}",
		game.CastSpellParams{Strict: true, Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(art, 0), effectstest.EscCardRef(ench, 1)}})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("two modes on two mana: %v, want ErrInsufficientMana", err)
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast left the hand")
	}
	if _, err := effectstest.CastEscalate(t, g, "Collective Resistance", "Instant", effectstest.CollectiveResistanceOracle, "{1}{G}",
		game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{effectstest.EscCardRef(art, 0)}}); err != nil {
		t.Fatalf("one mode on two mana: %v", err)
	}
}

// --- every card's modes ------------------------------------------------

func TestCollectiveResistanceModesDoWhatTheyPrint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	art := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)

	if _, err := effectstest.CastEscalate(t, g, "Collective Resistance", "Instant", effectstest.CollectiveResistanceOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{effectstest.EscCardRef(art, 0), effectstest.EscCardRef(ench, 1), effectstest.EscCardRef(bear, 2)}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(ench) {
		t.Error("the artifact and enchantment should be destroyed")
	}
	abilities := effectstest.EffectiveAbilities(t, g, bear)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	me.Hand.Cards = nil
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 4, 4)
	effectstest.HandCard(opp, "A", "Sorcery")
	effectstest.HandCard(opp, "B", "Sorcery")
	oppHand := opp.Hand.Size()

	if _, err := effectstest.CastEscalate(t, g, "Collective Defiance", "Sorcery", effectstest.CollectiveDefianceOracle, "",
		game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{
			effectstest.EscPlayerRef(opp.ID, 0), effectstest.EscCardRef(bear, 1), effectstest.EscPlayerRef(opp.ID, 2),
		}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
		theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
		if _, err := effectstest.CastEscalate(t, g, "Borrowed Malevolence", "Instant", effectstest.BorrowedMalevolenceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(mine, 0), effectstest.EscCardRef(theirs, 1)}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, mine); p != 3 {
			t.Errorf("+1/+1 on a 2/2: power %d, want 3", p)
		}
		if g.Battlefield.Contains(theirs) {
			t.Error("-1/-1 should kill a 1/1")
		}
	})
	t.Run("Borrowed Hostility", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.Hand.Cards = nil
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
		if _, err := effectstest.CastEscalate(t, g, "Borrowed Hostility", "Instant", effectstest.BorrowedHostilityOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{effectstest.EscCardRef(mine, 0), effectstest.EscCardRef(mine, 1)}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, mine); p != 5 {
			t.Errorf("+3/+0 on a 2/2: power %d, want 5", p)
		}
		found := false
		for _, a := range effectstest.EffectiveAbilities(t, g, mine) {
			found = found || a == "first strike"
		}
		if !found {
			t.Error("first strike was not granted")
		}
	})
	t.Run("Borrowed Grace", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.Hand.Cards = nil
		a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
		if _, err := effectstest.CastEscalate(t, g, "Borrowed Grace", "Instant", effectstest.BorrowedGraceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.EffectivePower(t, g, a) != 3 || effectstest.EffectiveToughness(t, g, a) != 3 {
			t.Errorf("a 1/1 under +2/+0 and +0/+2 should be 3/3, got %d/%d", effectstest.EffectivePower(t, g, a), effectstest.EffectiveToughness(t, g, a))
		}
	})
	t.Run("Savage Alliance", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
		x := effectstest.PushVanillaCreature(g, opp.ID, "X", 3, 3)
		y := effectstest.PushVanillaCreature(g, opp.ID, "Y", 1, 1)
		if _, err := effectstest.CastEscalate(t, g, "Savage Alliance", "Instant", effectstest.SavageAllianceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1, 2}, Targets: []game.TargetRef{
				effectstest.EscPlayerRef(me.ID, 0), effectstest.EscCardRef(x, 1), effectstest.EscPlayerRef(opp.ID, 2),
			}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
			t.Error("2 damage then 1 to each should have killed both of the opponent's creatures")
		}
		found := false
		for _, a := range effectstest.EffectiveAbilities(t, g, mine) {
			found = found || a == "trample"
		}
		if !found {
			t.Error("trample was not granted to the target player's creatures")
		}
	})
	t.Run("Blessed Alliance", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		me.Hand.Cards = nil
		tapped := effectstest.PushVanillaCreature(g, me.ID, "Tapped", 2, 2)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		if _, err := effectstest.CastEscalate(t, g, "Blessed Alliance", "Instant", effectstest.BlessedAllianceOracle, "",
			game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
				effectstest.EscPlayerRef(me.ID, 0), effectstest.EscCardRef(tapped, 1),
			}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != 44 {
			t.Errorf("life = %d, want 44", me.Life)
		}
		if c := effectstest.FindBattlefieldCardByID(g, tapped); c == nil || c.Tapped {
			t.Error("the creature should be untapped")
		}
		_ = opp
	})
}
