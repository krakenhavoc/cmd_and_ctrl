package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b626_test.go — Judgment cards, S82 OLD batch 626 (#5611), played
// through the real catalog entries.

const (
	o626Thriss     = "19ed8616-5d27-4f21-88a9-2af7cf5034ca"
	o626Liberated  = "c5081f2d-ebe2-4ff3-a8ef-fe50f8ecde60"
	o626Boiler     = "cbbca098-0b33-4693-8ea7-a2005b79af3d"
	o626Trainee    = "fce68545-c31d-4dc2-9f1b-aadf924a07d1"
	o626Crab       = "4216f9ce-bdd9-4213-8c3f-e2ef4cd848f9"
	o626Rats       = "6c14ecbf-0097-4fd8-97a3-a32b532c21e0"
	o626Turtle     = "771f9ba4-6c65-4497-ad48-869994435556"
	o626Newt       = "be014a79-5550-4f79-8789-b4d6783d7ec8"
	o626Epic       = "6f34ee2c-6e0b-4c71-9bdd-6ac55a323849"
	o626Bonds      = "708a76cc-d4ab-4dfd-98dd-815f5bc99364"
	o626Mist       = "e78673c4-a969-443f-b1a4-0eb37b8111af"
	o626PlanarChao = "f0864dbb-4d91-4282-a774-7b225e649ca8"
)

func TestS82OldB626CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o626Thriss: "Thriss, Nantuko Primus", o626Liberated: "Liberated Dwarf", o626Boiler: "Dwarven Bloodboiler",
		o626Trainee: "Cabal Trainee", o626Crab: "Wormfang Crab", o626Rats: "Earsplitting Rats",
		o626Turtle: "Wormfang Turtle", o626Newt: "Wormfang Newt", o626Epic: "Epic Struggle",
		o626Bonds: "Telekinetic Bonds", o626Mist: "Mist of Stagnation", o626PlanarChao: "Planar Chaos",
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

func o626Activate(t *testing.T, g *game.Game, who *game.Player, source uuid.UUID, p game.ActivateAbilityParams) {
	t.Helper()
	p.Strict = true
	if err := g.ActivateCatalogAbility(who.ID, source, 0, p); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func o626Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestS82OldB626ThrissPumpsFivePerActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thriss := effectstest.B12Push(g, me.ID, "Thriss, Nantuko Primus", "Legendary Creature — Insect Druid", o626Thriss, 5, 5)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.PayMana(t, g, me, "{G}")
	o626Activate(t, g, me, thriss, game.ActivateAbilityParams{Targets: o626Target(bear)})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("pumped Bear = %d/%d, want 7/7", p, tough)
	}
	if !effectstest.TappedOnBattlefield(t, g, thriss) {
		t.Error("Thriss did not tap")
	}
}

func TestS82OldB626LiberatedDwarfPumpsOnlyAGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dwarf := effectstest.B12Push(g, me.ID, "Liberated Dwarf", "Creature — Dwarf", o626Liberated, 1, 1)
	red := effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	green := effectstest.B27Push(g, me.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	effectstest.PayMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{Strict: true, Targets: o626Target(red)}); err == nil {
		t.Fatal("the Dwarf's ability targeted a non-green creature")
	}
	o626Activate(t, g, me, dwarf, game.ActivateAbilityParams{Targets: o626Target(green)})
	if p, tough := effectstest.EffectivePTOf(t, g, green); p != 3 || tough != 2 {
		t.Errorf("pumped green Bear = %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, green, "first strike") {
		t.Error("the green Bear has no first strike")
	}
	if effectstest.FindBattlefieldCardForTest(g, dwarf) != nil {
		t.Error("the Dwarf was not sacrificed")
	}
}

func TestS82OldB626BloodboilerTapsADwarfEvenItselfWhileSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boiler := effectstest.B12Push(g, me.ID, "Dwarven Bloodboiler", "Creature — Dwarf", o626Boiler, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == boiler {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	bear := effectstest.AuraBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, boiler, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}, Targets: o626Target(bear)}); err == nil {
		t.Fatal("a non-Dwarf paid the cost")
	}
	o626Activate(t, g, me, boiler, game.ActivateAbilityParams{TapIDs: []uuid.UUID{boiler}, Targets: o626Target(bear)})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("pumped Bear = %d/%d, want 4/2", p, tough)
	}
	if !effectstest.TappedOnBattlefield(t, g, boiler) {
		t.Error("the Bloodboiler did not tap itself")
	}
}

