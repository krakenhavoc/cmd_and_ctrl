package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znrFronts is every front face this file covers, with the name the catalog
// shows for it.
var znrFronts = []struct{ name, oracle string }{
	{"Akoum Warrior", effectstest.AkoumWarriorOracle},
	{"Beyeen Veil", effectstest.BeyeenVeilOracle},
	{"Blackbloom Rogue", effectstest.BlackbloomRogueOracle},
	{"Bloodsoaked Insight", effectstest.BloodsoakedInsightOracle},
	{"Drowner of Truth", effectstest.DrownerOfTruthOracle},
	{"Emeria's Call", effectstest.EmeriasCallOracle},
	{"Glasswing Grace", effectstest.GlasswingGraceOracle},
	{"Jwari Disruption", effectstest.JwariDisruptionOracle},
	{"Kabira Takedown", effectstest.KabiraTakedownOracle},
	{"Kazandu Mammoth", effectstest.KazanduMammothOracle},
	{"Khalni Ambush", effectstest.KhalniAmbushOracle},
	{"Legion Leadership", effectstest.LegionLeadershipOracle},
	{"Makindi Stampede", effectstest.MakindiStampedeOracle},
	{"Razorgrass Ambush", effectstest.RazorgrassAmbushOracle},
	{"Silundi Vision", effectstest.SilundiVisionOracle},
	{"Skyclave Cleric", effectstest.SkyclaveClericOracle},
	{"Song-Mad Treachery", effectstest.SongMadTreacheryOracle},
	{"Strength of the Harvest", effectstest.StrengthOfTheHarvestOracle},
	{"Tangled Florahedron", effectstest.TangledFlorahedronOracle},
	{"Turntimber Symbiosis", effectstest.TurntimberSymbiosisOracle},
	{"Umara Wizard", effectstest.UmaraWizardOracle},
	{"Vastwood Fortification", effectstest.VastwoodFortifOracle},
	{"Zof Consumption", effectstest.ZofConsumptionOracle},
}

// Both faces of every card are registered, Full, and say nothing to the
// player about a manual fallback.
func TestZNR_FrontsAreFullOnBothFaces(t *testing.T) {
	for _, tc := range znrFronts {
		for face := 0; face <= 1; face++ {
			key := game.CatalogKeyForFace(tc.oracle, face)
			spec, ok := effects.Lookup(key)
			if !ok {
				t.Errorf("%s: face %d (%s) is not in the catalog", tc.name, face, key)
				continue
			}
			if spec.Completeness != effects.CompletenessFull {
				t.Errorf("%s: face %d is %v, want Full", tc.name, face, spec.Completeness)
			}
			if len(spec.Caveats) != 0 {
				t.Errorf("%s: face %d carries caveats %v", tc.name, face, spec.Caveats)
			}
		}
		if spec, _ := effects.Lookup(tc.oracle); spec.Name != tc.name {
			t.Errorf("front face of %s is named %q", tc.name, spec.Name)
		}
	}
}

// --- Akoum Warrior ------------------------------------------------------

func TestZNR_AkoumWarriorIsATrampler(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Akoum Warrior", "Creature — Minotaur Warrior", effectstest.AkoumWarriorOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Akoum Warrior has no trample")
	}
}

// --- Beyeen Veil --------------------------------------------------------

func TestZNR_BeyeenVeilShrinksOnlyTheOpponentsCreaturesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	other := effectstest.CtrlPushCreature(g, g.Seats[2].ID, "Third Bear")
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")

	effectstest.CastCatalogSpell(t, g, "Beyeen Veil", "Instant", effectstest.BeyeenVeilOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"opponent's": theirs, "third seat's": other} {
		if got := effectstest.EffectivePower(t, g, id); got != 0 {
			t.Errorf("the %s creature has power %d, want 0 (2 - 2)", name, got)
		}
		if got := effectstest.EffectiveToughness(t, g, id); got != 2 {
			t.Errorf("the %s creature has toughness %d, want 2 (-2/-0 leaves it)", name, got)
		}
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("the caster's own creature has power %d, want 2", got)
	}

	// CR 611.2c: the set is the creatures there at resolution.
	later := effectstest.CtrlPushCreature(g, opp.ID, "Late Bear")
	if got := effectstest.EffectivePower(t, g, later); got != 2 {
		t.Errorf("a creature that arrived afterwards has power %d, want 2", got)
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("after the turn the opponent's creature has power %d, want 2", got)
	}
}

