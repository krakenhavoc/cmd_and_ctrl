package testsso

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsp "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/p"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func spm5CastResolve(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string) {
	t.Helper()
	effectstest.FloatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
		"Peter Parker":           cardsp.PeterParkerOracle,
		"Amazing Spider-Man":     cardsp.PeterParkerOracle + "#1",
		"Peter Parker's Camera":  spm5CameraOracle,
		"Pictures of Spider-Man": spm5PicturesOracle,
		"Prison Break":           spm5PrisonBreakOracle,
		"Professional Wrestler":  spm5WrestlerOracle,
	} {
		effectstest.RequireFull(t, oracle, name)
	}
}

// --- News Helicopter ---------------------------------------------------

func TestSpm5NewsHelicopterFliesAndMakesAHumanCitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.Spm5Hand(me, game.Card{Name: "News Helicopter", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{3}", OracleID: spm5NewsHelicopterOracle, Power: 1, Toughness: 1})
	spm5CastResolve(t, g, me, id, "CCC")
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Ominous Asylum", spm5OminousAsylumOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Ominous Asylum")

	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
	effectstest.Spm5TapForEach(t, g, me, id, 2, []string{"B", "R"})
}

func TestSpm5OminousAsylumSurveilsForFourAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Next Draw")
	land := effectstest.PushCatalogPermanent(g, me.ID, "Ominous Asylum", "Land", spm5OminousAsylumOracle, false)
	effectstest.AdvanceToMain(t, g)

	// Three mana is not enough.
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated with three mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil prompt")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the {T} cost was not paid")
	}
}

// --- Oscorp Industries -----------------------------------------------------

func spm5OscorpGame(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.MayhemHandCard(g, me, "Oscorp Industries", "Land", "", spm5OscorpIndustriesOracle)
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Oscorp Industries")
	if got := spm5Life(g, me); got != before {
		t.Errorf("life %d -> %d, want no change for a land played from hand", before, got)
	}
}

func TestSpm5OscorpIndustriesMayhemPlaysFromTheGraveyardAndCostsTwoLife(t *testing.T) {
	g, me, id := spm5OscorpGame(t)
	effectstest.MayhemDiscard(t, g, me, id)
	before := spm5Life(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("mayhem play from graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Oscorp Industries")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Oscorp Industries", "Land", spm5OscorpIndustriesOracle)
	effectstest.Spm5TapForEach(t, g, me, land, 3, []string{"U", "B", "R"})
}

// --- Oscorp Research Team ---------------------------------------------------

func TestSpm5OscorpResearchTeamDrawsTwoForSevenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	team := effectstest.B12Push(g, me.ID, "Oscorp Research Team", "Creature — Human Scientist", spm5ResearchTeamOracle, 1, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()

	effectstest.B06AddMana(me, "U", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, team, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with six mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, team, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2", before, got)
	}
}

// --- Passenger Ferry ---------------------------------------------------------

func spm5FerryAttack(t *testing.T, withOther bool) (s effectstest.CrewScene, ferry, other uuid.UUID) {
	t.Helper()
	s = effectstest.NewCrewScene(t)
	ferry = s.PushVehicle("Passenger Ferry", spm5PassengerFerryOracle, 4, 3)
	crewer := s.PushCreature("Crewer", "Creature — Human", 2, 2)
	if withOther {
		other = s.PushCreature("Other Attacker", "Creature — Human", 2, 2)
	}
	s.CrewAndResolve(t, ferry, 0, crewer)
	attackers := []uuid.UUID{ferry}
	if withOther {
		attackers = append(attackers, other)
	}
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, attackers...)
	// The pool empties at every step change, so the {U} is floated now.
	effectstest.B06AddMana(s.Me, "U")
	effectstest.PassPriorityAroundTable(t, s.G)
	return s, ferry, other
}

