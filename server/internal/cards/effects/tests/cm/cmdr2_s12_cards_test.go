package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s12_cards_test.go — batch 12 (#5368): Elementals, evoke and
// reveal-a-card lands.

const (
	s12SpectacleOracle = "dd963396-3f81-49ad-a68e-87b9764d7ac5"
	s12ChewerOracle    = "127795e6-3f0e-4318-95ad-ccb1912580ac"
	s12ShriekmawOracle = "717bbb9d-0a5b-4979-a24b-9d54aac7657a"
	s12VernalOracle    = "bf30b8ae-0125-4a00-bae8-ac1aa7703a08"
	s12SoulstokeOracle = "934bd9f9-937b-4c6b-a849-012ae6bd3321"
	s12HordeOracle     = "a7656948-4d7d-47f7-b8a1-fc9716155e37"
	s12EndRazeOracle   = "69b52907-8aff-4f2a-a391-7d1ab2669b5c"
	s12SmokeOracle     = "e970f2df-90cb-45b8-8ae5-71b762908925"
	s12BeyondOracle    = "541744d9-449d-420a-a5a1-2fffba18450f"
	s12VillageOracle   = "34a1eb04-08f6-49d8-a1d1-b987a76bd8b1"
	s12GlenOracle      = "09f52275-99e6-45e0-b2db-cafe26d5fb91"
)

func TestCmdr2S12ElementalSpectacleMakesATokenPerColorAndGainsLifePerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	effectstest.B27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Elemental Spectacle", "Sorcery", s12SpectacleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elemental" {
			tokens++
			if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 5 || tt != 5 {
				t.Errorf("token is %d/%d, want 5/5", p, tt)
			}
		}
	}
	if tokens != 2 {
		t.Fatalf("made %d tokens, want 2 (red and blue among permanents)", tokens)
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != 4 {
		t.Errorf("gained %d life, want 4 (two bears and two tokens)", got)
	}
}

func TestCmdr2S12ElementalSpectacleWithNoColorsStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.B12Creature(g, me.ID, "Colorless Golem", "Artifact Creature — Golem", 1, 1)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Elemental Spectacle", "Sorcery", s12SpectacleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" {
			t.Fatal("a token was made with no colors among permanents")
		}
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != 1 {
		t.Errorf("gained %d life, want 1", got)
	}
}

func TestCmdr2S12IngotChewerEvokedDestroysAnArtifactAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	id := effectstest.CastWithAltCost(t, g, "Ingot Chewer", "Creature — Elemental", s12ChewerOracle, "evoke")
	effectstest.S12Drive(t, g, me.ID, rock)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the evoked Chewer was not sacrificed")
	}
}

func TestCmdr2S12IngotChewerHardCastStaysAndHitsOnlyArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	id := effectstest.CastCatalogSpell(t, g, "Ingot Chewer", "Creature — Elemental", s12ChewerOracle, nil)
	effectstest.S12Drive(t, g, me.ID, rock)
	if g.Battlefield.Contains(rock) || !g.Battlefield.Contains(bear) {
		t.Error("expected the artifact destroyed and the bear untouched")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("a hard-cast Chewer was sacrificed")
	}
}

func TestCmdr2S12ShriekmawDestroysNonblackNonartifactOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	black := effectstest.B27Push(g, opp.ID, "Black Bear", "Creature — Bear", "", "{1}{B}", 2, 2, "B")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	green := effectstest.B27Push(g, opp.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	id := effectstest.CastWithAltCost(t, g, "Shriekmaw", "Creature — Elemental", s12ShriekmawOracle, "evoke")
	effectstest.S12Drive(t, g, me.ID, green, black, golem)
	if g.Battlefield.Contains(green) {
		t.Error("the green bear survived")
	}
	if !g.Battlefield.Contains(black) || !g.Battlefield.Contains(golem) {
		t.Error("an illegal target was destroyed")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the evoked Shriekmaw stayed")
	}
}

func TestCmdr2S12BatchIsRegisteredFull(t *testing.T) {
	for _, id := range []string{s12SpectacleOracle, s12ChewerOracle, s12ShriekmawOracle, s12VernalOracle,
		s12SoulstokeOracle, s12HordeOracle, s12EndRazeOracle, s12SmokeOracle, s12BeyondOracle,
		s12VillageOracle, s12GlenOracle} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: spec = %+v, %v", id, spec, ok)
		}
	}
}

func TestCmdr2S12VernalSovereignMakesAScalingElementalOnEnterAndAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sov := effectstest.Ecl6Cast(t, g, me, "Vernal Sovereign", "Creature — Elemental Elk", s12VernalOracle, 4, 4, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var token = func() []game.Card {
		var out []game.Card
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Elemental" {
				out = append(out, c)
			}
		}
		return out
	}
	toks := token()
	if len(toks) != 1 {
		t.Fatalf("%d Elemental tokens after entering, want 1", len(toks))
	}
	// Bear, Sovereign and the token itself: 3/3.
	if p, tt := effectstest.PtOf(t, g, toks[0].InstanceID); p != 3 || tt != 3 {
		t.Errorf("token is %d/%d, want 3/3", p, tt)
	}
	effectstest.B12Creature(g, me.ID, "Another", "Creature — Bear", 1, 1)
	if p, tt := effectstest.PtOf(t, g, toks[0].InstanceID); p != 4 || tt != 4 {
		t.Errorf("token is %d/%d after another creature, want 4/4 (live count)", p, tt)
	}
	if !g.Battlefield.Contains(sov) {
		t.Fatal("Sovereign left")
	}
}

