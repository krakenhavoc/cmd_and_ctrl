package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b3_cards_test.go — Marvel Super Heroes, half B, third batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB3CaptainMarvelOracle = "2e665a03-bece-43df-bd15-b4496e233f2e"
	mshB3DarkFortressOracle  = "40760bfa-a423-487c-ba29-043b2d00c736"
	mshB3BastionOracle       = "7785ffd4-f169-475d-9558-ce4877b3378a"
	mshB3DecoyPloyOracle     = "035cddbc-39c9-42af-9e26-c88419b43b49"
	mshB3DepowerOracle       = "33f89437-885b-4850-88b7-239f5dfcc8b1"
	mshB3DoomOracle          = "9e4fa592-71f8-4c71-8c91-78e6326cd736"
	mshB3EchoOracle          = "1091c83a-a4b9-457f-a69a-9a44a7e10759"
	mshB3EpicFightOracle     = "4bdfc580-de51-445b-b170-8c6832e44704"
	mshB3HarnessOracle       = "e69db5f0-beed-48e8-bc25-768b1e8be5ea"
	mshB3FinFangOracle       = "52ec472e-a745-4e25-b9a7-8e1eb8873594"
	mshB3ForgeOracle         = "0e5f97b5-bb8c-4688-a141-5db4d4eeccf0"
	mshB3GhostOracle         = "88f34d15-c338-40a7-b56c-c5b650847430"
)

func TestMshB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB3CaptainMarvelOracle: "Captain Marvel, Earth's Protector",
		mshB3DarkFortressOracle:  "Dark Fortress",
		mshB3BastionOracle:       "Gleaming Bastion",
		mshB3DecoyPloyOracle:     "Decoy Ploy",
		mshB3DepowerOracle:       "Depower",
		mshB3DoomOracle:          "Doctor Doom",
		mshB3EchoOracle:          "Echo, Perceptive Prodigy",
		mshB3EpicFightOracle:     "Epic Fight",
		mshB3HarnessOracle:       "Falcon's Wing Harness",
		mshB3FinFangOracle:       "Fin Fang Foom",
		mshB3ForgeOracle:         "Futurist Forge",
		mshB3GhostOracle:         "Ghost, Spectral Saboteur",
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

// --- Captain Marvel, Earth's Protector ----------------------------------------

func TestMshB3CaptainMarvelPowerUpPutsBothCountersOnceAndCostsLessTheTurnSheEntered(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	cm := castPowerUpCreature(t, g, me, "Captain Marvel, Earth's Protector", "Legendary Creature — Human Kree Hero",
		"{3}{W}{W}", "{C}{C}{C}{W}{W}", mshB3CaptainMarvelOracle, 5)
	for _, kw := range []string{"flash", "flying", "lifelink"} {
		if !hasEffectiveKeyword(t, g, cm, kw) {
			t.Errorf("Captain Marvel lacks %s", kw)
		}
	}
	if got := chargedPowerUp(t, g, me.ID, cm); got != "{2}" {
		t.Fatalf("the turn she entered: %q, want {2}", got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cm, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, cm); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := countersOn(g, cm, game.CounterIndestructible); got != 1 {
		t.Errorf("indestructible counters = %d, want 1", got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cm, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshB3CaptainMarvelPowerUpCostsItsPrintedPriceLater(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	cm := castPowerUpCreature(t, g, me, "Captain Marvel, Earth's Protector", "Legendary Creature — Human Kree Hero",
		"{3}{W}{W}", "{C}{C}{C}{W}{W}", mshB3CaptainMarvelOracle, 5)
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, cm); got != "{5}{W}{W}" {
		t.Errorf("a later turn: %q, want {5}{W}{W}", got)
	}
}

// --- Ghost, Spectral Saboteur ---------------------------------------------------

func TestMshB3GhostCannotBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghost := b12Push(g, me.ID, "Ghost, Spectral Saboteur", "Legendary Creature — Human Rogue Villain", mshB3GhostOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	brAttack(t, g, ghost, bear)
	brRefusal(t, g.DeclareBlocker(wall, ghost), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Errorf("the other attacker is blockable: %v", err)
	}
}

// --- Futurist Forge -------------------------------------------------------------

func TestMshB3FuturistForgeDrawsWhenItEntersAndTwoWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	forge := castCatalogSpell(t, g, "Futurist Forge", "Artifact", mshB3ForgeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("after entering: hand %d, want %d", got, hand+1)
	}
	floatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(forge) {
		t.Error("sacrificing is a cost: the Forge is gone before the ability resolves")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("after the sacrifice: hand %d, want %d", got, hand+3)
	}
}

// --- Falcon's Wing Harness ------------------------------------------------------

func TestMshB3HarnessAttachesOnEntryAndGrantsItsStats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	harness := castCatalogSpell(t, g, "Falcon's Wing Harness", "Artifact — Equipment", mshB3HarnessOracle, nil)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, harness); host.ID != bear {
		t.Fatalf("the Harness did not attach to the Bear: %+v", host)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped P/T %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the equipped creature has flying")
	}
}

