package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b7_cards_test.go — Final Fantasy (FIN) batch 7 (#4462). Each test
// plays the real catalog entry and asserts what its text says.

const (
	finB7RagnarokOracle   = "5c9f1d2e-0b7a-4e44-9d3a-1f0c2b6a7e11"
	finB7RelentlessOracle = "50994df7-300d-4e30-897a-0974dc0820ee"
	finB7InsomniaOracle   = "b7140f72-91b6-4be4-8d71-c019e64c2351"
	finB7ZodiarkOracle    = "f1190fb9-72a2-48f8-bc2d-61240f1e9119"
	finB7TellahOracle     = "1a080b26-c766-4f7c-9198-4974d2133f75"
	finB7FenrirOracle     = "3509171e-8d7d-4e9e-97a1-58e6c7d225ed"
	finB7TripleOracle     = "9567d963-4815-466b-a319-6d138b6a0cd2"
	finB7VanilleOracle    = "577725ae-290f-4081-8352-3317521a096f"
	finB7UniformOracle    = "17c0cf45-a031-40e8-a37f-eda98a8f785b"
)

func finB7RegisteredFull(t *testing.T, oracle, name string) Spec {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
	return spec
}

// --- Insomnia, Crown City -----------------------------------------------------------

func TestFinB7InsomniaEntersTappedAndTapsForWOrB(t *testing.T) {
	g := newCatalogGame(t)
	land := playLandFromHand(t, g, "Insomnia, Crown City", finB7InsomniaOracle)
	if c := findBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("Insomnia should enter tapped")
	}
	spec := finB7RegisteredFull(t, finB7InsomniaOracle, "Insomnia, Crown City")
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|B}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {W|B}", spec.ManaAbilities)
	}
}

// --- Relentless X-ATM092 ------------------------------------------------------------

func TestFinB7RelentlessCantBeBlockedExceptByThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	finB7RegisteredFull(t, finB7RelentlessOracle, "Relentless X-ATM092")
	robot := b12Push(g, me.ID, "Relentless X-ATM092", "Artifact Creature — Robot Spider", finB7RelentlessOracle, 6, 5)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	brAttack(t, g, robot)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: robot}, {Blocker: b, Attacker: robot},
	}), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: robot}, {Blocker: b, Attacker: robot}, {Blocker: c, Attacker: robot},
	}); err != nil {
		t.Fatalf("three blockers are legal: %v", err)
	}
}

func TestFinB7RelentlessReturnsTappedWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	robot := pushCatalogGraveyardCard(me, "Relentless X-ATM092", "Artifact Creature — Robot Spider", finB7RelentlessOracle, 6, 5)
	fillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, robot, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only seven mana")
	}
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, robot, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, robot)
	if !ok {
		t.Fatal("the robot did not return")
	}
	if !c.Tapped {
		t.Error("the robot should return tapped")
	}
	if got := countersOn(g, robot, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
}

// --- Zodiark, Umbral God ------------------------------------------------------------

func TestFinB7ZodiarkEdictsHalfTheNonGodCreaturesAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	var mine, theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		mine = append(mine, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	god := b12Creature(g, me.ID, "Stone God", "Legendary Creature — God", 2, 2)
	for i := 0; i < 3; i++ {
		theirs = append(theirs, b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2))
	}
	lone := b12Creature(g, third.ID, "Lone", "Creature — Bear", 1, 1)
	_ = fourth
	zod := castCatalogSpell(t, g, "Zodiark, Umbral God", "Legendary Creature — God", finB7ZodiarkOracle, nil)
	finB7RegisteredFull(t, finB7ZodiarkOracle, "Zodiark, Umbral God")
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, mine[:2]...)
	answerSacrifice(t, g, opp.ID, theirs[0])
	if sacrificeChoiceFor(g, third.ID) != nil {
		t.Error("a lone creature owes no sacrifice (half of one, rounded down)")
	}
	passPriorityAroundTable(t, g)
	for _, id := range mine[:2] {
		if g.Battlefield.Contains(id) {
			t.Error("a chosen creature survived")
		}
	}
	for _, id := range append(append([]uuid.UUID{god, lone, zod}, mine[2:]...), theirs[1:]...) {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed but was not chosen / is a God", id)
		}
	}
	if got := countersOn(g, zod, game.CounterPlusOne); got != 3 {
		t.Errorf("Zodiark has %d +1/+1 counters, want 3 (one per sacrificed creature)", got)
	}
	if !hasEffectiveKeyword(t, g, zod, "indestructible") {
		t.Error("Zodiark is not indestructible")
	}
}

// --- Tellah, Great Sage -------------------------------------------------------------

func finB7Heroes(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == "Hero" && c.IsToken() {
			n++
		}
	}
	return n
}

