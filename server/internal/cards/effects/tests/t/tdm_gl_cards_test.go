package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_gl_cards_test.go — Tarkir: Dragonstorm, cards from G to L (#4463).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tdmGreatArashinCityOracle    = "f40f374b-acaf-459d-9ccd-b0b22d1a3f28"
	tdmGurmagRakshasaOracle      = "2be43acc-a815-44bc-8803-d702e553fbc1"
	tdmHardenedTacticianOracle   = "ca1dc6f0-3ec2-4f62-a08a-1395201bd93f"
	tdmHighspireBellRingerOracle = "6d2d250b-cdce-4075-8724-b52ad31664fb"
	tdmHumblingElderOracle       = "013bcab8-e527-4700-82c9-4cb9cb8a4807"
	tdmInevitableDefeatOracle    = "d68439b5-7011-4b1d-a7da-14ef0f4d9994"
	tdmIridescentTigerOracle     = "ce09c5b0-9dd4-4f73-9f3f-dfbd829813e4"
	tdmJadeCastSentinelOracle    = "985192cd-ceb6-48ad-80b9-e10fc544260e"
	tdmJeskaiBrushmasterOracle   = "eff690a7-c2c7-4750-894c-0667d3151165"
	tdmJeskaiShrinekeeperOracle  = "bb50b4f4-f0da-4283-821b-6d74bd20752f"
	tdmKinTreeSeveranceOracle    = "211d1c59-0111-48dd-9ff6-d03e7ff2ad5c"
	tdmKishlaVillageOracle       = "4eb428ab-f5b0-46ca-98dd-b3466a91ef97"
	tdmKrotiqNestguardOracle     = "33117d34-ea70-4566-92a3-5dec9bfe529a"
	tdmLightfootTechniqueOracle  = "601b7481-19fc-4779-950e-883a9a22aec3"
	tdmLoxodonBattlePriestOracle = "8cf4ddb0-2a83-4d8f-b636-7ba3c81ea4ca"
)

func TestTdmGLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmGreatArashinCityOracle:    "Great Arashin City",
		tdmGurmagRakshasaOracle:      "Gurmag Rakshasa",
		tdmHardenedTacticianOracle:   "Hardened Tactician",
		tdmHighspireBellRingerOracle: "Highspire Bell-Ringer",
		tdmHumblingElderOracle:       "Humbling Elder",
		tdmInevitableDefeatOracle:    "Inevitable Defeat",
		tdmIridescentTigerOracle:     "Iridescent Tiger",
		tdmJadeCastSentinelOracle:    "Jade-Cast Sentinel",
		tdmJeskaiBrushmasterOracle:   "Jeskai Brushmaster",
		tdmJeskaiShrinekeeperOracle:  "Jeskai Shrinekeeper",
		tdmKinTreeSeveranceOracle:    "Kin-Tree Severance",
		tdmKishlaVillageOracle:       "Kishla Village",
		tdmKrotiqNestguardOracle:     "Krotiq Nestguard",
		tdmLightfootTechniqueOracle:  "Lightfoot Technique",
		tdmLoxodonBattlePriestOracle: "Loxodon Battle Priest",
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

// --- fixtures ------------------------------------------------------------------

// tdmGLActivate activates a catalog row strictly, paying from the pool.
func tdmGLActivate(t *testing.T, g *game.Game, who, source uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(who, source, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

// tdmGLCheckland plays a village from hand beside a land of `landType` and
// reports whether it entered tapped.
func tdmGLCheckland(t *testing.T, name, oracle, landType string) bool {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Other Land", landType)
	return effectstest.B16Tapped(t, g, effectstest.PlayLandFromHand(t, g, name, oracle))
}

// tdmGLSpirits lists the Spirit tokens `who` controls.
func tdmGLSpirits(g *game.Game, who uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == who && effects.IsToken(c) && c.HasSubtype("Spirit") {
			out = append(out, c)
		}
	}
	return out
}

// --- Great Arashin City --------------------------------------------------------

func TestTdmGreatArashinCityEntersTappedUnlessYouControlAForestOrPlains(t *testing.T) {
	for _, tc := range []struct {
		land   string
		tapped bool
	}{
		{"Basic Land — Island", true},
		{"Basic Land — Forest", false},
		{"Basic Land — Plains", false},
	} {
		if got := tdmGLCheckland(t, "Great Arashin City", tdmGreatArashinCityOracle, tc.land); got != tc.tapped {
			t.Errorf("beside a %s: tapped = %v, want %v", tc.land, got, tc.tapped)
		}
	}
}

func TestTdmGreatArashinCityExilesACreatureCardForASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	city := effectstest.PushCatalogPermanent(g, me.ID, "Great Arashin City", "Land", tdmGreatArashinCityOracle, false)
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	tdmGLActivate(t, g, me.ID, city, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}})
	if !g.Exile.Contains(bear) || !me.Graveyard.Contains(spell) {
		t.Fatal("the payment did not exile exactly the creature card")
	}
	if !effectstest.B16Tapped(t, g, city) {
		t.Error("{T} is part of the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := tdmGLSpirits(g, me.ID)
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	s := spirits[0]
	if s.Power != 1 || s.Toughness != 1 || len(s.Colors) != 1 || s.Colors[0] != "W" || len(s.Keywords) != 0 {
		t.Errorf("Spirit %d/%d %v %v, want a plain 1/1 white Spirit", s.Power, s.Toughness, s.Colors, s.Keywords)
	}
}

// --- Gurmag Rakshasa -----------------------------------------------------------

func TestTdmGurmagRakshasaShrinksTheirCreatureAndPumpsYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	demon := effectstest.RfdCast(t, g, "Gurmag Rakshasa", "Creature — Demon", tdmGurmagRakshasaOracle, 5, 5)
	if !effectstest.HasEffectiveKeyword(t, g, demon, "menace") {
		t.Error("Gurmag Rakshasa has no menace")
	}
	for i := 0; i < 4; i++ {
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			break
		}
		switch {
		case effectstest.HasID(p.PickTargetCards, theirs):
			if effectstest.HasID(p.PickTargetCards, mine) {
				t.Fatal("one clause offers both players' creatures")
			}
			effectstest.PickCard(t, g, me.ID, theirs)
		case effectstest.HasID(p.PickTargetCards, mine):
			effectstest.PickCard(t, g, me.ID, mine)
		default:
			t.Fatalf("unexpected target prompt %v", p.PickTargetCards)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their creature is %d/%d, want 1/1 (-2/-2)", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("my creature is %d/%d, want 4/4 (+2/+2)", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, demon); p != 5 || tt != 5 {
		t.Errorf("the Rakshasa is %d/%d, want 5/5", p, tt)
	}
}

