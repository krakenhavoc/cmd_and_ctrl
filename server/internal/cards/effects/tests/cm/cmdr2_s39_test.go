package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s39CommitMemoryOracle   = "56fd8895-3be2-4591-86fa-87567d9cdc14"
	s39HeavenEarthOracle    = "03d3fee8-f7de-4152-8c18-6d279cc6c125"
	s39ConniveConcoctOracle = "6e07e7ae-1bb7-4831-adb9-a5e71b7d79e3"
	s39DiscoveryOracle      = "8f1c1bd6-8420-4de2-81f4-5faa419da301"
)

func s39CommitMemory(p *game.Player, zone string) uuid.UUID {
	return effectstest.S39Split(p, s39CommitMemoryOracle, zone,
		game.Face{Name: "Commit", TypeLine: "Instant", ManaCost: "{3}{U}",
			OracleText: "Put target spell or nonland permanent into its owner's library second from the top."},
		effectstest.S39Aftermath("Memory", "{4}{U}{U}", "Each player shuffles their hand and graveyard into their library, then draws seven cards."))
}

func TestCommitPutsAPermanentSecondFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	top, _ := opp.Library.Top()
	id := s39CommitMemory(me, "hand")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast Commit: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, bear) {
		t.Fatal("the creature is still on the battlefield")
	}
	n := opp.Library.Size()
	if opp.Library.Cards[n-2].InstanceID != bear {
		t.Error("the permanent is not second from the top of its owner's library")
	}
	if opp.Library.Cards[n-1].InstanceID != top.InstanceID {
		t.Error("the old top card is no longer on top")
	}
}

func TestCommitRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID,
	})
	id := s39CommitMemory(me, "hand")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("Commit accepted a land")
	}
}

func TestCommitPutsASpellSecondFromTheTopWithoutCounteringIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.CastCatalogSpell(t, g, "Plain Instant", "Instant", "", nil)
	effectstest.S39Fund(g, me)
	top, _ := me.Library.Top()
	id := s39CommitMemory(me, "hand")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err != nil {
		t.Fatalf("cast Commit at a spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := me.Library.Size()
	if me.Library.Cards[n-2].InstanceID != spell {
		t.Error("the spell is not second from the top of its owner's library")
	}
	if me.Library.Cards[n-1].InstanceID != top.InstanceID {
		t.Error("the old top card is no longer on top")
	}
	if me.Graveyard.Contains(spell) {
		t.Error("the spell was countered into the graveyard")
	}
}

func TestMemoryIsAftermathAndWheelsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := s39CommitMemory(me, "graveyard")
	oppGrave := effectstest.PushGraveyardCardForTest(opp, "Their Fodder")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0, FromZone: "graveyard"}); err == nil {
		t.Fatal("Commit was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Memory from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := p.Hand.Size(); got != 7 {
			t.Errorf("%s: hand %d, want 7", p.Name, got)
		}
	}
	if opp.Graveyard.Contains(oppGrave) {
		t.Error("the opponent's graveyard card was not shuffled away")
	}
	if !g.Exile.Contains(id) {
		t.Error("Memory is not exiled")
	}
}

func s39HeavenEarth(p *game.Player, zone string) uuid.UUID {
	return effectstest.S39Split(p, s39HeavenEarthOracle, zone,
		game.Face{Name: "Heaven", TypeLine: "Instant", ManaCost: "{X}{G}",
			OracleText: "Heaven deals X damage to each creature with flying."},
		game.Face{Name: "Earth", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}",
			OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nEarth deals X damage to each creature without flying."})
}

func s39FlyerAndWalker(g *game.Game, owner uuid.UUID) (flyer, walker uuid.UUID) {
	flyer = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird",
		TypeLine: "Creature — Bird", Power: 1, Toughness: 2, Keywords: []string{"flying"}, Owner: owner, Controller: owner})
	walker = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: owner, Controller: owner})
	return
}

func TestHeavenHitsOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer, walker := s39FlyerAndWalker(g, opp.ID)
	id := s39HeavenEarth(me, "hand")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 0, XValue: 2}); err != nil {
		t.Fatalf("cast Heaven: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, flyer) {
		t.Error("the flyer survived Heaven for 2")
	}
	if !effectstest.OnBattlefieldNow(g, walker) {
		t.Error("a creature without flying was hit by Heaven")
	}
}

func TestEarthIsAftermathAndHitsOnlyNonFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer, walker := s39FlyerAndWalker(g, opp.ID)
	id := s39HeavenEarth(me, "graveyard")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0, FromZone: "graveyard", XValue: 2}); err == nil {
		t.Fatal("Heaven was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard", XValue: 2}); err != nil {
		t.Fatalf("cast Earth from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, walker) {
		t.Error("the ground creature survived Earth for 2")
	}
	if !effectstest.OnBattlefieldNow(g, flyer) {
		t.Error("a flyer was hit by Earth")
	}
	if !g.Exile.Contains(id) {
		t.Error("Earth is not exiled")
	}
}

func s39ConniveConcoct(p *game.Player, zone string) uuid.UUID {
	return effectstest.S39Split(p, s39ConniveConcoctOracle, zone,
		game.Face{Name: "Connive", TypeLine: "Sorcery", ManaCost: "{2}{U/B}{U/B}",
			OracleText: "Gain control of target creature with power 2 or less."},
		game.Face{Name: "Concoct", TypeLine: "Sorcery", ManaCost: "{3}{U}{B}",
			OracleText: "Surveil 3, then return a creature card from your graveyard to the battlefield."})
}

func TestConniveStealsASmallCreatureAndRefusesALargeOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 5)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	id := s39ConniveConcoct(me, "hand")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Fatal("Connive accepted a power 3 creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("cast Connive: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, small)
	if c.Controller != me.ID {
		t.Errorf("controller %v, want the caster", c.Controller)
	}
}

func TestConcoctSurveilsThenReanimatesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	// A creature card on top of the library, to be surveilled into the graveyard.
	surveiled := effectstest.SeedLibraryTop(me, "Surveiled Giant", "Creature — Giant", "")
	id := s39ConniveConcoct(me, "hand")
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Concoct: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	var keep []uuid.UUID
	for _, id := range c.ScryCards {
		if id != surveiled {
			keep = append(keep, id)
		}
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{surveiled}, keep); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature pick after the surveil")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[dead] || !offered[surveiled] {
		t.Fatalf("offer %v lacks the old or the just-surveilled creature", offered)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, surveiled)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, surveiled)
	if !ok {
		t.Fatal("the surveilled creature did not return")
	}
	if card.Tapped || card.Controller != me.ID {
		t.Errorf("tapped %v controller %v, want untapped under the caster", card.Tapped, card.Controller)
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("the other creature card left the graveyard")
	}
}

// s39Cast moves the table to a main phase, funds me, and casts card `id`
// from wherever it is with params.
func s39Cast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, params game.CastSpellParams) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.S39Fund(g, me)
	return g.CastSpell(me.ID, id, params)
}

func s39OnBattlefield(g *game.Game, owner uuid.UUID, name, typeLine, cost string, power, toughness int, kw ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		ManaCost: cost, Power: power, Toughness: toughness, Keywords: kw, Owner: owner, Controller: owner})
}

func TestDiscoverySurveilsTwoThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.S39Split(me, "8f1c1bd6-8420-4de2-81f4-5faa419da301", "hand",
		game.Face{Name: "Discovery", TypeLine: "Sorcery", ManaCost: "{1}{U/B}",
			OracleText: "Surveil 2, then draw a card."},
		game.Face{Name: "Dispersal", TypeLine: "Instant", ManaCost: "{3}{U}{B}",
			OracleText: "Each opponent returns a nonland permanent they control with the greatest mana value among permanents they control to its owner's hand, then discards a card."})
	if err := s39Cast(t, g, me, id, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatalf("cast Discovery: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	gone, kept := c.ScryCards[:1], c.ScryCards[1:]
	if err := g.ResolveSurveil(c.ID, me.ID, gone, kept); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if !me.Graveyard.Contains(gone[0]) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

func s39Dispersal(p *game.Player) uuid.UUID {
	return effectstest.S39Split(p, "8f1c1bd6-8420-4de2-81f4-5faa419da301", "hand",
		game.Face{Name: "Discovery", TypeLine: "Sorcery", ManaCost: "{1}{U/B}",
			OracleText: "Surveil 2, then draw a card."},
		game.Face{Name: "Dispersal", TypeLine: "Instant", ManaCost: "{3}{U}{B}",
			OracleText: "Each opponent returns a nonland permanent they control with the greatest mana value among permanents they control to its owner's hand, then discards a card."})
}

func TestDispersalReturnsTheGreatestNonlandThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, quiet := g.Seats[0], g.Seats[1], g.Seats[2]
	big := s39OnBattlefield(g, opp.ID, "Big", "Creature — Giant", "{4}{G}", 5, 5)
	small := s39OnBattlefield(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	land := s39OnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	quietLand := s39OnBattlefield(g, quiet.ID, "Island", "Basic Land — Island", "", 0, 0)
	myBig := s39OnBattlefield(g, me.ID, "Mine", "Creature — Giant", "{7}", 7, 7)
	oppHand, quietHand := opp.Hand.Size(), quiet.Hand.Size()
	if err := s39Cast(t, g, me, s39Dispersal(me), game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Dispersal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, big) || !effectstest.InHandOf(opp, big) {
		t.Error("the greatest-mana-value permanent did not return to its owner's hand")
	}
	if !effectstest.OnBattlefieldNow(g, small) || !effectstest.OnBattlefieldNow(g, land) || !effectstest.OnBattlefieldNow(g, quietLand) || !effectstest.OnBattlefieldNow(g, myBig) {
		t.Error("something other than the opponent's greatest nonland permanent returned")
	}
	// The returned card is in hand when the discard is chosen: it may be the one pitched.
	effectstest.AnswerDiscard(t, g, opp.ID, big)
	effectstest.AnswerDiscard(t, g, quiet.ID, quiet.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("opponent hand %d, want %d (one back, one discarded)", got, oppHand)
	}
	if !opp.Graveyard.Contains(big) {
		t.Error("the returned permanent was not the card discarded")
	}
	if got := quiet.Hand.Size(); got != quietHand-1 {
		t.Errorf("an opponent with only a land hand %d, want %d: no return, one discard", got, quietHand-1)
	}
}

func TestDispersalAsksOnATieAmongPermanentsAndCountsLandsAtZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := s39OnBattlefield(g, opp.ID, "A", "Artifact", "{3}", 0, 0)
	b := s39OnBattlefield(g, opp.ID, "B", "Creature — Bear", "{2}{G}", 3, 3)
	if err := s39Cast(t, g, me, s39Dispersal(me), game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Dispersal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("pick %+v, want a choice between the two tied permanents", pick)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	if !effectstest.InHandOf(opp, b) || !effectstest.OnBattlefieldNow(g, a) {
		t.Error("the chosen permanent should be the one returned")
	}
}

func TestRadicalIdeaJumpStartDiscardsAndExiles(t *testing.T) {
	const oracle = "fd51dce2-8e8a-4686-910e-1bbc2825d190"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "Radical Idea", "Instant", oracle)
	fodder := effectstest.HandCard(me, "Fodder", "Sorcery")
	hand := me.Hand.Size()
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start"}); err == nil {
		t.Fatal("jump-start was castable without discarding a card")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start", AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("jump-start cast: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discard is a cost, paid as the spell is cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (one discarded, one drawn)", got, hand)
	}
	if !g.Exile.Contains(id) {
		t.Error("the jump-started instant is not exiled")
	}
}

func TestElectricRevelationDiscardsThenDrawsTwiceAndFlashesBack(t *testing.T) {
	const oracle = "c92af0f4-e5f1-4731-a781-b2ca360867a7"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()
	effectstest.S39Fund(g, me)
	_, paid := effectstest.CastWithDiscard(t, g, "Electric Revelation", oracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatal("the discard should be paid as the spell is cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	// Flashback: the discard is still owed, and the card is exiled.
	id := effectstest.SeedGraveyardCard(t, g, "Electric Revelation 2", "Instant", oracle)
	discard := effectstest.HandCard(me, "Fodder", "Sorcery")
	effectstest.S39Fund(g, me)
	before = me.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashback was castable without the additional discard")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "flashback", DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("flashback drew %d, want 2", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card is not exiled")
	}
}

