package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bbd_5529_b1_cards_test.go — Battlebond, batch 1 (#5529). Every test
// plays the real catalog entry through the engine.

const (
	bbdArchonOracle          = "baff00b8-b109-4e31-942b-38aeb1e45ab2"
	bbdArenaRectorOracle     = "6a23963c-21dc-4d21-8d28-5abad7f859a1"
	bbdAuroraChampionOracle  = "a4499cf9-04a0-45c5-9c73-19f8e7e95da6"
	bbdBladeseekerOracle     = "ac4d065e-7eb6-4224-8a7c-e57f2c1a2733"
	bbdBrambleSovereignOracl = "5e26a688-67d7-48a5-8862-653851bf3784"
	bbdBrightlingOracle      = "909ff4a8-8bdd-4478-bd43-ce5175a02d68"
	bbdBullRushBruiserOracle = "d5a7412a-2258-4d3e-9e0f-a150cc69081e"
	bbdChakramSlingerOracle  = "5a2568fd-4c22-413e-87e2-a384002ded38"
	bbdComboAttackOracle     = "037eccdb-21c3-4048-af4f-b25dcbe99134"
	bbdDecoratedChampOracle  = "e7c0052e-6dda-4fcf-9368-7749aa9b9c5c"
	bbdInnerDemonOracle      = "2cee861f-d28c-47e6-baee-2da9b7fa89d8"
	bbdJungleWayfinderOracle = "794ad296-4718-45b7-ab85-6ae45903ada7"
	bbdLastOneStandingOracle = "b01465ea-cab9-4a59-81a6-8db4648b89f4"
	bbdNimbusChampionOracle  = "3a536065-d62a-44c9-886b-84209b6bce57"
	bbdOkaunOracle           = "7dec219b-96d1-4d4c-b3ac-15eda9d2ccc6"
	bbdRegnaOracle           = "0e14f59f-021a-4ad3-a26b-8f79e680c426"
	bbdRushbladeOracle       = "14613521-a939-426e-9bd5-183186b51583"
	bbdSaltwaterOracle       = "04472c4f-0df3-4de8-9cc2-882d95b822be"
	bbdSickleDancerOracle    = "836bfcad-014a-40cc-ab79-906fcab723a0"
	bbdSoaringShowOffOracle  = "5ca0f4f3-6133-4f53-9124-f63942796a07"
	bbdCorrupterOracle       = "8d708573-10b8-44fb-b61b-9ad65e3eb465"
	bbdRenewerOracle         = "7ccc2740-e659-452d-95e8-5dee958322fa"
	bbdStadiumVendorsOracle  = "84904dd3-8987-47c2-b508-f87519157b00"
	bbdStolenStrategyOracle  = "6807167a-e290-4259-940e-3cbbfa75d0a1"
	bbdThrasherBruteOracle   = "8ac09649-3b6c-40fa-81f8-583384c9e0bb"
	bbdVirtusOracle          = "6da4f16c-f187-479e-b19b-78a137b1abe5"
	bbdZndrspltOracle        = "502849a6-8e65-40f3-b348-a41c4f939768"

	bbdWarriorType = "Creature — Human Warrior"
	bbdBearType    = "Creature — Bear"
)

// bbdSettle resolves the stack, answering a trigger order in the order
// offered, every player target with `player`, every card target with
// those of `cards` it offers, and every yes/no with no.
func bbdSettle(t *testing.T, g *game.Game, player uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if effectstest.StackFullyEmpty(g) {
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			continue
		}
		var err error
		switch open.Kind {
		case game.PendingChoiceTriggerOrder:
			err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
		case game.PendingChoicePickTarget:
			if len(open.PickTargetPlayers) > 0 {
				err = g.ResolvePickTarget(open.ID, open.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player})
			} else {
				var picks []uuid.UUID
				for _, id := range cards {
					if effectstest.HasID(open.PickTargetCards, id) {
						picks = append(picks, id)
					}
				}
				err = g.ResolvePickTargets(open.ID, open.Chooser, effectstest.CardRefs(picks...))
			}
		case game.PendingChoiceConfirm:
			err = g.ResolveConfirm(open.ID, open.Chooser, false)
		default:
			t.Fatalf("unexpected %v prompt", open.Kind)
		}
		if err != nil {
			t.Fatalf("answering %v: %v", open.Kind, err)
		}
	}
	t.Fatal("the table did not settle")
}

