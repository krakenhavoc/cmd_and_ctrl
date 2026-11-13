package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_batch2_cards_test.go — Ravnica: City of Guilds, batch 2 (#5600).
// Each test plays the real catalog entry and asserts what its text says.

const (
	ravB2RiftcutterOracle = "e9612e4c-1527-4255-9f2f-b8d8cee65cc6"
	ravB2AncestorOracle   = "ee414d20-b18e-4bed-97fc-3c7535089ec1"
	ravB2ArchonOracle     = "82a70e4c-01ed-4280-80e4-86cb2bf6e63d"
	ravB2BlockbusterOracl = "11e50e35-bf4b-4089-b0c9-360d9c014584"
	ravB2BorosMageOracle  = "266c2f49-82eb-48e9-a7ea-e0da787fa40e"
	ravB2CaregiverOracle  = "7661484b-1faa-4ed7-b77b-fc1e625d23e3"
	ravB2SphinxOracle     = "296de325-cc1a-4cfa-a72d-80a2a5e18d10"
	ravB2ConsultOracle    = "c9edddb4-0d60-4d21-8887-51d943c6a31f"
	ravB2SnareOracle      = "6f5a86bf-caa7-48c9-a59f-b13fb483a9d4"
	ravB2DarkblastOracle  = "64446130-46d0-4129-a718-ed6a04218100"
	ravB2LightOracle      = "5334dc24-040e-4094-80f4-8d1b9de01230"
	ravB2CutpurseOracle   = "56c1a548-c80c-432b-bf92-d4ef6117b298"
	ravB2DimirMageOracle  = "1b6c67e5-5913-4db9-85d9-8dae57da02ee"
	ravB2StartingLife     = 40
)

func TestRavB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ravB2RiftcutterOracle: "Barbarian Riftcutter",
		ravB2AncestorOracle:   "Benevolent Ancestor",
		ravB2ArchonOracle:     "Blazing Archon",
		ravB2BlockbusterOracl: "Blockbuster",
		ravB2BorosMageOracle:  "Boros Guildmage",
		ravB2CaregiverOracle:  "Caregiver",
		ravB2SphinxOracle:     "Cerulean Sphinx",
		ravB2ConsultOracle:    "Consult the Necrosages",
		ravB2SnareOracle:      "Cyclopean Snare",
		ravB2DarkblastOracle:  "Darkblast",
		ravB2LightOracle:      "Devouring Light",
		ravB2CutpurseOracle:   "Dimir Cutpurse",
		ravB2DimirMageOracle:  "Dimir Guildmage",
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

// ravB2Activate floats the given mana and activates a catalog ability.
func ravB2Activate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams, mana ...string) error {
	t.Helper()
	b06AddMana(g.PlayerByIDForEffect(who), mana...)
	params.Strict = true
	return g.ActivateCatalogAbility(who, source, index, params)
}

// ravB2MustActivate is ravB2Activate that fails the test on an error and
// lets the ability resolve.
func ravB2MustActivate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams, mana ...string) {
	t.Helper()
	if err := ravB2Activate(t, g, who, source, index, params, mana...); err != nil {
		t.Fatalf("activate ability %d: %v", index, err)
	}
	passPriorityAroundTable(t, g)
}

// --- Barbarian Riftcutter -------------------------------------------------------

func TestRavB2BarbarianRiftcutterSacrificesToDestroyALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cutter := b12Push(g, me.ID, "Barbarian Riftcutter", "Creature — Human Barbarian", ravB2RiftcutterOracle, 3, 3)
	land := b12Creature(g, opp.ID, "Their Land", "Land", 0, 0)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, cutter, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "R"); err == nil {
		t.Fatal("the Riftcutter targeted a creature")
	}
	ravB2MustActivate(t, g, me.ID, cutter, 0, game.ActivateAbilityParams{Targets: cardRefs(land)}, "R")
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if !me.Graveyard.Contains(cutter) {
		t.Error("the Riftcutter was not sacrificed")
	}
}