func TestFinB7TellahMakesAHeroAndScalesWithManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB7RegisteredFull(t, finB7TellahOracle, "Tellah, Great Sage")
	tellah := socB17Perm(g, me.ID, "Tellah, Great Sage", "Legendary Creature — Human Wizard", finB7TellahOracle, "{3}{U}{R}", 3, 3)
	hand := me.Hand.Size()
	socB17CastSpellFor(t, g, "Instant", "{1}{U}", "UU")
	if n := finB7Heroes(g, me.ID); n != 1 {
		t.Errorf("%d Hero tokens after a two-mana spell, want 1", n)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after a two-mana spell, want no draw", hand, me.Hand.Size())
	}
	socB17CastSpellFor(t, g, "Creature — Bear", "{4}{U}", "UUUUU")
	if n := finB7Heroes(g, me.ID); n != 1 {
		t.Errorf("a creature spell made a Hero (%d)", n)
	}
	socB17CastSpellFor(t, g, "Sorcery", "{3}{U}", "UUUU")
	if n := finB7Heroes(g, me.ID); n != 2 {
		t.Errorf("%d Hero tokens after a four-mana spell, want 2", n)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d after a four-mana spell, want two drawn", hand, me.Hand.Size())
	}
	if !g.Battlefield.Contains(tellah) {
		t.Error("Tellah was sacrificed by a four-mana spell")
	}
}

func TestFinB7TellahSacrificesItselfAndBurnsEachOpponentForEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tellah := socB17Perm(g, me.ID, "Tellah, Great Sage", "Legendary Creature — Human Wizard", finB7TellahOracle, "{3}{U}{R}", 3, 3)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	socB17CastSpellFor(t, g, "Sorcery", "{6}{U}{U}", "UUUUUUUU")
	if g.Battlefield.Contains(tellah) {
		t.Error("Tellah survived an eight-mana spell")
	}
	for i, p := range g.Seats {
		want := lives[i]
		if i != g.Turn.ActiveSeat {
			want -= 8
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	if n := finB7Heroes(g, me.ID); n != 1 {
		t.Errorf("%d Hero tokens, want 1", n)
	}
}

// --- Summon: Fenrir -----------------------------------------------------------------

func TestFinB7FenrirFetchesATappedBasicThenBoostsTheNextCreatureThenMaybeDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB7RegisteredFull(t, finB7FenrirOracle, "Summon: Fenrir")
	forest := topLibraryCard(me, "Forest", "Basic Land — Forest")
	topLibraryCard(me, "Filler", "Sorcery")
	saga := finS3CastSaga(t, g, "Summon: Fenrir", "Enchantment Creature — Saga Wolf", finB7FenrirOracle, "{2}{G}", 3, 2)
	passPriorityAroundTable(t, g) // the creature, then chapter I
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Fatalf("the basic land is on the battlefield %v, tapped %v; want it tapped", ok, c.Tapped)
	}

	// Chapter II: a noncreature spell leaves the promise, the next creature spell uses it.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	first := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	second := castCatalogSpell(t, g, "Second Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, first, game.CounterPlusOne); got != 1 {
		t.Errorf("the first creature has %d +1/+1 counters, want 1", got)
	}
	if got := countersOn(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("the second creature has %d +1/+1 counters, want 0 (one use)", got)
	}

	// Chapter III: the Saga's 3 power ties or beats everything I control, but a 9/9
	// elsewhere is the greatest power, so no card.
	big := b12Creature(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Behemoth", "Creature — Beast", 9, 9)
	advanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	_ = saga
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d with an opponent holding the greatest power, want no draw", hand, me.Hand.Size())
	}
	_ = big
}

func TestFinB7FenrirChapterThreeDrawsWhenTiedForTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	b12Creature(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Rival", "Creature — Beast", 3, 3)
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Summon: Fenrir", TypeLine: "Enchantment Creature — Saga Wolf",
		OracleID: finB7FenrirOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 2},
	})
	advanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a draw for tying the greatest power", hand, me.Hand.Size())
	}
	_ = saga
}

// --- Triple Triad -------------------------------------------------------------------

func TestFinB7TripleTriadLetsYouPlayYourCardAndEveryLesserOneFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finB7RegisteredFull(t, finB7TripleOracle, "Triple Triad")
	s58p6Push(g, me.ID, "Triple Triad", "Enchantment", finB7TripleOracle, "{3}{R}{R}{R}", 0, 0)
	b12ToMyNextUpkeep(t, g)
	mine := seedLibraryTop(me, "Mine", "Creature — Bear", "{3}")
	// Seeded once the table is at my upkeep: the others' draws are behind us.
	lesser := seedLibraryTop(g.Seats[1], "Lesser", "Instant", "{2}")
	equal := seedLibraryTop(g.Seats[2], "Equal", "Sorcery", "{3}")
	land := seedLibraryTop(g.Seats[3], "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	for name, w := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"mine": {mine, true}, "lesser": {lesser, true}, "equal": {equal, false}, "land": {land, true}} {
		id, want := w.id, w.want
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", name)
			continue
		}
		perm := g.CastPermissionOnCardByIDForEffect(id)
		if got := perm != nil && perm.Player == me.ID; got != want {
			t.Errorf("card %s playable by me = %v, want %v", id, got, want)
		}
	}
}

