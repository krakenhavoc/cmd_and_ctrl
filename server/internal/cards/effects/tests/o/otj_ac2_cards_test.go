package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_ac2_cards_test.go — Outlaws of Thunder Junction (OTJ), the second
// A–C batch (#4468): the cards batch 1 (#5163) left. Each test plays the
// real catalog entry through the engine and asserts what the printed
// text says.

const (
	otjAC2AkulOracle         = "d6b56d6d-e2f6-47f5-a9d1-cd0e17f94a0d"
	otjAC2AnnieFlashOracle   = "858067f9-eb1f-4569-9831-30441cba5f90"
	otjAC2AridArchwayOracle  = "3be3d7e6-7860-438a-b8c8-ef154c18c163"
	otjAC2AtKnifepointOracle = "bddb2064-afdb-4184-ae4a-633745e64a3c"
	otjAC2BaronOracle        = "e25b516c-bf95-42bc-8b1e-04617a3d28df"
	otjAC2BetrayalOracle     = "f1dce3f0-30c8-488f-bbeb-b0533886c655"
	otjAC2BoomBoxOracle      = "b0fdf2e0-3025-4919-a9dc-d33100e487fd"
	otjAC2BrimstoneOracle    = "e79deeee-c14d-496d-9b0b-475a6a24274c"
	otjAC2BruseTarlOracle    = "c22589e1-1721-43a3-897e-990af6ebcffa"
	otjAC2CactarantulaOracle = "18a5fe34-0c97-486c-b09a-80535a99e982"
	otjAC2SureshotOracle     = "05f15b7b-28bf-4b63-a4ca-b57712a1a66a"
)

func TestOtjAC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjAC2AkulOracle:         "Akul the Unrepentant",
		otjAC2AnnieFlashOracle:   "Annie Flash, the Veteran",
		otjAC2AridArchwayOracle:  "Arid Archway",
		otjAC2AtKnifepointOracle: "At Knifepoint",
		otjAC2BaronOracle:        "Baron Bertram Graywater",
		otjAC2BetrayalOracle:     "Betrayal at the Vault",
		otjAC2BoomBoxOracle:      "Boom Box",
		otjAC2BrimstoneOracle:    "Brimstone Roundup",
		otjAC2BruseTarlOracle:    "Bruse Tarl, Roving Rancher",
		otjAC2CactarantulaOracle: "Cactarantula",
		otjAC2SureshotOracle:     "Cactusfolk Sureshot",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if sa := effectstest.PlotOf(otjAC2BrimstoneOracle); sa == nil || sa.Cost != "{2}{R}" {
		t.Errorf("Brimstone Roundup: plot %+v, want {2}{R}", sa)
	}
}

// otjAC2HandCreature puts a creature card in p's hand.
func otjAC2HandCreature(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Wurm", ManaCost: "{8}",
		Power: 8, Toughness: 8, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// otjAC2Damage reads the damage marked on a battlefield permanent.
func otjAC2Damage(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.DamageMarked
}

// --- Akul the Unrepentant ----------------------------------------------------

func TestOtjAC2AkulSacrificesThreeToPutACreatureFromHandOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	akul := effectstest.B12Push(g, me.ID, "Akul the Unrepentant", "Legendary Creature — Scorpion Dragon Rogue", otjAC2AkulOracle, 5, 5)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, akul, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	fodder := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Fodder C", "Creature — Bear", 1, 1),
	}
	fatty := otjAC2HandCreature(me, "Big Fatty")
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, akul, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{akul, fodder[0], fodder[1]}}); err == nil {
		t.Fatal("Akul sacrificed itself to its \"three other creatures\" cost")
	}
	effectstest.B16Activate(t, g, me.ID, akul, 0, game.ActivateAbilityParams{SacrificeIDs: fodder})
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no put-from-hand prompt")
	}
	if !effectstest.HasID(c.ChooseCards, fatty) || c.ChooseMin != 0 {
		t.Fatalf("you may put the creature card: %v in [%d,%d]", c.ChooseCards, c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fatty}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range fodder {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not sacrificed", id)
		}
	}
	if !g.Battlefield.Contains(fatty) {
		t.Fatal("the creature card did not reach the battlefield")
	}

	more := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Fodder D", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Fodder E", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Fodder F", "Creature — Bear", 1, 1),
	}
	if err := g.ActivateCatalogAbility(me.ID, akul, 0, game.ActivateAbilityParams{SacrificeIDs: more}); err == nil {
		t.Error("only once each turn: a second activation was accepted")
	}
}

// --- Annie Flash, the Veteran ------------------------------------------------

