package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b13_cards_test.go — S82 Innistrad: Crimson Vow, batch 13 (#5495):
// Geistlight Snare; Supernatural Rescue; Dreamshackle Geist; Steelclad
// Spirit; Wanderlight Spirit; Stormchaser Drake; Serpentine Ambush;
// Wandering Mind; By Invitation Only; Consuming Tide; Sigarda's Summons.

const (
	vowB13SnareOracle     = "45f159dc-a5f9-4335-896b-ad544c303801"
	vowB13RescueOracle    = "a77510ef-baf0-4cc5-b659-6aaaed0b02f4"
	vowB13GeistOracle     = "aa40d270-63d7-4022-ade9-67c1ab65ef70"
	vowB13SteelcladOracle = "3dc225a2-61b6-4ac9-a86b-df0c2b0b3898"
	vowB13WanderOracle    = "7b4c2dcf-86f4-409d-90c3-8fe83966cac0"
	vowB13DrakeOracle     = "a9530416-d5fb-4ddd-ae6f-0eefb9f276c5"
	vowB13AmbushOracle    = "822b6d56-939a-42b9-b50c-9a8ba50ec3ad"
	vowB13MindOracle      = "0fa886c2-9b58-4368-b2d0-e619c62a554f"
	vowB13InviteOracle    = "d64e96bf-6815-462c-93b4-c9c22dc43eda"
	vowB13TideOracle      = "24cd8f54-c1d4-4040-93bf-a7b9b719ac30"
	vowB13SummonsOracle   = "7fd06710-0721-4e3e-90ed-d0ccecb1919a"
)

func TestVowB13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB13SnareOracle:     "Geistlight Snare",
		vowB13RescueOracle:    "Supernatural Rescue",
		vowB13GeistOracle:     "Dreamshackle Geist",
		vowB13SteelcladOracle: "Steelclad Spirit",
		vowB13WanderOracle:    "Wanderlight Spirit",
		vowB13DrakeOracle:     "Stormchaser Drake",
		vowB13AmbushOracle:    "Serpentine Ambush",
		vowB13MindOracle:      "Wandering Mind",
		vowB13InviteOracle:    "By Invitation Only",
		vowB13TideOracle:      "Consuming Tide",
		vowB13SummonsOracle:   "Sigarda's Summons",
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

// --- Wanderlight Spirit ----------------------------------------------------

func TestVowB13WanderlightSpiritBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spirit := effectstest.B12Push(g, opp.ID, "Wanderlight Spirit", "Creature — Spirit", vowB13WanderOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "flying") {
		t.Error("the Spirit lacks flying")
	}
	if err := effectstest.PorBlock(t, g, bear, opp.ID, spirit); err == nil {
		t.Error("it blocked a creature without flying")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bird := effectstest.PorFlyer(g, me.ID, "Bird")
	spirit = effectstest.B12Push(g, opp.ID, "Wanderlight Spirit", "Creature — Spirit", vowB13WanderOracle, 2, 3)
	if err := effectstest.PorBlock(t, g, bird, opp.ID, spirit); err != nil {
		t.Errorf("it could not block a flyer: %v", err)
	}
}

// --- Serpentine Ambush and Stormchaser Drake -------------------------------

func TestVowB13SerpentineAmbushMakesABlueSerpentFiveFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Serpentine Ambush", "Instant", vowB13AmbushOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("%d/%d, want 5/5", p, tough)
	}
	c, _ := g.LookupCardForEffect(bear)
	if !c.HasSubtype("Serpent") || c.HasSubtype("Bear") {
		t.Errorf("want Serpent and no longer a Bear")
	}
	if cols := c.EffectiveColors(); len(cols) != 1 || cols[0] != "U" {
		t.Errorf("colours %v, want blue", cols)
	}
	// Until end of turn.
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d after the turn, want 2", p)
	}
}

func TestVowB13StormchaserDrakeDrawsOnlyForYourSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.B12Push(g, me.ID, "Stormchaser Drake", "Creature — Drake", vowB13DrakeOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake lacks flying")
	}
	effectstest.SeedLibrary(me, "A", "B")
	before := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Serpentine Ambush", "Instant", vowB13AmbushOracle, effectstest.CardTarget(drake))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 1 {
		t.Errorf("drew %d cards for your own spell, want 1", got)
	}

	// An opponent's spell does not draw.
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake = effectstest.B12Push(g, me.ID, "Stormchaser Drake", "Creature — Drake", vowB13DrakeOracle, 2, 1)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	before = len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Serpentine Ambush", "Instant", vowB13AmbushOracle, effectstest.CardTarget(drake))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 0 {
		t.Errorf("drew %d cards for %s's spell, want 0", got, opp.ID)
	}
}

