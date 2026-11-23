package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func finB7RegisteredFull(t *testing.T, oracle, name string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
	return spec
}

// --- Insomnia, Crown City -----------------------------------------------------------

func TestFinB7InsomniaEntersTappedAndTapsForWOrB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PlayLandFromHand(t, g, "Insomnia, Crown City", finB7InsomniaOracle)
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("Insomnia should enter tapped")
	}
	spec := finB7RegisteredFull(t, finB7InsomniaOracle, "Insomnia, Crown City")
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|B}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {W|B}", spec.ManaAbilities)
	}
}

// --- Relentless X-ATM092 ------------------------------------------------------------

func TestFinB7RelentlessCantBeBlockedExceptByThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	finB7RegisteredFull(t, finB7RelentlessOracle, "Relentless X-ATM092")
	robot := effectstest.B12Push(g, me.ID, "Relentless X-ATM092", "Artifact Creature — Robot Spider", finB7RelentlessOracle, 6, 5)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, robot)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: robot}, {Blocker: b, Attacker: robot},
	}), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: robot}, {Blocker: b, Attacker: robot}, {Blocker: c, Attacker: robot},
	}); err != nil {
		t.Fatalf("three blockers are legal: %v", err)
	}
}

func TestFinB7RelentlessReturnsTappedWithAFinalityCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	robot := effectstest.PushCatalogGraveyardCard(me, "Relentless X-ATM092", "Artifact Creature — Robot Spider", finB7RelentlessOracle, 6, 5)
	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, robot, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only seven mana")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, robot, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, robot)
	if !ok {
		t.Fatal("the robot did not return")
	}
	if !c.Tapped {
		t.Error("the robot should return tapped")
	}
	if got := effectstest.CountersOn(g, robot, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
}

// --- Zodiark, Umbral God ------------------------------------------------------------

func TestFinB7ZodiarkEdictsHalfTheNonGodCreaturesAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	var mine, theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		mine = append(mine, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	god := effectstest.B12Creature(g, me.ID, "Stone God", "Legendary Creature — God", 2, 2)
	for i := 0; i < 3; i++ {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2))
	}
	lone := effectstest.B12Creature(g, third.ID, "Lone", "Creature — Bear", 1, 1)
	_ = fourth
	zod := effectstest.CastCatalogSpell(t, g, "Zodiark, Umbral God", "Legendary Creature — God", finB7ZodiarkOracle, nil)
	finB7RegisteredFull(t, finB7ZodiarkOracle, "Zodiark, Umbral God")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine[:2]...)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs[0])
	if effectstest.SacrificeChoiceFor(g, third.ID) != nil {
		t.Error("a lone creature owes no sacrifice (half of one, rounded down)")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	if got := effectstest.CountersOn(g, zod, game.CounterPlusOne); got != 3 {
		t.Errorf("Zodiark has %d +1/+1 counters, want 3 (one per sacrificed creature)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, zod, "indestructible") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB7RegisteredFull(t, finB7TellahOracle, "Tellah, Great Sage")
	tellah := effectstest.SocB17Perm(g, me.ID, "Tellah, Great Sage", "Legendary Creature — Human Wizard", finB7TellahOracle, "{3}{U}{R}", 3, 3)
	hand := me.Hand.Size()
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{1}{U}", "UU")
	if n := finB7Heroes(g, me.ID); n != 1 {
		t.Errorf("%d Hero tokens after a two-mana spell, want 1", n)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after a two-mana spell, want no draw", hand, me.Hand.Size())
	}
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{4}{U}", "UUUUU")
	if n := finB7Heroes(g, me.ID); n != 1 {
		t.Errorf("a creature spell made a Hero (%d)", n)
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{3}{U}", "UUUU")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tellah := effectstest.SocB17Perm(g, me.ID, "Tellah, Great Sage", "Legendary Creature — Human Wizard", finB7TellahOracle, "{3}{U}{R}", 3, 3)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{6}{U}{U}", "UUUUUUUU")
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB7RegisteredFull(t, finB7FenrirOracle, "Summon: Fenrir")
	forest := effectstest.TopLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.TopLibraryCard(me, "Filler", "Sorcery")
	saga := effectstest.FinS3CastSaga(t, g, "Summon: Fenrir", "Enchantment Creature — Saga Wolf", finB7FenrirOracle, "{2}{G}", 3, 2)
	effectstest.PassPriorityAroundTable(t, g) // the creature, then chapter I
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || !c.Tapped {
		t.Fatalf("the basic land is on the battlefield %v, tapped %v; want it tapped", ok, c.Tapped)
	}

	// Chapter II: a noncreature spell leaves the promise, the next creature spell uses it.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.CastCatalogSpell(t, g, "Second Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, first, game.CounterPlusOne); got != 1 {
		t.Errorf("the first creature has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("the second creature has %d +1/+1 counters, want 0 (one use)", got)
	}

	// Chapter III: the Saga's 3 power ties or beats everything I control, but a 9/9
	// elsewhere is the greatest power, so no card.
	big := effectstest.B12Creature(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Behemoth", "Creature — Beast", 9, 9)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	_ = saga
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d with an opponent holding the greatest power, want no draw", hand, me.Hand.Size())
	}
	_ = big
}

func TestFinB7FenrirChapterThreeDrawsWhenTiedForTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.B12Creature(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Rival", "Creature — Beast", 3, 3)
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Summon: Fenrir", TypeLine: "Enchantment Creature — Saga Wolf",
		OracleID: finB7FenrirOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 2},
	})
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a draw for tying the greatest power", hand, me.Hand.Size())
	}
	_ = saga
}