// --- Blackbloom Rogue ---------------------------------------------------

func TestZNR_BlackbloomRogueGetsThreePowerWhileAnOpponentHasEightInTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	rogue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blackbloom Rogue", TypeLine: "Creature — Human Rogue",
		OracleID: effectstest.BlackbloomRogueOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, rogue, "menace") {
		t.Error("Blackbloom Rogue has no menace")
	}

	for i := 0; i < 7; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Filler", "Sorcery")
	}
	if got := effectstest.EffectivePower(t, g, rogue); got != 2 {
		t.Fatalf("seven cards in an opponent's graveyard: power %d, want 2", got)
	}

	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Eighth", "Sorcery")
	if got := effectstest.EffectivePower(t, g, rogue); got != 5 {
		t.Errorf("eight cards in one opponent's graveyard: power %d, want 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, rogue); got != 3 {
		t.Errorf("toughness %d, want 3 (+3/+0 only)", got)
	}

	// The count is per opponent, never the sum: 5 + 5 is not eight in either.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, third2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	rogue2 := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Blackbloom Rogue", TypeLine: "Creature — Human Rogue",
		OracleID: effectstest.BlackbloomRogueOracle, Power: 2, Toughness: 3, Owner: me2.ID, Controller: me2.ID,
	})
	for i := 0; i < 5; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g2, opp2.ID, "Filler", "Sorcery")
		effectstest.PushGraveyardCardWithTypeLine(g2, third2.ID, "Filler", "Sorcery")
	}
	if got := effectstest.EffectivePower(t, g2, rogue2); got != 2 {
		t.Errorf("five and five across two opponents: power %d, want 2", got)
	}

	// The controller's own graveyard does not count.
	for i := 0; i < 8; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g2, me2.ID, "Own", "Sorcery")
	}
	if got := effectstest.EffectivePower(t, g2, rogue2); got != 2 {
		t.Errorf("eight cards in the controller's own graveyard: power %d, want 2", got)
	}
	_ = third
}

// --- Bloodsoaked Insight ------------------------------------------------

func TestZNR_BloodsoakedInsightCostsOneLessForEachLifeAnOpponentLostThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const cost = "{5}{B/R}{B/R}"
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BloodsoakedInsightOracle, "Sorcery", cost); got != 7 {
		t.Fatalf("with no life lost the price is %d, want 7", got)
	}

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BloodsoakedInsightOracle, "Sorcery", cost); got != 4 {
		t.Errorf("after an opponent lost 3 the price is %d, want 4", got)
	}

	// Your own life loss is not an opponent's.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BloodsoakedInsightOracle, "Sorcery", cost); got != 4 {
		t.Errorf("after the caster lost 2 as well the price is %d, want 4 (only opponents count)", got)
	}

	// Two opponents add up; the coloured symbols are never reduced.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[2].ID, -10); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BloodsoakedInsightOracle, "Sorcery", cost); got != 2 {
		t.Errorf("after 13 life was lost the price is %d, want 2 (the two hybrid symbols stay)", got)
	}
}