func TestOtjAC2AnnieFlashReturnsASmallPermanentCardTappedWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	annie := effectstest.Msc17Play(t, g, me, game.Card{Name: "Annie Flash, the Veteran", TypeLine: "Legendary Creature — Human Rogue",
		ManaCost: "{3}{R}{G}{W}", OracleID: otjAC2AnnieFlashOracle, Power: 4, Toughness: 5})
	if !effectstest.HasEffectiveKeyword(t, g, annie, "flash") {
		t.Error("flash is printed")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("cast: no target prompt for the enters trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, small) || effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, spell) {
		t.Fatalf("a permanent card with mana value 3 or less: %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
		t.Fatalf("pick the land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the land card returns to the battlefield")
	}
	if !effectstest.B16Tapped(t, g, small) {
		t.Error("it returns tapped")
	}
}

func TestOtjAC2AnnieFlashExilesTwoToPlayWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three")
	annie := effectstest.B12Push(g, me.ID, "Annie Flash, the Veteran", "Legendary Creature — Human Rogue", otjAC2AnnieFlashOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, annie, opp.ID)
	for _, id := range ids[:2] {
		if !g.Exile.Contains(id) {
			t.Fatalf("%s: the top two cards are exiled", id)
		}
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
			t.Errorf("%s: permission %+v, want a live play permission for you", id, perm)
		}
	}
	if g.Exile.Contains(ids[2]) {
		t.Error("only two cards are exiled")
	}
}

// --- Arid Archway ------------------------------------------------------------

func TestOtjAC2AridArchwaySurveilsOnlyForAnotherDesert(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pick    func(archway, desert, forest uuid.UUID) uuid.UUID
		surveil bool
	}{
		{"another Desert", func(_, desert, _ uuid.UUID) uuid.UUID { return desert }, true},
		{"a Forest", func(_, _, forest uuid.UUID) uuid.UUID { return forest }, false},
		{"the Archway itself", func(archway, _, _ uuid.UUID) uuid.UUID { return archway }, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedLibrary(me, "Top", "Second")
		desert := effectstest.SeedLandOnBattlefield(g, me.ID, "Some Desert", "Land — Desert")
		forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
		archway := effectstest.PlayLandFromHand(t, g, "Arid Archway", otjAC2AridArchwayOracle)
		effectstest.Top100AssertEnteredTapped(t, g, archway, "Arid Archway")
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if prompt == nil {
			t.Fatalf("%s: the trigger asks which land to return", tc.name)
		}
		pick := tc.pick(archway, desert, forest)
		if err := g.ResolveOwnPermanents(prompt.ID, me.ID, []uuid.UUID{pick}); err != nil {
			t.Fatalf("%s: ResolveOwnPermanents: %v", tc.name, err)
		}
		if g.Battlefield.Contains(pick) || !me.Hand.Contains(pick) {
			t.Errorf("%s: the chosen land is back in hand", tc.name)
		}
		if got := effectstest.SurveilChoiceFor(g, me.ID) != nil; got != tc.surveil {
			t.Errorf("%s: surveil owed = %v, want %v", tc.name, got, tc.surveil)
		}
	}
	spec, _ := effects.Lookup(otjAC2AridArchwayOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}{C}" {
		t.Errorf("row produces %+v, want {C}{C}", spec.ManaAbilities)
	}
}

// --- At Knifepoint -----------------------------------------------------------

func TestOtjAC2AtKnifepointGivesYourOutlawsFirstStrikeOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "At Knifepoint", "Enchantment", otjAC2AtKnifepointOracle, false)
	rogue := effectstest.B12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, rogue, "first strike") {
		t.Error("your outlaw has first strike on your turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("a Bear is no outlaw")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("an opponent's Pirate is not an outlaw you control")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, rogue, "first strike") {
		t.Error("only during your turn")
	}
}

func TestOtjAC2AtKnifepointMakesAMercenaryOnTheFirstCrimeEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "At Knifepoint", "Enchantment", otjAC2AtKnifepointOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.OtjMPMercenaries(g, me.ID)); got != 0 {
		t.Fatalf("targeting your own creature is no crime: %d Mercenaries", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	merc := effectstest.OtjMPOneMercenary(t, g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, merc, "first strike") {
		t.Error("the Mercenary is an outlaw: it has first strike on your turn")
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := len(effectstest.OtjMPMercenaries(g, me.ID)); got != 1 {
		t.Errorf("only once each turn: %d Mercenaries after a second crime", got)
	}
}

// --- Baron Bertram Graywater -------------------------------------------------

func TestOtjAC2BaronMakesOneVampireRogueEachTurnTokensEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Baron Bertram Graywater", "Legendary Creature — Vampire Noble", otjAC2BaronOracle, 3, 4)
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Vampire Rogue"); got != 1 {
		t.Fatalf("%d Vampire Rogues after two tokens entered together, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire Rogue" {
			if !c.IsToken() || len(c.Colors) != 1 || c.Colors[0] != "B" || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink") {
				t.Errorf("Vampire Rogue %+v, want a black token with lifelink", c)
			}
			if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 1 || tt != 1 {
				t.Errorf("Vampire Rogue is %d/%d, want 1/1", p, tt)
			}
		}
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Vampire Rogue"); got != 1 {
		t.Errorf("only once each turn: %d Vampire Rogues", got)
	}
}

