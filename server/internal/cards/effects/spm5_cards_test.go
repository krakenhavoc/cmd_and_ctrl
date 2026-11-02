package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm5_cards_test.go — Marvel's Spider-Man, batch 5 (#4460): News
// Helicopter, Ominous Asylum, Oscorp Industries, Oscorp Research Team,
// Passenger Ferry, Peter Parker // Amazing Spider-Man, Peter Parker's
// Camera, Pictures of Spider-Man, Prison Break, Professional Wrestler.

const (
	spm5NewsHelicopterOracle   = "192dd1e2-b6c2-432a-a1b4-e52ff39b3db2"
	spm5OminousAsylumOracle    = "676141c3-a433-4aba-86fb-729628f96dfa"
	spm5OscorpIndustriesOracle = "f432eb6a-f1bf-4ce7-b915-1488dccb4bb9"
	spm5ResearchTeamOracle     = "4d2e233c-0173-417f-82a0-1e692a400ae1"
	spm5PassengerFerryOracle   = "e9832d58-0ee6-4641-a9fd-d34d54b3b11c"
	spm5CameraOracle           = "10777360-c046-43e3-ab44-9d0b926fbbf8"
	spm5PicturesOracle         = "4fd74a8c-a4e7-4988-ae02-9a73ab6ec89e"
	spm5PrisonBreakOracle      = "9c6a6332-37f0-4d7d-a86f-478cf0eeec91"
	spm5WrestlerOracle         = "3dd92eea-02ca-4e55-8f69-8c8059dd7ee6"
)

// spm5Hand seeds a card into a player's hand and returns its id.
func spm5Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func spm5CastResolve(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string) {
	t.Helper()
	floatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func spm5Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestSpm5CardsAreFullWithNoCaveat(t *testing.T) {
	for name, oracle := range map[string]string{
		"News Helicopter":        spm5NewsHelicopterOracle,
		"Ominous Asylum":         spm5OminousAsylumOracle,
		"Oscorp Industries":      spm5OscorpIndustriesOracle,
		"Oscorp Research Team":   spm5ResearchTeamOracle,
		"Passenger Ferry":        spm5PassengerFerryOracle,
		"Peter Parker":           peterParkerOracle,
		"Amazing Spider-Man":     peterParkerOracle + "#1",
		"Peter Parker's Camera":  spm5CameraOracle,
		"Pictures of Spider-Man": spm5PicturesOracle,
		"Prison Break":           spm5PrisonBreakOracle,
		"Professional Wrestler":  spm5WrestlerOracle,
	} {
		requireFull(t, oracle, name)
	}
}

// --- News Helicopter ---------------------------------------------------

func TestSpm5NewsHelicopterFliesAndMakesAHumanCitizen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := spm5Hand(me, game.Card{Name: "News Helicopter", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{3}", OracleID: spm5NewsHelicopterOracle, Power: 1, Toughness: 1})
	spm5CastResolve(t, g, me, id, "CCC")
	passPriorityAroundTable(t, g)

	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("News Helicopter has no flying")
	}
	toks := spm5Tokens(g, me.ID, "Human Citizen")
	if len(toks) != 1 {
		t.Fatalf("%d Human Citizen tokens, want 1", len(toks))
	}
	c := toks[0]
	if c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Human") || !c.HasSubtype("Citizen") ||
		!c.HasColor("G") || !c.HasColor("W") || c.HasColor("B") {
		t.Errorf("token = %+v, want a 1/1 green and white Human Citizen", c)
	}
}

// --- Ominous Asylum ------------------------------------------------------

func TestSpm5OminousAsylumEntersTappedAndMakesBlackOrRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Ominous Asylum", spm5OminousAsylumOracle)
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, id, "Ominous Asylum")

	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = false })
	spm5TapForEach(t, g, me, id, 2, []string{"B", "R"})
}

// spm5TapForEach taps a land for its mana ability once per wanted colour,
// answering the colour pick, and checks each pool.
func spm5TapForEach(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, options int, want []string) {
	t.Helper()
	for _, w := range want {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != options {
			t.Fatalf("colour options %+v, want %d", pick, options)
		}
		riderAnswerManaPicks(t, g, me.ID, w)
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != w {
			t.Errorf("pool %v, want [%s]", got, w)
		}
		me.ManaPool.EmptyPool()
		b08Untap(g, land)
	}
}