// bbdAnswerOption answers the open option pick owed by `chooser` with the
// option labelled `label`.
func bbdAnswerOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	c := effectstest.LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s", chooser)
	}
	for i, opt := range c.PickOptions {
		if opt.Label == label {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick(%q): %v", label, err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered: %v", label, effectstest.OptionWords(c))
}

// bbdCreatures counts the creatures on the battlefield.
func bbdCreatures(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

// --- every card is in the catalog ----------------------------------------

func TestBbd5529EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bbdArchonOracle: "Archon of Valor's Reach", bbdArenaRectorOracle: "Arena Rector",
		bbdAuroraChampionOracle: "Aurora Champion", bbdBladeseekerOracle: "Azra Bladeseeker",
		bbdBrambleSovereignOracl: "Bramble Sovereign", bbdBrightlingOracle: "Brightling",
		bbdBullRushBruiserOracle: "Bull-Rush Bruiser", bbdChakramSlingerOracle: "Chakram Slinger",
		bbdComboAttackOracle: "Combo Attack", bbdDecoratedChampOracle: "Decorated Champion",
		bbdInnerDemonOracle: "Inner Demon", bbdJungleWayfinderOracle: "Jungle Wayfinder",
		bbdLastOneStandingOracle: "Last One Standing", bbdNimbusChampionOracle: "Nimbus Champion",
		bbdOkaunOracle: "Okaun, Eye of Chaos", bbdRegnaOracle: "Regna, the Redeemer",
		bbdRushbladeOracle: "Rushblade Commander", bbdSaltwaterOracle: "Saltwater Stalwart",
		bbdSickleDancerOracle: "Sickle Dancer", bbdSoaringShowOffOracle: "Soaring Show-Off",
		bbdCorrupterOracle: "Soulblade Corrupter", bbdRenewerOracle: "Soulblade Renewer",
		bbdStadiumVendorsOracle: "Stadium Vendors", bbdStolenStrategyOracle: "Stolen Strategy",
		bbdThrasherBruteOracle: "Thrasher Brute", bbdVirtusOracle: "Virtus the Veiled",
		bbdZndrspltOracle: "Zndrsplt, Eye of Wisdom",
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

// --- printed keywords -------------------------------------------------------

func TestBbd5529KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Archon of Valor's Reach", "Creature — Archon", bbdArchonOracle, []string{"flying", "vigilance", "trample"}},
		{"Nimbus Champion", "Creature — Avatar Warrior", bbdNimbusChampionOracle, []string{"flying"}},
		{"Regna, the Redeemer", "Legendary Creature — Angel", bbdRegnaOracle, []string{"flying"}},
		{"Soaring Show-Off", "Creature — Bird Warrior", bbdSoaringShowOffOracle, []string{"flying"}},
		{"Soulblade Corrupter", bbdWarriorType, bbdCorrupterOracle, []string{"deathtouch"}},
		{"Virtus the Veiled", "Legendary Creature — Azra Assassin", bbdVirtusOracle, []string{"deathtouch"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- "if your team controls another Warrior" -----------------------------------

func TestBbd5529AuroraChampionTapsATargetWithAnotherWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Aurora Champion", "Creature — Elf Warrior", bbdAuroraChampionOracle, 3, 2)
	effectstest.B12Creature(g, me.ID, "Other Warrior", bbdWarriorType, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", bbdBearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, champ)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, victim) {
		t.Error("the target was not tapped")
	}
}

func TestBbd5529WarriorAttackTriggersNeedAnotherWarrior(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Aurora Champion", "Creature — Elf Warrior", bbdAuroraChampionOracle},
		{"Bull-Rush Bruiser", "Creature — Minotaur Warrior", bbdBullRushBruiserOracle},
		{"Sickle Dancer", bbdWarriorType, bbdSickleDancerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 2)
			effectstest.B12Creature(g, me.ID, "Bear", bbdBearType, 2, 2)
			effectstest.B12Creature(g, opp.ID, "Their Warrior", bbdWarriorType, 2, 2)
			effectstest.DeclareAttack(t, g, opp.ID, id)
			if n := effectstest.TriggersOnStackFrom(g, id); n != 0 {
				t.Errorf("%d triggers with no other Warrior of yours", n)
			}
			if effectstest.LatestPickTarget(g, me.ID) != nil {
				t.Error("a target was asked for with no other Warrior of yours")
			}
		})
	}
}

func TestBbd5529BullRushBruiserGainsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := effectstest.B12Push(g, me.ID, "Bull-Rush Bruiser", "Creature — Minotaur Warrior", bbdBullRushBruiserOracle, 4, 3)
	effectstest.B12Creature(g, me.ID, "Other Warrior", bbdWarriorType, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, bruiser)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bruiser, "first strike") {
		t.Error("the Bruiser has no first strike")
	}
}