// --- Benevolent Ancestor, Caregiver ---------------------------------------------

func TestRavB2BenevolentAncestorHasDefenderAndPreventsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ancestor := b12Push(g, me.ID, "Benevolent Ancestor", "Creature — Spirit", ravB2AncestorOracle, 0, 4)
	if !p1g2HasKeyword(g, ancestor, "defender") {
		t.Error("the Ancestor has no defender")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	ravB2MustActivate(t, g, me.ID, ancestor, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := damageMarkedOn(g, bear); got != 3 {
		t.Errorf("the shield is spent: marked %d, want 3", got)
	}
}

func TestRavB2BenevolentAncestorShieldsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ancestor := b12Push(g, me.ID, "Benevolent Ancestor", "Creature — Spirit", ravB2AncestorOracle, 0, 4)
	ravB2MustActivate(t, g, me.ID, ancestor, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID)})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 3) })
	if opp.Life != ravB2StartingLife-2 {
		t.Errorf("life %d, want %d", opp.Life, ravB2StartingLife-2)
	}
}

func TestRavB2CaregiverSacrificesACreatureToPreventOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	carer := b12Push(g, me.ID, "Caregiver", "Creature — Human Cleric", ravB2CaregiverOracle, 1, 1)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	ravB2MustActivate(t, g, me.ID, carer, 0, game.ActivateAbilityParams{Targets: playerRefs(me.ID), SacrificeIDs: []uuid.UUID{fodder}}, "W")
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 2) })
	if me.Life != ravB2StartingLife-1 {
		t.Errorf("life %d, want %d", me.Life, ravB2StartingLife-1)
	}
}

func TestRavB2CaregiverCantActivateWithoutACreatureToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	carer := b12Push(g, me.ID, "Caregiver", "Creature — Human Cleric", ravB2CaregiverOracle, 1, 1)
	land := b12Creature(g, me.ID, "My Land", "Land", 0, 0)
	if err := ravB2Activate(t, g, me.ID, carer, 0, game.ActivateAbilityParams{Targets: playerRefs(me.ID), SacrificeIDs: []uuid.UUID{land}}, "W"); err == nil {
		t.Fatal("a land paid \"sacrifice a creature\"")
	}
}

// --- Blazing Archon -------------------------------------------------------------

func TestRavB2BlazingArchonStopsAttacksOnItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	archon := b12Push(g, opp.ID, "Blazing Archon", "Creature — Archon", ravB2ArchonOracle, 5, 6)
	if !p1g2HasKeyword(g, archon, "flying") {
		t.Error("the Archon has no flying")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a creature attacked the Archon's controller")
	}
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatalf("attacking another player: %v", err)
	}
}

// --- Blockbuster ----------------------------------------------------------------

func TestRavB2BlockbusterHitsTappedCreaturesAndEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bb := b12Push(g, me.ID, "Blockbuster", "Enchantment", ravB2BlockbusterOracl, 0, 0)
	tapped := b12Creature(g, opp.ID, "Tired Bear", "Creature — Bear", 3, 3)
	untapped := b12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	ravB2MustActivate(t, g, me.ID, bb, 0, game.ActivateAbilityParams{}, "C", "R")
	if !me.Graveyard.Contains(bb) {
		t.Error("Blockbuster was not sacrificed")
	}
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 3/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(untapped) || damageMarkedOn(g, untapped) != 0 {
		t.Error("an untapped creature was dealt damage")
	}
	for _, p := range g.Seats {
		if p.Life != ravB2StartingLife-3 {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, ravB2StartingLife-3)
		}
	}
}

// --- Boros Guildmage ------------------------------------------------------------

