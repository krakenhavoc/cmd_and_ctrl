package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestWilheltSliceAIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.WrUndeadWarchiefOracle: "Undead Warchief",
		effectstest.WrWaterloggedOracle:    "Waterlogged Teachings",
		effectstest.WrVillageRitesOracle:   "Village Rites",
		effectstest.WrTaintedIsleOracle:    "Tainted Isle",
		effectstest.WrMurkySewerOracle:     "Murky Sewer",
		effectstest.WrVengefulDeadOracle:   "Vengeful Dead",
		effectstest.WrBeaconOfUnrestOracle: "Beacon of Unrest",
		effectstest.WrSyphonFleshOracle:    "Syphon Flesh",
		effectstest.WrEndlessRanksOracle:   "Endless Ranks of the Dead",
		effectstest.WrLichLordOracle:       "Lich Lord of Unx",
		effectstest.WrGhoulcallerGisaOrcl:  "Ghoulcaller Gisa",
		effectstest.WrUndercityInformerOrc: "Undercity Informer",
		effectstest.WrZulAshurOracle:       "Zul Ashur, Lich Lord",
		effectstest.WrCleaverSkaabOracle:   "Cleaver Skaab",
		effectstest.WrWilheltOracle:        "Wilhelt, the Rotcleaver",
		effectstest.WrCrowdedCryptOracle:   "Crowded Crypt",
		effectstest.WrGisaAndGeralfOracle:  "Gisa and Geralf",
		effectstest.WrRooftopStormOracle:   "Rooftop Storm",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// wrZombie puts a vanilla Zombie creature onto the battlefield.