func TestS82OldB626CabalTraineeShrinksAPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trainee := effectstest.B12Push(g, me.ID, "Cabal Trainee", "Creature — Human Minion", o626Trainee, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	o626Activate(t, g, me, trainee, game.ActivateAbilityParams{Targets: o626Target(theirs)})
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 3 {
		t.Errorf("shrunk creature = %d/%d, want 1/3", p, tough)
	}
	if effectstest.FindBattlefieldCardForTest(g, trainee) != nil {
		t.Error("the Trainee was not sacrificed")
	}
}

func TestS82OldB626EarsplittingRatsEveryPlayerDiscardsAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Earsplitting Rats", "Creature — Rat", o626Rats, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for round := 0; round < 8; round++ {
		answered := false
		for _, p := range g.Seats {
			if effectstest.DiscardChoiceFor(g, p.ID) != nil {
				effectstest.DiscardFromHand(t, g, p.ID)
				answered = true
			}
		}
		if !answered {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for i, p := range g.Seats {
		want := hands[i] - 1
		if p.Hand.Size() != want {
			t.Errorf("seat %d hand = %d, want %d", i, p.Hand.Size(), want)
		}
	}
	var rats uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == o626Rats {
			rats = c.InstanceID
		}
	}
	if rats == uuid.Nil {
		t.Fatal("the Rats did not enter")
	}
	pitch := effectstest.HandCardForTest(me, "Pitch", "Sorcery", "")
	o626Activate(t, g, me, rats, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}})
	if me.Hand.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rats) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, rats) == nil {
		t.Error("the Rats died instead of regenerating")
	}
}

func TestS82OldB626WormfangNewtAndTurtleExileALandAndReturnIt(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
	}{{"Wormfang Newt", o626Newt, "{1}{U}"}, {"Wormfang Turtle", o626Turtle, "{2}{U}"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
		theirs := effectstest.B12Push(g, g.Seats[1].ID, "Their Island", "Basic Land — Island", "", 0, 0)
		creature := effectstest.CastCatalogSpell(t, g, tc.name, "Creature — Nightmare Beast", tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if len(g.PendingChoices) == 0 {
			t.Fatalf("%s: no land choice was asked", tc.name)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.CountOnBattlefield(g, "Island", me.ID) != 0 {
			t.Fatalf("%s: the land is still on the battlefield", tc.name)
		}
		if effectstest.FindBattlefieldCardForTest(g, theirs) == nil {
			t.Errorf("%s: an opponent's land was exiled", tc.name)
		}
		_ = land
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(creature) })
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.CountOnBattlefield(g, "Island", me.ID) != 1 {
			t.Errorf("%s: the exiled land did not return when it left", tc.name)
		}
	}
}