func TestRavB2BorosGuildmageGrantsHasteAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := b12Push(g, me.ID, "Boros Guildmage", "Creature — Human Wizard", ravB2BorosMageOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}, "C", "R")
	if !p1g2HasKeyword(g, theirs, "haste") {
		t.Error("{1}{R} gave no haste")
	}
	if p1g2HasKeyword(g, theirs, "first strike") {
		t.Error("{1}{R} gave first strike")
	}
	ravB2MustActivate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: cardRefs(mage)}, "C", "W")
	if !p1g2HasKeyword(g, mage, "first strike") {
		t.Error("{1}{W} gave no first strike")
	}
	if err := ravB2Activate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: cardRefs(mage)}, "C", "R"); err == nil {
		t.Error("{1}{W} was paid with {1}{R}")
	}
}

// --- Cerulean Sphinx ------------------------------------------------------------

func TestRavB2CeruleanSphinxShufflesIntoItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Owned by me, controlled by the opponent, who activates it.
	sphinx := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cerulean Sphinx", TypeLine: "Creature — Sphinx", OracleID: ravB2SphinxOracle,
		Power: 5, Toughness: 5, Owner: me.ID, Controller: opp.ID,
	})
	if !p1g2HasKeyword(g, sphinx, "flying") {
		t.Error("the Sphinx has no flying")
	}
	mine, theirs := me.Library.Size(), opp.Library.Size()
	ravB2MustActivate(t, g, opp.ID, sphinx, 0, game.ActivateAbilityParams{}, "U")
	if g.Battlefield.Contains(sphinx) {
		t.Fatal("the Sphinx is still on the battlefield")
	}
	if !me.Library.Contains(sphinx) || me.Library.Size() != mine+1 {
		t.Error("the Sphinx did not go into its owner's library")
	}
	if opp.Library.Size() != theirs {
		t.Error("the controller's library changed")
	}
}

// --- Consult the Necrosages -----------------------------------------------------

func TestRavB2ConsultTheNecrosagesDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedLibrary(opp, "One", "Two", "Three")
	hand := opp.Hand.Size()
	b02cCastModalSpell(t, g, "Consult the Necrosages", "Sorcery", ravB2ConsultOracle, []int{0},
		[]game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
}

func TestRavB2ConsultTheNecrosagesDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b02cCastModalSpell(t, g, "Consult the Necrosages", "Sorcery", ravB2ConsultOracle, []int{1},
		[]game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 2 {
		t.Errorf("the target owes %d discards, want 2", got)
	}
}

// --- Cyclopean Snare ------------------------------------------------------------

func TestRavB2CyclopeanSnareTapsThenReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := b12Push(g, me.ID, "Cyclopean Snare", "Artifact", ravB2SnareOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, snare, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}, "C", "C", "C")
	if !tapCostTapped(g, theirs) {
		t.Error("the target was not tapped")
	}
	if g.Battlefield.Contains(snare) || !me.Hand.Contains(snare) {
		t.Error("the Snare did not return to its owner's hand")
	}
}

func TestRavB2CyclopeanSnareStaysWhenItsTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := b12Push(g, me.ID, "Cyclopean Snare", "Artifact", ravB2SnareOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, snare, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}, "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	hobDestroy(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(snare) {
		t.Error("the Snare returned to hand with no legal target")
	}
}

// --- Darkblast ------------------------------------------------------------------

func TestRavB2DarkblastShrinksAndKills(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Darkblast", "Instant", ravB2DarkblastOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the 3/3 is %d/%d, want 2/2", p, tough)
	}
	small := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Darkblast", "Instant", ravB2DarkblastOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 1/1 given -1/-1 did not die")
	}
}

func TestRavB2DarkblastDredgesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	blast := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: blast, Name: "Darkblast", TypeLine: "Instant",
		OracleID: ravB2DarkblastOracle, Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	_ = g.DrawCard(me.ID)
	offer := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if offer == nil || offer.DredgeOffer() != 3 {
		t.Fatalf("no dredge 3 offer: %+v", offer)
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Hand.Contains(blast) {
		t.Error("Darkblast did not return to hand")
	}
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d, want %d (milled three, drew none)", got, lib-3)
	}
}

