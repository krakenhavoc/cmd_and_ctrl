package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB32KessigOID    = "d7407920-7596-497a-a135-b9343909a0cb"
	rprB32DevoteeOID   = "b3b8c7d0-46b9-4519-808a-154e503a58b6"
	rprB32RidersOID    = "60dabe9e-4872-4aec-9e8e-35358b3b7dd5"
	rprB32DeathfangOID = "69c22788-cb90-4100-a64d-891124082dd4"
	rprB32SigardaOID   = "79b56cf8-d573-40cf-b797-6b4eef2bed23"
	rprB32ReaperOID    = "894176e3-af9c-41ef-9f35-59dfdda3c17f"
	rprB32GeralfOID    = "da394a25-79c9-4b75-9ce6-741f64ae6913"
	rprB32SummonerOID  = "874e0b54-bc24-4887-abe1-1ecfd3a3abae"
	rprB32EnvoyOID     = "fbdff913-8f86-4fcb-87c3-de2011fd036d"
	rprB32BeastmasOID  = "52058c39-4edf-40f5-af18-e6de93f302a3"
	rprB32ApexOID      = "d06eedb4-bd5a-4864-9505-a570a745d165"
)

func TestRprB32CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB32KessigOID: "Kessig Cagebreakers", rprB32DevoteeOID: "Liliana's Devotee",
		rprB32RidersOID: "Riders of Gavony", rprB32DeathfangOID: "Ruthless Deathfang",
		rprB32SigardaOID: "Sigarda, Heron's Grace", rprB32ReaperOID: "Spark Reaper",
		rprB32GeralfOID: "Stitcher Geralf", rprB32SummonerOID: "Trostani's Summoner",
		rprB32EnvoyOID: "Victory's Envoy", rprB32BeastmasOID: "Wild Beastmaster",
		rprB32ApexOID: "Apex of Power",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func rprB32Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

func rprB32Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && IsToken(c) && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestRprB32VictorysEnvoyPutsACounterOnEachOtherCreatureAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	envoy := rprB32Push(g, me.ID, "Victory's Envoy", "Creature — Human Cleric", rprB32EnvoyOID, 3, 3)
	a := rprB32Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	theirs := rprB32Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if counterCount(g, a, "+1/+1") != 0 || counterCount(g, envoy, "+1/+1") != 0 {
		t.Error("it is the Envoy's controller's upkeep that counts")
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if counterCount(g, a, "+1/+1") != 1 {
		t.Errorf("A has %d counters, want 1", counterCount(g, a, "+1/+1"))
	}
	if counterCount(g, envoy, "+1/+1") != 0 || counterCount(g, theirs, "+1/+1") != 0 {
		t.Error("the Envoy and an opponent's creature get none")
	}
}

func TestRprB32TrostanisSummonerMakesThreeTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12PlayFromHand(t, g, "Trostani's Summoner", "Creature — Elf Shaman", rprB32SummonerOID, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, w := range []struct {
		name  string
		p     int
		color string
		kw    string
	}{{"Knight", 2, "W", "vigilance"}, {"Centaur", 3, "G", ""}, {"Rhino", 4, "G", "trample"}} {
		toks := rprB32Tokens(g, me.ID, w.name)
		if len(toks) != 1 {
			t.Fatalf("%s tokens: %d, want 1", w.name, len(toks))
		}
		if toks[0].Power != w.p || toks[0].Toughness != w.p || !hasKeywordOnBattlefield(t, g, toks[0].InstanceID, w.kw) && w.kw != "" {
			t.Errorf("%s token is %d/%d, kw %q", w.name, toks[0].Power, toks[0].Toughness, w.kw)
		}
	}
}