func TestBbd5529SickleDancerGetsPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := effectstest.B12Push(g, me.ID, "Sickle Dancer", bbdWarriorType, bbdSickleDancerOracle, 3, 2)
	effectstest.B12Creature(g, me.ID, "Other Warrior", bbdWarriorType, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, dancer)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, dancer); p != 4 || tough != 3 {
		t.Errorf("the Dancer is %d/%d, want 4/3", p, tough)
	}
}

// --- Warrior tribal ------------------------------------------------------------

func TestBbd5529RushbladeCommanderHastesYourWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander := effectstest.B12Push(g, me.ID, "Rushblade Commander", "Creature — Azra Warrior", bbdRushbladeOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Warrior", bbdWarriorType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", bbdBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", bbdWarriorType, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, commander, "haste") || !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("a Warrior you control has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a non-Warrior has haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's Warrior has haste")
	}
}

func TestBbd5529DecoratedChampionGrowsForAnotherWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	champ := effectstest.B12Push(g, me.ID, "Decorated Champion", "Creature — Elf Warrior", bbdDecoratedChampOracle, 2, 2)
	effectstest.CastHoldCreature(t, g, "Bear", bbdBearType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, champ, game.CounterPlusOne); n != 0 {
		t.Fatalf("a non-Warrior entering put %d counters on it", n)
	}
	effectstest.CastHoldCreature(t, g, "New Warrior", bbdWarriorType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, champ, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters after a Warrior entered, want 1", n)
	}
}

func TestBbd5529ThrasherBruteDrainsForItselfAndAnotherWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastHoldCreature(t, g, "Thrasher Brute", "Creature — Orc Warrior", bbdThrasherBruteOracle, 4, 3)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Fatalf("after the Brute entered: you %d (want %d), opponent %d (want %d)", me.Life, myLife+1, opp.Life, oppLife-1)
	}
	effectstest.CastHoldCreature(t, g, "New Warrior", bbdWarriorType, "", 2, 2)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 || opp.Life != oppLife-2 {
		t.Errorf("after another Warrior entered: you %d (want %d), opponent %d (want %d)", me.Life, myLife+2, opp.Life, oppLife-2)
	}
}

// --- enters triggers ------------------------------------------------------------

func TestBbd5529AzraBladeseekerRummages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastHoldCreature(t, g, "Azra Bladeseeker", "Creature — Azra Warrior", bbdBladeseekerOracle, 3, 2)
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	pitch := me.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d", got, hand)
	}
	if got := me.Graveyard.Size(); got != graves+1 || !me.Graveyard.Contains(pitch) {
		t.Errorf("graveyard %d, want %d with the discarded card", got, graves+1)
	}
}

func TestBbd5529SoaringShowOffDrawsForEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	effectstest.CastHoldCreature(t, g, "Soaring Show-Off", "Creature — Bird Warrior", bbdSoaringShowOffOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != before[i]+1 {
			t.Errorf("seat %d hand %d, want %d", i, got, before[i]+1)
		}
	}
}

func TestBbd5529JungleWayfinderLetsEachPlayerFetchABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	want := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		want[p.ID] = effectstest.B07LibraryLand(p, "Forest", "Basic Land — Forest")
	}
	effectstest.CastHoldCreature(t, g, "Jungle Wayfinder", "Creature — Elf Warrior", bbdJungleWayfinderOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	// Seat 1 declines; everyone else takes the Forest.
	for i, p := range g.Seats {
		if i == 1 {
			effectstest.AnswerSearchFailToFind(t, g, p.ID)
			continue
		}
		effectstest.AnswerSearchByID(t, g, p.ID, want[p.ID])
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		inHand := p.Hand.Contains(want[p.ID])
		if i == 1 && inHand {
			t.Error("seat 1 declined but the Forest reached its hand")
		}
		if i != 1 && !inHand {
			t.Errorf("seat %d: the Forest is not in hand", i)
		}
	}
}

func TestBbd5529SoulbladeRenewerSupportsTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", bbdBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", bbdBearType, 2, 2)
	renewer := effectstest.CastHoldCreature(t, g, "Soulblade Renewer", "Creature — Elf Warrior", bbdRenewerOracle, 2, 2)
	bbdSettle(t, g, me.ID, mine, theirs, renewer)
	if effectstest.CountersOn(g, mine, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 1 {
		t.Error("support did not put a +1/+1 counter on each target")
	}
	if effectstest.CountersOn(g, renewer, game.CounterPlusOne) != 0 {
		t.Error("the Renewer supported itself")
	}
}

