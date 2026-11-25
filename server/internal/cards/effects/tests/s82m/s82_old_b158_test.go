package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b158_test.go — Core Set 2020, batch 158 (#5518): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB158LifechangerOracle = "24adb147-95b2-4854-86a5-17d43f904601"
	oldB158AltarOracle       = "67631753-f37a-4f41-a6e6-1edead385acc"
	oldB158SphinxOracle      = "29f61caf-1aba-40cd-aff6-c3685a0997d7"
	oldB158DrakeOracle       = "ecd6587c-3548-4bc1-89e4-9f7fbbc2d2ad"
	oldB158KaaliaOracle      = "6307ec34-bef6-4ac6-a363-f9bef0d03b3b"
	oldB158DenizenOracle     = "fc59aa9a-3abd-4479-a63e-ef0eeb35b5f2"
	oldB158GrowthOracle      = "ec714e79-10ec-43ea-a0ff-f8c9083fe015"
	oldB158MastiffOracle     = "f62d4437-bac6-49c8-9aeb-07de78551df7"
	oldB158ServantOracle     = "5bf9f397-0216-4ec9-a57b-406758dcc233"
	oldB158MaskOracle        = "dd86d11f-0bd8-49a9-936d-e407ea2c1818"
	oldB158EmbercatOracle    = "e75bec5c-538d-44ce-a6c7-85580d04fac8"
)

func TestOldB158CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB158LifechangerOracle: "Loxodon Lifechanter",
		oldB158AltarOracle:       "Bloodsoaked Altar",
		oldB158SphinxOracle:      "Riddlemaster Sphinx",
		oldB158DrakeOracle:       "Cerulean Drake",
		oldB158KaaliaOracle:      "Kaalia, Zenith Seeker",
		oldB158DenizenOracle:     "Sage's Row Denizen",
		oldB158GrowthOracle:      "Growth Cycle",
		oldB158MastiffOracle:     "Pack Mastiff",
		oldB158ServantOracle:     "Undead Servant",
		oldB158MaskOracle:        "Mask of Immolation",
		oldB158EmbercatOracle:    "Chandra's Embercat",
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

func oldB158Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB158Activate(g *game.Game, who, source uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(who, source, idx, params)
}

func oldB158Graveyard(g *game.Game, owner uuid.UUID, name, oracle string) {
	p := g.PlayerByID(owner)
	p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Zombie", OracleID: oracle, Owner: owner, Controller: owner})
}

func oldB158Tokens(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Loxodon Lifechanter ---

func TestOldB158LifechangerSetsLifeToTotalToughnessWhenYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Loxodon Lifechanter", TypeLine: "Creature — Elephant Cleric", OracleID: oldB158LifechangerOracle, Power: 4, Toughness: 6})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 10 {
		t.Errorf("life is %d, want 10 (the Wall's 4 and the Lifechanter's 6)", me.Life)
	}
}

func TestOldB158LifechangerLetsYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Loxodon Lifechanter", TypeLine: "Creature — Elephant Cleric", OracleID: oldB158LifechangerOracle, Power: 4, Toughness: 6})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life changed from %d to %d after declining", life, me.Life)
	}
}

func TestOldB158LifechangerPumpsByYourLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lc := effectstest.PushCatalogPermanent(g, me.ID, "Loxodon Lifechanter", "Creature — Elephant Cleric", oldB158LifechangerOracle, false)
	me.Life = 7
	effectstest.B06AddMana(me, "W", "W", "W", "W", "W", "W")
	if err := oldB158Activate(g, me.ID, lc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, lc); p != 8 || tt != 8 {
		t.Errorf("the Lifechanter is %d/%d, want 8/8 (1/1 plus 7)", p, tt)
	}
}

// --- Bloodsoaked Altar ---