func TestMshB3HarnessGivesWardOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	harness := b12Push(g, me.ID, "Falcon's Wing Harness", "Artifact — Equipment", mshB3HarnessOracle, 0, 0)
	equipTo(t, g, me.ID, harness, bear)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward counters the spell: the Bear should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is in its owner's graveyard")
	}
}

// --- Dark Fortress and Gleaming Bastion -----------------------------------------

func mshB3LandProduces(t *testing.T, g *game.Game, land uuid.UUID, color string) error {
	t.Helper()
	me := g.Seats[0]
	err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{})
	if err != nil {
		return err
	}
	b10ResolveAllManaPicks(t, g, me.ID, color)
	found := false
	for _, c := range batch01PoolColors(me) {
		if c == color {
			found = true
		}
	}
	if !found {
		t.Errorf("the pool %v lacks %s", batch01PoolColors(me), color)
	}
	return nil
}

func TestMshB3DualLandsWorkTheTurnTheyEnterAndOtherwiseNeedABasicLand(t *testing.T) {
	for _, tc := range []struct{ name, oracle, color string }{
		{"Dark Fortress", mshB3DarkFortressOracle, "R"},
		{"Gleaming Bastion", mshB3BastionOracle, "U"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			land := playLandFromHand(t, g, tc.name, tc.oracle)
			if err := mshB3LandProduces(t, g, land, tc.color); err != nil {
				t.Fatalf("the turn it entered: %v", err)
			}

			g2 := newCatalogGame(t)
			me2 := g2.Seats[0]
			land2 := pushCatalogPermanent(g2, me2.ID, tc.name, "Land", tc.oracle, false)
			// Not entered this turn (the push emitted an entry, so age it
			// by passing a full turn cycle), and no basic land yet.
			powerUpPassTo(t, g2, 0)
			if err := g2.ActivateManaAbility(me2.ID, land2, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
				t.Fatalf("no basic land and not new: err = %v, want ErrConditionNotMet", err)
			}
			if err := g2.ActivateManaAbility(me2.ID, land2, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("{T}: Add {C} is unconditional: %v", err)
			}
			_ = me
			untapForTest(g2, land2)
			b12Permanent(g2, me2.ID, "Forest", "Basic Land — Forest")
			if err := mshB3LandProduces(t, g2, land2, tc.color); err != nil {
				t.Errorf("with a basic land: %v", err)
			}
		})
	}
}

// --- Decoy Ploy -----------------------------------------------------------------

func TestMshB3DecoyPloyReturnsAVillainAndAHeroWhenBothAreChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	villain := seedGraveyardCard(t, g, "Bad Guy", "Creature — Human Villain", "")
	hero := seedGraveyardCard(t, g, "Good Guy", "Creature — Human Hero", "")
	bear := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")

	b02cCastModalSpell(t, g, "Decoy Ploy", "Instant", mshB3DecoyPloyOracle, []int{0, 1}, s58Refs(villain, hero))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(villain) || !me.Hand.Contains(hero) {
		t.Errorf("both cards return: villain %v hero %v", me.Hand.Contains(villain), me.Hand.Contains(hero))
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear stays")
	}
}

func TestMshB3DecoyPloyChoosesOneAndRefusesTheWrongType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	villain := seedGraveyardCard(t, g, "Bad Guy", "Creature — Human Villain", "")
	hero := seedGraveyardCard(t, g, "Good Guy", "Creature — Human Hero", "")

	b02cCastModalSpell(t, g, "Decoy Ploy", "Instant", mshB3DecoyPloyOracle, []int{1}, s58Refs(hero))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hero) || me.Hand.Contains(villain) {
		t.Errorf("only the Hero returns: hero %v villain %v", me.Hand.Contains(hero), me.Hand.Contains(villain))
	}

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Decoy Ploy", TypeLine: "Instant", OracleID: mshB3DecoyPloyOracle,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: s58Refs(villain)}); err == nil {
		t.Error("the Hero bullet accepted a Villain")
	}
}