// --- Steelclad Spirit ------------------------------------------------------

func TestVowB13SteelcladSpiritAttacksAfterAnEnchantmentEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := effectstest.B12Push(g, me.ID, "Steelclad Spirit", "Creature — Spirit", vowB13SteelcladOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "A Shrine", TypeLine: "Enchantment"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(spirit, opp.ID); err != nil {
		t.Errorf("it could not attack after an enchantment entered: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "defender") {
		t.Error("the waiver took the defender keyword away")
	}
}

func TestVowB13SteelcladSpiritCannotAttackOtherwise(t *testing.T) {
	// No enchantment: defender stops it.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := effectstest.B12Push(g, me.ID, "Steelclad Spirit", "Creature — Spirit", vowB13SteelcladOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(spirit, opp.ID); err == nil {
		t.Error("it attacked with no enchantment entering")
	}
	// An opponent's enchantment, or a non-enchantment of yours, does not.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	spirit = effectstest.B12Push(g, me.ID, "Steelclad Spirit", "Creature — Spirit", vowB13SteelcladOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.TrekB31Enter(t, g, opp, game.Card{Name: "Their Shrine", TypeLine: "Enchantment"})
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(spirit, opp.ID); err == nil {
		t.Error("it attacked off an opponent's enchantment or a creature")
	}
}

// --- Dreamshackle Geist ----------------------------------------------------

func TestVowB13DreamshackleGeistTapsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dreamshackle Geist", "Creature — Spirit", vowB13GeistOracle, 3, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, bear) {
		t.Error("the target was not tapped")
	}
}

func TestVowB13DreamshackleGeistFreezesATargetForOneUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dreamshackle Geist", "Creature — Spirit", vowB13GeistOracle, 3, 1)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, bear) {
		t.Fatal("the second mode must not untap it")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AssertTurnOf(t, g, 0, false, "my next turn")
	if !effectstest.TappedNow(g, bear) {
		t.Error("its controller's next untap step untapped it")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	if effectstest.TappedNow(g, bear) {
		t.Error("it stayed tapped through a second untap step")
	}
}

func TestVowB13DreamshackleGeistMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dreamshackle Geist", "Creature — Spirit", vowB13GeistOracle, 3, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for a mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) {
		t.Error("choosing no mode tapped something")
	}
}

// --- Wandering Mind --------------------------------------------------------

func TestVowB13WanderingMindTakesANoncreatureNonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	names := []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Bear", "Creature — Bear"},
		{"Shock", "Instant"},
		{"Bauble", "Artifact"},
		{"Plains", "Basic Land — Plains"},
		{"Wolf", "Creature — Wolf"},
		{"Seventh", "Sorcery"},
	}
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: names[i].Name, TypeLine: names[i].TypeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Wandering Mind", "Creature — Horror", vowB13MindOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger raised no pick")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[ids[2]] || !offered[ids[3]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the Shock and the artifact, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the Shock is not in hand")
	}
	// The seventh card was never looked at: it is the new top of the library
	// (index 0 is the bottom), and the five others went under it.
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[6] || me.Library.Cards[0].InstanceID == ids[6] {
		t.Error("the unseen card should now be on top, with the other cards below it")
	}
}

// --- By Invitation Only ----------------------------------------------------