func TestSpm5OminousAsylumSurveilsForFourAndTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Next Draw")
	land := pushCatalogPermanent(g, me.ID, "Ominous Asylum", "Land", spm5OminousAsylumOracle, false)
	advanceToMain(t, g)

	// Three mana is not enough.
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated with three mana")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil prompt")
	}
	if !b16Tapped(t, g, land) {
		t.Error("the {T} cost was not paid")
	}
}

// --- Oscorp Industries -----------------------------------------------------

func spm5OscorpGame(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Oscorp Industries", "Land", "", spm5OscorpIndustriesOracle)
	return g, me, id
}

func spm5Life(g *game.Game, p *game.Player) int {
	var n int
	g.WithWriteLock(func() { n = p.Life })
	return n
}

func TestSpm5OscorpIndustriesPlayedFromHandEntersTappedWithoutLifeLoss(t *testing.T) {
	g, me, id := spm5OscorpGame(t)
	before := spm5Life(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("play from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, id, "Oscorp Industries")
	if got := spm5Life(g, me); got != before {
		t.Errorf("life %d -> %d, want no change for a land played from hand", before, got)
	}
}

func TestSpm5OscorpIndustriesMayhemPlaysFromTheGraveyardAndCostsTwoLife(t *testing.T) {
	g, me, id := spm5OscorpGame(t)
	mayhemDiscard(t, g, me, id)
	before := spm5Life(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("mayhem play from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, id, "Oscorp Industries")
	if got := spm5Life(g, me); got != before-2 {
		t.Errorf("life %d -> %d, want a loss of 2 (entered from a graveyard)", before, got)
	}
}

func TestSpm5OscorpIndustriesNeedsToHaveBeenDiscardedThisTurn(t *testing.T) {
	g, me, id := spm5OscorpGame(t)
	// In the graveyard without being discarded (milled).
	g.WithWriteLock(func() {
		c, err := me.Hand.Remove(id)
		if err != nil {
			t.Fatalf("card not in hand: %v", err)
		}
		me.Graveyard.PushTop(c)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("played from the graveyard without having been discarded")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a refused play moved the card")
	}
}

func TestSpm5OscorpIndustriesTapsForUBOrR(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Oscorp Industries", "Land", spm5OscorpIndustriesOracle)
	spm5TapForEach(t, g, me, land, 3, []string{"U", "B", "R"})
}

// --- Oscorp Research Team ---------------------------------------------------

func TestSpm5OscorpResearchTeamDrawsTwoForSevenMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	team := b12Push(g, me.ID, "Oscorp Research Team", "Creature — Human Scientist", spm5ResearchTeamOracle, 1, 5)
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()

	b06AddMana(me, "U", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, team, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with six mana")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, team, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2", before, got)
	}
}

// --- Passenger Ferry ---------------------------------------------------------

func spm5FerryAttack(t *testing.T, withOther bool) (s crewScene, ferry, other uuid.UUID) {
	t.Helper()
	s = newCrewScene(t)
	ferry = s.pushVehicle("Passenger Ferry", spm5PassengerFerryOracle, 4, 3)
	crewer := s.pushCreature("Crewer", "Creature — Human", 2, 2)
	if withOther {
		other = s.pushCreature("Other Attacker", "Creature — Human", 2, 2)
	}
	s.crewAndResolve(t, ferry, 0, crewer)
	attackers := []uuid.UUID{ferry}
	if withOther {
		attackers = append(attackers, other)
	}
	declareAttack(t, s.g, s.opp.ID, attackers...)
	// The pool empties at every step change, so the {U} is floated now.
	b06AddMana(s.me, "U")
	passPriorityAroundTable(t, s.g)
	return s, ferry, other
}

func TestSpm5PassengerFerryPaysUToMakeAnotherAttackerUnblockable(t *testing.T) {
	s, ferry, other := spm5FerryAttack(t, true)
	answerPayUnless(t, s.g, s.me.ID, true)
	pick := pickTargetPrompt(t, s.g)
	if slices.Contains(pick.PickTargetCards, ferry) {
		t.Error("the Ferry offered itself: the clause says another creature")
	}
	if !slices.Contains(pick.PickTargetCards, other) {
		t.Fatalf("the other attacker was not offered: %v", pick.PickTargetCards)
	}
	answerPickTarget(t, s.g, other)
	passPriorityAroundTable(t, s.g)

	assertRestrictions(t, s.g, other, game.CantBeBlocked)
	if r := restrictionsOf(t, s.g, ferry); r != 0 {
		t.Errorf("the Ferry itself got %v", r.Names())
	}
	if len(poolColors(s.me)) != 0 {
		t.Errorf("pool = %v after paying {U}", poolColors(s.me))
	}
}

func TestSpm5PassengerFerryDecliningDoesNothing(t *testing.T) {
	s, _, other := spm5FerryAttack(t, true)
	answerPayUnless(t, s.g, s.me.ID, false)
	passPriorityAroundTable(t, s.g)
	if r := restrictionsOf(t, s.g, other); r != 0 {
		t.Errorf("restrictions = %v after declining", r.Names())
	}
	if len(poolColors(s.me)) != 1 {
		t.Errorf("the {U} was spent anyway: pool = %v", poolColors(s.me))
	}
}

func TestSpm5PassengerFerryAloneHasNoTargetAndTheTriggerIsRemoved(t *testing.T) {
	s, ferry, _ := spm5FerryAttack(t, false)
	answerPayUnless(t, s.g, s.me.ID, true)
	if latestPickTarget(s.g, s.me.ID) != nil {
		t.Fatal("a target prompt was raised with no other attacking creature")
	}
	passPriorityAroundTable(t, s.g)
	if r := restrictionsOf(t, s.g, ferry); r != 0 {
		t.Errorf("the Ferry made itself unblockable: %v", r.Names())
	}
}

func TestSpm5PassengerFerryCrewsAtTwoPower(t *testing.T) {
	s := newCrewScene(t)
	ferry := s.pushVehicle("Passenger Ferry", spm5PassengerFerryOracle, 4, 3)
	small := s.pushCreature("Small", "Creature — Human", 1, 1)
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, ferry, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crewed with total power 1")
	}
}

// --- Peter Parker // Amazing Spider-Man ---------------------------------------

func spm5PeterParkerCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   peterParkerOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name: "Peter Parker", TypeLine: "Legendary Creature — Human Scientist Hero",
				ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 0, Toughness: 1,
			},
			{
				Name: "Amazing Spider-Man", TypeLine: "Legendary Creature — Spider Human Hero",
				ManaCost: "{1}{G}{W}{U}", Colors: []string{"G", "W", "U"}, Power: 4, Toughness: 4,
			},
		},
	}
	c.SetFace(0)
	return c
}