// --- Depower --------------------------------------------------------------------

func TestMshB3DepowerCostsTwoLessAtAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	brAttack(t, g, attacker)
	if got := hobB3Price(t, g, opp, "Depower", "Instant", "{2}{U}", attacker); got != 1 {
		t.Errorf("at an attacker: %d, want 1", got)
	}
	if got := hobB3Price(t, g, opp, "Depower", "Instant", "{2}{U}", idle); got != 3 {
		t.Errorf("at a creature that is not attacking: %d, want 3", got)
	}
}

func TestMshB3DepowerShrinksPowerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Depower", "Instant", mshB3DepowerOracle, s58Refs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != -1 || tough != 3 {
		t.Errorf("P/T %d/%d, want -1/3", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Doctor Doom ----------------------------------------------------------------

func TestMshB3DoctorDoomMakesTwoDoombotsAndIsIndestructibleWithThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	doom := castCatalogSpell(t, g, "Doctor Doom", "Legendary Creature — Human Scientist Villain", mshB3DoomOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Doombot" {
			n++
			got := gaLayered(t, g, c.InstanceID)
			if got.Power != 3 || got.Toughness != 3 || !got.IsArtifact() || !got.HasSubtype("Robot") || !got.HasSubtype("Villain") || len(got.Colors) != 0 {
				t.Errorf("Doombot is a 3/3 colorless Robot Villain artifact creature: %+v", got)
			}
		}
	}
	if n != 2 {
		t.Fatalf("Doombot tokens = %d, want 2", n)
	}
	if !hasEffectiveKeyword(t, g, doom, "indestructible") {
		t.Error("Doom is indestructible while he controls an artifact creature")
	}
}

func TestMshB3DoctorDoomIndestructibleOnlyWithAnArtifactCreatureOrAPlan(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := b12Push(g, me.ID, "Doctor Doom", "Legendary Creature — Human Scientist Villain", mshB3DoomOracle, 3, 3)
	if hasEffectiveKeyword(t, g, doom, "indestructible") {
		t.Fatal("alone, Doom is not indestructible")
	}
	b12Push(g, me.ID, "Plain Artifact", "Artifact", "", 0, 0)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Push(g, opp.ID, "Their Robot", "Artifact Creature — Robot", "", 2, 2)
	if hasEffectiveKeyword(t, g, doom, "indestructible") {
		t.Fatal("a plain artifact, a plain creature and an opponent's artifact creature do not count")
	}
	b12Permanent(g, me.ID, "Some Plan", "Enchantment — Plan")
	if !hasEffectiveKeyword(t, g, doom, "indestructible") {
		t.Error("a Plan you control gives indestructible")
	}

	g2 := newCatalogGame(t)
	doom2 := b12Push(g2, g2.Seats[0].ID, "Doctor Doom", "Legendary Creature — Human Scientist Villain", mshB3DoomOracle, 3, 3)
	b12Push(g2, g2.Seats[0].ID, "Robot", "Artifact Creature — Robot", "", 2, 2)
	if !hasEffectiveKeyword(t, g2, doom2, "indestructible") {
		t.Error("an artifact creature you control gives indestructible")
	}
}

func TestMshB3DoctorDoomDrawsAndLosesLifeAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Doctor Doom", "Legendary Creature — Human Scientist Villain", mshB3DoomOracle, 3, 3)
	hand, life := me.Hand.Size(), me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Echo, Perceptive Prodigy -----------------------------------------------------