func TestOtjAC2BaronSacrificesAnotherCreatureOrArtifactToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	baron := effectstest.B12Push(g, me.ID, "Baron Bertram Graywater", "Legendary Creature — Vampire Noble", otjAC2BaronOracle, 3, 4)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "C")
	for _, bad := range []uuid.UUID{baron, land} {
		if err := g.ActivateCatalogAbility(me.ID, baron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bad}}); err == nil {
			t.Errorf("%s was accepted for \"another creature or artifact\"", bad)
		}
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, baron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

// --- Betrayal at the Vault ---------------------------------------------------

func TestOtjAC2BetrayalAtTheVaultHitsTwoOtherCreaturesForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "My Bruiser", "Creature — Bear", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, other.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Betrayal at the Vault", "Instant", otjAC2BetrayalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: mine, Slot: 1},
		{Kind: game.TargetCard, ID: small, Slot: 1},
	}); err == nil {
		t.Fatal("\"two other target creatures\" accepted the biter")
	}
	effectstest.CastCatalogSpell(t, g, "Betrayal at the Vault", "Instant", otjAC2BetrayalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: small, Slot: 1},
		{Kind: game.TargetCard, ID: big, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 took 3 and died")
	}
	if got := otjAC2Damage(t, g, big); got != 3 {
		t.Errorf("the 4/4 has %d damage, want 3", got)
	}
	if got := otjAC2Damage(t, g, mine); got != 0 {
		t.Errorf("the biter took %d damage, want none (not a fight)", got)
	}
}

// --- Boom Box ----------------------------------------------------------------

func TestOtjAC2BoomBoxDestroysUpToOneOfEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := effectstest.PushCatalogPermanent(g, me.ID, "Boom Box", "Artifact", otjAC2BoomBoxOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, box, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Slot: 0},
		{Kind: game.TargetCard, ID: bear, Slot: 1},
		{Kind: game.TargetCard, ID: land, Slot: 2},
	}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{box, rock, bear, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s is still on the battlefield", id)
		}
	}
}

func TestOtjAC2BoomBoxWithOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := effectstest.PushCatalogPermanent(g, me.ID, "Boom Box", "Artifact", otjAC2BoomBoxOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, box, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 1},
	}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact nobody targeted was destroyed")
	}
}

// --- Brimstone Roundup -------------------------------------------------------

func TestOtjAC2BrimstoneRoundupMakesAMercenaryOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Brimstone Roundup", "Enchantment", otjAC2BrimstoneOracle, false)
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.OtjMPMercenaries(g, me.ID)); got != 0 {
		t.Fatalf("the first spell made %d Mercenaries", got)
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OtjMPOneMercenary(t, g, me.ID)
	effectstest.B24CastNoncreature(t, g, "Third", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.OtjMPMercenaries(g, me.ID)); got != 1 {
		t.Errorf("the third spell triggered it: %d Mercenaries", got)
	}
}

// --- Bruse Tarl, Roving Rancher ----------------------------------------------

func TestOtjAC2BruseTarlGivesYourOxenDoubleStrikeAndMakesAnOxForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruse := effectstest.B12Push(g, me.ID, "Bruse Tarl, Roving Rancher", "Legendary Creature — Human Warrior", otjAC2BruseTarlOracle, 4, 3)
	ox := effectstest.B12Creature(g, me.ID, "My Ox", "Creature — Ox", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, ox, "double strike") {
		t.Error("your Ox has double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") || effectstest.HasEffectiveKeyword(t, g, bruse, "double strike") {
		t.Error("only Oxen you control")
	}
	land := effectstest.SeedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, bruse, opp.ID)
	if !g.Exile.Contains(land) {
		t.Fatal("attacking exiles the top card")
	}
	var oxen []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ox" && c.IsToken() && c.Controller == me.ID {
			oxen = append(oxen, c.InstanceID)
		}
	}
	if len(oxen) != 1 {
		t.Fatalf("%d Ox tokens for a land, want 1", len(oxen))
	}
	if p, tt := effectstest.PtOf(t, g, oxen[0]); p != 2 || tt != 2 {
		t.Errorf("the Ox is %d/%d, want 2/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, oxen[0], "double strike") {
		t.Error("the new Ox has double strike too")
	}
}