func TestSpm5PeterParkerMakesATwoOneReachSpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	peter := spm5PeterParkerCard(me.ID)
	me.Hand.PushTop(peter)
	spm5CastResolve(t, g, me, peter.InstanceID, "CW")
	passPriorityAroundTable(t, g)

	toks := spm5Tokens(g, me.ID, "Spider")
	if len(toks) != 1 {
		t.Fatalf("%d Spider tokens, want 1", len(toks))
	}
	if c := toks[0]; c.Power != 2 || c.Toughness != 1 || !c.HasColor("G") || !hasEffectiveKeyword(t, g, c.InstanceID, "reach") {
		t.Errorf("token = %+v, want a 2/1 green Spider with reach", c)
	}
}

func TestSpm5PeterParkerTransformsOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	peter := spm5PeterParkerCard(me.ID)
	g.Battlefield.PushTop(peter)

	// Not enough mana, then enough mana at instant speed.
	b06AddMana(me, "C", "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with three mana")
	}
	b06AddMana(me, "U")
	advanceTo(t, g, game.StepBeginCombat)
	b06AddMana(me, "C", "G", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}

	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	advanceTo(t, g, game.StepPostcombatMain)
	b06AddMana(me, "C", "G", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform in a main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, peter.InstanceID)
	if card == nil || card.ActiveFace != 1 || card.Name != "Amazing Spider-Man" {
		t.Fatalf("after the transform: %+v, want Amazing Spider-Man in place", card)
	}
	if !hasEffectiveKeyword(t, g, peter.InstanceID, "vigilance") || !hasEffectiveKeyword(t, g, peter.InstanceID, "reach") {
		t.Error("Amazing Spider-Man lacks vigilance or reach")
	}
}

func spm5AmazingSpiderMan(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := spm5PeterParkerCard(owner)
	c.SetFace(1)
	return apaPush(g, owner, owner, c)
}