func TestOldB158BloodsoakedAltarMakesAFlyingDemonForItsCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	altar := effectstest.PushCatalogPermanent(g, me.ID, "Bloodsoaked Altar", "Artifact", oldB158AltarOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	pitched := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitched, Name: "Pitched", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	life := me.Life
	if err := oldB158Activate(g, me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !me.Graveyard.Contains(pitched) {
		t.Error("the creature was not sacrificed or the card not discarded")
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 paid", life, me.Life)
	}
	demons := oldB158Tokens(g, me.ID, "Demon")
	if len(demons) != 1 {
		t.Fatalf("%d Demon tokens, want 1", len(demons))
	}
	if p, tt := effectstest.PtOf(t, g, demons[0].InstanceID); p != 5 || tt != 5 || !effectstest.HasEffectiveKeyword(t, g, demons[0].InstanceID, "flying") {
		t.Errorf("the token is %d/%d, want a 5/5 with flying", p, tt)
	}
}

func TestOldB158BloodsoakedAltarRefusesWithoutADiscardOrOutsideYourMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	altar := effectstest.PushCatalogPermanent(g, me.ID, "Bloodsoaked Altar", "Artifact", oldB158AltarOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	if err := oldB158Activate(g, me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("the Altar was activated with no card to discard")
	}
	pitched := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitched, Name: "Pitched", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := oldB158Activate(g, me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, DiscardIDs: []uuid.UUID{pitched}}); err == nil {
		t.Error("the Altar was activated at sorcery speed on an opponent's turn")
	}
}

// --- Riddlemaster Sphinx ---

func TestOldB158RiddlemasterSphinxMayReturnAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Riddlemaster Sphinx", TypeLine: "Creature — Sphinx", OracleID: oldB158SphinxOracle, Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opposing creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature left the battlefield")
	}
}

func TestOldB158RiddlemasterSphinxLetsYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Riddlemaster Sphinx", TypeLine: "Creature — Sphinx", OracleID: oldB158SphinxOracle, Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the creature was returned after the player declined")
	}
}

// --- Cerulean Drake ---

func TestOldB158CeruleanDrakeCountersASpellThatTargetsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Cerulean Drake", "Creature — Drake", oldB158DrakeOracle, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	life := me.Life
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := oldB158Activate(g, me.ID, drake, 0, game.ActivateAbilityParams{Targets: oldB158Target(bolt)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(drake) {
		t.Error("the Drake was not sacrificed")
	}
	if me.Life != life {
		t.Errorf("life %d -> %d: the countered spell still resolved", life, me.Life)
	}
	_ = opp
}

func TestOldB158CeruleanDrakeCantCounterASpellThatOnlyTargetsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Cerulean Drake", "Creature — Drake", oldB158DrakeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	if err := oldB158Activate(g, me.ID, drake, 0, game.ActivateAbilityParams{Targets: oldB158Target(bolt)}); err == nil {
		t.Error("the Drake targeted a spell that does not target its controller")
	}
}

func TestOldB158CeruleanDrakeIsProtectedFromRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Cerulean Drake", "Creature — Drake", oldB158DrakeOracle, false)
	for _, kw := range []string{"flying", "protection from red"} {
		if !effectstest.HasEffectiveKeyword(t, g, drake, kw) {
			t.Errorf("the Drake lacks %q", kw)
		}
	}
	_ = opp
}

// --- Kaalia, Zenith Seeker ---

func TestOldB158KaaliaTakesAnAngelADemonAndADragonAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := func(name, typeLine string) uuid.UUID {
		return effectstest.PushLibraryCard(me, game.Card{Name: name, TypeLine: typeLine, Power: 2, Toughness: 2})
	}
	deepA := mk("Deep Card A", "Sorcery")
	deepB := mk("Deep Card B", "Sorcery")
	rest1 := mk("Plain Bear", "Creature — Bear")
	dragon := mk("A Dragon", "Creature — Dragon")
	demon := mk("A Demon", "Creature — Demon")
	angel := mk("An Angel", "Creature — Angel")
	rest2 := mk("Another Bear", "Creature — Bear")
	rest3 := mk("Third Bear", "Creature — Bear")
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Kaalia, Zenith Seeker", TypeLine: "Legendary Creature — Human Cleric", OracleID: oldB158KaaliaOracle, Power: 3, Toughness: 3, Keywords: []string{"flying", "vigilance"}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, angel, demon, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Angel": angel, "Demon": demon, "Dragon": dragon} {
		if !me.Hand.Contains(id) {
			t.Errorf("the %s was not put into hand", name)
		}
	}
	for _, id := range []uuid.UUID{rest1, rest2, rest3} {
		if !me.Library.Contains(id) {
			t.Error("a card that was not taken left the library")
		}
	}
	// "The rest" go to the bottom, beneath the cards that were never looked at.
	cards := me.Library.Cards
	if cards[len(cards)-1].InstanceID != deepB || cards[len(cards)-2].InstanceID != deepA {
		t.Error("the cards that were not looked at are not on top")
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range cards[:3] {
		bottom[c.InstanceID] = true
	}
	if !bottom[rest1] || !bottom[rest2] || !bottom[rest3] {
		t.Error("the three cards left over are not the bottom three of the library")
	}
}

