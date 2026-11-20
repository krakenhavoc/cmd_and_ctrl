package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b08_test.go — The Brothers' War, batch 8 (S82, #5482): the
// sacrifice cards, the 3/3 Zombie artifact token and Scrap. Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	broGixianInfiltratorOracle = "2497926c-b253-4400-b540-776ee2cd807c"
	broBlastRunnerOracle       = "8b69da01-3ba5-4a61-9e26-e879b6d8b63e"
	broKillZoneAcrobatOracle   = "7c3982c5-6af3-4937-8d1f-e006f4e89342"
	broThraxodemonOracle       = "01cf689c-5a64-4ecf-b45d-3f385c9b212b"
	broStrongbullOracle        = "a4e34111-aefd-4c61-a6e3-ac8895799c3f"
	broFractureOracle          = "f13baf38-27fc-4c5d-8fee-459eefdb32da"
	broAltarOracle             = "239fea8d-c49d-40bc-96e1-b84d328a0e1e"
	broAshnodOracle            = "7dfae95e-f67e-49e3-b68a-40ea7e5099dd"
	broCrownOracle             = "767c82a7-ed42-4ef3-9d79-dd95e48f4326"
	broSupplyDropOracle        = "9594b8e0-a176-447a-a52e-e1a77175b748"
	broFaridOracle             = "83cb9ab7-6a8f-40b5-b076-6fa669ff15f2"
	broSardianOracle           = "6db90f86-205a-4e0b-944b-74742bc4e59d"
)

func TestBroB08EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broGixianInfiltratorOracle: "Gixian Infiltrator",
		broBlastRunnerOracle:       "Goblin Blast-Runner",
		broKillZoneAcrobatOracle:   "Kill-Zone Acrobat",
		broThraxodemonOracle:       "Thraxodemon",
		broStrongbullOracle:        "Penregon Strongbull",
		broFractureOracle:          "Powerstone Fracture",
		broAltarOracle:             "Transmogrant Altar",
		broAshnodOracle:            "Ashnod, Flesh Mechanist",
		broCrownOracle:             "Transmogrant's Crown",
		broSupplyDropOracle:        "Supply Drop",
		broFaridOracle:             "Farid, Enterprising Salvager",
		broSardianOracle:           "Sardian Avenger",
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

func broSac(t *testing.T, g *game.Game, who, id uuid.UUID) {
	t.Helper()
	if err := g.SacrificePermanent(who, id); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
}

func broPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Gixian Infiltrator -------------------------------------------------

func TestBroB08GixianInfiltratorGrowsOnOtherSacrificesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := effectstest.B27Push(g, me.ID, "Gixian Infiltrator", "Creature — Phyrexian Human", broGixianInfiltratorOracle, "{1}{B}", 2, 1, "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	broSac(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := broPT(t, g, inf); p != 2 {
		t.Fatalf("an opponent's sacrifice grew it: power %d", p)
	}
	broSac(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	broSac(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, inf); p != 4 || tt != 3 {
		t.Fatalf("after two other permanents: %d/%d, want 4/3", p, tt)
	}
}

func TestBroB08GixianInfiltratorIgnoresItsOwnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inf := effectstest.B27Push(g, me.ID, "Gixian Infiltrator", "Creature — Phyrexian Human", broGixianInfiltratorOracle, "{1}{B}", 2, 1, "B")
	before := g.Stack.Size()
	broSac(t, g, me.ID, inf)
	if got := g.Stack.Size(); got != before {
		t.Errorf("sacrificing itself put %d triggers on the stack", got-before)
	}
}

// --- Goblin Blast-Runner ------------------------------------------------

