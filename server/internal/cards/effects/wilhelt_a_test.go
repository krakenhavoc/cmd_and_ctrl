package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wilhelt_a_test.go — the Wilhelt, the Rotcleaver deck request
// (#3209), slice A: the cards existing seams carry, plus the Village
// Rites review.

const (
	wrUndeadWarchiefOracle = "e6af56bf-bd78-4801-8f6e-033cdd68de3d"
	wrWaterloggedOracle    = "e6ad1be9-f13d-4590-b3db-e2d0fff46f03"
	wrVillageRitesOracle   = "365548fb-5acc-4a8a-b20b-26d28b7d029f"
	wrTaintedIsleOracle    = "0222414f-98b5-458a-a0fd-831a66cd8b07"
	wrMurkySewerOracle     = "c2cdefeb-3176-4faf-be54-a62d31f777a5"
	wrVengefulDeadOracle   = "277c8ee9-0157-4e45-96ad-1b67716955ee"
	wrBeaconOfUnrestOracle = "aef874fa-93c0-4f86-8b0d-dc24ec51467d"
	wrSyphonFleshOracle    = "bc7c3151-ad07-429a-8af8-e4f24b7d9f6b"
	wrEndlessRanksOracle   = "69d4ecac-4735-4667-bfc1-c8800b436d08"
	wrLichLordOracle       = "6643fc8e-548e-48db-9f67-ffbca9e57016"
	wrGhoulcallerGisaOrcl  = "0dd894cb-1968-471c-af08-ea7ec5ce8428"
	wrUndercityInformerOrc = "ccf5a0d4-69e8-4607-a76c-cfca336899e4"
	wrZulAshurOracle       = "a7c6c9b3-91f2-4ecf-b2b0-b0edd96974d9"
	wrCleaverSkaabOracle   = "9fa66e03-7b20-43f6-8c53-1ce18169c382"
	wrWilheltOracle        = "eada96a1-4964-4ad0-b0e3-2df08d6ae188"
	wrCrowdedCryptOracle   = "25712827-626a-484d-bc52-f3bd6087857f"
	wrGisaAndGeralfOracle  = "20e94d0f-6887-45f8-a137-877011d786c6"
	wrRooftopStormOracle   = "6835bf6d-7197-481b-ac26-276ce363b4ad"
)

func TestWilheltSliceAIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		wrUndeadWarchiefOracle: "Undead Warchief",
		wrWaterloggedOracle:    "Waterlogged Teachings",
		wrVillageRitesOracle:   "Village Rites",
		wrTaintedIsleOracle:    "Tainted Isle",
		wrMurkySewerOracle:     "Murky Sewer",
		wrVengefulDeadOracle:   "Vengeful Dead",
		wrBeaconOfUnrestOracle: "Beacon of Unrest",
		wrSyphonFleshOracle:    "Syphon Flesh",
		wrEndlessRanksOracle:   "Endless Ranks of the Dead",
		wrLichLordOracle:       "Lich Lord of Unx",
		wrGhoulcallerGisaOrcl:  "Ghoulcaller Gisa",
		wrUndercityInformerOrc: "Undercity Informer",
		wrZulAshurOracle:       "Zul Ashur, Lich Lord",
		wrCleaverSkaabOracle:   "Cleaver Skaab",
		wrWilheltOracle:        "Wilhelt, the Rotcleaver",
		wrCrowdedCryptOracle:   "Crowded Crypt",
		wrGisaAndGeralfOracle:  "Gisa and Geralf",
		wrRooftopStormOracle:   "Rooftop Storm",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// wrZombie puts a vanilla Zombie creature onto the battlefield.
func wrZombie(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return pushTribalCreature(g, owner, name, "Creature — Zombie", p, tough)
}