func TestOldB158KaaliaRefusesTwoCardsOfOneKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushLibraryCard(me, game.Card{Name: "Angel One", TypeLine: "Creature — Angel", Power: 2, Toughness: 2})
	b := effectstest.PushLibraryCard(me, game.Card{Name: "Angel Two", TypeLine: "Creature — Angel", Power: 2, Toughness: 2})
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Kaalia, Zenith Seeker", TypeLine: "Legendary Creature — Human Cleric", OracleID: oldB158KaaliaOracle, Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Error("two Angel cards were accepted for the one Angel slot")
	}
}

// --- Sage's Row Denizen ---

func oldB158Denizen(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Sage's Row Denizen", "Creature — Vedalken Wizard", oldB158DenizenOracle, 2, 3)
}

func TestOldB158SagesRowDenizenMillsWhenAnotherBlueCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB158Denizen(g, me)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCard(opp, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	before := len(opp.Library.Cards)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Blue Bear", TypeLine: "Creature — Bear", Colors: []string{"U"}, Power: 2, Toughness: 2})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - len(opp.Library.Cards); got != 2 {
		t.Errorf("the target player milled %d cards, want 2", got)
	}
}

func TestOldB158SagesRowDenizenIgnoresNonblueAndOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB158Denizen(g, me)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCard(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	before := len(me.Library.Cards)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Power: 2, Toughness: 2})
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Blue Bear", TypeLine: "Creature — Bear", Colors: []string{"U"}, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(me.Library.Cards) != before {
		t.Error("the Denizen triggered for a creature that is not blue and mine")
	}
}

// --- Growth Cycle ---

func TestOldB158GrowthCycleCountsCopiesInYourGraveyard(t *testing.T) {
	for _, copies := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		for i := 0; i < copies; i++ {
			oldB158Graveyard(g, me.ID, "Growth Cycle", oldB158GrowthOracle)
		}
		// A copy in an opponent's graveyard is not counted.
		oldB158Graveyard(g, opp.ID, "Growth Cycle", oldB158GrowthOracle)
		effectstest.CastCatalogSpell(t, g, "Growth Cycle", "Instant", oldB158GrowthOracle, oldB158Target(bear))
		effectstest.PassPriorityAroundTable(t, g)
		want := 2 + 3 + 2*copies
		if p, tt := effectstest.PtOf(t, g, bear); p != want || tt != want {
			t.Errorf("with %d copies in the graveyard the bear is %d/%d, want %d/%d", copies, p, tt, want, want)
		}
	}
}

// --- Pack Mastiff ---

func TestOldB158PackMastiffPumpsEveryMastiffYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	one := effectstest.PushCatalogPermanent(g, me.ID, "Pack Mastiff", "Creature — Dog", oldB158MastiffOracle, false)
	two := effectstest.PushCatalogPermanent(g, me.ID, "Pack Mastiff", "Creature — Dog", oldB158MastiffOracle, false)
	other := effectstest.B12Creature(g, me.ID, "Other Dog", "Creature — Dog", 2, 2)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Pack Mastiff", "Creature — Dog", oldB158MastiffOracle, false)
	effectstest.B06AddMana(me, "R", "R")
	if err := oldB158Activate(g, me.ID, one, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	base := func(id uuid.UUID) int { p, _ := effectstest.PtOf(t, g, id); return p }
	if base(one) != 2 || base(two) != 2 {
		t.Errorf("my Mastiffs have power %d and %d, want 2 and 2 (1 base +1)", base(one), base(two))
	}
	if base(other) != 2 || base(theirs) != 1 {
		t.Errorf("a creature with another name has power %d and the opponent's Mastiff %d, want 2 and 1", base(other), base(theirs))
	}
}