func TestBroB08BlastRunnerNeedsYourOwnSacrificeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	run := effectstest.B27Push(g, me.ID, "Goblin Blast-Runner", "Creature — Goblin", broBlastRunnerOracle, "{R}", 1, 2, "R")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Theirs", "Artifact")

	if p, _ := broPT(t, g, run); p != 1 {
		t.Fatalf("power %d before any sacrifice", p)
	}
	if c := effectstest.LayeredCard(t, g, run); game.HasKeyword(&c, "menace") {
		t.Fatal("menace before any sacrifice")
	}
	broSac(t, g, opp.ID, theirs)
	if p, _ := broPT(t, g, run); p != 1 {
		t.Errorf("an opponent's sacrifice pumped it: %d", p)
	}
	broSac(t, g, me.ID, rock)
	if p, tt := broPT(t, g, run); p != 3 || tt != 2 {
		t.Errorf("after my sacrifice: %d/%d, want 3/2", p, tt)
	}
	if c := effectstest.LayeredCard(t, g, run); !game.HasKeyword(&c, "menace") {
		t.Error("no menace after my sacrifice")
	}
}

// --- Kill-Zone Acrobat --------------------------------------------------

func TestBroB08KillZoneAcrobatSacrificesForFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acro := effectstest.B27Push(g, me.ID, "Kill-Zone Acrobat", "Creature — Human Soldier", broKillZoneAcrobatOracle, "{2}{B}", 3, 2, "B")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.DeclareAttack(t, g, opp.ID, acro)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, rock) || effectstest.HasID(prompt.ChooseCards, acro) || effectstest.HasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if c := effectstest.LayeredCard(t, g, acro); !game.HasKeyword(&c, "flying") {
		t.Error("no flying after the sacrifice")
	}
}

func TestBroB08KillZoneAcrobatDecliningGivesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acro := effectstest.B27Push(g, me.ID, "Kill-Zone Acrobat", "Creature — Human Soldier", broKillZoneAcrobatOracle, "{2}{B}", 3, 2, "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, acro)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("declining still sacrificed")
	}
	if c := effectstest.LayeredCard(t, g, acro); game.HasKeyword(&c, "flying") {
		t.Error("flying without a sacrifice")
	}
}

// --- Thraxodemon --------------------------------------------------------

func TestBroB08ThraxodemonSacrificesAnotherToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	th := effectstest.B27Push(g, me.ID, "Thraxodemon", "Creature — Demon", broThraxodemonOracle, "{1}{B}", 2, 2, "B")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")

	for _, bad := range []uuid.UUID{th, land} {
		if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bad}, Strict: true}); err == nil {
			t.Fatalf("sacrificing %s was accepted", bad)
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if c := effectstest.LayeredCard(t, g, th); !c.Tapped {
		t.Error("the {T} was not paid")
	}
}

// --- Penregon Strongbull ------------------------------------------------

func TestBroB08StrongbullPumpsAndPingsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bull := effectstest.B27Push(g, me.ID, "Penregon Strongbull", "Creature — Minotaur", broStrongbullOracle, "{2}{R}", 2, 3, "R")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, bull, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a nonartifact creature was accepted")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bull, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bull); p != 3 || tt != 4 {
		t.Errorf("%d/%d, want 3/4", p, tt)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

// --- Powerstone Fracture ------------------------------------------------

func TestBroB08PowerstoneFractureSacrificesAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.AdvanceToMain(t, g)

	id := effectstest.B13HandCard(me, "Powerstone Fracture", "Sorcery", broFractureOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(target)}); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(target), SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land is no legal sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(land), SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("a land is no legal target")
	}
	if !g.Battlefield.Contains(rock) {
		t.Fatal("a refused cast paid its sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(target), SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(target) {
		t.Error("the artifact is sacrificed and the target destroyed")
	}
}

// --- Transmogrant Altar -------------------------------------------------