func TestS82OldB626WormfangCrabIsUnblockableAndAnOpponentExilesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	crab := effectstest.CastCatalogSpell(t, g, "Wormfang Crab", "Creature — Nightmare Crab", o626Crab, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.AuraRestrictions(t, g, crab); r&game.CantBeBlocked == 0 {
		t.Error("the Crab can be blocked")
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.AnswerChooseCards(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "My Bear", me.ID) != 0 {
		t.Fatal("the chosen permanent is still on the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, other) == nil || effectstest.FindBattlefieldCardForTest(g, crab) == nil {
		// the other Bear and the Crab stay
		t.Error("something other than the chosen permanent was exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(crab) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "My Bear", me.ID) != 1 {
		t.Error("the exiled permanent did not return when the Crab left")
	}
}

func TestS82OldB626EpicStruggleWinsAtTwentyCreatures(t *testing.T) {
	for _, tc := range []struct {
		creatures int
		wins      bool
	}{{19, false}, {20, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Epic Struggle", "Enchantment", o626Epic, false)
		for i := 0; i < tc.creatures; i++ {
			effectstest.B12Creature(g, me.ID, "Token", "Creature — Bear", 1, 1)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		seat, ok := g.WinnerSeat()
		won := ok && seat == 1
		if won != tc.wins {
			t.Errorf("%d creatures: won = %v, want %v", tc.creatures, won, tc.wins)
		}
	}
}

func TestS82OldB626TelekineticBondsTapsOrUntapsOnAnyPlayersDiscard(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pay    bool
		pick   int
		tapped bool
	}{{"pay and tap", true, 0, true}, {"pay and untap", true, 1, false}, {"decline", false, 0, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Telekinetic Bonds", "Enchantment", o626Bonds, false)
			bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear",
				TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: !tc.tapped && tc.pay, Owner: opp.ID, Controller: opp.ID})
			card := game.NewCard("Pitched Card", opp.ID)
			opp.Hand.PushTop(card)
			g.WithWriteLock(func() { g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp.ID, Source: uuid.New(), N: 1}) })
			effectstest.AnswerDiscard(t, g, opp.ID, card.InstanceID)
			effectstest.PassPriorityUntilChoice(t, g)
			if c := effectstest.LatestPickTarget(g, me.ID); c != nil {
				effectstest.AnswerPickTarget(t, g, bear)
				effectstest.PassPriorityUntilChoice(t, g)
			}
			if !effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatalf("an opponent's discard did not offer the Bonds' controller the payment: %+v", g.PendingChoices)
			}
			if tc.pay {
				effectstest.PayMana(t, g, me, "{U}{C}")
			}
			effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			effectstest.MonSettle(t, g)
			if tc.pay {
				effectstest.AnswerOptionPick(t, g, me.ID, tc.pick)
			}
			effectstest.PassPriorityAroundTable(t, g)
			tapped, _ := effectstest.BattlefieldCardTapped(g, bear)
			want := tc.tapped
			if !tc.pay {
				want = false
			}
			if tapped != want {
				t.Errorf("tapped = %v, want %v", tapped, want)
			}
		})
	}
}

func TestS82OldB626MistOfStagnationUntapsOnePermanentPerGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mist of Stagnation", "Enchantment", o626Mist, false)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
			TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID}))
	}
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Dead 1", "Sorcery")
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Dead 2", "Sorcery")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	for i, b := range bears {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, b); !tapped {
			t.Fatalf("bear %d untapped during the untap step", i)
		}
	}
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, bears[0], bears[2])
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range []bool{false, true, false} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, bears[i]); tapped != want {
			t.Errorf("bear %d tapped = %v, want %v", i, tapped, want)
		}
	}
}

// o626Flip answers the open coin call heads and reports whether the flip was
// won.
func o626Flip(t *testing.T, g *game.Game, who uuid.UUID) bool {
	t.Helper()
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if ask.Chooser != who {
		t.Fatalf("the coin call was asked of %s, want %s", ask.Chooser, who)
	}
	flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
	if err := g.ResolveCoinCall(ask.ID, who, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	evs := effectstest.RandomEvents(g, game.EventFlipCoin)
	if len(evs) != flips+1 {
		t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
	}
	return evs[len(evs)-1].Won
}

func TestS82OldB626PlanarChaosSacrificesOnALostUpkeepFlip(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		chaos := effectstest.PushCatalogPermanent(g, me.ID, "Planar Chaos", "Enchantment", o626PlanarChao, false)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		w := o626Flip(t, g, me.ID)
		here := effectstest.FindBattlefieldCardForTest(g, chaos) != nil
		if w {
			won++
		} else {
			lost++
		}
		if here != w {
			t.Errorf("won %v: the enchantment is on the battlefield = %v", w, here)
		}
	}
	if lost == 0 || won == 0 {
		t.Fatalf("saw %d lost and %d won flips; want both", lost, won)
	}
}

func TestS82OldB626PlanarChaosCountersASpellWhoseCasterLosesTheFlip(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[1], g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Planar Chaos", "Enchantment", o626PlanarChao, false)
		spell := effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		w := o626Flip(t, g, opp.ID)
		onBattlefield := effectstest.FindBattlefieldCardForTest(g, spell) != nil
		if w {
			won++
		} else {
			lost++
		}
		if onBattlefield != w {
			t.Errorf("won %v: the spell resolved = %v", w, onBattlefield)
		}
		if !w && !opp.Graveyard.Contains(spell) {
			t.Error("a countered spell is not in its owner's graveyard")
		}
	}
	if lost == 0 || won == 0 {
		t.Fatalf("saw %d lost and %d won flips; want both", lost, won)
	}
}