func TestMshB3EchoCopiesOnlyYourAbilitiesFromCreatureSources(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	echo := pushCatalogPermanent(g, me.ID, "Echo, Perceptive Prodigy", "Legendary Creature — Human Hero", mshB3EchoOracle, false)
	creature := pushCatalogPermanent(g, me.ID, "Source Bear", "Creature — Bear", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Source Rock", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	if err := g.AnnounceTrigger(me.ID, creature, game.AbilityParams{Label: "creature trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.AnnounceTrigger(me.ID, rock, game.AbilityParams{Label: "artifact trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.ActivateAbility(me.ID, creature, game.AbilityParams{Label: "creature activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	trig := acItemLabelled(g, "creature trigger")
	act := acItemLabelled(g, "creature activation")
	offered := acLegalAbilityTargets(g, me.ID, mshB3EchoOracle, 0)
	if !offered[trig] || !offered[act] {
		t.Errorf("a triggered and an activated ability of yours from a creature are offered: %v", offered)
	}
	if offered[acItemLabelled(g, "artifact trigger")] {
		t.Error("an ability from a noncreature source is not offered")
	}
	if offered[acItemLabelled(g, "their trigger")] {
		t.Error("an opponent's ability is not offered")
	}

	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, echo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	acPassUntilCopy(t, g)
	if n := acCopiesOnStack(g); n != 1 {
		t.Fatalf("Echo made %d copies, want 1", n)
	}
}

// --- Epic Fight -----------------------------------------------------------------

func TestMshB3EpicFightDoublesThenFightsWhenBothAreChosen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	troll := b12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 3)

	b02cCastModalSpell(t, g, "Epic Fight", "Sorcery", mshB3EpicFightOracle, []int{0, 1}, cardRefsOf(bear, bear, troll))
	passPriorityAroundTable(t, g)

	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("doubled Bear %d/%d, want 4/4", p, tough)
	}
	if g.Battlefield.Contains(troll) {
		t.Error("the doubled Bear (power 4) kills the 3/3 Troll")
	}
	if !g.Battlefield.Contains(bear) || damageMarkedOn(g, bear) != 3 {
		t.Errorf("the Bear survives with 3 damage: on battlefield %v, damage %d", g.Battlefield.Contains(bear), damageMarkedOn(g, bear))
	}
}

func TestMshB3EpicFightEachBulletStandsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	troll := b12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 3)
	b02cCastModalSpell(t, g, "Epic Fight", "Sorcery", mshB3EpicFightOracle, []int{0}, cardRefsOf(troll))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, troll); p != 6 || tough != 6 {
		t.Errorf("any creature can be doubled: %d/%d, want 6/6", p, tough)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	bear := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	mouse := b12Creature(g2, opp2.ID, "Mouse", "Creature — Mouse", 1, 1)
	b02cCastModalSpell(t, g2, "Epic Fight", "Sorcery", mshB3EpicFightOracle, []int{1}, cardRefsOf(bear, mouse))
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(mouse) || !g2.Battlefield.Contains(bear) {
		t.Errorf("the Bear fights and kills the Mouse: mouse %v bear %v", g2.Battlefield.Contains(mouse), g2.Battlefield.Contains(bear))
	}
	if p, _ := effectivePTOf(t, g2, bear); p != 2 {
		t.Errorf("the fight-only mode does not double: power %d", p)
	}
	_ = me
}

// --- Fin Fang Foom --------------------------------------------------------------

// mshB3SettleRetarget passes priority until the stack is empty, answering
// the copy's "choose new targets" prompt with `retarget`.
func mshB3SettleRetarget(t *testing.T, g *game.Game, chooser, retarget uuid.UUID) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			pickCard(t, g, chooser, retarget)
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("stack never settled")
}

func TestMshB3FinFangFoomCopiesASpellAtAnArtifactAndGrowsTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foom := b12Push(g, me.ID, "Fin Fang Foom", "Legendary Creature — Alien Dragon Villain", mshB3FinFangOracle, 3, 5)
	rock1 := b12Push(g, opp.ID, "Rock One", "Artifact", "", 0, 0)
	rock2 := b12Push(g, opp.ID, "Rock Two", "Artifact", "", 0, 0)

	castCatalogSpell(t, g, "Disenchant", "Instant", hobOracle(t, "Disenchant"), s58Refs(rock1))
	mshB3SettleRetarget(t, g, me.ID, rock2)

	if g.Battlefield.Contains(rock1) || g.Battlefield.Contains(rock2) {
		t.Errorf("the spell and its retargeted copy destroy both: rock1 %v rock2 %v", g.Battlefield.Contains(rock1), g.Battlefield.Contains(rock2))
	}
	if got := plusOnes(g, foom); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestMshB3FinFangFoomIgnoresASpellThatTargetsOnlyACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foom := b12Push(g, me.ID, "Fin Fang Foom", "Legendary Creature — Alien Dragon Villain", mshB3FinFangOracle, 3, 5)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Shock", "Instant", hobOracle(t, "Shock"), s58Refs(bear))
	passPriorityAroundTable(t, g)
	if n := acCopiesOnStack(g); n != 0 {
		t.Errorf("copies on the stack = %d, want 0", n)
	}
	if got := plusOnes(g, foom); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("Shock resolved once and killed the 2/2")
	}
}
