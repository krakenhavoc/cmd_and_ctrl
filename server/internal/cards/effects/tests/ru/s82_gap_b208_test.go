package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b208_test.go — cards missing from sets since AFR, batch 208 (#5948):
// Double Masters 2022 instants, sorceries and planeswalkers. Each test plays the
// real catalog entry and asserts what its text says.

const (
	gap208Touch    = "15692698-ef57-4672-bf76-5fe4a00c693a"
	gap208Aminatou = "3a30089d-cd2d-49be-9b06-7a2454117692"
	gap208Charm    = "2a9a05b4-1910-4f97-a0fe-0f57017f6df0"
	gap208Breath   = "e6fb4f0d-ce49-4913-a2ef-88365e4916cf"
	gap208Liliana  = "64a9a6b4-26b5-4d23-be86-85c54e9c4727"
	gap208Leak     = "c61fe162-2202-4e56-9ba0-393547f9875f"
	gap208Glory    = "e8aa12f8-8922-41ed-8cdb-219c9e4b9e36"
	gap208Blessing = "a7157ec5-9e6a-4771-b3db-6760640d6c6c"
	gap208Bloom    = "e5df4597-1647-4ac2-bdb3-a517598d1431"
	gap208Thrive   = "fe3c524e-10b7-485a-9285-973092eeb214"
)

func TestGap208EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap208Touch:    "Aethermage's Touch",
		gap208Aminatou: "Aminatou, the Fateshifter",
		gap208Charm:    "Jeskai Charm",
		gap208Breath:   "Last Breath",
		gap208Liliana:  "Liliana, the Last Hope",
		gap208Leak:     "Mana Leak",
		gap208Glory:    "Martial Glory",
		gap208Blessing: "Scale Blessing",
		gap208Bloom:    "Summer Bloom",
		gap208Thrive:   "Thrive",
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

// --- Mana Leak ---------------------------------------------------------------

func TestGap208ManaLeakCountersUnlessTheCasterPaysThree(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Mana Leak", "Instant", gap208Leak, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		if pay {
			effectstest.B06AddMana(me, "C", "C", "C")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		_, resolved := effectstest.BattlefieldCard(g, victim)
		if pay && !resolved {
			t.Error("the spell was countered although its controller paid {3}")
		}
		if !pay && (resolved || !me.Graveyard.Contains(victim)) {
			t.Error("the spell was not countered")
		}
	}
}

// --- Last Breath -------------------------------------------------------------

func TestGap208LastBreathExilesPowerTwoOrLessAndItsControllerGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Last Breath", "Instant", gap208Breath, effectstest.B16TargetCard(big)); err == nil {
		t.Fatal("Last Breath targeted a power-3 creature")
	}
	life := opp.Life
	effectstest.Ecl6Cast(t, g, me, "Last Breath", "Instant", gap208Breath, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(small), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the Wolf stayed")
	}
	if opp.Life != life+4 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life+4)
	}
}

// --- Martial Glory -----------------------------------------------------------

func TestGap208MartialGloryPumpsOnePowerAndAnotherToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Martial Glory", "Instant", gap208Glory, 0, 0, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, a); p != 5 || tt != 2 {
		t.Errorf("the first target is %d/%d, want 5/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, b); p != 1 || tt != 4 {
		t.Errorf("the second target is %d/%d, want 1/4", p, tt)
	}
}

// --- Thrive ------------------------------------------------------------------

func TestGap208ThrivePutsACounterOnEachOfXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Thrive", "Sorcery", gap208Thrive, 0, 0, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 1, b: 1, c: 0} {
		if card, _ := effectstest.BattlefieldCard(g, id); card.Counters[game.CounterPlusOne] != want {
			t.Errorf("%s has %d counters, want %d", card.Name, card.Counters[game.CounterPlusOne], want)
		}
	}
}

// --- Scale Blessing ----------------------------------------------------------