func TestOtjAC2BruseTarlLetsYouCastANonlandUntilTheEndOfYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.SeedLibraryTop(me, "Impulse", "Sorcery", "{R}")
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bruse Tarl, Roving Rancher", TypeLine: "Legendary Creature — Human Warrior",
		ManaCost: "{2}{R}{W}", OracleID: otjAC2BruseTarlOracle, Power: 4, Toughness: 3})
	if !g.Exile.Contains(top) {
		t.Fatal("entering exiles the top card")
	}
	if got := effectstest.B16CountNamed(g, "Ox"); got != 0 {
		t.Errorf("a nonland card made %d Oxen", got)
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID || !perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("permission %+v, want a live cast permission for you", perm)
	}
	for i := 0; i < 4; i++ {
		effectstest.EndTurn(t, g)
	}
	if g.CastPermissionOnCardByIDForEffect(top) == nil {
		t.Error("the permission lapsed before the end of your next turn")
	}
	effectstest.EndTurn(t, g)
	if g.CastPermissionOnCardByIDForEffect(top) != nil {
		t.Error("the permission outlasted the end of your next turn")
	}
}

// --- Cactarantula ------------------------------------------------------------

func TestOtjAC2CactarantulaCostsOneLessWithADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, otjAC2CactarantulaOracle, "Creature — Plant Spider", "{4}{G}{G}"); got != 6 {
		t.Errorf("no Desert: %d, want 6", got)
	}
	effectstest.SeedLandOnBattlefield(g, g.Seats[1].ID, "Their Desert", "Land — Desert")
	if got := effectstest.SelfPricedMV(t, g, me, otjAC2CactarantulaOracle, "Creature — Plant Spider", "{4}{G}{G}"); got != 6 {
		t.Errorf("an opponent's Desert: %d, want 6", got)
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "My Desert", "Land — Desert")
	if got := effectstest.SelfPricedMV(t, g, me, otjAC2CactarantulaOracle, "Creature — Plant Spider", "{4}{G}{G}"); got != 5 {
		t.Errorf("your Desert: %d, want 5", got)
	}
}

func TestOtjAC2CactarantulaMayDrawWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	spider := effectstest.B12Push(g, me.ID, "Cactarantula", "Creature — Plant Spider", otjAC2CactarantulaOracle, 6, 5)
	if !effectstest.HasEffectiveKeyword(t, g, spider, "reach") {
		t.Error("reach is printed")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(spider))
	prompt := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if prompt.Source != spider {
		t.Fatalf("the trigger prompt is from %v, want the Spider", prompt.Source)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer the Spider: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	mine := g2.Seats[0]
	s2 := effectstest.B12Push(g2, mine.ID, "Cactarantula", "Creature — Plant Spider", otjAC2CactarantulaOracle, 6, 5)
	effectstest.CastCatalogSpell(t, g2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(s2))
	if effectstest.TriggerOnStack(g2, s2) != nil || effectstest.LatestTriggerPrompt(g2, mine.ID) != nil {
		t.Error("your own spell triggered it")
	}
}

// --- Cactusfolk Sureshot -----------------------------------------------------

func TestOtjAC2SureshotGivesYourOtherBigCreaturesTrampleAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sureshot := effectstest.B12Push(g, me.ID, "Cactusfolk Sureshot", "Creature — Plant Mercenary", otjAC2SureshotOracle, 4, 4)
	big := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	if !effectstest.HasEffectiveKeyword(t, g, sureshot, "reach") {
		t.Error("reach is printed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, big, kw) {
			t.Errorf("your power 4 creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, small, kw) {
			t.Errorf("a power 3 creature gained %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature gained %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, sureshot, kw) {
			t.Errorf("\"other\": the Sureshot gave itself %s", kw)
		}
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, big, "trample") {
		t.Error("until end of turn only")
	}
}

func TestOtjAC2SureshotHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sureshot := effectstest.B12Push(g, me.ID, "Cactusfolk Sureshot", "Creature — Plant Mercenary", otjAC2SureshotOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, sureshot)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting the Sureshot")
	}
	if p := effectstest.AnswerPayUnless(t, g, opp.ID, false); p.PayCost != "{2}" {
		t.Errorf("ward cost %q, want {2}", p.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sureshot) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}