func TestRecklessSpiteDestroysTwoNonblackAndCostsFive(t *testing.T) {
	const oracle = "a684df3a-5441-4daa-86d1-c47a91b35e6a"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie",
		TypeLine: "Creature — Zombie", Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	targets := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Reckless Spite", "Instant", oracle, targets(a, black)); err == nil {
		t.Fatal("Reckless Spite accepted a black creature")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Reckless Spite", "Instant", oracle, targets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, a) || effectstest.OnBattlefieldNow(g, b) {
		t.Error("a target survived")
	}
	if !effectstest.OnBattlefieldNow(g, black) {
		t.Error("the black creature was destroyed")
	}
	if me.Life != life-5 {
		t.Errorf("life %d → %d, want -5", life, me.Life)
	}
}

func TestNightmareUnmakingExilesByPowerAgainstHandSize(t *testing.T) {
	const oracle = "285fecb7-f872-454d-ac2f-bc5615c70cd8"
	for _, tc := range []struct {
		mode     int
		survives []int // powers that stay, of 1, 3, 5
	}{
		{0, []int{1}},    // hand 2: power greater than 2 goes
		{1, []int{3, 5}}, // hand 2: power less than 2 goes
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Hand.Cards = nil
		effectstest.HandCard(me, "One", "Sorcery")
		effectstest.HandCard(me, "Two", "Sorcery")
		ids := map[int]uuid.UUID{}
		for _, p := range []int{1, 3, 5} {
			ids[p] = effectstest.B12Creature(g, opp.ID, "P", "Creature — Bear", p, 5)
		}
		effectstest.CastCatalogSpellWithModes(t, g, "Nightmare Unmaking", "Sorcery", oracle, []int{tc.mode})
		effectstest.PassPriorityAroundTable(t, g)
		for p, id := range ids {
			want := false
			for _, s := range tc.survives {
				want = want || s == p
			}
			if got := effectstest.OnBattlefieldNow(g, id); got != want {
				t.Errorf("mode %d: power %d on battlefield = %v, want %v", tc.mode, p, got, want)
			}
		}
	}
}

func TestCastOffDestroysEveryNonGiantAndLeavesTheAdventureCard(t *testing.T) {
	const oracle = "a1d17244-9c92-4094-9843-d7ee31c85ea4"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	card := game.Card{InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutAdventure, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Realm-Cloaked Giant", TypeLine: "Creature — Giant", ManaCost: "{5}{W}{W}", Power: 7, Toughness: 7, Keywords: []string{"vigilance"}},
			{Name: "Cast Off", TypeLine: "Sorcery — Adventure", ManaCost: "{3}{W}{W}"},
		}}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := 0; i < 5; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "W"})
		}
	})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Face: 1}); err != nil {
		t.Fatalf("cast Cast Off: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, bear) || effectstest.OnBattlefieldNow(g, mine) {
		t.Error("a non-Giant creature survived Cast Off")
	}
	if !effectstest.OnBattlefieldNow(g, giant) {
		t.Error("a Giant was destroyed")
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile, ready to be cast as the Giant")
	}
}