func TestSpm5PassengerFerryPaysUToMakeAnotherAttackerUnblockable(t *testing.T) {
	s, ferry, other := spm5FerryAttack(t, true)
	effectstest.AnswerPayUnless(t, s.G, s.Me.ID, true)
	pick := effectstest.PickTargetPrompt(t, s.G)
	if slices.Contains(pick.PickTargetCards, ferry) {
		t.Error("the Ferry offered itself: the clause says another creature")
	}
	if !slices.Contains(pick.PickTargetCards, other) {
		t.Fatalf("the other attacker was not offered: %v", pick.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, s.G, other)
	effectstest.PassPriorityAroundTable(t, s.G)

	effectstest.AssertRestrictions(t, s.G, other, game.CantBeBlocked)
	if r := effectstest.RestrictionsOf(t, s.G, ferry); r != 0 {
		t.Errorf("the Ferry itself got %v", r.Names())
	}
	if len(effectstest.PoolColors(s.Me)) != 0 {
		t.Errorf("pool = %v after paying {U}", effectstest.PoolColors(s.Me))
	}
}

func TestSpm5PassengerFerryDecliningDoesNothing(t *testing.T) {
	s, _, other := spm5FerryAttack(t, true)
	effectstest.AnswerPayUnless(t, s.G, s.Me.ID, false)
	effectstest.PassPriorityAroundTable(t, s.G)
	if r := effectstest.RestrictionsOf(t, s.G, other); r != 0 {
		t.Errorf("restrictions = %v after declining", r.Names())
	}
	if len(effectstest.PoolColors(s.Me)) != 1 {
		t.Errorf("the {U} was spent anyway: pool = %v", effectstest.PoolColors(s.Me))
	}
}

func TestSpm5PassengerFerryAloneHasNoTargetAndTheTriggerIsRemoved(t *testing.T) {
	s, ferry, _ := spm5FerryAttack(t, false)
	effectstest.AnswerPayUnless(t, s.G, s.Me.ID, true)
	if effectstest.LatestPickTarget(s.G, s.Me.ID) != nil {
		t.Fatal("a target prompt was raised with no other attacking creature")
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if r := effectstest.RestrictionsOf(t, s.G, ferry); r != 0 {
		t.Errorf("the Ferry made itself unblockable: %v", r.Names())
	}
}

func TestSpm5PassengerFerryCrewsAtTwoPower(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	ferry := s.PushVehicle("Passenger Ferry", spm5PassengerFerryOracle, 4, 3)
	small := s.PushCreature("Small", "Creature — Human", 1, 1)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, ferry, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crewed with total power 1")
	}
}

// --- Peter Parker // Amazing Spider-Man ---------------------------------------

func spm5PeterParkerCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cardsp.PeterParkerOracle,
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	peter := spm5PeterParkerCard(me.ID)
	me.Hand.PushTop(peter)
	spm5CastResolve(t, g, me, peter.InstanceID, "CW")
	effectstest.PassPriorityAroundTable(t, g)

	toks := spm5Tokens(g, me.ID, "Spider")
	if len(toks) != 1 {
		t.Fatalf("%d Spider tokens, want 1", len(toks))
	}
	if c := toks[0]; c.Power != 2 || c.Toughness != 1 || !c.HasColor("G") || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "reach") {
		t.Errorf("token = %+v, want a 2/1 green Spider with reach", c)
	}
}

func TestSpm5PeterParkerTransformsOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	peter := spm5PeterParkerCard(me.ID)
	g.Battlefield.PushTop(peter)

	// Not enough mana, then enough mana at instant speed.
	effectstest.B06AddMana(me, "C", "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with three mana")
	}
	effectstest.B06AddMana(me, "U")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "C", "G", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}

	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.B06AddMana(me, "C", "G", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, peter.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform in a main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card := effectstest.BattlefieldCardFor(g, peter.InstanceID)
	if card == nil || card.ActiveFace != 1 || card.Name != "Amazing Spider-Man" {
		t.Fatalf("after the transform: %+v, want Amazing Spider-Man in place", card)
	}
	if !effectstest.HasEffectiveKeyword(t, g, peter.InstanceID, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, peter.InstanceID, "reach") {
		t.Error("Amazing Spider-Man lacks vigilance or reach")
	}
}

func spm5AmazingSpiderMan(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := spm5PeterParkerCard(owner)
	c.SetFace(1)
	return effectstest.ApaPush(g, owner, owner, c)
}

func spm5Legend(p *game.Player, name, typeLine string, colors []string) uuid.UUID {
	return effectstest.Spm5Hand(p, game.Card{Name: name, TypeLine: typeLine, ManaCost: "{1}{G}", Colors: colors, Power: 2, Toughness: 2})
}