func TestZNR_BloodsoakedInsightExilesTheTopThreeAndLetsYouPlayThemWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := effectstest.ZnrStackLibrary(opp,
		game.Card{Name: "Stolen Bolt", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Stolen Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Stolen Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	fourth := opp.Library.Cards[len(opp.Library.Cards)-4]
	libBefore := opp.Library.Size()

	effectstest.CastCatalogSpell(t, g, "Bloodsoaked Insight", "Sorcery", effectstest.BloodsoakedInsightOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := opp.Library.Size(); got != libBefore-3 {
		t.Errorf("the opponent's library went %d -> %d, want exactly three gone", libBefore, got)
	}
	for _, id := range top {
		if z := effectstest.ZnrZoneOf(g, id); z != game.ZoneExile {
			t.Errorf("a top card is in %v, want exile", z)
		}
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID {
			t.Errorf("the play permission names %s, want the caster", perm.Player)
			continue
		}
		if perm.CastOnly {
			t.Error(`"you may play those cards" must not be cast-only: a land is playable`)
		}
		if !perm.AnyType || !perm.AnyColor {
			t.Errorf("mana of any type can be spent to cast it: AnyType=%v AnyColor=%v", perm.AnyType, perm.AnyColor)
		}
		if !effectstest.PermissionLive(g, perm, me.ID) {
			t.Error("the permission is not live for the caster")
		}
		if effectstest.PermissionLive(g, perm, opp.ID) {
			t.Error("the owner of the exiled card was given the permission")
		}
	}
	if opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != fourth.InstanceID {
		t.Error("the fourth card is not the new top of the library")
	}
}

// A card that cost {R} cast out of the exile zone with a pool of {C} only
// proves the any-type relaxation, which is the clause's second sentence.
func TestZNR_BloodsoakedInsightCastsAStolenSpellWithColourlessMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := effectstest.ZnrStackLibrary(opp,
		game.Card{Name: "Stolen Bolt", TypeLine: "Instant", ManaCost: "{R}"})

	effectstest.CastCatalogSpell(t, g, "Bloodsoaked Insight", "Sorcery", effectstest.BloodsoakedInsightOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{
		FromZone: "exile", Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Errorf("casting the stolen {R} spell off colourless mana: %v", err)
	}
}

func TestZNR_BloodsoakedInsightNeedsAnOpponentAsItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	err := effectstest.CastCatalogSpellErr(t, g, "Bloodsoaked Insight", "Sorcery", effectstest.BloodsoakedInsightOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Error("Bloodsoaked Insight accepted its caster as the target opponent")
	}
}

// --- Drowner of Truth ---------------------------------------------------

func znrDrownerBoard(t *testing.T, solRing bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lands := []string{"Forest", "Forest", "Mountain", "Mountain", "Mountain"}
	if solRing {
		effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", effectstest.WishSolRingOracle, false)
	} else {
		lands = append(lands, "Mountain", "Mountain")
	}
	basics := map[string]string{"Forest": "Basic Land — Forest", "Mountain": "Basic Land — Mountain"}
	for _, name := range lands {
		if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: name, TypeLine: basics[name]}, 1); err != nil {
			t.Fatalf("spawn %s: %v", name, err)
		}
	}
	return g, me
}

