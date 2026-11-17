package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// inv_b2_cards_test.go — Invasion, batch 2 (#5616). Each test plays the
// real catalog entry and asserts what its text says. The seeding and
// split-card helpers are batch 1's (inv_b1_cards_test.go).

const (
	invB2ShieldOracle    = "05b020fd-21be-495d-ae45-7de3b1224e6d"
	invB2CheetahOracle   = "a552edbe-bd5d-4281-8d9f-d03f4b10815d"
	invB2PantherOracle   = "bfb799d2-b2fa-4efb-91d7-a457230b370f"
	invB2KavuOracle      = "236fbf12-f908-4918-babb-1558e767cbe3"
	invB2CobraOracle     = "5b2e1105-5af0-4f34-bfe9-15a2f6881edd"
	invB2AssaultOracle   = "320820e4-3c7e-4e4d-b23c-63bf13a68e73"
	invB2WingsOracle     = "c1df6359-edf4-48cf-b8d1-6240ac291cf7"
	invB2RobeOracle      = "d82783d9-9bfe-4fcd-a7d2-6e2477f99edc"
	invB2ShadeOracle     = "9e120284-0fcc-432c-9471-9106eab71025"
	invB2ArtMutOracle    = "e77e8768-3ea8-44f9-aba6-1e023e006946"
	invB2AuraMutOracle   = "19bb51dd-2b93-45f0-927d-bbafddadd0a3"
	invB2SporesOracle    = "f79f5cb7-d238-41ab-85f0-eb21cc6bf561"
	invB2TarOracle       = "f268699f-c0a2-43f7-b933-7665599ab5ff"
	invB2PainOracle      = "433fcba3-06b8-4f1a-a709-300838251439"
	invB2StandOracle     = "6343aa78-2df6-41f3-9366-4933f9e6a93e"
	invB2SaprolingSearch = "Saproling"
)

func TestInvB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		invB2ShieldOracle:       "Angelic Shield",
		invB2CheetahOracle:      "Horned Cheetah",
		invB2PantherOracle:      "Noble Panther",
		invB2KavuOracle:         "Vicious Kavu",
		invB2CobraOracle:        "Voracious Cobra",
		invB2AssaultOracle:      "Reckless Assault",
		invB2WingsOracle:        "Wings of Hope",
		invB2RobeOracle:         "Sleeper's Robe",
		invB2ShadeOracle:        "Cinder Shade",
		invB2ArtMutOracle:       "Artifact Mutation",
		invB2AuraMutOracle:      "Aura Mutation",
		invB2SporesOracle:       "Plague Spores",
		invB2TarOracle:          "Smoldering Tar",
		invB2PainOracle:         "Pain",
		invB2PainOracle + "#1":  "Suffering",
		invB2StandOracle:        "Stand",
		invB2StandOracle + "#1": "Deliver",
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

// invB2Count counts the battlefield permanents named `name` that `who`
// controls.
func invB2Count(g *game.Game, who uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == who {
			n++
		}
	}
	return n
}

// --- Angelic Shield ------------------------------------------------------------------

func TestInvB2AngelicShieldGivesYourCreaturesPlusZeroPlusOneAndBounces(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	shield := invB1Push(g, me.ID, "Angelic Shield", "Enchantment", invB2ShieldOracle, 0, 0)
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 2/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	if err := invB1Activate(g, me.ID, shield, 0, cardTarget(theirs)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(shield) {
		t.Fatal("Angelic Shield was not sacrificed as the cost")
	}
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature kept the bonus after the Shield left: %d/%d", p, tough)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target did not go to its owner's hand")
	}
}

// --- Horned Cheetah ------------------------------------------------------------------

func TestInvB2HornedCheetahGainsLifeForAnyDamageItDeals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheetah := invB1Push(g, me.ID, "Horned Cheetah", "Creature — Cat", invB2CheetahOracle, 2, 2)
	theirs := seedBear(g, opp.ID)
	before := me.Life
	emitNoncombatDamage(g, cheetah, theirs, 2)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 2 {
		t.Errorf("damage to a creature gained %d life, want 2", got)
	}
	dealCombatDamageToPlayer(g, cheetah, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 4 {
		t.Errorf("combat damage to a player brought the gain to %d, want 4", got)
	}
	emitNoncombatDamage(g, theirs, cheetah, 2)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 4 {
		t.Errorf("damage dealt TO the Cheetah gained life: total %d, want 4", got)
	}
}

// --- Noble Panther -------------------------------------------------------------------

func TestInvB2NoblePantherBuysFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	panther := invB1Push(g, me.ID, "Noble Panther", "Creature — Cat", invB2PantherOracle, 3, 3)
	if err := invB1Activate(g, me.ID, panther, 0); err == nil {
		t.Fatal("the ability was activated without its {1}")
	}
	dcMana(t, g, me.ID, "{C}")
	if err := invB1Activate(g, me.ID, panther, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, panther, "first strike") {
		t.Error("Noble Panther does not have first strike")
	}
}

// --- Vicious Kavu --------------------------------------------------------------------