func TestSpm5AmazingSpiderManGrantsWebSlingingToColoredLegendarySpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	spm5AmazingSpiderMan(g, me.ID)
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	plain := spm5Legend(me, "Plain Elf", "Creature — Elf", []string{"G"})
	colorless := effectstest.Spm5Hand(me, game.Card{Name: "Legendary Relic", TypeLine: "Legendary Artifact", ManaCost: "{3}"})
	tapped := effectstest.PushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	untapped := effectstest.PushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)

	o, ok := effectstest.HandOffers(t, g, me, legend)[effects.GrantedAltCostWebSlinging]
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
	if _, ok := effectstest.HandOffers(t, g, me, plain)[effects.GrantedAltCostWebSlinging]; ok {
		t.Error("a nonlegendary spell was offered web-slinging")
	}
	if _, ok := effectstest.HandOffers(t, g, me, colorless)[effects.GrantedAltCostWebSlinging]; ok {
		t.Error("a colorless legendary spell was offered web-slinging")
	}
}

func TestSpm5AmazingSpiderManWebSlingingCastReturnsATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	spm5AmazingSpiderMan(g, me.ID)
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	tapped := effectstest.PushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	untapped := effectstest.PushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	cast := func(creature uuid.UUID) error {
		return g.CastSpell(me.ID, legend, game.CastSpellParams{
			Strict: true, AlternativeCost: effects.GrantedAltCostWebSlinging, AltCostIDs: []uuid.UUID{creature},
		})
	}

	effectstest.FloatForTest(g, me, "GWU")
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
	if item == nil || !effects.PaidWebSlinging(item.AltCost) {
		t.Fatalf("stack item = %+v, want a web-slinging cost", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(legend) {
		t.Error("the legendary creature did not resolve")
	}
}

func TestSpm5AmazingSpiderManGrantEndsWhenItLeavesAndOnlyForItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	spider := spm5AmazingSpiderMan(g, me.ID)
	mine := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	theirs := spm5Legend(opp, "Their Elf", "Legendary Creature — Elf", []string{"G"})

	if _, ok := effectstest.HandOffers(t, g, opp, theirs)[effects.GrantedAltCostWebSlinging]; ok {
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
	if _, ok := effectstest.HandOffers(t, g, me, mine)[effects.GrantedAltCostWebSlinging]; ok {
		t.Error("web-slinging is still offered after Amazing Spider-Man left")
	}
}

func TestSpm5PeterParkerFrontFaceGrantsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	g.Battlefield.PushTop(spm5PeterParkerCard(me.ID))
	legend := spm5Legend(me, "Legendary Elf", "Legendary Creature — Elf", []string{"G"})
	if _, ok := effectstest.HandOffers(t, g, me, legend)[effects.GrantedAltCostWebSlinging]; ok {
		t.Error("Peter Parker's front face granted web-slinging")
	}
}

// --- Peter Parker's Camera -------------------------------------------------------

func TestSpm5CameraEntersWithThreeFilmCountersAndCopiesAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	camera := effectstest.Spm5Hand(me, game.Card{Name: "Peter Parker's Camera", TypeLine: "Artifact", ManaCost: "{1}", OracleID: spm5CameraOracle})
	spm5CastResolve(t, g, me, camera, "C")
	if n := effectstest.CardByID(g, camera).Counters["film"]; n != 3 {
		t.Fatalf("film counters = %d, want 3", n)
	}

	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "my trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "my activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := effectstest.AcItemLabelled(g, "my trigger")
	activation := effectstest.AcItemLabelled(g, "my activation")
	notMine := effectstest.AcItemLabelled(g, "their trigger")

	offered := effectstest.AcLegalAbilityTargets(g, me.ID, spm5CameraOracle, 0)
	if !offered[mine] || !offered[activation] {
		t.Errorf("your triggered and activated abilities must be offered: %v", offered)
	}
	if offered[notMine] {
		t.Error("an opponent's ability is not 'you control'")
	}

	effectstest.FloatForTest(g, me, "CC")
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
	if n := effectstest.CardByID(g, camera).Counters["film"]; n != 2 {
		t.Errorf("film counters = %d after the activation, want 2 (paid as a cost)", n)
	}
	if !effectstest.CardByID(g, camera).Tapped {
		t.Error("the Camera is untapped")
	}
	effectstest.AcPassUntilCopy(t, g)
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Fatalf("%d copies on the stack, want 1", n)
	}
	if cp := effectstest.AcCopyOnStack(g); cp == nil || cp.Kind != game.StackItemTriggered || cp.Controller != me.ID {
		t.Errorf("copy = %+v, want a triggered ability under your control", cp)
	}
}