func TestZNR_DrownerOfTruthMakesTwoSpawnWhenColourlessManaPaidForIt(t *testing.T) {
	g, me := znrDrownerBoard(t, true)
	effectstest.CastFromHandForTest(t, g, me, "Drowner of Truth", "Creature — Eldrazi", "{5}{G/U}{G/U}", effectstest.DrownerOfTruthOracle,
		game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountOnBattlefield(g, "Eldrazi Spawn", me.ID); got != 2 {
		t.Errorf("%d Eldrazi Spawn after paying with a Sol Ring, want 2", got)
	}
	if got := effectstest.CountOnBattlefield(g, "Drowner of Truth", me.ID); got != 1 {
		t.Errorf("%d Drowner of Truth on the battlefield, want 1", got)
	}
}

func TestZNR_DrownerOfTruthMakesNothingWhenNoColourlessManaWasSpent(t *testing.T) {
	g, me := znrDrownerBoard(t, false)
	effectstest.CastFromHandForTest(t, g, me, "Drowner of Truth", "Creature — Eldrazi", "{5}{G/U}{G/U}", effectstest.DrownerOfTruthOracle,
		game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
	if len(g.PendingTriggers) != 0 || g.Stack.Size() != 1 {
		t.Errorf("an unpaid-with-{C} cast put a trigger on the stack (stack %d, pending %d)", g.Stack.Size(), len(g.PendingTriggers))
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountOnBattlefield(g, "Eldrazi Spawn", me.ID); got != 0 {
		t.Errorf("%d Eldrazi Spawn with no {C} spent, want 0", got)
	}
	if got := effectstest.CountOnBattlefield(g, "Drowner of Truth", me.ID); got != 1 {
		t.Errorf("%d Drowner of Truth on the battlefield, want 1", got)
	}
}

func TestZNR_DrownerOfTruthIsDevoid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Drowner of Truth", "Creature — Eldrazi", effectstest.DrownerOfTruthOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, id, game.KeywordDevoid) {
		t.Error("Drowner of Truth is not devoid")
	}
}

// --- Emeria's Call ------------------------------------------------------

func TestZNR_EmeriasCallMakesTwoAngelWarriorsAndProtectsTheNonAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	angel := effectstest.PushTribalCreature(g, me.ID, "My Angel", "Creature — Angel", 3, 3)
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")

	effectstest.CastCatalogSpell(t, g, "Emeria's Call", "Sorcery", effectstest.EmeriasCallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var tokens []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Angel Warrior" && c.Controller == me.ID {
				tokens = append(tokens, c.InstanceID)
			}
		}
	})
	if len(tokens) != 2 {
		t.Fatalf("%d Angel Warrior tokens, want 2", len(tokens))
	}
	for _, id := range tokens {
		c, _ := effectstest.ZnrBattlefieldCard(g, id)
		if effectstest.EffectivePower(t, g, id) != 4 || effectstest.EffectiveToughness(t, g, id) != 4 {
			t.Errorf("an Angel Warrior token is %d/%d, want 4/4", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
			t.Error("an Angel Warrior token has no flying")
		}
		if len(c.Colors) != 1 || c.Colors[0] != "W" {
			t.Errorf("an Angel Warrior token is %v, want white", c.Colors)
		}
		if !c.HasSubtypeWord("Angel") || !c.HasSubtypeWord("Warrior") {
			t.Errorf("an Angel Warrior token is %q", c.TypeLine)
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Error("an Angel token gained indestructible: it is an Angel")
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a non-Angel creature you control did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, angel, "indestructible") {
		t.Error("an Angel you control gained indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}

	// CR 611.2c: a creature that arrives later is outside the set.
	late := effectstest.CtrlPushCreature(g, me.ID, "Late Bear")
	if effectstest.HasEffectiveKeyword(t, g, late, "indestructible") {
		t.Error("a creature that arrived after resolution gained indestructible")
	}

	// Until YOUR next turn: through every other seat's turn, gone on mine.
	for seat := 1; seat <= 3; seat++ {
		effectstest.AdvancePastCleanupForTest(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
			t.Fatalf("indestructible ended early, before seat %d's turn was over", seat)
		}
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible outlived the start of the caster's next turn")
	}
}

// --- Glasswing Grace ----------------------------------------------------

func TestZNR_GlasswingGraceGivesPlusTwoFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.CtrlPushCreature(g, me.ID, "Bear")

	effectstest.CastCatalogSpell(t, g, "Glasswing Grace", "Enchantment — Aura", effectstest.GlasswingGraceOracle,
		[]game.TargetRef{effectstest.CardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.EffectivePower(t, g, bear) != 4 || effectstest.EffectiveToughness(t, g, bear) != 4 {
		t.Errorf("the host is %d/%d, want 4/4", effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear))
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the host has no %s", kw)
		}
	}
}

// --- Jwari Disruption ---------------------------------------------------

func znrJwariBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster, them *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	caster, them = g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	disruption := effectstest.HandCardFull(them, "Jwari Disruption", "Instant", "", effectstest.JwariDisruptionOracle, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, disruption, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Jwari Disruption: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, caster.ID)
	return g, victim, caster, them
}

func TestZNR_JwariDisruptionCountersUnlessItsControllerPaysOne(t *testing.T) {
	g, victim, caster, _ := znrJwariBoard(t)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("the tax is %q, want {1}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestZNR_JwariDisruptionLetsTheSpellResolveWhenTheTaxIsPaid(t *testing.T) {
	g, victim, caster, _ := znrJwariBoard(t)
	caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

// --- Kabira Takedown ----------------------------------------------------

func TestZNR_KabiraTakedownDealsDamageEqualToYourCreatureCount(t *testing.T) {
	for _, tc := range []struct {
		mine     int
		wantDead bool
	}{
		{2, false},
		{3, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < tc.mine; i++ {
			effectstest.CtrlPushCreature(g, me.ID, "Mine")
		}
		wall := effectstest.PushTribalCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 3)

		effectstest.CastCatalogSpell(t, g, "Kabira Takedown", "Instant", effectstest.KabiraTakedownOracle,
			[]game.TargetRef{effectstest.CardRef(wall)})
		effectstest.PassPriorityAroundTable(t, g)

		if gone := !g.Battlefield.Contains(wall); gone != tc.wantDead {
			t.Errorf("with %d creatures the 0/3 died = %v, want %v", tc.mine, gone, tc.wantDead)
		}
	}
}

func TestZNR_KabiraTakedownHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.CtrlPushCreature(g, me.ID, "Mine")
	}
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Kabira Takedown", "Instant", effectstest.KabiraTakedownOracle,
		[]game.TargetRef{effectstest.CardRef(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty is %d, want 2 (5 - 3)", got)
	}
}

// --- Kazandu Mammoth ----------------------------------------------------

func TestZNR_KazanduMammothGetsTwoTwoForEachLandThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mammoth := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kazandu Mammoth", TypeLine: "Creature — Elephant",
		OracleID: effectstest.KazanduMammothOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, mammoth) != 5 || effectstest.EffectiveToughness(t, g, mammoth) != 5 {
		t.Errorf("after one land: %d/%d, want 5/5", effectstest.EffectivePower(t, g, mammoth), effectstest.EffectiveToughness(t, g, mammoth))
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mammoth); got != 7 {
		t.Errorf("after two lands: power %d, want 7", got)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.EffectivePower(t, g, mammoth); got != 3 {
		t.Errorf("next turn: power %d, want 3", got)
	}
}

func TestZNR_KazanduMammothIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mammoth := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kazandu Mammoth", TypeLine: "Creature — Elephant",
		OracleID: effectstest.KazanduMammothOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if _, err := g.SpawnCards(opp.ID, opp.ID, game.ZoneBattlefield, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, 1); err != nil {
		t.Fatalf("spawn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mammoth); got != 3 {
		t.Errorf("an opponent's land gave the Mammoth power %d, want 3", got)
	}
}

// --- Khalni Ambush ------------------------------------------------------

func TestZNR_KhalniAmbushMakesYourCreatureFightTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushTribalCreature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Khalni Ambush", "Instant", effectstest.KhalniAmbushOracle,
		[]game.TargetRef{effectstest.CardRef(mine), effectstest.CardRef(theirs)})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived a fight with a 3/3")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the 3/3 died to a 2/2's two damage")
	}
}