func spm5Legend(p *game.Player, name, typeLine string, colors []string) uuid.UUID {
	return spm5Hand(p, game.Card{Name: name, TypeLine: typeLine, ManaCost: "{1}{G}", Colors: colors, Power: 2, Toughness: 2})
}

func TestSpm5AmazingSpiderManGrantsWebSlingingToColoredLegendarySpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spm5AmazingSpiderMan(g, me.ID)
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	plain := spm5Legend(me, "Plain Elf", "Creature — Elf", []string{"G"})
	colorless := spm5Hand(me, game.Card{Name: "Legendary Relic", TypeLine: "Legendary Artifact", ManaCost: "{3}"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)

	o, ok := handOffers(t, g, me, legend)[GrantedAltCostWebSlinging]
	if !ok {
		t.Fatal("no granted web-slinging offer on a colored legendary spell")
	}
	if !strings.Contains(o.Label, "Web-slinging {G}{W}{U}") {
		t.Errorf("label = %q", o.Label)
	}
	if o.PayOptions == nil || !slices.Contains(o.PayOptions.Cards, tapped.String()) ||
		slices.Contains(o.PayOptions.Cards, untapped.String()) {
		t.Errorf("return options = %+v, want the tapped creature only", o.PayOptions)
	}
	if _, ok := handOffers(t, g, me, plain)[GrantedAltCostWebSlinging]; ok {
		t.Error("a nonlegendary spell was offered web-slinging")
	}
	if _, ok := handOffers(t, g, me, colorless)[GrantedAltCostWebSlinging]; ok {
		t.Error("a colorless legendary spell was offered web-slinging")
	}
}

func TestSpm5AmazingSpiderManWebSlingingCastReturnsATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spm5AmazingSpiderMan(g, me.ID)
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	cast := func(creature uuid.UUID) error {
		return g.CastSpell(me.ID, legend, game.CastSpellParams{
			Strict: true, AlternativeCost: GrantedAltCostWebSlinging, AltCostIDs: []uuid.UUID{creature},
		})
	}

	floatForTest(g, me, "GWU")
	if err := cast(untapped); err == nil {
		t.Fatal("web-slinging paid with an untapped creature")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("a refused payment moved the creature")
	}
	if err := cast(tapped); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to its owner's hand")
	}
	item := g.StackMeta[legend]
	if item == nil || !PaidWebSlinging(item.AltCost) {
		t.Fatalf("stack item = %+v, want a web-slinging cost", item)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(legend) {
		t.Error("the legendary creature did not resolve")
	}
}

func TestSpm5AmazingSpiderManGrantEndsWhenItLeavesAndOnlyForItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	spider := spm5AmazingSpiderMan(g, me.ID)
	mine := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	theirs := spm5Legend(opp, "Their Elf", "Legendary Creature — Elf", []string{"G"})

	if _, ok := handOffers(t, g, opp, theirs)[GrantedAltCostWebSlinging]; ok {
		t.Error("an opponent's spell was offered web-slinging")
	}
	g.WithWriteLock(func() {
		c, err := g.Battlefield.Remove(spider)
		if err != nil {
			t.Fatalf("Amazing Spider-Man is gone: %v", err)
		}
		me.Graveyard.PushTop(c)
		g.RecomputeLayersIfStaleLocked()
	})
	if _, ok := handOffers(t, g, me, mine)[GrantedAltCostWebSlinging]; ok {
		t.Error("web-slinging is still offered after Amazing Spider-Man left")
	}
}

func TestSpm5PeterParkerFrontFaceGrantsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	g.Battlefield.PushTop(spm5PeterParkerCard(me.ID))
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	if _, ok := handOffers(t, g, me, legend)[GrantedAltCostWebSlinging]; ok {
		t.Error("Peter Parker's front face granted web-slinging")
	}
}

// --- Peter Parker's Camera -------------------------------------------------------