// --- Devouring Light ------------------------------------------------------------

// ravB2CastDevouringLight casts Devouring Light from the active seat's hand
// under the strict mana gate, paying only by tapping `taps` (convoke).
func ravB2CastDevouringLight(t *testing.T, g *game.Game, target uuid.UUID, taps []uuid.UUID) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Devouring Light", TypeLine: "Instant", ManaCost: "{1}{W}{W}",
		OracleID: ravB2LightOracle, Colors: []string{"W"}, Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps, Targets: cardRefs(target)})
}

func TestRavB2DevouringLightConvokesAndExilesAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	crtBlock(g, blocker, attacker, me.ID)
	taps := ravB1Tappers(g, me.ID, 3, "W")
	if err := ravB2CastDevouringLight(t, g, attacker, taps); err != nil {
		t.Fatalf("convoke cast with three white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
}

func TestRavB2DevouringLightNeedsACreatureInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	taps := ravB1Tappers(g, me.ID, 3, "W")
	if err := ravB2CastDevouringLight(t, g, idle, taps); err == nil {
		t.Fatal("Devouring Light targeted a creature that is not attacking or blocking")
	}
}

func TestRavB2DevouringLightNeedsItsWhitePips(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	crtBlock(g, blocker, attacker, me.ID)
	taps := append(ravB1Tappers(g, me.ID, 1, "W"), ravB1Tappers(g, me.ID, 2, "G")...)
	if err := ravB2CastDevouringLight(t, g, attacker, taps); err == nil {
		t.Fatal("one white creature paid {W}{W}")
	}
}

// --- Dimir Cutpurse -------------------------------------------------------------

func TestRavB2DimirCutpurseDiscardsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := blbB1Seats(g)
	purse := b12Push(g, me.ID, "Dimir Cutpurse", "Creature — Spirit", ravB2CutpurseOracle, 2, 2)
	keep := pushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := pushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	hand := me.Hand.Size()

	dealNoncombatDamage(g, me.ID, purse, opp.ID)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil || me.Hand.Size() != hand {
		t.Fatal("noncombat damage triggered the Cutpurse")
	}

	dealCombatDamage(g, me.ID, purse, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if discardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	answerDiscard(t, g, opp.ID, pitch)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("the Cutpurse's controller drew %d, want 1", got)
	}
}

func TestRavB2DimirCutpurseStillDrawsAgainstAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	purse := b12Push(g, me.ID, "Dimir Cutpurse", "Creature — Spirit", ravB2CutpurseOracle, 2, 2)
	for opp.Hand.Size() > 0 {
		opp.Graveyard.PushTop(opp.Hand.Cards[0])
		opp.Hand.Cards = opp.Hand.Cards[1:]
	}
	hand := me.Hand.Size()
	dealCombatDamage(g, me.ID, purse, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Dimir Guildmage ------------------------------------------------------------

func TestRavB2DimirGuildmageDrawsAndDiscardsAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := b12Push(g, me.ID, "Dimir Guildmage", "Creature — Human Wizard", ravB2DimirMageOracle, 2, 2)
	// The harness starts in the draw step, where a sorcery-speed row is shut.
	if err := ravB2Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID)}, "C", "C", "C", "U"); err == nil {
		t.Fatal("activated a sorcery-speed ability in the draw step")
	}
	me.ManaPool = game.ManaPool{}
	advanceToMain(t, g)
	hand := opp.Hand.Size()
	ravB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID)}, "C", "C", "C", "U")
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Errorf("the target drew %d, want 1", got)
	}
	pushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	ravB2MustActivate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: playerRefs(opp.ID)}, "C", "C", "C", "B")
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target owes %d discards, want 1", n)
	}
}
