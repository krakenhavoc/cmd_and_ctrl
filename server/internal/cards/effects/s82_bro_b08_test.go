package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
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
	c := layeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Gixian Infiltrator -------------------------------------------------

func TestBroB08GixianInfiltratorGrowsOnOtherSacrificesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := b27Push(g, me.ID, "Gixian Infiltrator", "Creature — Phyrexian Human", broGixianInfiltratorOracle, "{1}{B}", 2, 1, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	broSac(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, _ := broPT(t, g, inf); p != 2 {
		t.Fatalf("an opponent's sacrifice grew it: power %d", p)
	}
	broSac(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	broSac(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, inf); p != 4 || tt != 3 {
		t.Fatalf("after two other permanents: %d/%d, want 4/3", p, tt)
	}
}

func TestBroB08GixianInfiltratorIgnoresItsOwnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inf := b27Push(g, me.ID, "Gixian Infiltrator", "Creature — Phyrexian Human", broGixianInfiltratorOracle, "{1}{B}", 2, 1, "B")
	before := g.Stack.Size()
	broSac(t, g, me.ID, inf)
	if got := g.Stack.Size(); got != before {
		t.Errorf("sacrificing itself put %d triggers on the stack", got-before)
	}
}

// --- Goblin Blast-Runner ------------------------------------------------

func TestBroB08BlastRunnerNeedsYourOwnSacrificeThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	run := b27Push(g, me.ID, "Goblin Blast-Runner", "Creature — Goblin", broBlastRunnerOracle, "{R}", 1, 2, "R")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Theirs", "Artifact")

	if p, _ := broPT(t, g, run); p != 1 {
		t.Fatalf("power %d before any sacrifice", p)
	}
	if c := layeredCard(t, g, run); game.HasKeyword(&c, "menace") {
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
	if c := layeredCard(t, g, run); !game.HasKeyword(&c, "menace") {
		t.Error("no menace after my sacrifice")
	}
}

// --- Kill-Zone Acrobat --------------------------------------------------

func TestBroB08KillZoneAcrobatSacrificesForFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acro := b27Push(g, me.ID, "Kill-Zone Acrobat", "Creature — Human Soldier", broKillZoneAcrobatOracle, "{2}{B}", 3, 2, "B")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	declareAttack(t, g, opp.ID, acro)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, rock) || hasID(prompt.ChooseCards, acro) || hasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	answerOwnPermanents(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if c := layeredCard(t, g, acro); !game.HasKeyword(&c, "flying") {
		t.Error("no flying after the sacrifice")
	}
}

func TestBroB08KillZoneAcrobatDecliningGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acro := b27Push(g, me.ID, "Kill-Zone Acrobat", "Creature — Human Soldier", broKillZoneAcrobatOracle, "{2}{B}", 3, 2, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, acro)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("declining still sacrificed")
	}
	if c := layeredCard(t, g, acro); game.HasKeyword(&c, "flying") {
		t.Error("flying without a sacrifice")
	}
}

// --- Thraxodemon --------------------------------------------------------

func TestBroB08ThraxodemonSacrificesAnotherToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	th := b27Push(g, me.ID, "Thraxodemon", "Creature — Demon", broThraxodemonOracle, "{1}{B}", 2, 2, "B")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	floatMana(t, g, me, "{C}{C}{C}")

	for _, bad := range []uuid.UUID{th, land} {
		if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bad}, Strict: true}); err == nil {
			t.Fatalf("sacrificing %s was accepted", bad)
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if c := layeredCard(t, g, th); !c.Tapped {
		t.Error("the {T} was not paid")
	}
}

// --- Penregon Strongbull ------------------------------------------------

func TestBroB08StrongbullPumpsAndPingsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bull := b27Push(g, me.ID, "Penregon Strongbull", "Creature — Minotaur", broStrongbullOracle, "{2}{R}", 2, 3, "R")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, bull, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a nonartifact creature was accepted")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bull, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bull); p != 3 || tt != 4 {
		t.Errorf("%d/%d, want 3/4", p, tt)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

// --- Powerstone Fracture ------------------------------------------------

func TestBroB08PowerstoneFractureSacrificesAnArtifactOrCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	advanceToMain(t, g)

	id := b13HandCard(me, "Powerstone Fracture", "Sorcery", broFractureOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(target)}); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(target), SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land is no legal sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(land), SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("a land is no legal target")
	}
	if !g.Battlefield.Contains(rock) {
		t.Fatal("a refused cast paid its sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(target), SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(target) {
		t.Error("the artifact is sacrificed and the target destroyed")
	}
}