func wrZombie(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.PushTribalCreature(g, owner, name, "Creature — Zombie", p, tough)
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	chief := effectstest.PushCatalogPermanent(g, me.ID, "Undead Warchief", "Creature — Zombie", effectstest.WrUndeadWarchiefOracle, false)
	zombie := wrZombie(g, me.ID, "Walker", 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := wrZombie(g, them.ID, "Their Walker", 2, 2)

	if got := effectstest.PriceInHand(t, g, me, "Gravecrawler", "Creature — Zombie", "{2}{B}"); got != 2 {
		t.Errorf("own Zombie spell: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own non-Zombie spell: %d, want 2 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Zombie", "Creature — Zombie", "{2}{B}"); got != 3 {
		t.Errorf("an opponent's Zombie spell: %d, want 3 (undiscounted)", got)
	}
	for id, want := range map[uuid.UUID][2]int{chief: {3, 2}, zombie: {4, 3}, bear: {2, 2}, theirs: {2, 2}} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%v: %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

func TestWaterloggedTeachingsFindsAnInstantOrACardWithFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant"})
	snapper := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ambush Snapper", TypeLine: "Creature — Turtle", Keywords: []string{"flash"}})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})

	effectstest.CastCatalogSpell(t, g, "Waterlogged Teachings", "Instant", effectstest.WrWaterloggedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if !effectstest.HasID(c.SearchCards, bolt) || !effectstest.HasID(c.SearchCards, snapper) || effectstest.HasID(c.SearchCards, bear) {
		t.Errorf("offered %v; want the instant and the flash creature, not the bear", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, snapper)
	if !me.Hand.Contains(snapper) {
		t.Error("the chosen card did not reach the hand")
	}
}

func TestMurkySewerEntersTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if !effectstest.B13Tapped(t, g, effectstest.B12PlayFromHand(t, g, "Murky Sewer", "Land", effectstest.WrMurkySewerOracle, game.CastSpellParams{})) {
		t.Error("everyone at 40: the land entered untapped")
	}
	g.Seats[2].Life = 13
	if effectstest.B13Tapped(t, g, effectstest.B12PlayFromHand(t, g, "Murky Sewer", "Land", effectstest.WrMurkySewerOracle, game.CastSpellParams{})) {
		t.Error("a player at 13: the land entered tapped")
	}
}

func TestVengefulDeadDrainsOnAnyZombieAndOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	dead := effectstest.PushCatalogPermanent(g, me.ID, "Vengeful Dead", "Creature — Zombie", effectstest.WrVengefulDeadOracle, false)
	theirZombie := wrZombie(g, them.ID, "Their Walker", 2, 2)
	bear := effectstest.PushVanillaCreature(g, them.ID, "Bear", 2, 2)
	before := them.Life

	wrDestroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before {
		t.Fatalf("a non-Zombie died: opponent lost %d", before-them.Life)
	}
	wrDestroy(t, g, theirZombie)
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-1 {
		t.Fatalf("an opponent's Zombie died: opponent lost %d, want 1", before-them.Life)
	}
	wrDestroy(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-2 {
		t.Errorf("Vengeful Dead itself died: opponent lost %d in all, want 2", before-them.Life)
	}
	if g.Seats[2].Life != 40-2 || me.Life != 40 {
		t.Errorf("each opponent loses, the controller does not: seat 2 %d, me %d", g.Seats[2].Life, me.Life)
	}
}

func TestBeaconOfUnrestTakesAnOpponentsArtifactAndShufflesItselfAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.PushGraveyardCardWithTypeLineID(them, "Mind Stone", "Artifact")

	beacon := effectstest.CastCatalogSpell(t, g, "Beacon of Unrest", "Sorcery", effectstest.WrBeaconOfUnrestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, rock)
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
	g := effectstest.NewCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	a, b := g.Seats[(active+1)%4], g.Seats[(active+2)%4]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	aBear := effectstest.PushVanillaCreature(g, a.ID, "A Bear", 2, 2)
	bBear := effectstest.PushVanillaCreature(g, b.ID, "B Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Syphon Flesh", "Sorcery", effectstest.WrSyphonFleshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, a.ID, aBear)
	effectstest.AnswerSacrifice(t, g, b.ID, bBear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's creature was sacrificed")
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("%d Zombies, want 2 (one per creature sacrificed)", got)
	}
}

func TestEndlessRanksMakesHalfYourZombiesRoundedDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Endless Ranks of the Dead", "Enchantment", effectstest.WrEndlessRanksOracle, false)
	for _, n := range []string{"Z1", "Z2", "Z3"} {
		wrZombie(g, me.ID, n, 2, 2)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("three Zombies: %d tokens, want 1", got)
	}
}

func TestLichLordOfUnxMakesWizardsAndDrainsByZombieCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lord := effectstest.PushCatalogPermanent(g, me.ID, "Lich Lord of Unx", "Creature — Zombie Wizard", effectstest.WrLichLordOracle, false)
	wrZombie(g, me.ID, "Walker", 2, 2)

	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.B16Activate(t, g, me.ID, lord, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	tok, ok := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Zombie Wizard"))
	if !ok || !tok.HasColor("U") || !tok.HasColor("B") {
		t.Fatal("no blue and black Zombie Wizard token")
	}

	// Three Zombies now: the Lord, the Walker and the token.
	before, library := them.Life, them.Library.Size()
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, lord, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-3 || them.Library.Size() != library-3 {
		t.Errorf("lost %d life and milled %d, want 3 and 3", before-them.Life, library-them.Library.Size())
	}
}

func TestGhoulcallerGisaMakesZombiesEqualToTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gisa := effectstest.PushCatalogPermanent(g, me.ID, "Ghoulcaller Gisa", "Legendary Creature — Human Wizard", effectstest.WrGhoulcallerGisaOrcl, false)
	// A 4/4 with two +1/+1 counters: power 6 as it last existed.
	big := effectstest.PushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)
	want := 6

	effectstest.FillPoolColored(me, "B", 1)
	effectstest.B16Activate(t, g, me.ID, gisa, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != want {
		t.Errorf("%d Zombies, want %d (the sacrificed creature's power)", got, want)
	}
}

func TestUndercityInformerMillsATargetPlayerToALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	informer := effectstest.PushCatalogPermanent(g, me.ID, "Undercity Informer", "Creature — Human Rogue", effectstest.WrUndercityInformerOrc, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	land := uuid.New()
	them.Library.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: them.ID, Controller: them.ID})
	spell := uuid.New()
	them.Library.PushTop(game.Card{InstanceID: spell, Name: "Ponder", TypeLine: "Sorcery", Owner: them.ID, Controller: them.ID})

	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, informer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !them.Graveyard.Contains(spell) || !them.Graveyard.Contains(land) {
		t.Error("the revealed run, the land included, was not milled")
	}
}

func TestZulAshurLetsYouCastAZombieFromYourGraveyardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zul := effectstest.PushCatalogPermanent(g, me.ID, "Zul Ashur, Lich Lord", "Legendary Creature — Zombie Warlock", effectstest.WrZulAshurOracle, false)
	zombie := effectstest.PtGraveCard(me, "Dead Walker", "Creature — Zombie")
	bear := effectstest.PtGraveCard(me, "Dead Bear", "Creature — Bear")

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
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Zombie from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(zombie) {
		t.Error("the Zombie did not resolve onto the battlefield")
	}
}