func TestGap208ScaleBlessingBolstersThenGrowsEveryCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	small := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	marked := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == marked {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Scale Blessing", "Instant", gap208Blessing, 0, 0, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{small: 2, marked: 2, plain: 0} {
		if card, _ := effectstest.BattlefieldCard(g, id); card.Counters[game.CounterPlusOne] != want {
			t.Errorf("%s has %d counters, want %d", card.Name, card.Counters[game.CounterPlusOne], want)
		}
	}
}

// --- Summer Bloom ------------------------------------------------------------

func TestGap208SummerBloomGrantsThreeExtraLandPlays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 0 {
		t.Fatalf("extra land plays before the spell = %d", got)
	}
	effectstest.Ecl6Cast(t, g, me, "Summer Bloom", "Sorcery", gap208Bloom, 0, 0, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 3 {
		t.Errorf("extra land plays = %d, want 3", got)
	}
}

// --- Jeskai Charm ------------------------------------------------------------

func TestGap208JeskaiCharmTucksDamagesAndTeamPumps(t *testing.T) {
	// Mode 0: put target creature on top of its owner's library.
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpellWithModesTargets(t, g, gap208Charm, []int{0}, effectstest.B16TargetCard(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, foe); ok {
		t.Error("the Wolf stayed on the battlefield")
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1]; top.InstanceID != foe && opp.Library.Cards[0].InstanceID != foe {
		t.Error("the Wolf is not on top of its owner's library")
	}

	// Mode 1: 4 damage to target opponent.
	g2 := effectstest.NewCatalogGame(t)
	_, opp2 := effectstest.Ecl6Seats(g2)
	effectstest.AdvanceToMain(t, g2)
	life := opp2.Life
	effectstest.CastCatalogSpellWithModesTargets(t, g2, gap208Charm, []int{1}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	effectstest.PassPriorityAroundTable(t, g2)
	if opp2.Life != life-4 {
		t.Errorf("the opponent is at %d, want %d", opp2.Life, life-4)
	}

	// Mode 2: creatures you control get +1/+1 and lifelink.
	g3 := effectstest.NewCatalogGame(t)
	me3, opp3 := effectstest.Ecl6Seats(g3)
	mine := effectstest.B12Creature(g3, me3.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g3, opp3.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g3)
	effectstest.CastCatalogSpellWithModesTargets(t, g3, gap208Charm, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g3)
	if p, tt := effectstest.PtOf(t, g3, mine); p != 3 || tt != 3 {
		t.Errorf("my creature is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g3, mine, "lifelink") {
		t.Error("my creature has no lifelink")
	}
	if p, _ := effectstest.PtOf(t, g3, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g3, theirs, "lifelink") {
		t.Error("an opponent's creature was changed")
	}
	_ = me
}

// --- Liliana, the Last Hope --------------------------------------------------

func TestGap208LilianaShrinksMillsAndReturnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, the Last Hope", "Legendary Planeswalker — Liliana", gap208Liliana, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	effectstest.B16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(foe)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, lili) != 4 {
		t.Errorf("loyalty %d after the +1, want 4", effectstest.LoyaltyCount(g, lili))
	}
	if p, tt := effectstest.PtOf(t, g, foe); p != 1 || tt != 2 {
		t.Errorf("the Wolf is %d/%d, want 1/2", p, tt)
	}
	// It lasts until my next turn: through the opponent's turn.
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, _ := effectstest.PtOf(t, g, foe); p != 1 {
		t.Errorf("the shrink ended early: power %d", p)
	}
}

func TestGap208LilianaMinusTwoMillsTwoThenMayReturnACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, the Last Hope", "Legendary Planeswalker — Liliana", gap208Liliana, 3)
	effectstest.SeedLibrary(me, "Mill A", "Mill B", "Deep")
	dead := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	effectstest.B16Activate(t, g, me.ID, lili, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, lili) != 1 {
		t.Errorf("loyalty %d after the -2, want 1", effectstest.LoyaltyCount(g, lili))
	}
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no prompt to return a creature card")
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.InHandOf(me, dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestGap208LilianaUltimateEmblemMakesTwoPlusZombiesEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, the Last Hope", "Legendary Planeswalker — Liliana", gap208Liliana, 7)
	effectstest.B16Activate(t, g, me.ID, lili, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Zombie", me.ID); n != 2 {
		t.Fatalf("%d Zombies after the first end step, want 2", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, indexOf208(g, me))
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Zombie", me.ID); n != 6 {
		t.Errorf("%d Zombies after the second end step, want 2 + (2 + 2) = 6", n)
	}
}

func indexOf208(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s == p {
			return i
		}
	}
	return -1
}