func TestSpm5CameraEntersWithThreeFilmCountersAndCopiesAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	camera := spm5Hand(me, game.Card{Name: "Peter Parker's Camera", TypeLine: "Artifact", ManaCost: "{1}", OracleID: spm5CameraOracle})
	spm5CastResolve(t, g, me, camera, "C")
	if n := cardByID(g, camera).Counters["film"]; n != 3 {
		t.Fatalf("film counters = %d, want 3", n)
	}

	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "my trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "my activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := acItemLabelled(g, "my trigger")
	activation := acItemLabelled(g, "my activation")
	notMine := acItemLabelled(g, "their trigger")

	offered := acLegalAbilityTargets(g, me.ID, spm5CameraOracle, 0)
	if !offered[mine] || !offered[activation] {
		t.Errorf("your triggered and activated abilities must be offered: %v", offered)
	}
	if offered[notMine] {
		t.Error("an opponent's ability is not 'you control'")
	}

	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, camera, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: notMine}},
	}); err == nil {
		t.Fatal("the Camera targeted an opponent's ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, camera, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if n := cardByID(g, camera).Counters["film"]; n != 2 {
		t.Errorf("film counters = %d after the activation, want 2 (paid as a cost)", n)
	}
	if !cardByID(g, camera).Tapped {
		t.Error("the Camera is untapped")
	}
	acPassUntilCopy(t, g)
	if n := acCopiesOnStack(g); n != 1 {
		t.Fatalf("%d copies on the stack, want 1", n)
	}
	if cp := acCopyOnStack(g); cp == nil || cp.Kind != game.StackItemTriggered || cp.Controller != me.ID {
		t.Errorf("copy = %+v, want a triggered ability under your control", cp)
	}
}

func TestSpm5CameraWithNoFilmCountersCannotBeActivated(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	camera := pushCatalogPermanent(g, me.ID, "Peter Parker's Camera", "Artifact", spm5CameraOracle, false)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "my trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := acItemLabelled(g, "my trigger")
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, camera, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err == nil {
		t.Fatal("activated with no film counter")
	}
	if cardByID(g, camera).Tapped {
		t.Error("a refused activation tapped the Camera")
	}
}

// --- Pictures of Spider-Man ---------------------------------------------------

func spm5Library(me *game.Player, cards ...game.Card) []uuid.UUID {
	me.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = me.ID, me.ID
		ids[i] = cards[i].InstanceID
		me.Library.PushTop(cards[i])
	}
	return ids
}

func spm5PicturesTable(t *testing.T) (g *game.Game, me *game.Player, lib []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	toMainForCost(t, g)
	// Top to bottom: creature, land, creature, creature, instant | creature.
	lib = spm5Library(me,
		game.Card{Name: "Creature A", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Creature B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Creature C", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Creature D", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	pics := spm5Hand(me, game.Card{Name: "Pictures of Spider-Man", TypeLine: "Artifact", ManaCost: "{2}{G}",
		OracleID: spm5PicturesOracle, Colors: []string{"G"}})
	spm5CastResolve(t, g, me, pics, "CCG")
	passPriorityAroundTable(t, g)
	return g, me, lib
}

func TestSpm5PicturesTakesUpToTwoCreatureCardsAndBottomsTheRest(t *testing.T) {
	g, me, lib := spm5PicturesTable(t)
	a, forest, b, c, bolt, d := lib[0], lib[1], lib[2], lib[3], lib[4], lib[5]
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no choose-cards prompt")
	}

	// Three is over the ceiling; a land is not a creature card.
	pick := chooseCardsChoiceFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b, c}); err == nil {
		t.Fatal("three cards accepted for 'up to two'")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{forest}); err == nil {
		t.Fatal("a land accepted as a creature card")
	}
	answerChooseCards(t, g, me.ID, a, c)

	if !me.Hand.Contains(a) || !me.Hand.Contains(c) {
		t.Error("the chosen creature cards are not in hand")
	}
	if me.Hand.Contains(b) {
		t.Error("an unchosen creature card went to hand")
	}
	for _, id := range []uuid.UUID{forest, b, bolt} {
		if !me.Library.Contains(id) {
			t.Errorf("an unchosen looked-at card left the library")
		}
	}
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != d {
		t.Errorf("library top = %v, want the sixth card (the five looked-at cards go to the bottom)", top.Name)
	}
}

func TestSpm5PicturesMayTakeNone(t *testing.T) {
	g, me, lib := spm5PicturesTable(t)
	answerChooseCards(t, g, me.ID)
	for _, id := range lib[:5] {
		if me.Hand.Contains(id) {
			t.Error("a card went to hand on an empty pick")
		}
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != lib[5] {
		t.Error("the looked-at cards were not all put on the bottom")
	}
}

func TestSpm5PicturesSacrificesForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pics := pushCatalogPermanent(g, me.ID, "Pictures of Spider-Man", "Artifact", spm5PicturesOracle, false)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, pics, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, pics, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pics) {
		t.Error("Pictures was not sacrificed")
	}
	if len(spm5Tokens(g, me.ID, "Treasure")) != 1 {
		t.Error("no Treasure token")
	}
}