func TestInvB2ViciousKavuGetsPlusTwoPlusZeroWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	kavu := invB1Push(g, me.ID, "Vicious Kavu", "Creature — Kavu", invB2KavuOracle, 2, 2)
	if p, tough := ptOf(t, g, kavu); p != 2 || tough != 2 {
		t.Fatalf("Vicious Kavu is %d/%d before attacking, want 2/2", p, tough)
	}
	declareAttack(t, g, them.ID, kavu)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, kavu); p != 4 || tough != 2 {
		t.Errorf("Vicious Kavu is %d/%d while attacking, want 4/2", p, tough)
	}
}

// --- Voracious Cobra -----------------------------------------------------------------

func TestInvB2VoraciousCobraDestroysACreatureItDealsCombatDamageTo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cobra := invB1Push(g, me.ID, "Voracious Cobra", "Creature — Snake", invB2CobraOracle, 2, 2)
	if !hasKeywordOnBattlefield(t, g, cobra, "first strike") {
		t.Error("Voracious Cobra does not have first strike")
	}
	wall := invB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	mrd3BlockedBy(t, g, opp.ID, cobra, wall)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the 0/4 blocker survived the Cobra's 2 combat damage")
	}
	if !g.Battlefield.Contains(cobra) {
		t.Error("the Cobra died")
	}
}

func TestInvB2VoraciousCobraIgnoresNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cobra := invB1Push(g, me.ID, "Voracious Cobra", "Creature — Snake", invB2CobraOracle, 2, 2)
	wall := invB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	emitNoncombatDamage(g, cobra, wall, 2)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("noncombat damage destroyed the creature")
	}
}

// --- Reckless Assault ----------------------------------------------------------------

func TestInvB2RecklessAssaultPaysManaAndTwoLifeForOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	assault := invB1Push(g, me.ID, "Reckless Assault", "Enchantment", invB2AssaultOracle, 0, 0)
	mine, theirs := me.Life, opp.Life
	dcMana(t, g, me.ID, "{C}")
	if err := invB1Activate(g, me.ID, assault, 0, playerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != mine-2 {
		t.Errorf("life after paying the cost %d, want %d", me.Life, mine-2)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-1 {
		t.Errorf("the target is at %d, want %d", opp.Life, theirs-1)
	}
	if err := invB1Activate(g, me.ID, assault, 0, playerTarget(opp.ID)); err == nil {
		t.Error("the ability was activated without its {1}")
	}
}

func TestInvB2RecklessAssaultNeedsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	assault := invB1Push(g, me.ID, "Reckless Assault", "Enchantment", invB2AssaultOracle, 0, 0)
	me.Life = 1
	dcMana(t, g, me.ID, "{C}")
	if err := invB1Activate(g, me.ID, assault, 0, playerTarget(opp.ID)); err == nil {
		t.Error("the ability was activated at 1 life")
	}
}

// --- Wings of Hope -------------------------------------------------------------------

func TestInvB2WingsOfHopeGivesPlusOnePlusThreeAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	bear := seedBear(g, g.Seats[0].ID)
	castCatalogSpell(t, g, "Wings of Hope", auraTypeLine, invB2WingsOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the enchanted creature is %d/%d, want 3/5", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the enchanted creature does not have flying")
	}
}

// --- Sleeper's Robe ------------------------------------------------------------------

func TestInvB2SleepersRobeGrantsFearAndDrawsOnCombatDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Sleeper's Robe", auraTypeLine, invB2RobeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Error("the enchanted creature does not have fear")
	}

	emitNoncombatDamage(g, bear, opp.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("noncombat damage asked about a draw")
	}
	dealCombatDamageToPlayer(g, bear, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("combat damage to the Robe's controller asked about a draw")
	}

	before := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}

	before = me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Error("declining the draw still drew a card")
	}
}

// --- Cinder Shade --------------------------------------------------------------------

func TestInvB2CinderShadePumpsThenSacrificesForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	shade := invB1Push(g, me.ID, "Cinder Shade", "Creature — Shade", invB2ShadeOracle, 1, 1)
	ogre := invB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	for i := 0; i < 2; i++ {
		dcMana(t, g, me.ID, "{B}")
		if err := invB1Activate(g, me.ID, shade, 0); err != nil {
			t.Fatalf("pump %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := ptOf(t, g, shade); p != 3 || tough != 3 {
		t.Fatalf("Cinder Shade is %d/%d after two pumps, want 3/3", p, tough)
	}
	dcMana(t, g, me.ID, "{R}")
	if err := invB1Activate(g, me.ID, shade, 1, cardTarget(ogre)...); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if g.Battlefield.Contains(shade) {
		t.Fatal("Cinder Shade was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 3/3 survived 3 damage from a pumped Shade")
	}
}

// --- Artifact Mutation and Aura Mutation ---------------------------------------------

func TestInvB2ArtifactMutationDestroysThroughRegenerationAndMakesSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: opp.ID, Controller: opp.ID,
	})
	ench := invB1Push(g, opp.ID, "Ench", "Enchantment", "", 0, 0)
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, relic); c != nil {
			c.RegenerationShields = 1
		}
	})
	if err := castCatalogSpellErr(t, g, "Artifact Mutation", "Instant", invB2ArtMutOracle, cardTarget(ench)); err == nil {
		t.Fatal("an enchantment was accepted as the target")
	}
	castCatalogSpell(t, g, "Artifact Mutation", "Instant", invB2ArtMutOracle, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("a regeneration shield saved the artifact")
	}
	if n := invB2Count(g, me.ID, invB2SaprolingSearch); n != 3 {
		t.Errorf("the caster has %d Saprolings, want 3 (the artifact's mana value)", n)
	}
	if n := invB2Count(g, opp.ID, invB2SaprolingSearch); n != 0 {
		t.Errorf("the artifact's controller got %d Saprolings, want 0", n)
	}
}