func TestZNR_KhalniAmbushWillNotTakeTwoOfYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.CtrlPushCreature(g, me.ID, "A")
	b := effectstest.CtrlPushCreature(g, me.ID, "B")
	if err := effectstest.CastCatalogSpellErr(t, g, "Khalni Ambush", "Instant", effectstest.KhalniAmbushOracle,
		[]game.TargetRef{effectstest.CardRef(a), effectstest.CardRef(b)}); err == nil {
		t.Error("Khalni Ambush accepted a creature you control as its second target")
	}
}

// --- Legion Leadership --------------------------------------------------

func TestZNR_LegionLeadershipDoublesPowerAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 3, 2)
	if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("counter: %v", err)
	}

	effectstest.CastCatalogSpell(t, g, "Legion Leadership", "Instant", effectstest.LegionLeadershipOracle,
		[]game.TargetRef{effectstest.CardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)

	// 3 base + a +1/+1 counter = 4, doubled = 8; the toughness stays 3.
	if got := effectstest.PowerOf(t, g, bear); got != 8 {
		t.Errorf("power %d, want 8 (4 doubled)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the creature did not gain first strike")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.PowerOf(t, g, bear); got != 4 {
		t.Errorf("next turn power %d, want 4", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlived the turn")
	}
}

// --- Makindi Stampede ---------------------------------------------------

func TestZNR_MakindiStampedeGivesYourCreaturesPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.CtrlPushCreature(g, me.ID, "A")
	b := effectstest.CtrlPushCreature(g, me.ID, "B")
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Theirs")

	effectstest.CastCatalogSpell(t, g, "Makindi Stampede", "Sorcery", effectstest.MakindiStampedeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if effectstest.EffectivePower(t, g, id) != 4 || effectstest.EffectiveToughness(t, g, id) != 4 {
			t.Errorf("one of mine is %d/%d, want 4/4", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
		}
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", got)
	}
	late := effectstest.CtrlPushCreature(g, me.ID, "Late")
	if got := effectstest.EffectivePower(t, g, late); got != 2 {
		t.Errorf("a creature that arrived afterwards has power %d, want 2", got)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 2 {
		t.Errorf("next turn power %d, want 2", got)
	}
}

// --- Razorgrass Ambush --------------------------------------------------

func TestZNR_RazorgrassAmbushDealsThreeToAnAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushTribalCreature(g, me.ID, "Attacker", "Creature — Bear", 2, 3)
	bystander := effectstest.PushTribalCreature(g, opp.ID, "Bystander", "Creature — Bear", 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	// A creature that is not attacking or blocking is not a legal target.
	if err := func() error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Razorgrass Ambush", TypeLine: "Instant",
			OracleID: effectstest.RazorgrassAmbushOracle, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(bystander)}})
	}(); err == nil {
		t.Error("Razorgrass Ambush accepted a creature that is neither attacking nor blocking")
	}

	effectstest.CastFromHandForTest(t, g, me, "Razorgrass Ambush", "Instant", "", effectstest.RazorgrassAmbushOracle,
		game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(attacker)}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("three damage did not kill a 2/3 attacker")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("the bystander died")
	}
}

