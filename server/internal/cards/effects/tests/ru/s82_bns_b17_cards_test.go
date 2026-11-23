package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b17_cards_test.go — Ravnica: Clue Edition batch 17 (S82, #5929),
// played through the real catalog entries.

const (
	bnsTithebearerGiantOracle = "475aaf7c-b0f8-40f3-ad91-3e7d414de933"
	bnsToweringThunderOracle  = "d68adf9b-57b1-4c14-97b1-c5906dc5d29b"
	bnsTransluminantOracle    = "f4d54a89-8409-4fbc-b01f-3b03f352820f"
	bnsTribuneOfRotOracle     = "cd1b3ef3-73ac-4f06-8cfb-5f2388481379"
	bnsTrustedPegasusOracle   = "18729459-8669-4c56-a78d-fd5982c3b2ac"
	bnsTurretOgreOracle       = "fb4a044b-f25d-47ff-b272-754e577ce50a"
	bnsTwilightPantherOracle  = "5e19aa52-5b76-40d7-ba9c-0e5d64f2b773"
	bnsUndercoverButlerOracle = "ff135dc7-31c8-41d6-a5d5-626217844d09"
	bnsUnrulyKrasisOracle     = "8420f73d-fdcf-4425-b9a7-c9d359fc028d"
	bnsUrbisProtectorOracle   = "564dd5e7-6c1e-4959-a594-60e08552aec9"
	bnsUtvaraScalperOracle    = "58e754ad-88b4-47c5-b105-8b50e54143c8"
	bnsVenomousHierophantOrcl = "e235d1ca-3099-46c1-8f33-798286292d71"
)

func TestBnsB17AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		bnsTithebearerGiantOracle: "Tithebearer Giant", bnsToweringThunderOracle: "Towering Thunderfist",
		bnsTransluminantOracle: "Transluminant", bnsTribuneOfRotOracle: "Tribune of Rot",
		bnsTrustedPegasusOracle: "Trusted Pegasus", bnsTurretOgreOracle: "Turret Ogre",
		bnsTwilightPantherOracle: "Twilight Panther", bnsUndercoverButlerOracle: "Undercover Butler",
		bnsUnrulyKrasisOracle: "Unruly Krasis", bnsUrbisProtectorOracle: "Urbis Protector",
		bnsUtvaraScalperOracle: "Utvara Scalper", bnsVenomousHierophantOrcl: "Venomous Hierophant",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func bnsMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func bnsGraveyardSize(p *game.Player) int { return len(p.Graveyard.Cards) }

// --- Tithebearer Giant -----------------------------------------------------

func TestBnsB17TithebearerGiantDrawsAndLosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"))
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Tithebearer Giant", "Creature — Giant Warrior", bnsTithebearerGiantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		// the Giant joined the hand and left it as it was cast; the draw added one
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Urbis Protector -------------------------------------------------------

func TestBnsB17UrbisProtectorMakesAFourFourAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Urbis Protector", "Creature — Human Cleric", bnsUrbisProtectorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Angel"); got != 1 {
		t.Fatalf("%d Angels, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" {
			if effectstest.EffectivePower(t, g, c.InstanceID) != 4 || effectstest.EffectiveToughness(t, g, c.InstanceID) != 4 ||
				!effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("the Angel is not a 4/4 flier")
			}
		}
	}
}

// --- Venomous Hierophant ---------------------------------------------------

func TestBnsB17VenomousHierophantMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("D"))
	before := bnsGraveyardSize(me)
	effectstest.CastCatalogSpell(t, g, "Venomous Hierophant", "Creature — Gorgon Cleric", bnsVenomousHierophantOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := bnsGraveyardSize(me) - before; got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	if len(me.Library.Cards) != 1 {
		t.Errorf("library %d, want 1", len(me.Library.Cards))
	}
}

// --- Tribune of Rot --------------------------------------------------------

func TestBnsB17TribuneOfRotMakesASaprolingPerMilledCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tribune := effectstest.B12Push(g, me.ID, "Tribune of Rot", "Creature — Elf Shaman", bnsTribuneOfRotOracle, 3, 3)
	effectstest.Mm2Library(me, effectstest.Mm2CreatureCard("Bear", 2), effectstest.Mm2Land("Forest"), effectstest.Mm2CreatureCard("Ox", 3))
	effectstest.AttackWith(t, g, opp.ID, tribune)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); got != 1 {
		t.Errorf("%d Saprolings, want 1 (one creature among the two milled)", got)
	}
}

func TestBnsB17TribuneOfRotNoCreatureNoSaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tribune := effectstest.B12Push(g, me.ID, "Tribune of Rot", "Creature — Elf Shaman", bnsTribuneOfRotOracle, 3, 3)
	effectstest.Mm2Library(me, effectstest.Mm2Land("Forest"), effectstest.Mm2Spell("Shock"), effectstest.Mm2CreatureCard("Ox", 3))
	effectstest.AttackWith(t, g, opp.ID, tribune)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); got != 0 {
		t.Errorf("%d Saprolings, want 0", got)
	}
	if got := bnsGraveyardSize(me); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

// --- Trusted Pegasus -------------------------------------------------------

func TestBnsB17TrustedPegasusGrantsFlyingToAnAttackerWithout(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peg := effectstest.B12Push(g, me.ID, "Trusted Pegasus", "Creature — Pegasus", bnsTrustedPegasusOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the bear starts with flying")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{peg, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the attacking bear did not gain flying")
	}
}

func TestBnsB17TrustedPegasusIgnoresANonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peg := effectstest.B12Push(g, me.ID, "Trusted Pegasus", "Creature — Pegasus", bnsTrustedPegasusOracle, 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, peg)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, home, "flying") {
		t.Error("a creature that did not attack gained flying")
	}
}

// --- Turret Ogre -----------------------------------------------------------

func TestBnsB17TurretOgreDealsTwoWithAnotherPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		before[s.ID] = s.Life
	}
	effectstest.CastCatalogSpell(t, g, "Turret Ogre", "Creature — Ogre Warrior", bnsTurretOgreOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != before[s.ID]-2 {
			t.Errorf("opponent at %d, want %d", s.Life, before[s.ID]-2)
		}
	}
}

func TestBnsB17TurretOgreNeedsPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	before := g.Seats[1].Life
	effectstest.CastCatalogSpell(t, g, "Turret Ogre", "Creature — Ogre Warrior", bnsTurretOgreOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before {
		t.Errorf("opponent lost life with only a 3-power creature")
	}
}

// --- Twilight Panther / Towering Thunderfist -------------------------------

func TestBnsB17TwilightPantherGainsDeathtouchForB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Twilight Panther", "Creature — Cat Spirit", bnsTwilightPantherOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	bnsMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("no deathtouch")
	}
}

func TestBnsB17ToweringThunderfistGainsVigilanceForW(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Towering Thunderfist", "Creature — Giant Soldier", bnsToweringThunderOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	bnsMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {W} with red mana")
	}
	bnsMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Transluminant ---------------------------------------------------------

func TestBnsB17TransluminantSacrificesNowAndMakesASpiritAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Transluminant", "Creature — Dryad Shaman", bnsTransluminantOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {W}")
	}
	bnsMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, id) {
		t.Error("the Dryad was not sacrificed")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 0 {
		t.Fatalf("%d Spirits before the end step, want 0", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 1 {
		t.Fatalf("%d Spirits after the end step began, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Error("the Spirit has no flying")
		}
	}
}

// --- Undercover Butler -----------------------------------------------------

func bnsButlerAttack(t *testing.T, life1, life2, life3 int, defender int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life = life1, life2, life3
	id := effectstest.B12Push(g, me.ID, "Undercover Butler", "Creature — Human Rogue", bnsUndercoverButlerOracle, 2, 3)
	effectstest.AttackWith(t, g, g.Seats[defender].ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	return g, id
}

func TestBnsB17UndercoverButlerUnblockableAgainstTheLifeLeader(t *testing.T) {
	g, id := bnsButlerAttack(t, 20, 40, 30, 1)
	if effectstest.RestrictionsOf(t, g, id)&game.CantBeBlocked == 0 {
		t.Error("the Butler can be blocked after attacking the player with the most life")
	}
}

func TestBnsB17UndercoverButlerTiedForMostLifeCounts(t *testing.T) {
	g, id := bnsButlerAttack(t, 20, 40, 40, 2)
	if effectstest.RestrictionsOf(t, g, id)&game.CantBeBlocked == 0 {
		t.Error("tied for most life should count")
	}
}

func TestBnsB17UndercoverButlerBlockableAgainstOthers(t *testing.T) {
	g, id := bnsButlerAttack(t, 20, 40, 30, 2)
	if effectstest.RestrictionsOf(t, g, id)&game.CantBeBlocked != 0 {
		t.Error("the Butler attacked a player without the most life and still cannot be blocked")
	}
}

// --- Utvara Scalper --------------------------------------------------------

func TestBnsB17UtvaraScalperMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Utvara Scalper", "Creature — Goblin Scout", bnsUtvaraScalperOracle, 1, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != id {
		t.Fatalf("pass with the Scalper home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Scalper attacking: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
}

// --- Unruly Krasis ---------------------------------------------------------

func TestBnsB17UnrulyKrasisSetsAnotherCreaturesBaseSizeToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krasis := effectstest.B12Push(g, me.ID, "Unruly Krasis", "Creature — Shark Octopus Lizard", bnsUnrulyKrasisOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sncCounterForBns(g, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(krasis, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// base 4/4 plus the +1/+1 counter
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 5 {
		t.Errorf("the bear has power %d, want 5 (base 4 plus its counter)", p)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, krasis); p != 4 {
		t.Errorf("the Krasis itself changed to %d", p)
	}
}

func TestBnsB17UnrulyKrasisDeclinedLeavesTheSizeAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krasis := effectstest.B12Push(g, me.ID, "Unruly Krasis", "Creature — Shark Octopus Lizard", bnsUnrulyKrasisOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(krasis, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 2 {
		t.Errorf("the bear is %d-power after declining, want 2", p)
	}
}

func TestBnsB17UnrulyKrasisAdaptsOnlyWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	krasis := effectstest.B12Push(g, me.ID, "Unruly Krasis", "Creature — Shark Octopus Lizard", bnsUnrulyKrasisOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	bnsMana(me, "C", "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, krasis, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {3}{G}{U} without blue")
	}
	bnsMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, krasis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, krasis); p != 7 {
		t.Fatalf("power %d after adapt, want 7", p)
	}
	bnsMana(me, "C", "C", "C", "G", "U")
	if err := g.ActivateCatalogAbility(me.ID, krasis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, krasis); p != 7 {
		t.Errorf("power %d after a second adapt, want still 7", p)
	}
}

func sncCounterForBns(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
}
