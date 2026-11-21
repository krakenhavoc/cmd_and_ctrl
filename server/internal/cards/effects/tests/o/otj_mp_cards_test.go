package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_mp_cards_test.go — Outlaws of Thunder Junction (OTJ), the M–P batch
// (#4468): the cards that need no new engine seam, and the set's Mercenary
// token. Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjMPMaraudingSphinxOracle    = "70bea459-f6e9-411a-bd33-339bf95ea47a"
	otjMPMineRaiderOracle         = "5d575b78-3d7b-4860-abc3-e99e429279a1"
	otjMPMournersSurpriseOracle   = "567fa6e0-fc7b-4502-9c51-f2538c411003"
	otjMPNezumiLinkbreakerOracle  = "5d24fb6b-7176-407c-b917-6c88a6da40df"
	otjMPNimbleBrigandOracle      = "8b153452-b4f2-4310-99e2-c729f28be6b9"
	otjMPOasisGardenerOracle      = "210a130c-d5d3-43bf-8a80-d796c37242e8"
	otjMPOmenportVigilanteOracle  = "852e3d89-64a0-4b32-be2b-b1d6cd7c2f64"
	otjMPOutlawMedicOracle        = "637be5f7-ab39-4e3a-a5eb-c6e48fca00a9"
	otjMPOverzealousMuscleOracle  = "4cb63198-29b5-4607-bd4a-f449807fb764"
	otjMPPeerlessRopemasterOracle = "d7aaffed-a143-49d6-a516-b9bc8b503aee"
	otjMPPricklyPairOracle        = "4ce8e0d0-6a3e-4c47-a7b7-c1b1ea1b9426"
	otjMPProsperityTycoonOracle   = "07acb7e5-7aae-4d89-aabf-cbd329b3480b"
)

func TestOtjMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjMPMaraudingSphinxOracle:    "Marauding Sphinx",
		otjMPMineRaiderOracle:         "Mine Raider",
		otjMPMournersSurpriseOracle:   "Mourner's Surprise",
		otjMPNezumiLinkbreakerOracle:  "Nezumi Linkbreaker",
		otjMPNimbleBrigandOracle:      "Nimble Brigand",
		otjMPOasisGardenerOracle:      "Oasis Gardener",
		otjMPOmenportVigilanteOracle:  "Omenport Vigilante",
		otjMPOutlawMedicOracle:        "Outlaw Medic",
		otjMPOverzealousMuscleOracle:  "Overzealous Muscle",
		otjMPPeerlessRopemasterOracle: "Peerless Ropemaster",
		otjMPPricklyPairOracle:        "Prickly Pair",
		otjMPProsperityTycoonOracle:   "Prosperity Tycoon",
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
}

// --- the Mercenary token -----------------------------------------------------

func TestOtjMPMercenaryPumpsACreatureYouControlAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merc := effectstest.PushToken(g, me.ID, effects.MercenaryToken())
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// Activate only as a sorcery: not while a spell is on the stack.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)}); err == nil {
		t.Fatal("activate only as a sorcery: refused with a spell on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	effectstest.B16Activate(t, g, me.ID, merc, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tt)
	}
	if !effectstest.B16Tapped(t, g, merc) {
		t.Error("the Mercenary taps to activate")
	}
}

// --- Nezumi Linkbreaker ------------------------------------------------------

func TestOtjMPNezumiLinkbreakerMakesAMercenaryWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rat := effectstest.B12Push(g, me.ID, "Nezumi Linkbreaker", "Creature — Rat Warlock", otjMPNezumiLinkbreakerOracle, 1, 1)
	if len(effectstest.OtjMPMercenaries(g, me.ID)) != 0 {
		t.Fatal("a Mercenary before the Linkbreaker died")
	}
	effectstest.P1Destroy(t, g, rat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OtjMPOneMercenary(t, g, me.ID)
}

// --- Prickly Pair ------------------------------------------------------------

func TestOtjMPPricklyPairMakesAMercenaryThatIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pair := effectstest.Msc17Play(t, g, me, game.Card{Name: "Prickly Pair", TypeLine: "Creature — Plant Mercenary",
		ManaCost: "{2}{R}", OracleID: otjMPPricklyPairOracle, Power: 2, Toughness: 2})
	merc := effectstest.OtjMPOneMercenary(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(pair)}); err == nil {
		t.Error("the new Mercenary tapped for its ability the turn it was created")
	}
}

// --- Prosperity Tycoon -------------------------------------------------------