func TestBbd5529StadiumVendorsGiveTheChosenPlayerTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastHoldCreature(t, g, "Stadium Vendors", "Creature — Goblin", bbdStadiumVendorsOracle, 3, 3)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	p := effectstest.ManaPickFor(g, opp.ID)
	if p == nil {
		t.Fatal("the chosen player was not asked for a colour")
	}
	if err := g.ResolveManaChoice(p.ID, opp.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := opp.ManaPool.Total(); got != 2 {
		t.Errorf("the chosen player has %d mana, want 2", got)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("the Vendors' controller has %d mana, want 0", got)
	}
}

func TestBbd5529ArchonOfValorsReachBansTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastHoldCreature(t, g, "Archon of Valor's Reach", "Creature — Archon", bbdArchonOracle, 5, 6)
	bbdAnswerOption(t, g, me.ID, "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Some Instant", "Instant", "", nil); err == nil {
		t.Error("an instant was cast with instant chosen")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Some Sorcery", "Sorcery", "", nil); err != nil {
		t.Errorf("a sorcery was refused with instant chosen: %v", err)
	}
}

// --- combat and damage triggers ---------------------------------------------------

func TestBbd5529SoulbladeCorrupterGivesCounteredAttackersDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Soulblade Corrupter", bbdWarriorType, bbdCorrupterOracle, 3, 3)
	countered := effectstest.B12Creature(g, me.ID, "Countered Bear", bbdBearType, 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", bbdBearType, 2, 2)
	effectstest.B28AddCounter(t, g, countered, 1)
	effectstest.DeclareAttack(t, g, opp.ID, countered, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, countered, "deathtouch") {
		t.Error("the attacker with a +1/+1 counter has no deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "deathtouch") {
		t.Error("an attacker with no counter gained deathtouch")
	}
}

func TestBbd5529NimbusChampionBouncesWhatTheWarriorsCanCover(t *testing.T) {
	for _, tc := range []struct {
		name     string
		warriors int
		bounced  bool
	}{
		{"three Warriors", 2, true},
		{"one Warrior", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			nimbus := effectstest.B12Push(g, me.ID, "Nimbus Champion", "Creature — Avatar Warrior", bbdNimbusChampionOracle, 4, 4)
			for i := 0; i < tc.warriors; i++ {
				effectstest.B12Creature(g, me.ID, "Other Warrior", bbdWarriorType, 1, 1)
			}
			victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
			effectstest.DeclareAttack(t, g, opp.ID, nimbus)
			effectstest.AnswerPickTarget(t, g, victim)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if got := opp.Hand.Contains(victim); got != tc.bounced {
				t.Errorf("in its owner's hand: %v, want %v", got, tc.bounced)
			}
		})
	}
}

func TestBbd5529SaltwaterStalwartMakesTheTargetDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	stalwart := effectstest.B12Push(g, me.ID, "Saltwater Stalwart", "Creature — Merfolk Warrior", bbdSaltwaterOracle, 2, 4)
	hand := opp.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, stalwart, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("the target's hand is %d, want %d", got, hand+1)
	}
}

func TestBbd5529VirtusHalvesTheDamagedPlayersLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	virtus := effectstest.B12Push(g, me.ID, "Virtus the Veiled", "Legendary Creature — Azra Assassin", bbdVirtusOracle, 1, 1)
	life := opp.Life
	effectstest.DealCombatDamage(g, me.ID, virtus, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if want := life - (life+1)/2; opp.Life != want {
		t.Errorf("life %d, want %d", opp.Life, want)
	}
}

func TestBbd5529RegnaMakesWarriorsOnlyAfterLifeGain(t *testing.T) {
	for _, gained := range []bool{true, false} {
		name := "gained"
		if !gained {
			name = "no gain"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Regna, the Redeemer", "Legendary Creature — Angel", bbdRegnaOracle, 4, 4)
			if gained {
				g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			want := 0
			if gained {
				want = 2
			}
			if got := effectstest.CountOnBattlefield(g, "Warrior", me.ID); got != want {
				t.Errorf("%d Warrior tokens, want %d", got, want)
			}
		})
	}
}

// --- activated abilities -------------------------------------------------------------

func TestBbd5529ChakramSlingerPingsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slinger := effectstest.B12Push(g, me.ID, "Chakram Slinger", bbdWarriorType, bbdChakramSlingerOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, slinger, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if !effectstest.HmlTapped(t, g, slinger) {
		t.Error("the Slinger did not tap")
	}
}