func TestBroB08AltarManaAbilityMakesThreeColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Transmogrant Altar", "Artifact", broAltarOracle)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activated without paying {B}")
	}
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{altar}}); err == nil {
		t.Fatal("sacrificing the Altar to its own cost was accepted")
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool holds %d, want 3", me.ManaPool.Total())
	}
	for i, u := range me.ManaPool.Units() {
		if u.Color != "C" {
			t.Errorf("unit %d is %q, want C", i, u.Color)
		}
	}
	if c, ok := effectstest.BattlefieldCard(g, altar); !ok || !c.Tapped {
		t.Error("the Altar did not tap")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the creature was not sacrificed")
	}
}

func TestBroB08AltarMakesAZombieAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Transmogrant Altar", "Artifact", broAltarOracle)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	zs := effectstest.TokensNamed(g, "Zombie")
	if len(zs) != 1 {
		t.Fatalf("%d Zombies, want 1", len(zs))
	}
	z := zs[0]
	if z.Controller != me.ID || !z.IsArtifact() || !z.IsCreature() || z.Power != 3 || z.Toughness != 3 || len(z.Colors) != 0 || z.Tapped {
		t.Errorf("Zombie %+v, want an untapped 3/3 colorless artifact creature", z)
	}

	// Not at sorcery speed: during the opponent's turn it is refused.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	_ = opp
	extra := effectstest.SeedCreature(g, "Extra", me.ID)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{extra}, Strict: true}); err == nil {
		t.Error("the token ability worked off my own turn")
	}
}

// --- Ashnod, Flesh Mechanist -------------------------------------------

func TestBroB08AshnodSacrificesForATappedPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := effectstest.B27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, ash)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, bear) || effectstest.HasID(prompt.ChooseCards, ash) || effectstest.HasID(prompt.ChooseCards, rock) {
		t.Fatalf("want a may-sacrifice over another creature: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	ps := effectstest.TokensNamed(g, "Powerstone")
	if len(ps) != 1 || !ps[0].Tapped || ps[0].Controller != me.ID {
		t.Fatalf("Powerstones %+v, want one tapped one of mine", ps)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
}

func TestBroB08AshnodDecliningMakesNoPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := effectstest.B27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, ash)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || len(effectstest.TokensNamed(g, "Powerstone")) != 0 {
		t.Error("declining still sacrificed or made a Powerstone")
	}
}

func TestBroB08AshnodExilesACreatureCardForATappedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ash := effectstest.B27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	var creatureCard, spellCard uuid.UUID
	g.WithWriteLock(func() {
		creatureCard, spellCard = uuid.New(), uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: creatureCard, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: spellCard, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ash, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spellCard}, Strict: true}); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, ash, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{creatureCard}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	zs := effectstest.TokensNamed(g, "Zombie")
	if len(zs) != 1 || !zs[0].Tapped || !zs[0].IsArtifact() || zs[0].Power != 3 || zs[0].Toughness != 3 {
		t.Fatalf("Zombies %+v, want one tapped 3/3 artifact creature", zs)
	}
}

// --- Transmogrant's Crown ----------------------------------------------

func TestBroB08CrownEquipsForEitherCostAndDrawsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	crown := effectstest.SeedPermanentWithOracle(g, me.ID, "Transmogrant's Crown", "Artifact — Equipment", broCrownOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("equip {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bear); p != 4 || tt != 2 {
		t.Fatalf("equipped %d/%d, want 4/2", p, tt)
	}
	hand := me.Hand.Size()
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}

}

func TestBroB08CrownEquipsForTwoGenericToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	crown := effectstest.SeedPermanentWithOracle(g, me.ID, "Transmogrant's Crown", "Artifact — Equipment", broCrownOracle)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear2), Strict: true}); err != nil {
		t.Fatalf("equip {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := broPT(t, g, bear2); p != 4 {
		t.Errorf("equipped power %d, want 4", p)
	}
}

// --- Supply Drop --------------------------------------------------------