// --- Prison Break -----------------------------------------------------------------

func spm5PrisonBreakTable(t *testing.T) (g *game.Game, me *game.Player, bear, prison uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	toMainForCost(t, g)
	bear = uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Fallen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	prison = spm5Hand(me, game.Card{Name: "Prison Break", TypeLine: "Sorcery", ManaCost: "{4}{B}",
		OracleID: spm5PrisonBreakOracle, Colors: []string{"B"}})
	return g, me, bear, prison
}

func spm5CastPrison(g *game.Game, me *game.Player, prison, target uuid.UUID, params game.CastSpellParams) error {
	params.Strict = true
	params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	return g.CastSpell(me.ID, prison, params)
}

func TestSpm5PrisonBreakReturnsACreatureWithAnExtraCounter(t *testing.T) {
	g, me, bear, prison := spm5PrisonBreakTable(t)
	floatForTest(g, me, "CCCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return")
	}
	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if got := effectivePowerWithCounters(t, g, bear); got != 3 {
		t.Errorf("power = %d, want 3 (2/2 with a counter)", got)
	}
}

func TestSpm5PrisonBreakRefusesANonCreatureAndAnOpponentsGraveyard(t *testing.T) {
	g, me, _, prison := spm5PrisonBreakTable(t)
	opp := g.Seats[1]
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID})
	floatForTest(g, me, "CCCCB")
	if err := spm5CastPrison(g, me, prison, land, game.CastSpellParams{}); err == nil {
		t.Error("targeted a land card")
	}
	if err := spm5CastPrison(g, me, prison, theirs, game.CastSpellParams{}); err == nil {
		t.Error("targeted a creature card in an opponent's graveyard")
	}
	if !me.Hand.Contains(prison) {
		t.Error("a refused cast moved Prison Break")
	}
}

func TestSpm5PrisonBreakMayhemCastsFromTheGraveyardForFour(t *testing.T) {
	g, me, bear, prison := spm5PrisonBreakTable(t)

	// Not discarded: the graveyard is shut.
	g.WithWriteLock(func() {
		c, _ := me.Hand.Remove(prison)
		me.Graveyard.PushTop(c)
	})
	floatForTest(g, me, "CCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("cast from the graveyard without having been discarded")
	}

	// Discarded this turn: {3}{B}.
	g.WithWriteLock(func() {
		c, _ := me.Graveyard.Remove(prison)
		me.Hand.PushTop(c)
		me.ManaPool.EmptyPool()
	})
	mayhemDiscard(t, g, me, prison)
	floatForTest(g, me, "CCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	if item := g.StackMeta[prison]; item == nil || item.AltCost != "mayhem" {
		t.Errorf("stack item = %+v, want AltCost mayhem", item)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || hobA8Counters(t, g, bear) != 1 {
		t.Error("the mayhem cast did not reanimate the creature with a counter")
	}
}

// --- Professional Wrestler ----------------------------------------------------------

func TestSpm5ProfessionalWrestlerMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := spm5Hand(me, game.Card{Name: "Professional Wrestler", TypeLine: "Creature — Human Warrior Performer",
		ManaCost: "{3}{G}", OracleID: spm5WrestlerOracle, Colors: []string{"G"}, Power: 4, Toughness: 4})
	spm5CastResolve(t, g, me, id, "CCCG")
	passPriorityAroundTable(t, g)
	if len(spm5Tokens(g, me.ID, "Treasure")) != 1 {
		t.Error("no Treasure token")
	}
}

func TestSpm5ProfessionalWrestlerCannotBeBlockedByMoreThanOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wrestler := b12Push(g, me.ID, "Professional Wrestler", "Creature — Human Warrior Performer", spm5WrestlerOracle, 4, 4)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttack(t, g, wrestler)

	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: wrestler}, {Blocker: b, Attacker: wrestler},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlocker(a, wrestler); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
	brRefusal(t, g.DeclareBlocker(b, wrestler), game.BlockReasonTooManyBlockers)
}