func TestCleaverSkaabCopiesTheSacrificedZombieTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	skaab := effectstest.PushCatalogPermanent(g, me.ID, "Cleaver Skaab", "Creature — Zombie Horror", effectstest.WrCleaverSkaabOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 3, 1)

	effectstest.FillPoolColored(me, "C", 3)
	effectstest.B16Activate(t, g, me.ID, skaab, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{walker}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the Zombie was not sacrificed")
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Walker"); got != 2 {
		t.Fatalf("%d copies named Walker, want 2", got)
	}
	if id := effectstest.FindBattlefieldByName(g, "Walker"); effectstest.EffectivePower(t, g, id) != 3 {
		t.Error("the copy is not a 3/1")
	}
}

func TestWilheltMakesADecayedZombieOnlyForAZombieWithoutDecayed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Wilhelt, the Rotcleaver", "Legendary Creature — Zombie Warrior", effectstest.WrWilheltOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 2, 2)

	wrDestroy(t, g, walker)
	effectstest.PassPriorityAroundTable(t, g)
	token := effectstest.FindBattlefieldByName(g, "Zombie")
	if token == uuid.Nil {
		t.Fatal("no decayed Zombie for a Zombie without decayed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, game.KeywordDecayed) {
		t.Error("the token has no decayed")
	}
	wrDestroy(t, g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a decayed Zombie died: %d new tokens, want 0", got)
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wrDestroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a non-Zombie died: %d tokens, want 0", got)
	}
}

func TestWilheltSacrificesAZombieAtYourEndStepToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wilhelt := effectstest.PushCatalogPermanent(g, me.ID, "Wilhelt, the Rotcleaver", "Legendary Creature — Zombie Warrior", effectstest.WrWilheltOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 {
		t.Fatal("the sacrifice is asked as the trigger resolves, with a floor of zero")
	}
	if effectstest.HasID(p.ChooseCards, bear) || !effectstest.HasID(p.ChooseCards, wilhelt) || !effectstest.HasID(p.ChooseCards, walker) {
		t.Error("the Zombies, Wilhelt included, are offered and the Bear is not")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the chosen Zombie was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestCrowdedCryptCountsDeathsThenCashesThemIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	crypt := effectstest.PushCatalogPermanent(g, me.ID, "Crowded Crypt", "Artifact", effectstest.WrCrowdedCryptOracle, false)
	for _, n := range []string{"A", "B", "C"} {
		wrDestroy(t, g, effectstest.PushVanillaCreature(g, me.ID, n, 1, 1))
		effectstest.PassPriorityAroundTable(t, g)
	}
	wrDestroy(t, g, effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Theirs", 1, 1))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, crypt, "corpse"); got != 3 {
		t.Fatalf("%d corpse counters, want 3 (an opponent's creature adds none)", got)
	}
	effectstest.FillPoolColored(me, "C", 4)
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.B16Activate(t, g, me.ID, crypt, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(crypt) {
		t.Error("the Crypt was not sacrificed")
	}
	if got := effectstest.B29CountNamed(g, me.ID, "Zombie"); got != 3 {
		t.Errorf("%d Zombies, want 3", got)
	}
}

func TestGisaAndGeralfMillsFourAndCastsOneZombieAPerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	library := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Gisa and Geralf", "Legendary Creature — Human Wizard", effectstest.WrGisaAndGeralfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library-4 {
		t.Fatalf("milled %d, want 4", library-me.Library.Size())
	}

	first := effectstest.PtGraveCard(me, "First Walker", "Creature — Zombie")
	second := effectstest.PtGraveCard(me, "Second Walker", "Creature — Zombie")
	bear := effectstest.PtGraveCard(me, "Dead Bear", "Creature — Bear")
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err == nil {
		t.Error("a non-Zombie creature was cast from the graveyard")
	}
	if err := effectstest.MuldrothaCast(g, me, first, ""); err != nil {
		t.Fatalf("the first Zombie from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.MuldrothaCast(g, me, second, ""); err == nil {
		t.Error("a second Zombie was cast from the graveyard in the same turn")
	}
}

func TestRooftopStormOffersAFreeCastForZombieCreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rooftop Storm", "Enchantment", effectstest.WrRooftopStormOracle, false)
	zombie := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: zombie, Name: "Big Zombie", TypeLine: "Creature — Zombie", ManaCost: "{6}{B}", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})

	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, bear, game.ZoneHand), effects.GrantedAltCostZeroZombieCreatures) != nil {
		t.Error("a non-Zombie creature spell is offered the free cast")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, zombie, game.ZoneHand), effects.GrantedAltCostZeroZombieCreatures) == nil {
		t.Fatal("no free-cast offer for a Zombie creature spell")
	}
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostZeroZombieCreatures}); err != nil {
		t.Fatalf("cast for {0}: %v", err)
	}
	if me.Hand.Contains(zombie) {
		t.Error("the Zombie is still in hand")
	}
}