// --- Transmogrant Altar -------------------------------------------------

func TestBroB08AltarManaAbilityMakesThreeColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	altar := seedPermanentWithOracle(g, me.ID, "Transmogrant Altar", "Artifact", broAltarOracle)
	bear := seedCreature(g, "Bear", me.ID)
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activated without paying {B}")
	}
	floatMana(t, g, me, "{B}")
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
	if c, ok := battlefieldCard(g, altar); !ok || !c.Tapped {
		t.Error("the Altar did not tap")
	}
	if _, ok := battlefieldCard(g, bear); ok {
		t.Error("the creature was not sacrificed")
	}
}

func TestBroB08AltarMakesAZombieAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := seedPermanentWithOracle(g, me.ID, "Transmogrant Altar", "Artifact", broAltarOracle)
	bear := seedCreature(g, "Bear", me.ID)
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	zs := tokensNamed(g, "Zombie")
	if len(zs) != 1 {
		t.Fatalf("%d Zombies, want 1", len(zs))
	}
	z := zs[0]
	if z.Controller != me.ID || !z.IsArtifact() || !z.IsCreature() || z.Power != 3 || z.Toughness != 3 || len(z.Colors) != 0 || z.Tapped {
		t.Errorf("Zombie %+v, want an untapped 3/3 colorless artifact creature", z)
	}

	// Not at sorcery speed: during the opponent's turn it is refused.
	advanceToNextSeatsTurn(t, g)
	_ = opp
	extra := seedCreature(g, "Extra", me.ID)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{extra}, Strict: true}); err == nil {
		t.Error("the token ability worked off my own turn")
	}
}

// --- Ashnod, Flesh Mechanist -------------------------------------------

func TestBroB08AshnodSacrificesForATappedPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := b27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	declareAttack(t, g, opp.ID, ash)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, bear) || hasID(prompt.ChooseCards, ash) || hasID(prompt.ChooseCards, rock) {
		t.Fatalf("want a may-sacrifice over another creature: %+v", prompt)
	}
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	ps := tokensNamed(g, "Powerstone")
	if len(ps) != 1 || !ps[0].Tapped || ps[0].Controller != me.ID {
		t.Fatalf("Powerstones %+v, want one tapped one of mine", ps)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
}

func TestBroB08AshnodDecliningMakesNoPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := b27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, ash)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || len(tokensNamed(g, "Powerstone")) != 0 {
		t.Error("declining still sacrificed or made a Powerstone")
	}
}

func TestBroB08AshnodExilesACreatureCardForATappedZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ash := b27Push(g, me.ID, "Ashnod, Flesh Mechanist", "Legendary Creature — Human Artificer", broAshnodOracle, "{B}", 1, 1, "B")
	var creatureCard, spellCard uuid.UUID
	g.WithWriteLock(func() {
		creatureCard, spellCard = uuid.New(), uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: creatureCard, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: spellCard, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ash, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spellCard}, Strict: true}); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, ash, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{creatureCard}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	zs := tokensNamed(g, "Zombie")
	if len(zs) != 1 || !zs[0].Tapped || !zs[0].IsArtifact() || zs[0].Power != 3 || zs[0].Toughness != 3 {
		t.Fatalf("Zombies %+v, want one tapped 3/3 artifact creature", zs)
	}
}

// --- Transmogrant's Crown ----------------------------------------------

func TestBroB08CrownEquipsForEitherCostAndDrawsOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	crown := seedPermanentWithOracle(g, me.ID, "Transmogrant's Crown", "Artifact — Equipment", broCrownOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("equip {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bear); p != 4 || tt != 2 {
		t.Fatalf("equipped %d/%d, want 4/2", p, tt)
	}
	hand := me.Hand.Size()
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}

}

func TestBroB08CrownEquipsForTwoGenericToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	crown := seedPermanentWithOracle(g, me.ID, "Transmogrant's Crown", "Artifact — Equipment", broCrownOracle)
	bear2 := b12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, game.ActivateAbilityParams{Targets: cardRefs(bear2), Strict: true}); err != nil {
		t.Fatalf("equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := broPT(t, g, bear2); p != 4 {
		t.Errorf("equipped power %d, want 4", p)
	}
}