// wrDestroy destroys a permanent through the engine, so its death
// fires triggers.
func wrDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func TestUndeadWarchiefDiscountsYourZombieSpellsAndPumpsYourZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	chief := pushCatalogPermanent(g, me.ID, "Undead Warchief", "Creature — Zombie", wrUndeadWarchiefOracle, false)
	zombie := wrZombie(g, me.ID, "Walker", 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := wrZombie(g, them.ID, "Their Walker", 2, 2)

	if got := priceInHand(t, g, me, "Gravecrawler", "Creature — Zombie", "{2}{B}"); got != 2 {
		t.Errorf("own Zombie spell: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own non-Zombie spell: %d, want 2 (undiscounted)", got)
	}
	if got := priceInHand(t, g, them, "Their Zombie", "Creature — Zombie", "{2}{B}"); got != 3 {
		t.Errorf("an opponent's Zombie spell: %d, want 3 (undiscounted)", got)
	}
	for id, want := range map[uuid.UUID][2]int{chief: {3, 2}, zombie: {4, 3}, bear: {2, 2}, theirs: {2, 2}} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%v: %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

func TestWaterloggedTeachingsFindsAnInstantOrACardWithFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := pushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant"})
	snapper := pushLibraryCardForTest(me, game.Card{Name: "Ambush Snapper", TypeLine: "Creature — Turtle", Keywords: []string{"flash"}})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})

	castCatalogSpell(t, g, "Waterlogged Teachings", "Instant", wrWaterloggedOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if !hasID(c.SearchCards, bolt) || !hasID(c.SearchCards, snapper) || hasID(c.SearchCards, bear) {
		t.Errorf("offered %v; want the instant and the flash creature, not the bear", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, snapper)
	if !me.Hand.Contains(snapper) {
		t.Error("the chosen card did not reach the hand")
	}
}

func TestMurkySewerEntersTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	g := newCatalogGame(t)
	if !b13Tapped(t, g, b12PlayFromHand(t, g, "Murky Sewer", "Land", wrMurkySewerOracle, game.CastSpellParams{})) {
		t.Error("everyone at 40: the land entered untapped")
	}
	g.Seats[2].Life = 13
	if b13Tapped(t, g, b12PlayFromHand(t, g, "Murky Sewer", "Land", wrMurkySewerOracle, game.CastSpellParams{})) {
		t.Error("a player at 13: the land entered tapped")
	}
}

func TestVengefulDeadDrainsOnAnyZombieAndOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	dead := pushCatalogPermanent(g, me.ID, "Vengeful Dead", "Creature — Zombie", wrVengefulDeadOracle, false)
	theirZombie := wrZombie(g, them.ID, "Their Walker", 2, 2)
	bear := pushVanillaCreature(g, them.ID, "Bear", 2, 2)
	before := them.Life

	wrDestroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if them.Life != before {
		t.Fatalf("a non-Zombie died: opponent lost %d", before-them.Life)
	}
	wrDestroy(t, g, theirZombie)
	passPriorityAroundTable(t, g)
	if them.Life != before-1 {
		t.Fatalf("an opponent's Zombie died: opponent lost %d, want 1", before-them.Life)
	}
	wrDestroy(t, g, dead)
	passPriorityAroundTable(t, g)
	if them.Life != before-2 {
		t.Errorf("Vengeful Dead itself died: opponent lost %d in all, want 2", before-them.Life)
	}
	if g.Seats[2].Life != 40-2 || me.Life != 40 {
		t.Errorf("each opponent loses, the controller does not: seat 2 %d, me %d", g.Seats[2].Life, me.Life)
	}
}

func TestBeaconOfUnrestTakesAnOpponentsArtifactAndShufflesItselfAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushGraveyardCardWithTypeLineID(them, "Mind Stone", "Artifact")

	beacon := castCatalogSpell(t, g, "Beacon of Unrest", "Sorcery", wrBeaconOfUnrestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, rock)
	if !ok {
		t.Fatal("the artifact did not return")
	}
	if c.Controller != me.ID {
		t.Error("the artifact is not under the caster's control")
	}
	if !me.Library.Contains(beacon) || me.Graveyard.Contains(beacon) {
		t.Error("Beacon was not shuffled into its owner's library")
	}
}

func TestSyphonFleshMakesAZombiePerCreatureSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	a, b := g.Seats[(active+1)%4], g.Seats[(active+2)%4]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	aBear := pushVanillaCreature(g, a.ID, "A Bear", 2, 2)
	bBear := pushVanillaCreature(g, b.ID, "B Bear", 2, 2)

	castCatalogSpell(t, g, "Syphon Flesh", "Sorcery", wrSyphonFleshOracle, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to sacrifice")
	}
	answerSacrifice(t, g, a.ID, aBear)
	answerSacrifice(t, g, b.ID, bBear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's creature was sacrificed")
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("%d Zombies, want 2 (one per creature sacrificed)", got)
	}
}

