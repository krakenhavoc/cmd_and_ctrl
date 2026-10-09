package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_spell_a_test.go — the Reality Fracture instants and
// sorceries of slice fra-spell-a.

const (
	rfArtifistAcumenOracle    = "129b41cf-8327-42f0-b186-1111b42c08ab"
	rfBestialIncursionOracle  = "bc00f29a-e9d2-4b86-bb22-33b686ac4365"
	rfCastAwayDoubtOracle     = "bcf29dd7-6d5a-4139-958e-7c59434c0770"
	rfChargeTheSanctumOracle  = "9d835137-584c-46bd-952e-4b2812bba1a3"
	rfClashOfElementsOracle   = "78304a06-4c9a-4cdf-b67f-067d88762381"
	rfCommandTheStageOracle   = "13bfc51f-d079-40ef-a4ba-47d06d3150a3"
	rfCruelCalculationsOracle = "69b96fe5-9733-4b43-bd66-075742e142d9"
	rfGenerousRevivalOracle   = "f8601f5f-0f7c-4689-9d15-137f1e82c49c"
	rfGerminateRecruitsOracle = "2c3a1313-a66e-4317-9732-e4ff9288b0db"
	rfIcyReceptionOracle      = "612c7807-be70-4e2f-afd2-3d3c03eddbbe"
	rfKindredJudgmentOracle   = "b5dce42a-a769-4d7d-b29e-8722f79d4092"
	rfKonstrariCharmOracle    = "40961b23-b153-434a-aeb9-b9c2ff613dcb"
	rfMultiplyByZeroOracle    = "d940611f-f84d-43fa-82d6-eecb4fb54164"
	rfPerfectedTheoryOracle   = "d824a319-fa74-42e4-9cbe-dba1576d6bad"
	rfPreciseRedactionOracle  = "c251b676-0e98-4047-bed1-d72c83aa8da0"
)

func rfCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// rfCastModal casts a modal instant from the active seat's hand with
// the given modes and targets.
func rfCastModal(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, "Instant", "", oracle, nil)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// rfPT is the creature's power and toughness INCLUDING counters, which
// effectivePower / effectiveToughness (layer output only) leave out.
func rfPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.CurrentPower(), c.CurrentToughness()
}

func rfFirstStrike(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return hasAbility(effectiveAbilities(t, g, id), "first strike")
}

func rfCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return allCountersOn(t, g, id)[game.CounterPlusOne]
}

func TestCastAwayDoubtDrawsTwoAndHitsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Cast Away Doubt", "Sorcery", rfCastAwayDoubtOracle, nil)
	passPriorityAroundTable(t, g)
	// The spell leaves the hand (-1) and draws two.
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand size = %d, want %d (the cast card is pushed into hand first, then +2 drawn)", got, before+2)
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("seat %s life = %d, want %d", p.ID, p.Life, lives[p.ID]-2)
		}
	}
}

func TestArtifistAcumenGivesFirstStrikeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Artifist Acumen", "Sorcery", rfArtifistAcumenOracle, nil)
	passPriorityAroundTable(t, g)
	if !rfFirstStrike(t, g, mine) {
		t.Error("my creature should have first strike")
	}
	if rfFirstStrike(t, g, theirs) {
		t.Error("an opponent's creature must not gain first strike")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d (one card drawn net of the cast)", got, before+1)
	}
}

func TestPreciseRedactionCountersWhiteAndRefusesGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	white := handCardFull(me, "White Thing", "Sorcery", "", "", []string{"W"})
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white spell: %v", err)
	}
	castCatalogSpell(t, g, "Precise Redaction", "Instant", rfPreciseRedactionOracle, rfCardTarget(white))
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(white) {
		t.Error("the white spell should have been countered into the graveyard")
	}

	green := handCardFull(me, "Green Thing", "Sorcery", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast green spell: %v", err)
	}
	if err := castCatalogSpellErr(t, g, "Precise Redaction", "Instant", rfPreciseRedactionOracle, rfCardTarget(green)); err == nil {
		t.Error("a green spell is not a legal target")
	}
}

func TestPreciseRedactionCountersBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	black := handCardFull(me, "Black Thing", "Sorcery", "", "", []string{"B", "U"})
	if err := g.CastSpell(me.ID, black, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	castCatalogSpell(t, g, "Precise Redaction", "Instant", rfPreciseRedactionOracle, rfCardTarget(black))
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(black) {
		t.Error("a black-and-blue spell is a legal target and should be countered")
	}
}

func TestMultiplyByZeroSetsBaseToZeroUnderCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plain := seedCreature(g, "Plain", me.ID)
	counted := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 2})

	castCatalogSpell(t, g, "Multiply by Zero", "Instant", rfMultiplyByZeroOracle, rfCardTarget(counted))
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, counted); p != 2 || tough != 2 {
		t.Errorf("a creature with two +1/+1 counters = %d/%d, want 2/2 over a 0/0 base", p, tough)
	}

	castCatalogSpell(t, g, "Multiply by Zero", "Instant", rfMultiplyByZeroOracle, rfCardTarget(plain))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plain) {
		t.Error("a creature with no counters and 0 toughness should die to state-based actions")
	}
}

func TestPerfectedTheoryBothModes(t *testing.T) {
	for _, tc := range []struct {
		mode       int
		power, tgh int
	}{{0, 1, 1}, {1, 4, 5}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		c := seedCreatureWithCounters(g, me.ID, nil)
		if err := rfCastModal(t, g, "Perfected Theory", rfPerfectedTheoryOracle, []int{tc.mode}, rfCardTarget(c)); err != nil {
			t.Fatalf("mode %d: %v", tc.mode, err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := effectivePower(t, g, c), effectiveToughness(t, g, c); p != tc.power || tough != tc.tgh {
			t.Errorf("mode %d = %d/%d, want %d/%d", tc.mode, p, tough, tc.power, tc.tgh)
		}
	}
}

func TestKonstrariCharmDamagesOnlyAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 6,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	ground := seedCreature(g, "Ground", opp.ID)

	if err := rfCastModal(t, g, "Konstrari Charm", rfKonstrariCharmOracle, []int{0}, rfCardTarget(ground)); err == nil {
		t.Error("a creature without flying is not a legal target of the damage mode")
	}
	if err := rfCastModal(t, g, "Konstrari Charm", rfKonstrariCharmOracle, []int{0}, rfCardTarget(flyer)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("6 damage should kill the 6-toughness flyer")
	}
}

func TestKonstrariCharmCountersAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[0].ID)
	if err := rfCastModal(t, g, "Konstrari Charm", rfKonstrariCharmOracle, []int{1}, rfCardTarget(c)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfCounters(t, g, c); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !hasAbility(effectiveAbilities(t, g, c), "trample") {
		t.Error("the creature should gain trample")
	}
}

func TestKonstrariCharmAddsThreeColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := rfCastModal(t, g, "Konstrari Charm", rfKonstrariCharmOracle, []int{2}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 3 {
		t.Errorf("mana pool = %v, want three colorless", got)
	}
}

func TestChargeTheSanctumModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := seedCreature(g, "A", me.ID)
	b := seedCreature(g, "B", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	if err := rfCastModal(t, g, "Charge the Sanctum", rfChargeTheSanctumOracle, []int{0}, nil); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, a) != 4 || effectivePower(t, g, b) != 4 {
		t.Errorf("my creatures should be 4/2: %d, %d", effectivePower(t, g, a), effectivePower(t, g, b))
	}
	if effectivePower(t, g, theirs) != 2 {
		t.Errorf("an opponent's creature must not be pumped: %d", effectivePower(t, g, theirs))
	}

	g2 := newCatalogGame(t)
	c := seedCreature(g2, "C", g2.Seats[0].ID)
	if err := rfCastModal(t, g2, "Charge the Sanctum", rfChargeTheSanctumOracle, []int{1}, rfCardTarget(c)); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if p, tough := rfPT(t, g2, c); p != 5 || tough != 3 {
		t.Errorf("mode 1 = %d/%d, want 5/3 (2/2 +2/+0 and a counter)", p, tough)
	}
	if !rfFirstStrike(t, g2, c) {
		t.Error("mode 1 grants first strike")
	}
}