// --- Triple Triad -------------------------------------------------------------------

func TestFinB7TripleTriadLetsYouPlayYourCardAndEveryLesserOneFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	finB7RegisteredFull(t, finB7TripleOracle, "Triple Triad")
	effectstest.S58p6Push(g, me.ID, "Triple Triad", "Enchantment", finB7TripleOracle, "{3}{R}{R}{R}", 0, 0)
	effectstest.B12ToMyNextUpkeep(t, g)
	mine := effectstest.SeedLibraryTop(me, "Mine", "Creature — Bear", "{3}")
	// Seeded once the table is at my upkeep: the others' draws are behind us.
	lesser := effectstest.SeedLibraryTop(g.Seats[1], "Lesser", "Instant", "{2}")
	equal := effectstest.SeedLibraryTop(g.Seats[2], "Equal", "Sorcery", "{3}")
	land := effectstest.SeedLibraryTop(g.Seats[3], "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.S58p6Push(g, me.ID, "Triple Triad", "Enchantment", finB7TripleOracle, "{3}{R}{R}{R}", 0, 0)
	mine := effectstest.SeedLibraryTop(me, "Mine", "Land", "")
	other := effectstest.SeedLibraryTop(g.Seats[1], "Other", "Creature — Bear", "{1}")
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if perm := g.CastPermissionOnCardByIDForEffect(mine); perm == nil {
		t.Error("my own card should be playable")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(other); perm != nil {
		t.Error("a card with greater mana value than mine was made playable")
	}
}

// --- Vanille, Cheerful l'Cie --------------------------------------------------------

func TestFinB7VanilleMillsTwoThenReturnsAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB7RegisteredFull(t, finB7VanilleOracle, "Vanille, Cheerful l'Cie")
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	bear := effectstest.PushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(me, "Old Spell", "Instant", "", 0, 0)
	library := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Vanille, Cheerful l'Cie", "Legendary Creature — Human Cleric", finB7VanilleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	print := finB7FangMeldPrint()
	vanille := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanille, Cheerful l'Cie", OracleID: finB7VanilleOracle,
		TypeLine: "Legendary Creature — Human Cleric", ManaCost: "{3}{G}", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: print, ScryfallID: uuid.NewString(),
	})
	fang := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fang, Fearless l'Cie", TypeLine: "Legendary Creature — Human Warrior",
		ManaCost: "{3}{B}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
		Layout: game.LayoutMeld, Meld: print, ScryfallID: uuid.NewString(),
	})
	finB7ToMyNextMain(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "BGCCC")
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the meld ability asked for no payment")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanille, Cheerful l'Cie", OracleID: finB7VanilleOracle,
		TypeLine: "Legendary Creature — Human Cleric", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: finB7FangMeldPrint(), ScryfallID: uuid.NewString(),
	})
	finB7ToMyNextMain(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the meld ability was offered with no Fang")
	}
}

// finB7ToMyNextMain walks the table to `me`'s next precombat main and lets the
// beginning-of-main triggers go on the stack.
func finB7ToMyNextMain(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i, p := range g.Seats {
		if p.ID == me.ID {
			effectstest.AdvanceToPrecombatMainOf(t, g, i)
		}
	}
}

// --- Stolen Uniform -----------------------------------------------------------------

func TestFinB7StolenUniformTakesAndAttachesTheEquipmentThenUnattachesWhenItGoesHome(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	finB7RegisteredFull(t, finB7UniformOracle, "Stolen Uniform")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := effectstest.B12Permanent(g, opp.ID, "Sword", "Artifact — Equipment")
	effectstest.CastCatalogSpell(t, g, "Stolen Uniform", "Instant", finB7UniformOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: sword},
	})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, sword)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the Equipment is controlled by %v, want me", c.Controller)
	}
	if !c.IsAttachedTo(bear) {
		t.Fatal("the Equipment was not attached to the chosen creature")
	}
	// Control returns at cleanup; the Equipment is unattached.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ = effectstest.BattlefieldCard(g, sword)
	if c.Controller != opp.ID {
		t.Errorf("the Equipment is controlled by %v after the turn, want its owner", c.Controller)
	}
	if c.IsAttached() {
		t.Error("the Equipment is still attached after I lost control of it")
	}
}