func TestEndlessRanksMakesHalfYourZombiesRoundedDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Endless Ranks of the Dead", "Enchantment", wrEndlessRanksOracle, false)
	for _, n := range []string{"Z1", "Z2", "Z3"} {
		wrZombie(g, me.ID, n, 2, 2)
	}
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("three Zombies: %d tokens, want 1", got)
	}
}

func TestLichLordOfUnxMakesWizardsAndDrainsByZombieCount(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lord := pushCatalogPermanent(g, me.ID, "Lich Lord of Unx", "Creature — Zombie Wizard", wrLichLordOracle, false)
	wrZombie(g, me.ID, "Walker", 2, 2)

	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "B", 1)
	b16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	tok, ok := battlefieldCard(g, findBattlefieldByName(g, "Zombie Wizard"))
	if !ok || !tok.HasColor("U") || !tok.HasColor("B") {
		t.Fatal("no blue and black Zombie Wizard token")
	}

	// Three Zombies now: the Lord, the Walker and the token.
	before, library := them.Life, them.Library.Size()
	fillPoolColored(me, "U", 2)
	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, lord, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != before-3 || them.Library.Size() != library-3 {
		t.Errorf("lost %d life and milled %d, want 3 and 3", before-them.Life, library-them.Library.Size())
	}
}

func TestGhoulcallerGisaMakesZombiesEqualToTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gisa := pushCatalogPermanent(g, me.ID, "Ghoulcaller Gisa", "Legendary Creature — Human Wizard", wrGhoulcallerGisaOrcl, false)
	// A 4/4 with two +1/+1 counters: power 6 as it last existed.
	big := pushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)
	want := 6

	fillPoolColored(me, "B", 1)
	b16Activate(t, g, me.ID, gisa, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != want {
		t.Errorf("%d Zombies, want %d (the sacrificed creature's power)", got, want)
	}
}

func TestUndercityInformerMillsATargetPlayerToALand(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	informer := pushCatalogPermanent(g, me.ID, "Undercity Informer", "Creature — Human Rogue", wrUndercityInformerOrc, false)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	land := uuid.New()
	them.Library.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: them.ID, Controller: them.ID})
	spell := uuid.New()
	them.Library.PushTop(game.Card{InstanceID: spell, Name: "Ponder", TypeLine: "Sorcery", Owner: them.ID, Controller: them.ID})

	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, informer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !them.Graveyard.Contains(spell) || !them.Graveyard.Contains(land) {
		t.Error("the revealed run, the land included, was not milled")
	}
}

func TestZulAshurLetsYouCastAZombieFromYourGraveyardThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	zul := pushCatalogPermanent(g, me.ID, "Zul Ashur, Lich Lord", "Legendary Creature — Zombie Warlock", wrZulAshurOracle, false)
	zombie := ptGraveCard(me, "Dead Walker", "Creature — Zombie")
	bear := ptGraveCard(me, "Dead Bear", "Creature — Bear")

	if err := g.ActivateCatalogAbility(me.ID, zul, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a non-Zombie creature card was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, zul, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombie}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Zombie from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(zombie) {
		t.Error("the Zombie did not resolve onto the battlefield")
	}
}

func TestCleaverSkaabCopiesTheSacrificedZombieTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skaab := pushCatalogPermanent(g, me.ID, "Cleaver Skaab", "Creature — Zombie Horror", wrCleaverSkaabOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 3, 1)

	fillPoolColored(me, "C", 3)
	b16Activate(t, g, me.ID, skaab, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{walker}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the Zombie was not sacrificed")
	}
	if got := b29CountNamed(g, me.ID, "Walker"); got != 2 {
		t.Fatalf("%d copies named Walker, want 2", got)
	}
	if id := findBattlefieldByName(g, "Walker"); effectivePower(t, g, id) != 3 {
		t.Error("the copy is not a 3/1")
	}
}