// --- Silundi Vision -----------------------------------------------------

func TestZNR_SilundiVisionTakesAnInstantOrSorceryAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.ZnrStackLibrary(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	seventh := me.Library.Cards[len(me.Library.Cards)-7].InstanceID

	effectstest.CastCatalogSpell(t, g, "Silundi Vision", "Instant", effectstest.SilundiVisionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	want := map[uuid.UUID]bool{top[1]: true, top[3]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the two instants and sorceries", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a card that is not an instant or sorcery: %s", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}

	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[3]) || me.Hand.Size() != handBefore+1 {
		t.Error("the chosen sorcery is not the one new card in hand")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != seventh {
		t.Error("the seventh card is not the new top of the library: the rest were not bottomed")
	}
	for _, id := range []uuid.UUID{top[0], top[1], top[2], top[4], top[5]} {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("a card that was not chosen is not in the library: %s", id)
		}
	}
}

func TestZNR_SilundiVisionMayTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.ZnrStackLibrary(me,
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	effectstest.CastCatalogSpell(t, g, "Silundi Vision", "Instant", effectstest.SilundiVisionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != handBefore {
		t.Error("declining still took a card")
	}
	for _, id := range top {
		if !me.Library.Contains(id) {
			t.Error("a looked-at card left the library")
		}
	}
}

// --- Skyclave Cleric ----------------------------------------------------

func TestZNR_SkyclaveClericGainsTwoLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOf(g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Skyclave Cleric", "Creature — Kor Cleric", effectstest.SkyclaveClericOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d -> %d, want +2", before, got)
	}
}