func TestVowB13ByInvitationOnlyEveryPlayerSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var all [][]uuid.UUID
	for _, p := range g.Seats {
		var ids []uuid.UUID
		for i := 0; i < 3; i++ {
			ids = append(ids, effectstest.B12Creature(g, p.ID, "Fodder", "Creature — Bear", 1, 1))
		}
		all = append(all, ids)
	}
	effectstest.CastCatalogSpell(t, g, "By Invitation Only", "Sorcery", vowB13InviteOracle, nil)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Fatal("nothing is asked while the spell is on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no number prompt")
	}
	if pa := c.PayAmount; pa.Min != 0 || pa.Max != 13 {
		t.Fatalf("prompt %+v, want 0 through 13", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}
	asked := map[uuid.UUID]int{}
	for guard := 0; guard < 16; guard++ {
		progressed := false
		for _, p := range g.Seats {
			if prompt := effectstest.SacrificeChoiceFor(g, p.ID); prompt != nil {
				effectstest.AnswerSacrifice(t, g, p.ID, effectstest.SacrificeOptions(prompt)[0])
				asked[p.ID]++
				progressed = true
				break
			}
		}
		if !progressed {
			break
		}
	}
	for i, p := range g.Seats {
		if asked[p.ID] != 2 {
			t.Errorf("seat %d was asked %d times, want 2", i, asked[p.ID])
		}
		left := 0
		for _, id := range all[i] {
			if g.Battlefield.Contains(id) {
				left++
			}
		}
		if left != 1 {
			t.Errorf("seat %d has %d of its 3 creatures left, want 1", i, left)
		}
	}
}

func TestVowB13ByInvitationOnlyZeroSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "By Invitation Only", "Sorcery", vowB13InviteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a 0 sacrificed a creature")
	}
}

// --- Consuming Tide --------------------------------------------------------

func TestVowB13ConsumingTideKeepsOneNonlandPermanentPerPlayerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vowB13ClearHands(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	for i := 0; i < 5; i++ {
		g.WithWriteLock(func() {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		})
	}
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myOgre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	myLand := pushTappedLand(g, me.ID)
	theirA := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	theirB := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Soldier", TypeLine: "Token Creature — Soldier", TokenKey: "test/soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Consuming Tide", "Sorcery", vowB13TideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, step := range []struct {
		p    *game.Player
		keep uuid.UUID
	}{{me, myOgre}, {opp, theirB}} {
		pick := effectstest.LatestChooseCards(g, step.p.ID)
		if pick == nil {
			t.Fatalf("%s was not asked", step.p.ID)
		}
		if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
			t.Errorf("min %d max %d, want exactly one", pick.ChooseMin, pick.ChooseMax)
		}
		for _, id := range pick.ChooseCards {
			if id == myLand {
				t.Error("a land was offered")
			}
		}
		if err := g.ResolveChooseCards(pick.ID, step.p.ID, []uuid.UUID{step.keep}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myOgre) || !g.Battlefield.Contains(theirB) || !g.Battlefield.Contains(myLand) {
		t.Error("the chosen permanents and the land must stay")
	}
	if g.Battlefield.Contains(myBear) || g.Battlefield.Contains(theirA) || g.Battlefield.Contains(token) {
		t.Error("every other nonland permanent must return")
	}
	if !me.Hand.Contains(myBear) {
		t.Error("my Bear is not in my hand")
	}
	if !opp.Hand.Contains(theirA) {
		t.Error("their Bear is not in their hand")
	}
	// Hands after the return: me 1 (the Bear); the first opponent 6; the two
	// others 0. Only one opponent has more cards than me, so I draw one.
	if got := len(me.Hand.Cards); got != 2 {
		t.Errorf("my hand is %d cards, want 2 (the returned Bear and one draw)", got)
	}
}

func TestVowB13ConsumingTideDrawsNothingWhenNoOpponentHasMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vowB13ClearHands(g)
	effectstest.SeedLibrary(me, "A", "B")
	only := effectstest.B12Creature(g, me.ID, "My Only", "Creature — Bear", 2, 2)
	_ = opp
	effectstest.CastCatalogSpell(t, g, "Consuming Tide", "Sorcery", vowB13TideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("a player with one nonland permanent was asked to choose")
	}
	if !g.Battlefield.Contains(only) {
		t.Error("the only permanent was returned")
	}
	if got := len(me.Hand.Cards); got != 0 {
		t.Errorf("drew %d, want 0", got)
	}
}

// --- Sigarda's Summons -----------------------------------------------------