func TestBbd5529BrightlingAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Brightling", "Creature — Shapeshifter", bbdBrightlingOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("the grants are missing")
	}
	effectstest.B16Activate(t, g, me.ID, id, 3, game.ActivateAbilityParams{})
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 2 {
		t.Errorf("after +1/-1 it is %d/%d, want 4/2", p, tough)
	}
	effectstest.B16Activate(t, g, me.ID, id, 2, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
		t.Error("the Brightling did not return to its owner's hand")
	}
}

// --- spells ------------------------------------------------------------------------

func TestBbd5529ComboAttackHitsWithBothPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear", bbdBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Combo Attack", "Sorcery", bbdComboAttackOracle, effectstest.CardRefs(a, b, victim))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("five damage did not kill the 0/4")
	}
}

func TestBbd5529LastOneStandingLeavesOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Bear", bbdBearType, 2, 2)
	effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Last One Standing", "Sorcery", bbdLastOneStandingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := bbdCreatures(g); n != 1 {
		t.Errorf("%d creatures left, want 1", n)
	}
}

func TestBbd5529InnerDemonPumpsTheHostAndShrinksTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Bear", bbdBearType, 2, 2)
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Inner Demon", "Enchantment — Aura", bbdInnerDemonOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("the enchanted Demon is %d/%d, want 4/4 (spared the shrink)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature has no flying")
	}
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(goblin) {
		t.Error("the 2/2 survived -2/-2")
	}
}

// --- other triggers ----------------------------------------------------------------

func TestBbd5529StolenStrategyExilesEachOpponentsTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Stolen Strategy", "Enchantment", bbdStolenStrategyOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	libs := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}
	exiled := g.Exile.Size()
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := libs[i] - 1
		if i == 0 {
			want = libs[i]
		}
		if got := p.Library.Size(); got != want {
			t.Errorf("seat %d library %d, want %d", i, got, want)
		}
	}
	if got := g.Exile.Size(); got != exiled+len(g.Seats)-1 {
		t.Errorf("exile %d, want %d", got, exiled+len(g.Seats)-1)
	}
}

func TestBbd5529BrambleSovereignCopiesAnEnteringCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bramble Sovereign", "Creature — Dryad", bbdBrambleSovereignOracl, 4, 4)
	effectstest.CastHoldCreature(t, g, "Bear", bbdBearType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Bear", me.ID); got != 2 {
		t.Errorf("%d Bears, want the creature and its token copy", got)
	}
}

func TestBbd5529OkaunDoublesOnAnyWonFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	okaun := effectstest.B12Push(g, me.ID, "Okaun, Eye of Chaos", "Legendary Creature — Cyclops Berserker", bbdOkaunOracle, 3, 3)
	effectstest.EmitCoinFlip(g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, okaun); p != 3 || tough != 3 {
		t.Fatalf("a lost flip made it %d/%d", p, tough)
	}
	effectstest.EmitCoinFlip(g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, okaun); p != 6 || tough != 6 {
		t.Errorf("after an opponent's won flip it is %d/%d, want 6/6", p, tough)
	}
}

func TestBbd5529ZndrspltDrawsOnAnyWonFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Zndrsplt, Eye of Wisdom", "Legendary Creature — Homunculus", bbdZndrspltOracle, 1, 4)
	hand := me.Hand.Size()
	effectstest.EmitCoinFlip(g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestBbd5529FlipUntilYouLoseAsksAtBeginningOfCombat(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Okaun, Eye of Chaos", "Legendary Creature — Cyclops Berserker", bbdOkaunOracle},
		{"Zndrsplt, Eye of Wisdom", "Legendary Creature — Homunculus", bbdZndrspltOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			effectstest.PassPriorityAroundTable(t, g)
			call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
			if call == nil || call.Chooser != me.ID {
				t.Fatal("no coin call for the controller at the beginning of combat")
			}
		})
	}
}

func TestBbd5529ArenaRectorTradesItselfForAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 3})
	rector := effectstest.PushDiesCreatureForTest(g, me.ID, "Arena Rector", bbdArenaRectorOracle, "Creature — Human Cleric", 1, 2)
	effectstest.B18Kill(t, g, rector)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rector) {
		t.Error("the Rector was not exiled")
	}
	if me.Library.Contains(walker) || !g.Battlefield.Contains(walker) {
		t.Error("the planeswalker did not reach the battlefield")
	}
}