// --- Song-Mad Treachery -------------------------------------------------

func TestZNR_SongMadTreacheryStealsUntapsAndHastesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	effectstest.CastCatalogSpell(t, g, "Song-Mad Treachery", "Sorcery", effectstest.SongMadTreacheryOracle,
		[]game.TargetRef{effectstest.CardRef(victim)})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want the caster", got)
	}
	if c, _ := effectstest.ZnrBattlefieldCard(g, victim); c.Tapped {
		t.Error("the stolen creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("the stolen creature has no haste")
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after the turn the controller is %s, want the owner", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("haste outlived the turn")
	}
}

// --- Strength of the Harvest -------------------------------------------

func TestZNR_StrengthOfTheHarvestCountsCreaturesAndEnchantmentsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	effectstest.CtrlPushCreature(g, me.ID, "Mine")
	effectstest.PushCatalogPermanent(g, me.ID, "Sphere", "Enchantment", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Dryad", "Enchantment Creature — Dryad", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Ward", "Enchantment", "", false)

	effectstest.CastCatalogSpell(t, g, "Strength of the Harvest", "Enchantment — Aura", effectstest.StrengthOfTheHarvestOracle,
		[]game.TargetRef{effectstest.CardRef(host)})
	effectstest.PassPriorityAroundTable(t, g)

	// Mine, the Sphere, the Dryad (once) and the Aura itself: four. The
	// host is the opponent's, and so is their Ward.
	if effectstest.EffectivePower(t, g, host) != 6 || effectstest.EffectiveToughness(t, g, host) != 6 {
		t.Errorf("the host is %d/%d, want 6/6 (2/2 +4/+4)", effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host))
	}
}

// --- Tangled Florahedron ------------------------------------------------

func TestZNR_TangledFlorahedronTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Tangled Florahedron", "Creature — Elemental", effectstest.TangledFlorahedronOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping for mana: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestZNR_TangledFlorahedronIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Tangled Florahedron", "Creature — Elemental", effectstest.TangledFlorahedronOracle, true)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Florahedron tapped for mana")
	}
}

func TestZNR_TurntimberSymbiosisPutsACheapCreatureOntoTheBattlefieldWithThreeCounters(t *testing.T) {
	g, me, top := effectstest.ZnrTurntimberBoard(t)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	want := map[uuid.UUID]bool{top[0]: true, top[2]: true, top[4]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the three creature cards", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a card that is not a creature card: %s", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}
	eighth := me.Library.Cards[len(me.Library.Cards)-8].InstanceID

	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[4]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	c, ok := effectstest.ZnrBattlefieldCard(g, top[4])
	if !ok {
		t.Fatal("the chosen creature is not on the battlefield")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("a mana value 3 creature entered with %d +1/+1 counters, want 3", got)
	}
	if got := effectstest.PowerOf(t, g, top[4]); got != 6 {
		t.Errorf("the 3/3 is a %d-power creature, want 6", got)
	}
	// The other six are on the bottom in some order, the eighth card on top.
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != eighth {
		t.Error("the eighth card is not the new top of the library")
	}
	for _, id := range []uuid.UUID{top[0], top[1], top[2], top[3], top[5], top[6]} {
		if !me.Library.Contains(id) {
			t.Errorf("an unchosen card is not in the library: %s", id)
		}
	}
}