func TestRprB32WildBeastmasterPumpsOthersByItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bm := rprB32Push(g, me.ID, "Wild Beastmaster", "Creature — Human Shaman", rprB32BeastmasOID, 3, 1)
	a := rprB32Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	theirs := rprB32Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, bm)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, a); p != 5 {
		t.Errorf("A has power %d, want 5", p)
	}
	if p := effectivePower(t, g, bm); p != 3 {
		t.Errorf("the Beastmaster itself has power %d, want 3", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestRprB32KessigCagebreakersMakesAWolfPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kc := rprB32Push(g, me.ID, "Kessig Cagebreakers", "Creature — Human Rogue", rprB32KessigOID, 3, 4)
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	declareAttack(t, g, opp.ID, kc)
	passPriorityAroundTable(t, g)
	wolves := rprB32Tokens(g, me.ID, "Wolf")
	if len(wolves) != 2 {
		t.Fatalf("wolves: %d, want 2", len(wolves))
	}
	for _, w := range wolves {
		if !w.Tapped || w.AttackingTarget != opp.ID || w.Power != 2 {
			t.Errorf("wolf must be a tapped, attacking 2/2: %+v", w)
		}
	}
	if n := triggersOnStackFrom(g, kc); n != 0 {
		t.Errorf("tokens were never declared as attackers: %d triggers", n)
	}
}

func TestRprB32SparkReaperSacrificesACreatureOrPlaneswalkerToGainAndDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	reaper := rprB32Push(g, me.ID, "Spark Reaper", "Creature — Zombie", rprB32ReaperOID, 2, 3)
	fodder := rprB32Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	wall := rprB32Push(g, me.ID, "Wall", "Artifact — Equipment", "", 0, 0)
	fillPoolColored(me, "G", 3)
	// An artifact is not a legal sacrifice.
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err == nil {
		t.Fatal("an artifact paid the sacrifice")
	}
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d->%d, hand %d->%d; want +1 each", life, me.Life, hand, me.Hand.Size())
	}
	if _, ok := battlefieldCard(g, fodder); ok {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

func TestRprB32SparkReaperRefusesWithoutThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	reaper := rprB32Push(g, me.ID, "Spark Reaper", "Creature — Zombie", rprB32ReaperOID, 2, 3)
	pushLands(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{reaper}}); err == nil {
		t.Error("two lands paid {3}")
	}
}

func TestRprB32RuthlessDeathfangMakesATargetOpponentSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	fang := rprB32Push(g, me.ID, "Ruthless Deathfang", "Creature — Dragon", rprB32DeathfangOID, 4, 4)
	reaper := rprB32Push(g, me.ID, "Spark Reaper", "Creature — Zombie", rprB32ReaperOID, 2, 3)
	theirs := rprB32Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	other := rprB32Push(g, opp.ID, "Other", "Creature — Bear", "", 2, 2)
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{reaper}}); err != nil {
		t.Fatal(err)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayerTarget(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent kept the creature they chose to sacrifice")
	}
	if _, ok := battlefieldCard(g, other); !ok {
		t.Error("the opponent lost a creature they did not choose")
	}
	if _, ok := battlefieldCard(g, fang); !ok {
		t.Error("the Deathfang left")
	}
}

func TestRprB32LilianasDevoteeAnthemsZombiesAndPaysForAZombieAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dev := rprB32Push(g, me.ID, "Liliana's Devotee", "Creature — Human Warlock", rprB32DevoteeOID, 2, 3)
	zom := rprB32Push(g, me.ID, "Zed", "Creature — Zombie", "", 2, 2)
	bear := rprB32Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if effectivePower(t, g, zom) != 3 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, dev) != 2 {
		t.Errorf("anthem: zombie %d bear %d devotee %d", effectivePower(t, g, zom), effectivePower(t, g, bear), effectivePower(t, g, dev))
	}
	// Nothing died: no trigger, no token.
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(rprB32Tokens(g, me.ID, "Zombie")); n != 0 {
		t.Errorf("zombie tokens without a death: %d", n)
	}
}

func TestRprB32LilianasDevoteeOffersThePaymentAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rprB32Push(g, me.ID, "Liliana's Devotee", "Creature — Human Warlock", rprB32DevoteeOID, 2, 3)
	victim := rprB32Push(g, me.ID, "Victim", "Creature — Bear", "", 1, 1)
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment offer after a creature died")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	toks := rprB32Tokens(g, me.ID, "Zombie")
	if len(toks) != 1 || toks[0].Power != 2 {
		t.Errorf("zombie tokens: %d", len(toks))
	}
}

func TestRprB32RidersOfGavonyGivesHumansProtectionFromTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	riders := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Riders of Gavony", TypeLine: "Creature — Human Knight", OracleID: rprB32RidersOID,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, NamedTribe: "Vampire",
	})
	human := rprB32Push(g, me.ID, "Hu", "Creature — Human Soldier", "", 1, 1)
	bear := rprB32Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	vamp := rprB32Push(g, opp.ID, "Vamp", "Creature — Vampire", "", 2, 2)
	elf := rprB32Push(g, opp.ID, "Elf", "Creature — Elf", "", 2, 2)
	kindred := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Kin", TypeLine: "Kindred Enchantment — Vampire", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, c := range []struct {
		id   uuid.UUID
		src  uuid.UUID
		want bool
		note string
	}{
		{human, vamp, false, "a Vampire creature can't target a Human"},
		{human, elf, true, "an Elf can"},
		{bear, vamp, true, "a non-Human has no protection"},
		{riders, vamp, false, "the Riders are Humans"},
		{human, kindred, true, "a non-creature carrying the type does not match"},
	} {
		tgt := battlefieldCardPtr(t, g, c.id)
		src := battlefieldCardPtr(t, g, c.src)
		if got := game.CanBeTargetedBy(tgt, game.ZoneBattlefield, game.SourceObject(opp.ID, src)); got != c.want {
			t.Errorf("%s: targetable = %v, want %v", c.note, got, c.want)
		}
	}
}