func TestSpm5CameraWithNoFilmCountersCannotBeActivated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	camera := effectstest.PushCatalogPermanent(g, me.ID, "Peter Parker's Camera", "Artifact", spm5CameraOracle, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "my trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := effectstest.AcItemLabelled(g, "my trigger")
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, camera, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err == nil {
		t.Fatal("activated with no film counter")
	}
	if effectstest.CardByID(g, camera).Tapped {
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
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.ToMainForCost(t, g)
	// Top to bottom: creature, land, creature, creature, instant | creature.
	lib = spm5Library(me,
		game.Card{Name: "Creature A", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Creature B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Creature C", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Creature D", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	pics := effectstest.Spm5Hand(me, game.Card{Name: "Pictures of Spider-Man", TypeLine: "Artifact", ManaCost: "{2}{G}",
		OracleID: spm5PicturesOracle, Colors: []string{"G"}})
	spm5CastResolve(t, g, me, pics, "CCG")
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, lib
}

func TestSpm5PicturesTakesUpToTwoCreatureCardsAndBottomsTheRest(t *testing.T) {
	g, me, lib := spm5PicturesTable(t)
	a, forest, b, c, bolt, d := lib[0], lib[1], lib[2], lib[3], lib[4], lib[5]
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no choose-cards prompt")
	}

	// Three is over the ceiling; a land is not a creature card.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b, c}); err == nil {
		t.Fatal("three cards accepted for 'up to two'")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{forest}); err == nil {
		t.Fatal("a land accepted as a creature card")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, c)

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
	effectstest.AnswerChooseCards(t, g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pics := effectstest.PushCatalogPermanent(g, me.ID, "Pictures of Spider-Man", "Artifact", spm5PicturesOracle, false)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, pics, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, pics, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear = uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Fallen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	prison = effectstest.Spm5Hand(me, game.Card{Name: "Prison Break", TypeLine: "Sorcery", ManaCost: "{4}{B}",
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
	effectstest.FloatForTest(g, me, "CCCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return")
	}
	if n := effectstest.HobA8Counters(t, g, bear); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, bear); got != 3 {
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
	effectstest.FloatForTest(g, me, "CCCCB")
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
	effectstest.FloatForTest(g, me, "CCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("cast from the graveyard without having been discarded")
	}

	// Discarded this turn: {3}{B}.
	g.WithWriteLock(func() {
		c, _ := me.Graveyard.Remove(prison)
		me.Hand.PushTop(c)
		me.ManaPool.EmptyPool()
	})
	effectstest.MayhemDiscard(t, g, me, prison)
	effectstest.FloatForTest(g, me, "CCCB")
	if err := spm5CastPrison(g, me, prison, bear, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	if item := g.StackMeta[prison]; item == nil || item.AltCost != "mayhem" {
		t.Errorf("stack item = %+v, want AltCost mayhem", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || effectstest.HobA8Counters(t, g, bear) != 1 {
		t.Error("the mayhem cast did not reanimate the creature with a counter")
	}
}

// --- Professional Wrestler ----------------------------------------------------------

func TestSpm5ProfessionalWrestlerMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.Spm5Hand(me, game.Card{Name: "Professional Wrestler", TypeLine: "Creature — Human Warrior Performer",
		ManaCost: "{3}{G}", OracleID: spm5WrestlerOracle, Colors: []string{"G"}, Power: 4, Toughness: 4})
	spm5CastResolve(t, g, me, id, "CCCG")
	effectstest.PassPriorityAroundTable(t, g)
	if len(spm5Tokens(g, me.ID, "Treasure")) != 1 {
		t.Error("no Treasure token")
	}
}

func TestSpm5ProfessionalWrestlerCannotBeBlockedByMoreThanOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wrestler := effectstest.B12Push(g, me.ID, "Professional Wrestler", "Creature — Human Warrior Performer", spm5WrestlerOracle, 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, wrestler)

	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: wrestler}, {Blocker: b, Attacker: wrestler},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlocker(a, wrestler); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(b, wrestler), game.BlockReasonTooManyBlockers)
}