func TestZNR_TurntimberSymbiosisGivesAnExpensiveCreatureNoCounters(t *testing.T) {
	g, me, top := effectstest.ZnrTurntimberBoard(t)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	c, ok := effectstest.ZnrBattlefieldCard(g, top[2])
	if !ok {
		t.Fatal("the chosen creature is not on the battlefield")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("a mana value 5 creature entered with %d counters, want 0", got)
	}
}

func TestZNR_TurntimberSymbiosisMayPutNothing(t *testing.T) {
	g, me, top := effectstest.ZnrTurntimberBoard(t)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	for _, id := range top {
		if !me.Library.Contains(id) {
			t.Errorf("a looked-at card left the library: %s", id)
		}
	}
}

// Counters gated on the picked card belong to a clause that puts one card:
// anything that could put more would hand every card the first one's
// counters.
func TestZNR_PutFromLibraryRefusesPerCardCountersOnAClauseThatPutsMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	counters := effects.PutCounters{Kind: game.CounterPlusOne, N: 1, IfManaValueAtMost: true, ManaValueAtMost: 3}
	ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
	for name, p := range map[string]effects.PutFromLibraryOntoBattlefield{
		"any number": {Player: me.ID, Max: 0, Optional: true, Counters: counters},
		"all":        {Player: me.ID, Max: 1, All: true, Counters: counters},
		"up to two":  {Player: me.ID, Max: 2, Optional: true, Counters: counters},
	} {
		var err error
		g.WithWriteLock(func() { err = p.Apply(ctx) })
		if err == nil {
			t.Errorf("%s: gated counters were accepted on a clause that can put more than one card", name)
		}
	}
}

// --- Umara Wizard -------------------------------------------------------

func TestZNR_UmaraWizardGainsFlyingWhenYouCastAnInstantSorceryOrWizard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantFlying     bool
	}{
		{"an instant", "Instant", true},
		{"a sorcery", "Sorcery", true},
		{"a Wizard creature", "Creature — Human Wizard", true},
		{"a non-Wizard creature", "Creature — Human Knight", false},
		{"an artifact", "Artifact", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			wizard := effectstest.PushCatalogPermanent(g, me.ID, "Umara Wizard", "Creature — Merfolk Wizard", effectstest.UmaraWizardOracle, false)
			if effectstest.HasEffectiveKeyword(t, g, wizard, "flying") {
				t.Fatal("the Wizard starts with flying")
			}
			effectstest.CastCatalogSpell(t, g, "Test spell", tc.typeLine, "", nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.HasEffectiveKeyword(t, g, wizard, "flying"); got != tc.wantFlying {
				t.Errorf("flying after casting %s = %v, want %v", tc.name, got, tc.wantFlying)
			}
			effectstest.AdvancePastCleanupForTest(t, g)
			if effectstest.HasEffectiveKeyword(t, g, wizard, "flying") {
				t.Error("flying outlived the turn")
			}
		})
	}
}

// --- Vastwood Fortification ---------------------------------------------

func TestZNR_VastwoodFortificationPutsAPlusOnePlusOneCounterOnTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Vastwood Fortification", "Instant", effectstest.VastwoodFortifOracle,
		[]game.TargetRef{effectstest.CardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.ZnrBattlefieldCard(g, bear)
	if got := c.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectstest.PowerOf(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
}

// --- Zof Consumption ----------------------------------------------------

func TestZNR_ZofConsumptionDrainsEachOpponentForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.LifeOf(g, me.ID)
	before := effectstest.LifeOfOpponents(g)

	effectstest.CastCatalogSpell(t, g, "Zof Consumption", "Sorcery", effectstest.ZofConsumptionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LifeOf(g, me.ID); got != mine+4 {
		t.Errorf("caster life %d -> %d, want +4 (once, not per opponent)", mine, got)
	}
	after := effectstest.LifeOfOpponents(g)
	for i := range before {
		if after[i] != before[i]-4 {
			t.Errorf("opponent %d life %d -> %d, want -4", i, before[i], after[i])
		}
	}
}