// --- Hardened Tactician --------------------------------------------------------

func TestTdmHardenedTacticianSacrificesATokenToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tactician := effectstest.B12Push(g, me.ID, "Hardened Tactician", "Creature — Human Warrior", tdmHardenedTacticianOracle, 2, 4)
	nonToken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	token := effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	effectstest.SeedLibrary(me, "Next")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tactician, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{nonToken}}); err == nil {
		t.Fatal("a nontoken creature paid \"sacrifice a token\"")
	}
	hand := len(me.Hand.Cards)
	tdmGLActivate(t, g, me.ID, tactician, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{token}})
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
	if effectstest.B16Tapped(t, g, tactician) {
		t.Error("the ability has no {T} in its cost")
	}
}

// --- Highspire Bell-Ringer -----------------------------------------------------

func TestTdmHighspireBellRingerDiscountsOnlyTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ringer := effectstest.B12Push(g, me.ID, "Highspire Bell-Ringer", "Creature — Djinn Monk", tdmHighspireBellRingerOracle, 1, 4)
	if !effectstest.HasEffectiveKeyword(t, g, ringer, "flying") {
		t.Error("Highspire Bell-Ringer has no flying")
	}
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("first spell: %d, want 3 (undiscounted)", got)
	}
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("second spell: %d, want 2 ({1} less)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Giant Growth", "Instant", "{G}"); got != 1 {
		t.Errorf("a second spell with no generic mana: %d, want 1 (only generic is reduced)", got)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("third spell: %d, want 3 (undiscounted)", got)
	}
}

// --- Humbling Elder ------------------------------------------------------------

func TestTdmHumblingElderShrinksAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.RfdCast(t, g, "Humbling Elder", "Creature — Human Monk", tdmHumblingElderOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("the controller's own creature is offered")
	}
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, victim); p != 1 || tt != 3 {
		t.Errorf("the target is %d/%d, want 1/3 (-2/-0)", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("my creature's power %d, want 2", p)
	}
}

func TestTdmHumblingElderHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.RfdHandCard(me, "Humbling Elder", "Creature — Human Monk", "", tdmHumblingElderOracle, 1, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
}

// --- Inevitable Defeat ---------------------------------------------------------

func TestTdmInevitableDefeatExilesAndSwingsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Inevitable Defeat", "Instant", tdmInevitableDefeatOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ogre) {
		t.Error("the Ogre was not exiled")
	}
	if opp.Life != theirs-3 {
		t.Errorf("its controller's life %d -> %d, want -3", theirs, opp.Life)
	}
	if me.Life != mine+3 {
		t.Errorf("my life %d -> %d, want +3", mine, me.Life)
	}
	if spec, _ := effects.Lookup(tdmInevitableDefeatOracle); !spec.CantBeCountered {
		t.Error("Inevitable Defeat can be countered")
	}
}

func TestTdmInevitableDefeatCantTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Their Land", "Basic Land — Swamp")
	if err := effectstest.CastCatalogSpellErr(t, g, "Inevitable Defeat", "Instant", tdmInevitableDefeatOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

// --- Iridescent Tiger ----------------------------------------------------------

func TestTdmIridescentTigerAddsFiveColorsOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Iridescent Tiger", "Creature — Cat", tdmIridescentTigerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if got := effectstest.McPoolCount(me, c); got != 1 {
			t.Errorf("%s in the pool = %d, want 1", c, got)
		}
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Iridescent Tiger", TypeLine: "Creature — Cat", OracleID: tdmIridescentTigerOracle,
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if got := effectstest.McPoolCount(me, c); got != 0 {
			t.Errorf("a Tiger that was not cast added %d %s", got, c)
		}
	}
}