func TestCalamityOfCindersConvokesAndHitsOnlyUntappedCreatures(t *testing.T) {
	const oracle = "fbcf0c1c-29a7-4e83-8b06-869f83be49cd"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := s39OnBattlefield(g, me.ID, "Red A", "Creature — Goblin", "{R}", 2, 2)
	b := s39OnBattlefield(g, me.ID, "Red B", "Creature — Goblin", "{R}", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	tappedGiant := s39OnBattlefield(g, opp.ID, "Tapped Giant", "Creature — Giant", "{6}", 6, 6)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tappedGiant {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Calamity of Cinders", TypeLine: "Sorcery", ManaCost: "{5}{R}{R}",
		OracleID: oracle, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for i := 0; i < 5; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("five mana and one convoked creature paid a seven-mana spell")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("cast with convoke: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, bystander) {
		t.Error("an untapped creature survived 6 damage")
	}
	if !effectstest.OnBattlefieldNow(g, a) || !effectstest.OnBattlefieldNow(g, b) {
		t.Error("a creature tapped for convoke was damaged")
	}
	if !effectstest.OnBattlefieldNow(g, tappedGiant) {
		t.Error("a tapped creature was damaged")
	}
}

func TestStarstormDamagesEveryCreatureAndCycles(t *testing.T) {
	const oracle = "34d112c3-cc2d-4428-bd7f-568bda61bd30"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 2)
	flyer := s39OnBattlefield(g, me.ID, "Bird", "Creature — Bird", "", 1, 2, "flying")
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 3)
	effectstest.PaCast(t, g, "Starstorm", "Instant", "{X}{R}{R}", oracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, small) || effectstest.OnBattlefieldNow(g, flyer) {
		t.Error("a creature with toughness 2 survived X=2")
	}
	if !effectstest.OnBattlefieldNow(g, big) {
		t.Error("a toughness 3 creature died to X=2")
	}
	g2 := effectstest.NewCatalogGame(t)
	hand := g2.Seats[0].Hand.Size()
	id, p := effectstest.CycleFromHand(t, g2, "Starstorm", "Instant", oracle, "{C}{C}{C}")
	effectstest.PassPriorityAroundTable(t, g2)
	if !p.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	if got := p.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after cycling, want %d (cycled card out, one drawn)", got, hand+1)
	}
}

func TestDecreeOfPainDestroysAllDrawsPerCreatureAndCyclesForMinusTwo(t *testing.T) {
	const oracle = "932668fa-d6e3-41c0-ad0c-8e0a00e68d11"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	}
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Decree of Pain", "Sorcery", oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			t.Errorf("%s survived Decree of Pain", c.Name)
		}
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d: a card for each of three creatures", got, hand+3)
	}

	g2 := effectstest.NewCatalogGame(t)
	small := effectstest.B12Creature(g2, g2.Seats[1].ID, "Small", "Creature — Bear", 2, 2)
	tough := effectstest.B12Creature(g2, g2.Seats[1].ID, "Tough", "Creature — Wall", 0, 4)
	id, p := effectstest.CycleFromHand(t, g2, "Decree of Pain", "Sorcery", oracle, "{C}{C}{C}{B}{B}")
	effectstest.PassPriorityAroundTable(t, g2)
	if !p.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	if effectstest.OnBattlefieldNow(g2, small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if !effectstest.OnBattlefieldNow(g2, tough) {
		t.Fatal("the 0/4 died")
	}
	if c := effectstest.S58p6View(t, g2, tough); c.CurrentToughness() != 2 {
		t.Errorf("toughness %d, want 4-2", c.CurrentToughness())
	}
}

func TestShatterTheSkyDrawsForPowerFourControllersThenDestroysAll(t *testing.T) {
	const oracle = "2ffc7509-c7ce-44c4-a8cf-90ed161b5d10"
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Creature(g, me.ID, "Mine Small", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Big A", "Creature — Giant", 4, 4)
	effectstest.B12Creature(g, opp.ID, "Big B", "Creature — Giant", 6, 6)
	effectstest.B12Creature(g, third.ID, "Third Small", "Creature — Bear", 2, 2)
	hands := []int{me.Hand.Size(), opp.Hand.Size(), third.Hand.Size()}
	effectstest.CastCatalogSpell(t, g, "Shatter the Sky", "Sorcery", oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hands[0] || third.Hand.Size() != hands[2] {
		t.Error("a player with no power 4 creature drew")
	}
	if got := opp.Hand.Size(); got != hands[1]+1 {
		t.Errorf("opponent hand %d, want %d: one card though they control two big creatures", got, hands[1]+1)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			t.Errorf("%s survived", c.Name)
		}
	}
}