// --- Aminatou, the Fateshifter -----------------------------------------------

func TestGap208AminatouLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	am := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Aminatou, the Fateshifter", "Legendary Planeswalker — Aminatou", gap208Aminatou, 3)
	hand := effectstest.HandCount(g, me)
	effectstest.B16Activate(t, g, me.ID, am, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no prompt to put a card back")
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{me.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := effectstest.HandCount(g, me); got != hand {
		t.Errorf("hand %d, want %d (drew one, put one back)", got, hand)
	}
}

func TestGap208AminatouMinusOneBlinksAnotherPermanentYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	am := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Aminatou, the Fateshifter", "Legendary Planeswalker — Aminatou", gap208Aminatou, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	effectstest.B16Activate(t, g, me.ID, am, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, am) != 2 {
		t.Errorf("loyalty %d after the -1, want 2", effectstest.LoyaltyCount(g, am))
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" {
			n++
			if c.Counters[game.CounterPlusOne] != 0 {
				t.Error("the Bear kept its counters, so it never left")
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Bears back, want 1", n)
	}
}

func TestGap208AminatouMinusSixRotatesNonlandPermanentsAroundTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.ToMain(t, g)
	am := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Aminatou, the Fateshifter", "Legendary Planeswalker — Aminatou", gap208Aminatou, 7)
	owned := map[int]uuid.UUID{}
	lands := map[int]uuid.UUID{}
	for i, s := range g.Seats {
		owned[i] = effectstest.B12Creature(g, s.ID, "Creature", "Creature — Test", 1, 1)
		lands[i] = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: s.ID, Controller: s.ID})
	}
	effectstest.B16Activate(t, g, me.ID, am, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			pick = c
		}
	}
	if pick == nil {
		t.Fatal("no left/right prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == game.AttackDirectionLeftWord {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no left option")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Left is the next player in turn order: each player gets the permanents
	// the next seat controlled.
	for i := range g.Seats {
		next := (i + 1) % len(g.Seats)
		if got := effectstest.ControllerOf(t, g, owned[next]); got != g.Seats[i].ID {
			t.Errorf("seat %d's creature is controlled by %s, want seat %d", next, got, i)
		}
		if got := effectstest.ControllerOf(t, g, lands[i]); got != g.Seats[i].ID {
			t.Errorf("seat %d's land changed hands", i)
		}
	}
	if got := effectstest.ControllerOf(t, g, am); got != me.ID {
		t.Error("Aminatou changed hands")
	}
}

// --- Aethermage's Touch ------------------------------------------------------

func TestGap208AethermagesTouchPutsACreatureOntoTheBattlefieldThatReturnsAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Deep Two", "Deep One")
	var creature uuid.UUID
	g.WithWriteLock(func() {
		creature = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: creature, Name: "Revealed Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		for _, n := range []string{"Top C", "Top B", "Top A"} {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.Ecl6Cast(t, g, me, "Aethermage's Touch", "Instant", gap208Touch, 0, 0, game.CastSpellParams{})
	effectstest.Ecl6ResolveSpell(t, g)
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no prompt to put a creature card onto the battlefield")
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// The rest go on the bottom in an order the player picks: answer it if asked.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary && c.Chooser == me.ID {
			if err := g.ResolvePutInLibrary(c.ID, me.ID, c.ScryCards, nil); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	var bear uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Revealed Bear" {
			bear = c.InstanceID
		}
	}
	if bear == uuid.Nil {
		t.Fatal("the creature card did not enter")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the creature stayed after the end step")
	}
	var back bool
	for _, c := range me.Hand.Cards {
		if c.Name == "Revealed Bear" {
			back = true
		}
	}
	if !back {
		t.Error("the creature did not return to its owner's hand")
	}
}