// --- Supply Drop --------------------------------------------------------

func TestBroB08SupplyDropPumpsAsItEntersAndDrawsWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := b13HandCard(me, "Supply Drop", "Artifact", broSupplyDropOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, bear); p != 4 || tt != 4 {
		t.Fatalf("pumped %d/%d, want 4/4", p, tt)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	hand := me.Hand.Size()
	// The artifact entered this turn, but it has no {T} symbol on a creature:
	// an artifact that is not a creature has no summoning sickness.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Supply Drop was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Farid, Enterprising Salvager --------------------------------------

func TestBroB08FaridMakesScrapForNontokenArtifactsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Farid, Enterprising Salvager", "Legendary Creature — Human Soldier", broFaridOracle, "{2}{R}", 3, 3, "R")
	mine := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	b27Kill(g, theirs)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if n := len(tokensNamed(g, "Scrap")); n != 0 {
		t.Fatalf("%d Scrap from an opponent's artifact and a nonartifact", n)
	}
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	scraps := tokensNamed(g, "Scrap")
	if len(scraps) != 1 || scraps[0].Controller != me.ID || !scraps[0].IsArtifact() || scraps[0].IsCreature() || len(scraps[0].Colors) != 0 {
		t.Fatalf("Scrap %+v, want one colorless artifact token of mine", scraps)
	}
	// A token artifact (the Scrap itself) dying makes no more.
	b27Kill(g, scraps[0].InstanceID)
	passPriorityAroundTable(t, g)
	if n := len(tokensNamed(g, "Scrap")); n != 0 {
		t.Errorf("%d Scrap after a token artifact died, want 0", n)
	}
}

func TestBroB08FaridModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farid := b27Push(g, me.ID, "Farid, Enterprising Salvager", "Legendary Creature — Human Soldier", broFaridOracle, "{2}{R}", 3, 3, "R")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Mode 1: counter and menace.
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	floatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, farid); p != 4 || tt != 4 {
		t.Errorf("Farid %d/%d, want 4/4", p, tt)
	}
	if c := layeredCard(t, g, farid); !game.HasKeyword(&c, "menace") {
		t.Error("no menace")
	}

	// Mode 2: goad.
	rock2 := b12Permanent(g, me.ID, "Rock Two", "Artifact")
	floatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock2}, Targets: cardRefs(theirs), Strict: true}); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, theirs); !c.IsGoaded() {
		t.Error("the creature was not goaded")
	}

	// Mode 3: discard, then draw.
	rock3 := b12Permanent(g, me.ID, "Rock Three", "Artifact")
	pitch := b13HandCard(me, "Old Card", "Sorcery", "")
	floatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, farid, 2, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock3}, Strict: true}); err != nil {
		t.Fatalf("mode 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, pitch)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discarding one and drawing one, want %d", got, hand)
	}
	if !graveyardHas(me, pitch) {
		t.Error("the chosen card was not discarded")
	}
}

// --- Sardian Avenger ----------------------------------------------------

func TestBroB08SardianAvengerGrowsPerOpposingArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	av := b27Push(g, me.ID, "Sardian Avenger", "Creature — Goblin Warrior", broSardianOracle, "{1}{R}", 1, 1, "R")
	b12Permanent(g, opp.ID, "Rock A", "Artifact")
	b12Permanent(g, opp.ID, "Rock B", "Artifact")
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	if c := layeredCard(t, g, av); !game.HasKeyword(&c, "first strike") || !game.HasKeyword(&c, "trample") {
		t.Fatal("missing first strike or trample")
	}
	declareAttack(t, g, opp.ID, av)
	passPriorityAroundTable(t, g)
	if p, tt := broPT(t, g, av); p != 3 || tt != 1 {
		t.Errorf("attacking %d/%d, want 3/1", p, tt)
	}
}

func TestBroB08SardianAvengerPingsTheArtifactsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Sardian Avenger", "Creature — Goblin Warrior", broSardianOracle, "{1}{R}", 1, 1, "R")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	oppLife, myLife := opp.Life, me.Life
	b27Kill(g, mine)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Fatalf("life changed on my artifact or their nonartifact: me %d opp %d", me.Life, opp.Life)
	}
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife {
		t.Errorf("life me %d opp %d, want me %d opp %d", me.Life, opp.Life, myLife, oppLife-1)
	}
}

func graveyardHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