func TestWilheltMakesADecayedZombieOnlyForAZombieWithoutDecayed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Wilhelt, the Rotcleaver", "Legendary Creature — Zombie Warrior", wrWilheltOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 2, 2)

	wrDestroy(t, g, walker)
	passPriorityAroundTable(t, g)
	token := findBattlefieldByName(g, "Zombie")
	if token == uuid.Nil {
		t.Fatal("no decayed Zombie for a Zombie without decayed")
	}
	if !hasEffectiveKeyword(t, g, token, game.KeywordDecayed) {
		t.Error("the token has no decayed")
	}
	wrDestroy(t, g, token)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a decayed Zombie died: %d new tokens, want 0", got)
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wrDestroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a non-Zombie died: %d tokens, want 0", got)
	}
}

func TestWilheltSacrificesAZombieAtYourEndStepToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wilhelt := pushCatalogPermanent(g, me.ID, "Wilhelt, the Rotcleaver", "Legendary Creature — Zombie Warrior", wrWilheltOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 {
		t.Fatal("the sacrifice is asked as the trigger resolves, with a floor of zero")
	}
	if hasID(p.ChooseCards, bear) || !hasID(p.ChooseCards, wilhelt) || !hasID(p.ChooseCards, walker) {
		t.Error("the Zombies, Wilhelt included, are offered and the Bear is not")
	}
	answerOwnPermanents(t, g, me.ID, walker)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the chosen Zombie was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestCrowdedCryptCountsDeathsThenCashesThemIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	crypt := pushCatalogPermanent(g, me.ID, "Crowded Crypt", "Artifact", wrCrowdedCryptOracle, false)
	for _, n := range []string{"A", "B", "C"} {
		wrDestroy(t, g, pushVanillaCreature(g, me.ID, n, 1, 1))
		passPriorityAroundTable(t, g)
	}
	wrDestroy(t, g, pushVanillaCreature(g, g.Seats[1].ID, "Theirs", 1, 1))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, crypt, "corpse"); got != 3 {
		t.Fatalf("%d corpse counters, want 3 (an opponent's creature adds none)", got)
	}
	fillPoolColored(me, "C", 4)
	fillPoolColored(me, "B", 2)
	b16Activate(t, g, me.ID, crypt, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(crypt) {
		t.Error("the Crypt was not sacrificed")
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 3 {
		t.Errorf("%d Zombies, want 3", got)
	}
}

func TestGisaAndGeralfMillsFourAndCastsOneZombieAPerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	library := me.Library.Size()
	castCatalogSpell(t, g, "Gisa and Geralf", "Legendary Creature — Human Wizard", wrGisaAndGeralfOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-4 {
		t.Fatalf("milled %d, want 4", library-me.Library.Size())
	}

	first := ptGraveCard(me, "First Walker", "Creature — Zombie")
	second := ptGraveCard(me, "Second Walker", "Creature — Zombie")
	bear := ptGraveCard(me, "Dead Bear", "Creature — Bear")
	if err := muldrothaCast(g, me, bear, ""); err == nil {
		t.Error("a non-Zombie creature was cast from the graveyard")
	}
	if err := muldrothaCast(g, me, first, ""); err != nil {
		t.Fatalf("the first Zombie from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := muldrothaCast(g, me, second, ""); err == nil {
		t.Error("a second Zombie was cast from the graveyard in the same turn")
	}
}

func TestRooftopStormOffersAFreeCastForZombieCreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Rooftop Storm", "Enchantment", wrRooftopStormOracle, false)
	zombie := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: zombie, Name: "Big Zombie", TypeLine: "Creature — Zombie", ManaCost: "{6}{B}", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})

	if offerWithKey(grantedCardOffers(g, me.ID, bear, game.ZoneHand), GrantedAltCostZeroZombieCreatures) != nil {
		t.Error("a non-Zombie creature spell is offered the free cast")
	}
	if offerWithKey(grantedCardOffers(g, me.ID, zombie, game.ZoneHand), GrantedAltCostZeroZombieCreatures) == nil {
		t.Fatal("no free-cast offer for a Zombie creature spell")
	}
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostZeroZombieCreatures}); err != nil {
		t.Fatalf("cast for {0}: %v", err)
	}
	if me.Hand.Contains(zombie) {
		t.Error("the Zombie is still in hand")
	}
}