// --- Undead Servant ---

func TestOldB158UndeadServantMakesAZombiePerCopyInYourGraveyard(t *testing.T) {
	for _, copies := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < copies; i++ {
			oldB158Graveyard(g, me.ID, "Undead Servant", oldB158ServantOracle)
		}
		oldB158Graveyard(g, opp.ID, "Undead Servant", oldB158ServantOracle)
		effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Undead Servant", TypeLine: "Creature — Zombie", OracleID: oldB158ServantOracle, Power: 3, Toughness: 2})
		effectstest.PassPriorityAroundTable(t, g)
		zombies := oldB158Tokens(g, me.ID, "Zombie")
		if len(zombies) != copies {
			t.Errorf("with %d copies in the graveyard %d Zombie tokens, want %d", copies, len(zombies), copies)
		}
		for _, z := range zombies {
			if p, tt := effectstest.PtOf(t, g, z.InstanceID); p != 2 || tt != 2 {
				t.Errorf("a Zombie token is %d/%d, want 2/2", p, tt)
			}
		}
	}
}

// --- Mask of Immolation ---

func TestOldB158MaskOfImmolationMakesAnElementalAndAttachesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mask := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Mask of Immolation", TypeLine: "Artifact — Equipment", OracleID: oldB158MaskOracle})
	effectstest.PassPriorityAroundTable(t, g)
	elementals := oldB158Tokens(g, me.ID, "Elemental")
	if len(elementals) != 1 {
		t.Fatalf("%d Elemental tokens, want 1", len(elementals))
	}
	if !effectstest.AttachedTo(t, g, mask, elementals[0].InstanceID) {
		t.Error("the Mask is not attached to the token it made")
	}
}

func TestOldB158MaskOfImmolationLetsTheEquippedCreatureSacrificeForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Mask of Immolation", TypeLine: "Artifact — Equipment", OracleID: oldB158MaskOracle})
	effectstest.PassPriorityAroundTable(t, g)
	elemental := oldB158Tokens(g, me.ID, "Elemental")[0].InstanceID
	life := opp.Life
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if err := oldB158Activate(g, me.ID, elemental, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elemental) {
		t.Error("the equipped creature was not sacrificed")
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want 1 damage", life, opp.Life)
	}
}

func TestOldB158MaskOfImmolationGrantsNothingOnceItIsUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mask := effectstest.PushCatalogPermanent(g, me.ID, "Mask of Immolation", "Artifact — Equipment", oldB158MaskOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB158Activate(g, me.ID, bear, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("an unequipped creature sacrificed itself for damage")
	}
	effectstest.B06AddMana(me, "R", "R")
	if err := oldB158Activate(g, me.ID, mask, 0, game.ActivateAbilityParams{Targets: oldB158Target(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, mask, bear) {
		t.Fatal("the Mask did not attach to the bear")
	}
}

// --- Chandra's Embercat ---

func TestOldB158EmbercatManaPaysForAnElementalOrAChandraOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cat := effectstest.PushCatalogPermanent(g, me.ID, "Chandra's Embercat", "Creature — Elemental Cat", oldB158EmbercatOracle, false)
	hand := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{R}", Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
		return id
	}
	plain := hand("Goblin", "Creature — Goblin")
	jace := hand("Jace Test", "Legendary Planeswalker — Jace")
	elemental := hand("Spark", "Creature — Elemental")
	chandra := hand("Chandra Test", "Legendary Planeswalker — Chandra")
	cast := func(id uuid.UUID) error {
		effectstest.TapForMana(t, g, me.ID, cat)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if err != nil {
			g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == cat {
						g.Battlefield.Cards[i].Tapped = false
					}
				}
			})
		}
		return err
	}
	if err := cast(plain); err == nil {
		t.Error("the Embercat's mana paid for a creature that is not an Elemental")
	}
	if err := cast(jace); err == nil {
		t.Error("the Embercat's mana paid for a planeswalker that is not a Chandra")
	}
	if err := cast(elemental); err != nil {
		t.Errorf("the Embercat's mana did not pay for an Elemental spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cat {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := cast(chandra); err != nil {
		t.Errorf("the Embercat's mana did not pay for a Chandra planeswalker spell: %v", err)
	}
}