func TestCmdr2S12IncandescentSoulstokeAnthemAndPutFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	stoke := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Incandescent Soulstoke", TypeLine: "Creature — Elemental Shaman",
		OracleID: s12SoulstokeOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	other := effectstest.B12Creature(g, me.ID, "Other Elemental", "Creature — Elemental", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if p, _ := effectstest.PtOf(t, g, other); p != 2 {
		t.Errorf("another Elemental has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 1 {
		t.Errorf("a Bear has power %d, want 1", p)
	}
	if p, _ := effectstest.PtOf(t, g, stoke); p != 2 {
		t.Errorf("Soulstoke has power %d, want 2 (other Elementals only)", p)
	}
	elem := effectstest.RfdHandCard(me, "Hand Elemental", "Creature — Elemental", "", "", 3, 3)
	notElem := effectstest.RfdHandCard(me, "Hand Bear", "Creature — Bear", "", "", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, stoke, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the Elemental card")
	}
	if effectstest.HasID(pick.ChooseCards, notElem) || !effectstest.HasID(pick.ChooseCards, elem) {
		t.Errorf("offered %v: want only the Elemental card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elem) {
		t.Fatal("the Elemental did not enter")
	}
	if !effectstest.HasEffectiveKeyword(t, g, elem, "haste") {
		t.Error("the Elemental did not gain haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elem) {
		t.Error("the Elemental was not sacrificed at the next end step")
	}
}

func TestCmdr2S12HordeOfNotionsPlaysAnElementalFromTheGraveyardFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	horde := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Horde of Notions", TypeLine: "Legendary Creature — Elemental",
		OracleID: s12HordeOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	elem := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: elem, Name: "Dead Elemental", TypeLine: "Creature — Elemental",
		ManaCost: "{4}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, horde, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a non-Elemental card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, horde, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elem}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	s12Resolve(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, elem)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, elem, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elem) {
		t.Error("the Elemental did not come back from the graveyard")
	}
}

func TestCmdr2S12EndRazeForerunnersPumpsOthersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	boar := effectstest.Ecl6Cast(t, g, me, "End-Raze Forerunners", "Creature — Boar", s12EndRazeOracle, 7, 7, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("my Bear is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("my Bear lacks vigilance or trample")
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, boar); p != 7 || tt != 7 {
		t.Errorf("the Boar is %d/%d, want 7/7 (other creatures only)", p, tt)
	}
}

func TestCmdr2S12SmokebraiderManaPaysOnlyForElementals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	braider := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Smokebraider", OracleID: s12SmokeOracle,
		TypeLine: "Creature — Elemental Shaman", Power: 1, Toughness: 1})
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, braider, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R"); got != 2 {
		t.Fatalf("%d colour picks, want 2", got)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool holds %d mana, want 2", got)
	}
	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{R}{R}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var short *game.InsufficientManaError
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Bear with Smokebraider's mana: err = %v, want InsufficientManaError", err)
	}
	elem := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: elem, Name: "Elemental", TypeLine: "Creature — Elemental", ManaCost: "{R}{R}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, elem, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an Elemental with Smokebraider's mana: %v", err)
	}
}

func TestCmdr2S12PrimalBeyondTapsForColorlessAndRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.HandCard(me, "Elemental In Hand", "Creature — Elemental")
	land := effectstest.PlayLandFromHand(t, g, "Primal Beyond", s12BeyondOracle)
	effectstest.AnswerEntryReveal(t, g, me.ID, func() uuid.UUID { return effectstest.EntryRevealChoiceFor(g, me.ID).ChooseCards[0] }())
	if tapped, _ := effectstest.BattlefieldCardTapped(g, land); tapped {
		t.Error("the land entered tapped after revealing an Elemental")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for coloured mana: %v", err)
	}
	effectstest.AshlingPickColor(t, g, me.ID, "W")
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	var short *game.InsufficientManaError
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Bear with the restricted mana: err = %v, want InsufficientManaError", err)
	}
}

func TestCmdr2S12FlamekinVillageRevealsAnElemental(t *testing.T) {
	effectstest.S12RevealLandCase(t, "Flamekin Village", s12VillageOracle, "Creature — Elemental", "Creature — Faerie")
}

func TestCmdr2S12SecludedGlenRevealsAFaerie(t *testing.T) {
	effectstest.S12RevealLandCase(t, "Secluded Glen", s12GlenOracle, "Creature — Faerie", "Creature — Elemental")
}

func TestCmdr2S12FlamekinVillageGrantsHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	village := effectstest.PushCatalogPermanent(g, me.ID, "Flamekin Village", "Land", s12VillageOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, village, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear did not gain haste")
	}
}

// s12Resolve passes priority until the stack empties or a prompt opens,
// leaving the prompt for the caller.
func s12Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16 && !effectstest.StackFullyEmpty(g); i++ {
		if len(g.PendingChoices) > 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) || len(g.PendingChoices) > 0 {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