func TestRprB32SigardaGivesYouAndHumansHexproofAndMakesHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sig := rprB32Push(g, me.ID, "Sigarda, Heron's Grace", "Legendary Creature — Angel", rprB32SigardaOID, 4, 5)
	human := rprB32Push(g, me.ID, "Hu", "Creature — Human Soldier", "", 1, 1)
	bear := rprB32Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !playerHasHexproof(g, me) || playerHasHexproof(g, opp) {
		t.Error("you have hexproof, your opponents don't")
	}
	if !hasKeywordOnBattlefield(t, g, human, KeywordHexproof) || hasKeywordOnBattlefield(t, g, bear, KeywordHexproof) || hasKeywordOnBattlefield(t, g, sig, KeywordHexproof) {
		t.Error("exactly the Humans have hexproof")
	}
	fillPoolColored(me, "G", 2)
	gy := game.Card{InstanceID: uuid.New(), Name: "Old", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(gy)
	if err := g.ActivateCatalogAbility(me.ID, sig, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the exile cost was skipped")
	}
	if err := g.ActivateCatalogAbility(me.ID, sig, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{gy.InstanceID}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	toks := rprB32Tokens(g, me.ID, "Human Soldier")
	if len(toks) != 1 || toks[0].Power != 1 {
		t.Errorf("Human Soldier tokens: %d", len(toks))
	}
	if me.Graveyard.Size() != 0 {
		t.Error("the card was not exiled")
	}
}

func TestRprB32StitcherGeralfMillsExilesCreaturesAndMakesAZombieOfTheirPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	geralf := rprB32Push(g, me.ID, "Stitcher Geralf", "Legendary Creature — Human Wizard", rprB32GeralfOID, 3, 4)
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "G", 2)
	var wantTotal int
	var big []uuid.UUID
	for i, s := range g.Seats {
		// Three known cards on top of every library: a land, a 2-power
		// creature and a 3-power creature (the last pushed is the top).
		isle := game.Card{InstanceID: uuid.New(), Name: "Isle", TypeLine: "Basic Land — Island", Owner: s.ID, Controller: s.ID}
		small := game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: s.ID, Controller: s.ID}
		bigC := game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: s.ID, Controller: s.ID}
		s.Library.PushTop(isle)
		s.Library.PushTop(small)
		s.Library.PushTop(bigC)
		if i == 0 {
			big = append(big, bigC.InstanceID, small.InstanceID)
			wantTotal = 5
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, geralf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	pc := pendingChooseCards(t, g, me.ID)
	if len(pc.ChooseCards) != 8 {
		t.Errorf("candidates: %d, want the 8 milled creature cards", len(pc.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, big...)
	passPriorityAroundTable(t, g)
	toks := rprB32Tokens(g, me.ID, "Zombie")
	if len(toks) != 1 || toks[0].Power != wantTotal || toks[0].Toughness != wantTotal {
		t.Fatalf("zombie tokens %+v, want one %d/%d", toks, wantTotal, wantTotal)
	}
	for _, id := range big {
		if !inExile(g, id) {
			t.Error("a picked card was not exiled")
		}
	}
	if len(g.Exile.Cards) != 2 {
		t.Errorf("exile holds %d cards, want 2", len(g.Exile.Cards))
	}
}

func TestRprB32ApexOfPowerExilesSevenAddsTenFromHandOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Apex of Power", "Sorcery", rprB32ApexOID, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-7 {
		t.Errorf("library %d -> %d, want 7 fewer", lib, me.Library.Size())
	}
	if len(g.Exile.Cards) != 7 {
		t.Errorf("exile holds %d, want 7", len(g.Exile.Cards))
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no color choice for the ten mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("R"); n != 10 || me.ManaPool.Total() != 10 || manaPickFor(g, me.ID) != nil {
		t.Errorf("mana pool holds %d red of %d, want ten of one color", n, me.ManaPool.Total())
	}
}

func TestRprB32StitcherGeralfWithNoCreatureCardMilledMakesADyingZeroZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	geralf := rprB32Push(g, me.ID, "Stitcher Geralf", "Legendary Creature — Human Wizard", rprB32GeralfOID, 3, 4)
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, geralf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if toks := rprB32Tokens(g, me.ID, "Zombie"); len(toks) != 0 {
		t.Errorf("a 0/0 Zombie survived: %d", len(toks))
	}
	if len(g.Exile.Cards) != 0 {
		t.Error("nothing was to be exiled")
	}
}
