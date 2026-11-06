package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// otjMPMercenaries lists the Mercenary tokens `owner` controls.
func otjMPMercenaries(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Mercenary" && c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// otjMPOneMercenary asserts `owner` controls exactly one Mercenary token,
// printed as the set makes it, and returns it.
func otjMPOneMercenary(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	ids := otjMPMercenaries(g, owner)
	if len(ids) != 1 {
		t.Fatalf("%d Mercenary tokens, want 1", len(ids))
	}
	c, ok := battlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("the Mercenary is not on the battlefield")
	}
	if game.TokenTextForCard(c) == "" {
		t.Error("the Mercenary ships no printed text")
	}
	if !c.IsCreature() || !c.HasSubtype("Mercenary") || len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Mercenary %q colors %v, want a red Mercenary creature", c.TypeLine, c.Colors)
	}
	if p, tt := ptOf(t, g, ids[0]); p != 1 || tt != 1 {
		t.Errorf("Mercenary is %d/%d, want 1/1", p, tt)
	}
	return ids[0]
}

// --- the Mercenary token -----------------------------------------------------

func TestOtjMPMercenaryPumpsACreatureYouControlAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merc := pushToken(g, me.ID, MercenaryToken())
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	// Activate only as a sorcery: not while a spell is on the stack.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: b16TargetCard(mine)}); err == nil {
		t.Fatal("activate only as a sorcery: refused with a spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	b16Activate(t, g, me.ID, merc, 0, game.ActivateAbilityParams{Targets: b16TargetCard(mine)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tt)
	}
	if !b16Tapped(t, g, merc) {
		t.Error("the Mercenary taps to activate")
	}
}

// --- Nezumi Linkbreaker ------------------------------------------------------

func TestOtjMPNezumiLinkbreakerMakesAMercenaryWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rat := b12Push(g, me.ID, "Nezumi Linkbreaker", "Creature — Rat Warlock", otjMPNezumiLinkbreakerOracle, 1, 1)
	if len(otjMPMercenaries(g, me.ID)) != 0 {
		t.Fatal("a Mercenary before the Linkbreaker died")
	}
	p1Destroy(t, g, rat)
	passPriorityAroundTable(t, g)
	otjMPOneMercenary(t, g, me.ID)
}

// --- Prickly Pair ------------------------------------------------------------

func TestOtjMPPricklyPairMakesAMercenaryThatIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pair := msc17Play(t, g, me, game.Card{Name: "Prickly Pair", TypeLine: "Creature — Plant Mercenary",
		ManaCost: "{2}{R}", OracleID: otjMPPricklyPairOracle, Power: 2, Toughness: 2})
	merc := otjMPOneMercenary(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, merc, 0, game.ActivateAbilityParams{Targets: b16TargetCard(pair)}); err == nil {
		t.Error("the new Mercenary tapped for its ability the turn it was created")
	}
}

// --- Prosperity Tycoon -------------------------------------------------------

func TestOtjMPProsperityTycoonSacrificesATokenToBecomeIndestructibleAndTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tycoon := msc17Play(t, g, me, game.Card{Name: "Prosperity Tycoon", TypeLine: "Creature — Human Noble",
		ManaCost: "{3}{W}", OracleID: otjMPProsperityTycoonOracle, Power: 4, Toughness: 2})
	merc := otjMPOneMercenary(t, g, me.ID)
	nontoken := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tycoon, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{nontoken}}); err == nil {
		t.Fatal("a nontoken creature paid \"Sacrifice a token\"")
	}
	b16Activate(t, g, me.ID, tycoon, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{merc}})
	if g.Battlefield.Contains(merc) {
		t.Error("the Mercenary was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, tycoon, "indestructible") {
		t.Error("the Tycoon gains indestructible")
	}
	if !b16Tapped(t, g, tycoon) {
		t.Error("tap it: the Tycoon is tapped")
	}
	p1Destroy(t, g, tycoon)
	if !g.Battlefield.Contains(tycoon) {
		t.Error("an indestructible Tycoon was destroyed")
	}
}