func TestBroB08SupplyDropPumpsAsItEntersAndDrawsWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B13HandCard(me, "Supply Drop", "Artifact", broSupplyDropOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bear); p != 4 || tt != 4 {
		t.Fatalf("pumped %d/%d, want 4/4", p, tt)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	hand := me.Hand.Size()
	// The artifact entered this turn, but it has no {T} symbol on a creature:
	// an artifact that is not a creature has no summoning sickness.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Supply Drop was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Farid, Enterprising Salvager --------------------------------------

func TestBroB08FaridMakesScrapForNontokenArtifactsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Farid, Enterprising Salvager", "Legendary Creature — Human Soldier", broFaridOracle, "{2}{R}", 3, 3, "R")
	mine := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.B27Kill(g, theirs)
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TokensNamed(g, "Scrap")); n != 0 {
		t.Fatalf("%d Scrap from an opponent's artifact and a nonartifact", n)
	}
	effectstest.B27Kill(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	scraps := effectstest.TokensNamed(g, "Scrap")
	if len(scraps) != 1 || scraps[0].Controller != me.ID || !scraps[0].IsArtifact() || scraps[0].IsCreature() || len(scraps[0].Colors) != 0 {
		t.Fatalf("Scrap %+v, want one colorless artifact token of mine", scraps)
	}
	// A token artifact (the Scrap itself) dying makes no more.
	effectstest.B27Kill(g, scraps[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TokensNamed(g, "Scrap")); n != 0 {
		t.Errorf("%d Scrap after a token artifact died, want 0", n)
	}
}

func TestBroB08FaridModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farid := effectstest.B27Push(g, me.ID, "Farid, Enterprising Salvager", "Legendary Creature — Human Soldier", broFaridOracle, "{2}{R}", 3, 3, "R")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Mode 1: counter and menace.
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, farid); p != 4 || tt != 4 {
		t.Errorf("Farid %d/%d, want 4/4", p, tt)
	}
	if c := effectstest.LayeredCard(t, g, farid); !game.HasKeyword(&c, "menace") {
		t.Error("no menace")
	}

	// Mode 2: goad.
	rock2 := effectstest.B12Permanent(g, me.ID, "Rock Two", "Artifact")
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock2}, Targets: effectstest.CardRefs(theirs), Strict: true}); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, theirs); !c.IsGoaded() {
		t.Error("the creature was not goaded")
	}

	// Mode 3: discard, then draw.
	rock3 := effectstest.B12Permanent(g, me.ID, "Rock Three", "Artifact")
	pitch := effectstest.B13HandCard(me, "Old Card", "Sorcery", "")
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 2, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock3}, Strict: true}); err != nil {
		t.Fatalf("mode 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discarding one and drawing one, want %d", got, hand)
	}
	if !effectstest.GraveyardHas(me, pitch) {
		t.Error("the chosen card was not discarded")
	}
}

// --- Sardian Avenger ----------------------------------------------------

func TestBroB08SardianAvengerGrowsPerOpposingArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	av := effectstest.B27Push(g, me.ID, "Sardian Avenger", "Creature — Goblin Warrior", broSardianOracle, "{1}{R}", 1, 1, "R")
	effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Rock B", "Artifact")
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if c := effectstest.LayeredCard(t, g, av); !game.HasKeyword(&c, "first strike") || !game.HasKeyword(&c, "trample") {
		t.Fatal("missing first strike or trample")
	}
	effectstest.DeclareAttack(t, g, opp.ID, av)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, av); p != 3 || tt != 1 {
		t.Errorf("attacking %d/%d, want 3/1", p, tt)
	}
}

func TestBroB08SardianAvengerPingsTheArtifactsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Sardian Avenger", "Creature — Goblin Warrior", broSardianOracle, "{1}{R}", 1, 1, "R")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	oppLife, myLife := opp.Life, me.Life
	effectstest.B27Kill(g, mine)
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Fatalf("life changed on my artifact or their nonartifact: me %d opp %d", me.Life, opp.Life)
	}
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife {
		t.Errorf("life me %d opp %d, want me %d opp %d", me.Life, opp.Life, myLife, oppLife-1)
	}
}