func TestVowB13SigardasSummonsMakesCounteredCreaturesAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sigarda's Summons", "Enchantment", vowB13SummonsOracle, 0, 0)
	counted := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 1, 1)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.PutCounters(t, g, counted, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)
	// 4/4 base plus the counter.
	if p := effectstest.EffectivePowerWithCounters(t, g, counted); p != 5 {
		t.Errorf("counted creature has power %d, want 5 (base 4 plus the counter)", p)
	}
	if p := effectstest.EffectivePower(t, g, counted); p != 4 {
		t.Errorf("counted creature has base power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, counted, "flying") {
		t.Error("the counted creature lacks flying")
	}
	if c, _ := g.LookupCardForEffect(counted); !c.HasSubtype("Angel") || !c.HasSubtype("Bear") {
		t.Errorf("want a Bear and an Angel")
	}
	if effectstest.EffectivePowerWithCounters(t, g, plain) != 1 || effectstest.HasEffectiveKeyword(t, g, plain, "flying") {
		t.Error("a creature without a counter changed")
	}
	if effectstest.EffectivePowerWithCounters(t, g, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's creature was affected")
	}
}

// --- Geistlight Snare ------------------------------------------------------

// vowB13ClearHands empties every seat's hand.
func vowB13ClearHands(g *game.Game) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
}

func vowB13SnareTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, _ := effectstest.P7Table(t)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	snare := effectstest.FlashIfHandCard(g, me, "Geistlight Snare", "Instant", "{2}{U}", vowB13SnareOracle)
	return g, me, spell, snare
}

func TestVowB13GeistlightSnareCostsLessForSpiritAndEnchantment(t *testing.T) {
	for _, tc := range []struct {
		name     string
		spirit   bool
		ench     bool
		pay      []string
		wantFail bool
	}{
		{"neither: full price", false, false, []string{"U", "C"}, true},
		{"neither paid in full", false, false, []string{"U", "C", "C"}, false},
		{"a Spirit saves one", true, false, []string{"U", "C"}, false},
		{"an enchantment saves one", false, true, []string{"U", "C"}, false},
		{"both save two", true, true, []string{"U"}, false},
		{"both, but nothing paid", true, true, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, spell, snare := vowB13SnareTable(t)
			if tc.spirit {
				effectstest.B12Creature(g, me.ID, "A Spirit", "Creature — Spirit", 1, 1)
			}
			if tc.ench {
				effectstest.B12Push(g, me.ID, "An Enchantment", "Enchantment", "", 0, 0)
			}
			effectstest.FlashIfFund(g, me, tc.pay...)
			err := g.CastSpell(me.ID, snare, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(spell)})
			if tc.wantFail != (err != nil) {
				t.Errorf("cast err = %v, want failure %v", err, tc.wantFail)
			}
		})
	}
}

func TestVowB13GeistlightSnareCountersUnlessThreeIsPaid(t *testing.T) {
	g, me, spell, snare := vowB13SnareTable(t)
	effectstest.FlashIfFund(g, me, "U", "C", "C")
	if err := g.CastSpell(me.ID, snare, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(spell)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

// --- Supernatural Rescue ---------------------------------------------------

func TestVowB13SupernaturalRescueFlashNeedsASpirit(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rescue := effectstest.FlashIfHandCard(g, me, "Supernatural Rescue", "Enchantment — Aura", "{3}{W}", vowB13RescueOracle)
	effectstest.SsToBeginCombat(t, g)
	effectstest.FlashIfFund(g, me, "W", "C", "C", "C")
	err := g.CastSpell(me.ID, rescue, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("no Spirit: err = %v, want ErrSorcerySpeedRequired", err)
	}
	effectstest.B12Creature(g, me.ID, "A Spirit", "Creature — Spirit", 1, 1)
	if err := g.CastSpell(me.ID, rescue, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatalf("with a Spirit: %v", err)
	}
}

func TestVowB13SupernaturalRescueTapsTwoAndEnchants(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Their C", "Creature — Bear", 2, 2)
	ownTapped := effectstest.B12Creature(g, me.ID, "Mine Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	rescue := effectstest.FlashIfHandCard(g, me, "Supernatural Rescue", "Enchantment — Aura", "{3}{W}", vowB13RescueOracle)
	effectstest.FlashIfFund(g, me, "W", "C", "C", "C")
	if err := g.CastSpell(me.ID, rescue, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger asked for no targets")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ownTapped}); err == nil {
		t.Error("a creature you control was accepted as a target")
	}
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, a) || !effectstest.TappedNow(g, b) {
		t.Error("the two targets were not tapped")
	}
	if effectstest.TappedNow(g, c) {
		t.Error("an untargeted creature was tapped")
	}
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 4 {
		t.Errorf("enchanted creature is %d/%d, want 3/4", p, tough)
	}
}