func TestOtjMPProsperityTycoonSacrificesATokenToBecomeIndestructibleAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tycoon := effectstest.Msc17Play(t, g, me, game.Card{Name: "Prosperity Tycoon", TypeLine: "Creature — Human Noble",
		ManaCost: "{3}{W}", OracleID: otjMPProsperityTycoonOracle, Power: 4, Toughness: 2})
	merc := effectstest.OtjMPOneMercenary(t, g, me.ID)
	nontoken := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tycoon, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{nontoken}}); err == nil {
		t.Fatal("a nontoken creature paid \"Sacrifice a token\"")
	}
	effectstest.B16Activate(t, g, me.ID, tycoon, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{merc}})
	if g.Battlefield.Contains(merc) {
		t.Error("the Mercenary was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, tycoon, "indestructible") {
		t.Error("the Tycoon gains indestructible")
	}
	if !effectstest.B16Tapped(t, g, tycoon) {
		t.Error("tap it: the Tycoon is tapped")
	}
	effectstest.P1Destroy(t, g, tycoon)
	if !g.Battlefield.Contains(tycoon) {
		t.Error("an indestructible Tycoon was destroyed")
	}
}

// --- Mourner's Surprise ------------------------------------------------------

func TestOtjMPMournersSurpriseReturnsACreatureCardAndMakesAMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	spell := effectstest.Batch01GraveyardCard(me, "Old Shock", "Instant")
	theirs := effectstest.Batch01GraveyardCard(opp, "Their Dead Bear", "Creature — Bear")
	for _, bad := range []uuid.UUID{spell, theirs} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, effectstest.B16TargetCard(bad)); err == nil {
			t.Fatal("only a creature card from your own graveyard may be targeted")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, effectstest.B16TargetCard(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dead) {
		t.Error("the creature card did not return to your hand")
	}
	effectstest.OtjMPOneMercenary(t, g, me.ID)
}

func TestOtjMPMournersSurpriseWithNoTargetStillMakesAMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OtjMPOneMercenary(t, g, me.ID)
}

// --- Outlaw Medic ------------------------------------------------------------

func TestOtjMPOutlawMedicHasLifelinkAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Drawn")
	medic := effectstest.B12Push(g, me.ID, "Outlaw Medic", "Creature — Human Rogue", otjMPOutlawMedicOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, medic, "lifelink") {
		t.Error("lifelink is printed")
	}
	hand := me.Hand.Size()
	effectstest.P1Destroy(t, g, medic)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d when the Medic died, want 1", got)
	}
}

// --- Oasis Gardener ----------------------------------------------------------

func TestOtjMPOasisGardenerGainsTwoOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Oasis Gardener", TypeLine: "Artifact Creature — Scarecrow",
		ManaCost: "{3}", OracleID: otjMPOasisGardenerOracle, Power: 2, Toughness: 2})
	if me.Life != before+2 {
		t.Errorf("you gain 2 life: %d → %d", before, me.Life)
	}
}

func TestOtjMPOasisGardenerTapsForAnyColour(t *testing.T) {
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		gardener := effectstest.B12Push(g, me.ID, "Oasis Gardener", "Artifact Creature — Scarecrow", otjMPOasisGardenerOracle, 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.B28TapForMana(t, g, me.ID, gardener, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}
}

// --- Peerless Ropemaster -----------------------------------------------------

func TestOtjMPPeerlessRopemasterBouncesUpToOneTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Their Tapped Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, tapped)
	untapped := effectstest.B12Creature(g, opp.ID, "Their Untapped Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Peerless Ropemaster", TypeLine: "Creature — Human Rogue",
		ManaCost: "{4}{U}", OracleID: otjMPPeerlessRopemasterOracle, Power: 4, Toughness: 4})
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, tapped) || effectstest.HasID(p.PickTargetCards, untapped) {
		t.Fatalf("only tapped creatures are offered: %v", p.PickTargetCards)
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 1 {
		t.Errorf("up to one: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: tapped}); err != nil {
		t.Fatalf("pick the tapped Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !effectstest.HandHasID(opp, tapped) {
		t.Error("the tapped creature returns to its owner's hand")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature stays")
	}
}

// --- Mine Raider -------------------------------------------------------------

func TestOtjMPMineRaiderMakesATreasureOnlyWithAnotherOutlaw(t *testing.T) {
	for _, tc := range []struct {
		other    string
		treasure int
	}{
		{"", 0},
		{"Creature — Bear", 0},
		{"Creature — Human Warlock", 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if tc.other != "" {
			effectstest.B12Creature(g, me.ID, "Other", tc.other, 1, 1)
		}
		effectstest.AdvanceToMain(t, g)
		raider := effectstest.Msc17Play(t, g, me, game.Card{Name: "Mine Raider", TypeLine: "Creature — Human Rogue",
			ManaCost: "{2}{R}", OracleID: otjMPMineRaiderOracle, Power: 3, Toughness: 2})
		if !effectstest.HasEffectiveKeyword(t, g, raider, "trample") {
			t.Errorf("%q: trample is printed", tc.other)
		}
		if got := effectstest.B16CountNamed(g, "Treasure"); got != tc.treasure {
			t.Errorf("other %q: %d Treasures, want %d", tc.other, got, tc.treasure)
		}
	}
}

// --- Marauding Sphinx --------------------------------------------------------

func TestOtjMPMaraudingSphinxSurveilsTwoOnTheFirstCrimeEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 10)
	sphinx := effectstest.B12Push(g, me.ID, "Marauding Sphinx", "Creature — Sphinx Rogue", otjMPMaraudingSphinxOracle, 3, 5)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, sphinx, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Fatal("targeting your own creature is no crime")
	}

	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SurveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("a crime: no surveil prompt")
	}
	if len(prompt.ScryCards) != 2 {
		t.Errorf("surveil 2 looks at %d cards", len(prompt.ScryCards))
	}
	var rest []uuid.UUID
	for _, id := range prompt.ScryCards {
		if id != top {
			rest = append(rest, id)
		}
	}
	if err := g.ResolveSurveil(prompt.ID, me.ID, []uuid.UUID{top}, rest); err != nil {
		t.Fatalf("surveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}

	otjB1BoltOpponent(t, g, opp.ID)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("a second crime the same turn surveilled again")
	}
}

func TestOtjMPMaraudingSphinxHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := effectstest.B12Push(g, me.ID, "Marauding Sphinx", "Creature — Sphinx Rogue", otjMPMaraudingSphinxOracle, 3, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, sphinx)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting the Sphinx")
	}
	if p := effectstest.AnswerPayUnless(t, g, opp.ID, false); p.PayCost != "{2}" {
		t.Errorf("ward cost %q, want {2}", p.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphinx) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Overzealous Muscle ------------------------------------------------------

func TestOtjMPOverzealousMuscleGainsIndestructibleFromACrimeOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muscle := effectstest.B12Push(g, me.ID, "Overzealous Muscle", "Creature — Ogre Mercenary", otjMPOverzealousMuscleOracle, 5, 4)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Fatal("indestructible before any crime")
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if !effectstest.HasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Fatal("a crime on your turn: indestructible")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Error("the grant lasts until end of turn only")
	}
}

func TestOtjMPOverzealousMuscleIgnoresACrimeOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muscle := effectstest.B12Push(g, opp.ID, "Overzealous Muscle", "Creature — Ogre Mercenary", otjMPOverzealousMuscleOracle, 5, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: effectstest.B16TargetPlayer(me.ID)}); err != nil {
		t.Fatalf("the Muscle's controller bolts you on your turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Error("a crime during another player's turn does not trigger it")
	}
}

// --- Omenport Vigilante ------------------------------------------------------

func TestOtjMPOmenportVigilanteHasDoubleStrikeAfterACrimeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vigilante := effectstest.B12Push(g, me.ID, "Omenport Vigilante", "Creature — Human Mercenary", otjMPOmenportVigilanteOracle, 2, 2)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("double strike before any crime")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("targeting your own creature is no crime")
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if !effectstest.HasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("you've committed a crime this turn: double strike")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Error("a crime last turn does not count this turn")
	}
}

// --- Nimble Brigand ----------------------------------------------------------

func TestOtjMPNimbleBrigandCanBeBlockedWithoutACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigand := effectstest.B12Push(g, me.ID, "Nimble Brigand", "Creature — Human Rogue", otjMPNimbleBrigandOracle, 1, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.Pr7bAttack(t, g, opp.ID, brigand)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, brigand); err != nil {
		t.Errorf("with no crime this turn the Brigand can be blocked: %v", err)
	}
}

func TestOtjMPNimbleBrigandIsUnblockableAfterACrimeAndDrawsOnHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Drawn")
	brigand := effectstest.B12Push(g, me.ID, "Nimble Brigand", "Creature — Human Rogue", otjMPNimbleBrigandOracle, 1, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	effectstest.Pr7bAttack(t, g, opp.ID, brigand)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, brigand); err == nil {
		t.Fatal("after a crime this turn the Brigand can't be blocked")
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand, life := me.Hand.Size(), opp.Life
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the Brigand deals 1: %d → %d", life, opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on combat damage to a player, want 1", got)
	}
}