func TestFinB7TripleTriadGrantsNothingElseWhenYourOwnCardCostsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s58p6Push(g, me.ID, "Triple Triad", "Enchantment", finB7TripleOracle, "{3}{R}{R}{R}", 0, 0)
	mine := seedLibraryTop(me, "Mine", "Land", "")
	other := seedLibraryTop(g.Seats[1], "Other", "Creature — Bear", "{1}")
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if perm := g.CastPermissionOnCardByIDForEffect(mine); perm == nil {
		t.Error("my own card should be playable")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(other); perm != nil {
		t.Error("a card with greater mana value than mine was made playable")
	}
}

// --- Vanille, Cheerful l'Cie --------------------------------------------------------

func TestFinB7VanilleMillsTwoThenReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB7RegisteredFull(t, finB7VanilleOracle, "Vanille, Cheerful l'Cie")
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	bear := pushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	pushCatalogGraveyardCard(me, "Old Spell", "Instant", "", 0, 0)
	library := me.Library.Size()
	castCatalogSpell(t, g, "Vanille, Cheerful l'Cie", "Legendary Creature — Human Cleric", finB7VanilleOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-2 {
		t.Errorf("library %d → %d, want two milled", library, me.Library.Size())
	}
	if !me.Hand.Contains(bear) {
		t.Error("the only permanent card in the graveyard was not returned to hand")
	}
}

func finB7FangMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   finB7RagnarokOracle,
		ResultScryfallID: uuid.NewString(),
		Result:           game.Face{Name: "Ragnarok, Divine Deliverance", TypeLine: "Legendary Creature — Beast Avatar", Power: 7, Toughness: 6},
	}
}

func TestFinB7VanilleMeldsWithFangForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	print := finB7FangMeldPrint()
	vanille := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanille, Cheerful l'Cie", OracleID: finB7VanilleOracle,
		TypeLine: "Legendary Creature — Human Cleric", ManaCost: "{3}{G}", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: print, ScryfallID: uuid.NewString(),
	})
	fang := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fang, Fearless l'Cie", TypeLine: "Legendary Creature — Human Warrior",
		ManaCost: "{3}{B}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
		Layout: game.LayoutMeld, Meld: print, ScryfallID: uuid.NewString(),
	})
	finB7ToMyNextMain(t, g, me)
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "BGCCC")
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the meld ability asked for no payment")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vanille) || g.Battlefield.Contains(fang) {
		t.Error("the two cards were not exiled")
	}
	for _, c := range g.Battlefield.Cards {
		t.Logf("battlefield: %q oracle %s", c.Name, c.OracleID)
	}
	for _, c := range g.Exile.Cards {
		t.Logf("exile: %q oracle %s", c.Name, c.OracleID)
	}
	melded := 0
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == finB7RagnarokOracle {
			melded++
		}
	}
	if melded != 1 {
		t.Errorf("%d Ragnarok on the battlefield, want 1", melded)
	}
}

func TestFinB7VanilleMeldAbilityIsOnlyOfferedWithFang(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanille, Cheerful l'Cie", OracleID: finB7VanilleOracle,
		TypeLine: "Legendary Creature — Human Cleric", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: finB7FangMeldPrint(), ScryfallID: uuid.NewString(),
	})
	finB7ToMyNextMain(t, g, me)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the meld ability was offered with no Fang")
	}
}

// finB7ToMyNextMain walks the table to `me`'s next precombat main and lets the
// beginning-of-main triggers go on the stack.
func finB7ToMyNextMain(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i, p := range g.Seats {
		if p.ID == me.ID {
			advanceToPrecombatMainOf(t, g, i)
		}
	}
}

// --- Stolen Uniform -----------------------------------------------------------------

func TestFinB7StolenUniformTakesAndAttachesTheEquipmentThenUnattachesWhenItGoesHome(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	finB7RegisteredFull(t, finB7UniformOracle, "Stolen Uniform")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := b12Permanent(g, opp.ID, "Sword", "Artifact — Equipment")
	castCatalogSpell(t, g, "Stolen Uniform", "Instant", finB7UniformOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: sword},
	})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, sword)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the Equipment is controlled by %v, want me", c.Controller)
	}
	if !c.IsAttachedTo(bear) {
		t.Fatal("the Equipment was not attached to the chosen creature")
	}
	// Control returns at cleanup; the Equipment is unattached.
	advanceToMainOf(t, g, 1)
	passPriorityAroundTable(t, g)
	c, _ = battlefieldCard(g, sword)
	if c.Controller != opp.ID {
		t.Errorf("the Equipment is controlled by %v after the turn, want its owner", c.Controller)
	}
	if c.IsAttached() {
		t.Error("the Equipment is still attached after I lost control of it")
	}
}