// --- Mourner's Surprise ------------------------------------------------------

func TestOtjMPMournersSurpriseReturnsACreatureCardAndMakesAMercenary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	spell := batch01GraveyardCard(me, "Old Shock", "Instant")
	theirs := batch01GraveyardCard(opp, "Their Dead Bear", "Creature — Bear")
	for _, bad := range []uuid.UUID{spell, theirs} {
		if err := castCatalogSpellErr(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, b16TargetCard(bad)); err == nil {
			t.Fatal("only a creature card from your own graveyard may be targeted")
		}
	}
	castCatalogSpell(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, b16TargetCard(dead))
	passPriorityAroundTable(t, g)
	if !handHasID(me, dead) {
		t.Error("the creature card did not return to your hand")
	}
	otjMPOneMercenary(t, g, me.ID)
}

func TestOtjMPMournersSurpriseWithNoTargetStillMakesAMercenary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Mourner's Surprise", "Sorcery", otjMPMournersSurpriseOracle, nil)
	passPriorityAroundTable(t, g)
	otjMPOneMercenary(t, g, me.ID)
}

// --- Outlaw Medic ------------------------------------------------------------

func TestOtjMPOutlawMedicHasLifelinkAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	medic := b12Push(g, me.ID, "Outlaw Medic", "Creature — Human Rogue", otjMPOutlawMedicOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, medic, "lifelink") {
		t.Error("lifelink is printed")
	}
	hand := me.Hand.Size()
	p1Destroy(t, g, medic)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d when the Medic died, want 1", got)
	}
}

// --- Oasis Gardener ----------------------------------------------------------

func TestOtjMPOasisGardenerGainsTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	before := me.Life
	msc17Play(t, g, me, game.Card{Name: "Oasis Gardener", TypeLine: "Artifact Creature — Scarecrow",
		ManaCost: "{3}", OracleID: otjMPOasisGardenerOracle, Power: 2, Toughness: 2})
	if me.Life != before+2 {
		t.Errorf("you gain 2 life: %d → %d", before, me.Life)
	}
}

func TestOtjMPOasisGardenerTapsForAnyColour(t *testing.T) {
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		gardener := b12Push(g, me.ID, "Oasis Gardener", "Artifact Creature — Scarecrow", otjMPOasisGardenerOracle, 2, 2)
		advanceToMain(t, g)
		b28TapForMana(t, g, me.ID, gardener, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}
}

// --- Peerless Ropemaster -----------------------------------------------------

func TestOtjMPPeerlessRopemasterBouncesUpToOneTappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Their Tapped Bear", "Creature — Bear", 2, 2)
	b16Tap(g, tapped)
	untapped := b12Creature(g, opp.ID, "Their Untapped Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	msc17Play(t, g, me, game.Card{Name: "Peerless Ropemaster", TypeLine: "Creature — Human Rogue",
		ManaCost: "{4}{U}", OracleID: otjMPPeerlessRopemasterOracle, Power: 4, Toughness: 4})
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !hasID(p.PickTargetCards, tapped) || hasID(p.PickTargetCards, untapped) {
		t.Fatalf("only tapped creatures are offered: %v", p.PickTargetCards)
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 1 {
		t.Errorf("up to one: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: tapped}); err != nil {
		t.Fatalf("pick the tapped Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !handHasID(opp, tapped) {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		if tc.other != "" {
			b12Creature(g, me.ID, "Other", tc.other, 1, 1)
		}
		advanceToMain(t, g)
		raider := msc17Play(t, g, me, game.Card{Name: "Mine Raider", TypeLine: "Creature — Human Rogue",
			ManaCost: "{2}{R}", OracleID: otjMPMineRaiderOracle, Power: 3, Toughness: 2})
		if !hasEffectiveKeyword(t, g, raider, "trample") {
			t.Errorf("%q: trample is printed", tc.other)
		}
		if got := b16CountNamed(g, "Treasure"); got != tc.treasure {
			t.Errorf("other %q: %d Treasures, want %d", tc.other, got, tc.treasure)
		}
	}
}

// --- Marauding Sphinx --------------------------------------------------------

func TestOtjMPMaraudingSphinxSurveilsTwoOnTheFirstCrimeEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 10)
	sphinx := b12Push(g, me.ID, "Marauding Sphinx", "Creature — Sphinx Rogue", otjMPMaraudingSphinxOracle, 3, 5)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, sphinx, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("targeting your own creature is no crime")
	}

	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	prompt := surveilChoiceFor(g, me.ID)
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
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}

	otjB1BoltOpponent(t, g, opp.ID)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("a second crime the same turn surveilled again")
	}
}

func TestOtjMPMaraudingSphinxHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := b12Push(g, me.ID, "Marauding Sphinx", "Creature — Sphinx Rogue", otjMPMaraudingSphinxOracle, 3, 5)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, sphinx)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting the Sphinx")
	}
	if p := answerPayUnless(t, g, opp.ID, false); p.PayCost != "{2}" {
		t.Errorf("ward cost %q, want {2}", p.PayCost)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphinx) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Overzealous Muscle ------------------------------------------------------

func TestOtjMPOverzealousMuscleGainsIndestructibleFromACrimeOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muscle := b12Push(g, me.ID, "Overzealous Muscle", "Creature — Ogre Mercenary", otjMPOverzealousMuscleOracle, 5, 4)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Fatal("indestructible before any crime")
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if !hasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Fatal("a crime on your turn: indestructible")
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Error("the grant lasts until end of turn only")
	}
}

func TestOtjMPOverzealousMuscleIgnoresACrimeOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muscle := b12Push(g, opp.ID, "Overzealous Muscle", "Creature — Ogre Mercenary", otjMPOverzealousMuscleOracle, 5, 4)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: b16TargetPlayer(me.ID)}); err != nil {
		t.Fatalf("the Muscle's controller bolts you on your turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, muscle, "indestructible") {
		t.Error("a crime during another player's turn does not trigger it")
	}
}

// --- Omenport Vigilante ------------------------------------------------------

func TestOtjMPOmenportVigilanteHasDoubleStrikeAfterACrimeThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vigilante := b12Push(g, me.ID, "Omenport Vigilante", "Creature — Human Mercenary", otjMPOmenportVigilanteOracle, 2, 2)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("double strike before any crime")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("targeting your own creature is no crime")
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if !hasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Fatal("you've committed a crime this turn: double strike")
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, vigilante, "double strike") {
		t.Error("a crime last turn does not count this turn")
	}
}

// --- Nimble Brigand ----------------------------------------------------------

func TestOtjMPNimbleBrigandCanBeBlockedWithoutACrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigand := b12Push(g, me.ID, "Nimble Brigand", "Creature — Human Rogue", otjMPNimbleBrigandOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	pr7bAttack(t, g, opp.ID, brigand)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, brigand); err != nil {
		t.Errorf("with no crime this turn the Brigand can be blocked: %v", err)
	}
}

func TestOtjMPNimbleBrigandIsUnblockableAfterACrimeAndDrawsOnHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Drawn")
	brigand := b12Push(g, me.ID, "Nimble Brigand", "Creature — Human Rogue", otjMPNimbleBrigandOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	advanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	pr7bAttack(t, g, opp.ID, brigand)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, brigand); err == nil {
		t.Fatal("after a crime this turn the Brigand can't be blocked")
	}
	finishBlockDeclarations(t, g)
	passPriorityAroundTable(t, g)
	hand, life := me.Hand.Size(), opp.Life
	pr7bDamageStep(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the Brigand deals 1: %d → %d", life, opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on combat damage to a player, want 1", got)
	}
}