// --- Jade-Cast Sentinel --------------------------------------------------------

func TestTdmJadeCastSentinelTucksAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinel := effectstest.B12Push(g, me.ID, "Jade-Cast Sentinel", "Artifact Creature — Ape Snake", tdmJadeCastSentinelOracle, 1, 5)
	if !effectstest.HasEffectiveKeyword(t, g, sentinel, "reach") {
		t.Error("Jade-Cast Sentinel has no reach")
	}
	dead := effectstest.PushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	tdmGLActivate(t, g, me.ID, sentinel, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	if !effectstest.B16Tapped(t, g, sentinel) {
		t.Error("{T} is part of the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
}

// --- Jeskai Brushmaster --------------------------------------------------------

func TestTdmJeskaiBrushmasterHasDoubleStrikeAndProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.B12Push(g, me.ID, "Jeskai Brushmaster", "Creature — Orc Monk", tdmJeskaiBrushmasterOracle, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, monk, "double strike") {
		t.Error("Jeskai Brushmaster has no double strike")
	}
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, monk); p != 3 || tt != 5 {
		t.Errorf("after a noncreature spell the Brushmaster is %d/%d, want 3/5 (prowess)", p, tt)
	}
}

// --- Jeskai Shrinekeeper -------------------------------------------------------

func TestTdmJeskaiShrinekeeperGainsAndDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Jeskai Shrinekeeper", "Creature — Dragon", tdmJeskaiShrinekeeperOracle, 3, 3)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("Jeskai Shrinekeeper has no %s", kw)
		}
	}
	effectstest.SeedLibrary(me, "Next")
	life, hand := me.Life, len(me.Hand.Cards)
	effectstest.AttackWith(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

// --- Kin-Tree Severance --------------------------------------------------------

func TestTdmKinTreeSeveranceExilesOnlyManaValueThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Kin-Tree Severance", "Instant", tdmKinTreeSeveranceOracle, effectstest.CardRefs(small)); err == nil {
		t.Error("a mana value 2 permanent was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Kin-Tree Severance", "Instant", tdmKinTreeSeveranceOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(big) {
		t.Error("the mana value 3 permanent was not exiled")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the untargeted Bear left the battlefield")
	}
}

// --- Kishla Village ------------------------------------------------------------

func TestTdmKishlaVillageEntersTappedUnlessYouControlAnIslandOrSwamp(t *testing.T) {
	for _, tc := range []struct {
		land   string
		tapped bool
	}{
		{"Basic Land — Forest", true},
		{"Basic Land — Island", false},
		{"Basic Land — Swamp", false},
	} {
		if got := tdmGLCheckland(t, "Kishla Village", tdmKishlaVillageOracle, tc.land); got != tc.tapped {
			t.Errorf("beside a %s: tapped = %v, want %v", tc.land, got, tc.tapped)
		}
	}
}

func TestTdmKishlaVillageSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Kishla Village", "Land", tdmKishlaVillageOracle, false)
	effectstest.SeedLibrary(me, "Top", "Second", "Third")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{G}")
	tdmGLActivate(t, g, me.ID, village, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, village) {
		t.Error("{T} is part of the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looks at %d cards, want 2", len(c.ScryCards))
	}
}

// --- Krotiq Nestguard ----------------------------------------------------------

func TestTdmKrotiqNestguardPaysToAttackThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	insect := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Krotiq Nestguard", OracleID: tdmKrotiqNestguardOracle,
		TypeLine: "Creature — Insect", Power: 4, Toughness: 4})
	effectstest.AdvanceToMainOf(t, g, 0)
	before := g.Clone()
	effectstest.AdvanceTo(t, before, game.StepDeclareAttackers)
	effectstest.WantDefenderRefused(t, before, insect, b.ID)

	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	tdmGLActivate(t, g, me.ID, insect, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(insect, b.ID); err != nil {
		t.Fatalf("attack after the activation: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, insect, "defender") {
		t.Error("the Nestguard lost defender; it should only be waived")
	}
}

// --- Lightfoot Technique -------------------------------------------------------

func TestTdmLightfootTechniqueCountersAndGrantsFlyingAndIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Lightfoot Technique", "Instant", tdmLightfootTechniqueOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	for _, kw := range []string{"flying", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
	}
}

// --- Loxodon Battle Priest -----------------------------------------------------

func TestTdmLoxodonBattlePriestCountersAnotherCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.B12Push(g, me.ID, "Loxodon Battle Priest", "Creature — Elephant Cleric", tdmLoxodonBattlePriestOracle, 3, 5)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, priest) {
		t.Error("the Priest offered itself as \"another\" creature")
	}
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, priest, game.CounterPlusOne); got != 0 {
		t.Errorf("the Priest has %d +1/+1 counters, want 0", got)
	}
}