func TestIcyReceptionMinusFive(t *testing.T) {
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[1].ID)
	if err := rfCastModal(t, g, "Icy Reception", rfIcyReceptionOracle, []int{1}, rfCardTarget(c)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, c); p != -3 {
		t.Errorf("effective power = %d, want -3", p)
	}
}

func TestIcyReceptionCountersACreatureSpellUnlessPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := handCardFull(me, "Bear Cub", "Creature — Bear", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	if err := rfCastModal(t, g, "Icy Reception", rfIcyReceptionOracle, []int{0}, rfCardTarget(bear)); err != nil {
		t.Fatalf("cast Icy Reception: %v", err)
	}
	passUntilTaxed(t, g, me.ID)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("declining the {3} should counter the creature spell")
	}
}

func TestIcyReceptionRefusesANonlegendaryNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spell := handCardFull(me, "Plain Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := rfCastModal(t, g, "Icy Reception", rfIcyReceptionOracle, []int{0}, rfCardTarget(spell)); err == nil {
		t.Error("a nonlegendary noncreature spell is not a legal target")
	}
}

func TestBestialIncursionMakesATramplingBeastAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Bestial Incursion", "Sorcery", rfBestialIncursionOracle, nil)
	passPriorityAroundTable(t, g)
	beast := findBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("no Beast token")
	}
	if effectivePower(t, g, beast) != 4 || effectiveToughness(t, g, beast) != 4 {
		t.Error("the Beast should be 4/4")
	}
	if !hasAbility(effectiveAbilities(t, g, beast), "trample") {
		t.Error("the Beast should have trample")
	}

	id := seedGraveyardCard(t, g, "Bestial Incursion", "Sorcery", rfBestialIncursionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Beast"); got != 2 {
		t.Errorf("Beasts = %d, want 2 after flashback", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card should be exiled")
	}
}

func TestGenerousRevivalReturnsASmallCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := pushGraveyardCreature(g, me.ID, "Small Bear", "{2}{G}")
	big := pushGraveyardCreature(g, me.ID, "Big Bear", "{3}{G}")

	if err := castCatalogSpellErr(t, g, "Generous Revival", "Sorcery", rfGenerousRevivalOracle, rfCardTarget(big)); err == nil {
		t.Error("a mana value 4 creature is not a legal target")
	}
	castCatalogSpell(t, g, "Generous Revival", "Sorcery", rfGenerousRevivalOracle, rfCardTarget(small))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the small creature should be on the battlefield")
	}
	if got := rfCounters(t, g, small); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestGenerousRevivalFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := pushGraveyardCreature(g, me.ID, "Small Bear", "{1}{G}")
	id := seedGraveyardCard(t, g, "Generous Revival", "Sorcery", rfGenerousRevivalOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: rfCardTarget(small),
	}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) || !g.Exile.Contains(id) {
		t.Error("flashback should return the creature and exile the spell")
	}
}

func TestGerminateRecruitsMakesACadetPerLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	castCatalogSpell(t, g, "Germinate Recruits", "Instant", rfGerminateRecruitsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Cadet"); got != 3 {
		t.Errorf("Cadets = %d, want 3", got)
	}
	cadet := findBattlefieldByName(g, "Cadet")
	if cadet == uuid.Nil {
		t.Fatal("no Cadet")
	}
	if effectivePower(t, g, cadet) != 2 || effectiveToughness(t, g, cadet) != 2 {
		t.Error("Cadet should be 2/2")
	}
}

func TestGerminateRecruitsWithNoLifeGainedMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Germinate Recruits", "Instant", rfGerminateRecruitsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, g.Seats[0].ID, "Cadet"); got != 0 {
		t.Errorf("Cadets = %d, want 0", got)
	}
}

func TestCommandTheStageGrowsOtherWizardTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wizard := func(owner uuid.UUID) uuid.UUID {
		c := TokenCard("2/2 colorless Wizard Soldier named Cadet")
		c.InstanceID, c.Owner, c.Controller = uuid.New(), owner, owner
		return pushBattlefieldCardWithTimestamp(g, c)
	}
	old := wizard(me.ID)
	theirs := wizard(opp.ID)
	bear := seedCreature(g, "Bear", me.ID)

	castCatalogSpell(t, g, "Command the Stage", "Sorcery", rfCommandTheStageOracle, nil)
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Cadet"); got != 2 {
		t.Fatalf("my Cadets = %d, want 2", got)
	}
	if rfCounters(t, g, old) != 1 {
		t.Errorf("the existing Wizard token should get a counter, has %d", rfCounters(t, g, old))
	}
	if rfCounters(t, g, theirs) != 0 || rfCounters(t, g, bear) != 0 {
		t.Error("an opponent's token and a non-token must not get a counter")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Cadet" && c.InstanceID != old && c.Counters[game.CounterPlusOne] != 0 {
			t.Error("the new Cadet is excluded from the counter")
		}
	}
}

func TestClashOfElementsOwnerChoosesTopWithDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	c := seedCreature(g, "Victim", opp.ID)
	life := opp.Life
	castCatalogSpell(t, g, "Clash of Elements", "Instant", rfClashOfElementsOracle, rfCardTarget(c))
	passPriorityAroundTable(t, g)
	ask := latestConfirmFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the permanent's owner is asked")
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c) {
		t.Error("the permanent should leave the battlefield")
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1]; top.InstanceID != c {
		t.Error("the permanent should be on top of its owner's library")
	}
	if opp.Life != life-2 {
		t.Errorf("owner life = %d, want %d", opp.Life, life-2)
	}
}

func TestClashOfElementsOwnerDeclinesToBottomNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	c := seedCreature(g, "Victim", opp.ID)
	life := opp.Life
	castCatalogSpell(t, g, "Clash of Elements", "Instant", rfClashOfElementsOracle, rfCardTarget(c))
	passPriorityAroundTable(t, g)
	ask := latestConfirmFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the permanent's owner is asked")
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := len(opp.Library.Cards)
	if n == 0 || opp.Library.Cards[0].InstanceID != c {
		t.Error("the permanent should be on the bottom of its owner's library")
	}
	if opp.Life != life {
		t.Errorf("owner life = %d, want unchanged %d", opp.Life, life)
	}
}

func TestClashOfElementsRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := castCatalogSpellErr(t, g, "Clash of Elements", "Instant", rfClashOfElementsOracle, rfCardTarget(land)); err == nil {
		t.Error("a land is not a legal target")
	}
}

func TestKindredJudgmentDestroysEverythingNotOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	myElf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", opp.ID)
	myBear := seedCreature(g, "My Bear", me.ID)
	shapeshifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordChangeling}, Owner: opp.ID, Controller: opp.ID,
	})

	castAndPause(t, g, "Kindred Judgment", "Sorcery", rfKindredJudgmentOracle)
	answerCreatureType(t, g, me.ID, "Elf")
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{elf, myElf, shapeshifter} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("an Elf (or a changeling) %s should survive", id)
		}
	}
	for _, id := range []uuid.UUID{bear, myBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a non-Elf %s should be destroyed", id)
		}
	}
}

func TestCruelCalculationsDrawsPerCardMilledFromTheTargetsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		_ = g.MillNForEffect(opp.ID, 3)
		_ = g.MillNForEffect(g.Seats[2].ID, 2)
	})
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Cruel Calculations", "Sorcery", rfCruelCalculationsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	// +1 for the spell pushed into hand by the harness, +3 drawn.
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand size = %d, want %d (3 cards milled from the target's library)", got, before+3)
	}
}

func TestCruelCalculationsWithNothingMilledDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Cruel Calculations", "Sorcery", rfCruelCalculationsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size = %d, want %d (nothing milled)", got, before)
	}
}