func TestInvB2AuraMutationDestroysAnEnchantmentAndMakesSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ench", TypeLine: "Enchantment", ManaCost: "{1}{W}",
		Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Aura Mutation", "Instant", invB2AuraMutOracle, cardTarget(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if n := invB2Count(g, me.ID, invB2SaprolingSearch); n != 2 {
		t.Errorf("the caster has %d Saprolings, want 2", n)
	}
}

// --- Plague Spores -------------------------------------------------------------------

func TestInvB2PlagueSporesDestroysTheCreatureAndTheLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Troll", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, green); c != nil {
			c.RegenerationShields = 1
		}
	})
	if err := castCatalogSpellErr(t, g, "Plague Spores", "Sorcery", invB2SporesOracle,
		append(cardTarget(black), cardTarget(land)...)); err == nil {
		t.Fatal("a black creature was accepted as the first target")
	}
	castCatalogSpell(t, g, "Plague Spores", "Sorcery", invB2SporesOracle,
		append(cardTarget(green), cardTarget(land)...))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("a regeneration shield saved the creature")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

// --- Smoldering Tar ------------------------------------------------------------------

func TestInvB2SmolderingTarMakesTargetPlayerLoseOneAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner, victim := g.Seats[1], g.Seats[2]
	invB1Push(g, owner.ID, "Smoldering Tar", "Enchantment", invB2TarOracle, 0, 0)
	before, mine := victim.Life, owner.Life
	advanceToUpkeepOf(t, g, 1)
	pickPlayerTarget(t, g, owner.ID, victim.ID)
	passPriorityAroundTable(t, g)
	if victim.Life != before-1 {
		t.Errorf("the target is at %d, want %d", victim.Life, before-1)
	}
	if owner.Life != mine {
		t.Errorf("the Tar's controller is at %d, want %d", owner.Life, mine)
	}
}

func TestInvB2SmolderingTarSacrificesForFourDamageAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	tar := invB1Push(g, me.ID, "Smoldering Tar", "Enchantment", invB2TarOracle, 0, 0)
	big := invB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	if err := invB1Activate(g, me.ID, tar, 0, cardTarget(big)...); err == nil {
		t.Fatal("the sacrifice was activated outside a main phase")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := invB1Activate(g, me.ID, tar, 0, cardTarget(big)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(tar) {
		t.Fatal("Smoldering Tar was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- Pain // Suffering ---------------------------------------------------------------

var invB2PainSufferingFaces = [2]game.Face{
	{Name: "Pain", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}},
	{Name: "Suffering", TypeLine: "Sorcery", ManaCost: "{3}{R}", Colors: []string{"R"}},
}

func TestInvB2PainMakesTargetPlayerDiscardACardOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	keep := pushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := pushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	invB1CastHalf(t, g, invB2PainOracle, invB2PainSufferingFaces, 0, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the target owes %d discards, want 1", n)
	}
	answerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestInvB2SufferingDestroysTargetLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	bear := seedBear(g, opp.ID)
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: invB2PainOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: invB2PainSufferingFaces[:],
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("a creature was accepted as Suffering's target")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardTarget(land)}); err != nil {
		t.Fatalf("cast Suffering: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

// --- Stand // Deliver ----------------------------------------------------------------

var invB2StandDeliverFaces = [2]game.Face{
	{Name: "Stand", TypeLine: "Instant", ManaCost: "{W}", Colors: []string{"W"}},
	{Name: "Deliver", TypeLine: "Instant", ManaCost: "{2}{U}", Colors: []string{"U"}},
}

func TestInvB2StandPreventsTheNextTwoDamageToTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	bear := invB1Push(g, g.Seats[0].ID, "Ogre", "Creature — Ogre", "", 3, 5)
	invB1CastHalf(t, g, invB2StandOracle, invB2StandDeliverFaces, 0, cardTarget(bear))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("3 damage into a 2-shield: marked %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("the spent shield still prevented: marked %d, want 2", got)
	}
}

func TestInvB2DeliverReturnsTargetPermanentToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	relic := invB1Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	invB1CastHalf(t, g, invB2StandOracle, invB2StandDeliverFaces, 1, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !opp.Hand.Contains(relic) {
		t.Error("the permanent did not go to its owner's hand")
	}
}
