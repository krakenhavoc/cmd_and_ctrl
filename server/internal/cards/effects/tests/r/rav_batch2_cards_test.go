package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Barbarian Riftcutter -------------------------------------------------------

func TestRavB2BarbarianRiftcutterSacrificesToDestroyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cutter := effectstest.B12Push(g, me.ID, "Barbarian Riftcutter", "Creature — Human Barbarian", ravB2RiftcutterOracle, 3, 3)
	land := effectstest.B12Creature(g, opp.ID, "Their Land", "Land", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, cutter, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "R"); err == nil {
		t.Fatal("the Riftcutter targeted a creature")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, cutter, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)}, "R")
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if !me.Graveyard.Contains(cutter) {
		t.Error("the Riftcutter was not sacrificed")
	}
}

// --- Benevolent Ancestor, Caregiver ---------------------------------------------

func TestRavB2BenevolentAncestorHasDefenderAndPreventsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ancestor := effectstest.B12Push(g, me.ID, "Benevolent Ancestor", "Creature — Spirit", ravB2AncestorOracle, 0, 4)
	if !effectstest.P1g2HasKeyword(g, ancestor, "defender") {
		t.Error("the Ancestor has no defender")
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	effectstest.RavB2MustActivate(t, g, me.ID, ancestor, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 3 {
		t.Errorf("the shield is spent: marked %d, want 3", got)
	}
}

func TestRavB2BenevolentAncestorShieldsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ancestor := effectstest.B12Push(g, me.ID, "Benevolent Ancestor", "Creature — Spirit", ravB2AncestorOracle, 0, 4)
	effectstest.RavB2MustActivate(t, g, me.ID, ancestor, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 3) })
	if opp.Life != ravB2StartingLife-2 {
		t.Errorf("life %d, want %d", opp.Life, ravB2StartingLife-2)
	}
}

func TestRavB2CaregiverSacrificesACreatureToPreventOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	carer := effectstest.B12Push(g, me.ID, "Caregiver", "Creature — Human Cleric", ravB2CaregiverOracle, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	effectstest.RavB2MustActivate(t, g, me.ID, carer, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(me.ID), SacrificeIDs: []uuid.UUID{fodder}}, "W")
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 2) })
	if me.Life != ravB2StartingLife-1 {
		t.Errorf("life %d, want %d", me.Life, ravB2StartingLife-1)
	}
}

func TestRavB2CaregiverCantActivateWithoutACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	carer := effectstest.B12Push(g, me.ID, "Caregiver", "Creature — Human Cleric", ravB2CaregiverOracle, 1, 1)
	land := effectstest.B12Creature(g, me.ID, "My Land", "Land", 0, 0)
	if err := effectstest.RavB2Activate(t, g, me.ID, carer, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(me.ID), SacrificeIDs: []uuid.UUID{land}}, "W"); err == nil {
		t.Fatal("a land paid \"sacrifice a creature\"")
	}
}

// --- Blazing Archon -------------------------------------------------------------

func TestRavB2BlazingArchonStopsAttacksOnItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	archon := effectstest.B12Push(g, opp.ID, "Blazing Archon", "Creature — Archon", ravB2ArchonOracle, 5, 6)
	if !effectstest.P1g2HasKeyword(g, archon, "flying") {
		t.Error("the Archon has no flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a creature attacked the Archon's controller")
	}
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatalf("attacking another player: %v", err)
	}
}

// --- Blockbuster ----------------------------------------------------------------

func TestRavB2BlockbusterHitsTappedCreaturesAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bb := effectstest.B12Push(g, me.ID, "Blockbuster", "Enchantment", ravB2BlockbusterOracl, 0, 0)
	tapped := effectstest.B12Creature(g, opp.ID, "Tired Bear", "Creature — Bear", 3, 3)
	untapped := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	effectstest.RavB2MustActivate(t, g, me.ID, bb, 0, game.ActivateAbilityParams{}, "C", "R")
	if !me.Graveyard.Contains(bb) {
		t.Error("Blockbuster was not sacrificed")
	}
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 3/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(untapped) || effectstest.DamageMarkedOn(g, untapped) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Boros Guildmage", "Creature — Human Wizard", ravB2BorosMageOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "C", "R")
	if !effectstest.P1g2HasKeyword(g, theirs, "haste") {
		t.Error("{1}{R} gave no haste")
	}
	if effectstest.P1g2HasKeyword(g, theirs, "first strike") {
		t.Error("{1}{R} gave first strike")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mage)}, "C", "W")
	if !effectstest.P1g2HasKeyword(g, mage, "first strike") {
		t.Error("{1}{W} gave no first strike")
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mage)}, "C", "R"); err == nil {
		t.Error("{1}{W} was paid with {1}{R}")
	}
}

// --- Cerulean Sphinx ------------------------------------------------------------

func TestRavB2CeruleanSphinxShufflesIntoItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Owned by me, controlled by the opponent, who activates it.
	sphinx := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cerulean Sphinx", TypeLine: "Creature — Sphinx", OracleID: ravB2SphinxOracle,
		Power: 5, Toughness: 5, Owner: me.ID, Controller: opp.ID,
	})
	if !effectstest.P1g2HasKeyword(g, sphinx, "flying") {
		t.Error("the Sphinx has no flying")
	}
	mine, theirs := me.Library.Size(), opp.Library.Size()
	effectstest.RavB2MustActivate(t, g, opp.ID, sphinx, 0, game.ActivateAbilityParams{}, "U")
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	hand := opp.Hand.Size()
	effectstest.B02cCastModalSpell(t, g, "Consult the Necrosages", "Sorcery", ravB2ConsultOracle, []int{0},
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
}

func TestRavB2ConsultTheNecrosagesDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B02cCastModalSpell(t, g, "Consult the Necrosages", "Sorcery", ravB2ConsultOracle, []int{1},
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
		t.Errorf("the target owes %d discards, want 2", got)
	}
}

// --- Cyclopean Snare ------------------------------------------------------------

func TestRavB2CyclopeanSnareTapsThenReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := effectstest.B12Push(g, me.ID, "Cyclopean Snare", "Artifact", ravB2SnareOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, snare, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "C", "C", "C")
	if !effectstest.TapCostTapped(g, theirs) {
		t.Error("the target was not tapped")
	}
	if g.Battlefield.Contains(snare) || !me.Hand.Contains(snare) {
		t.Error("the Snare did not return to its owner's hand")
	}
}

func TestRavB2CyclopeanSnareStaysWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snare := effectstest.B12Push(g, me.ID, "Cyclopean Snare", "Artifact", ravB2SnareOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, snare, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.HobDestroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(snare) {
		t.Error("the Snare returned to hand with no legal target")
	}
}

// --- Darkblast ------------------------------------------------------------------

func TestRavB2DarkblastShrinksAndKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Darkblast", "Instant", ravB2DarkblastOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the 3/3 is %d/%d, want 2/2", p, tough)
	}
	small := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Darkblast", "Instant", ravB2DarkblastOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 1/1 given -1/-1 did not die")
	}
}

func TestRavB2DarkblastDredgesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	blast := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: blast, Name: "Darkblast", TypeLine: "Instant",
		OracleID: ravB2DarkblastOracle, Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	_ = g.DrawCard(me.ID)
	offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
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
	effectstest.ToMainForCost(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps, Targets: effectstest.CardRefs(target)})
}

func TestRavB2DevouringLightConvokesAndExilesAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	effectstest.CrtBlock(g, blocker, attacker, me.ID)
	taps := ravB1Tappers(g, me.ID, 3, "W")
	if err := ravB2CastDevouringLight(t, g, attacker, taps); err != nil {
		t.Fatalf("convoke cast with three white creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
}

func TestRavB2DevouringLightNeedsACreatureInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	taps := ravB1Tappers(g, me.ID, 3, "W")
	if err := ravB2CastDevouringLight(t, g, idle, taps); err == nil {
		t.Fatal("Devouring Light targeted a creature that is not attacking or blocking")
	}
}

func TestRavB2DevouringLightNeedsItsWhitePips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	effectstest.CrtBlock(g, blocker, attacker, me.ID)
	taps := append(ravB1Tappers(g, me.ID, 1, "W"), ravB1Tappers(g, me.ID, 2, "G")...)
	if err := ravB2CastDevouringLight(t, g, attacker, taps); err == nil {
		t.Fatal("one white creature paid {W}{W}")
	}
}

// --- Dimir Cutpurse -------------------------------------------------------------

func TestRavB2DimirCutpurseDiscardsThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, other := effectstest.BlbB1Seats(g)
	purse := effectstest.B12Push(g, me.ID, "Dimir Cutpurse", "Creature — Spirit", ravB2CutpurseOracle, 2, 2)
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	hand := me.Hand.Size()

	effectstest.DealNoncombatDamage(g, me.ID, purse, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil || me.Hand.Size() != hand {
		t.Fatal("noncombat damage triggered the Cutpurse")
	}

	effectstest.DealCombatDamage(g, me.ID, purse, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if effectstest.DiscardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("the Cutpurse's controller drew %d, want 1", got)
	}
}

func TestRavB2DimirCutpurseStillDrawsAgainstAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	purse := effectstest.B12Push(g, me.ID, "Dimir Cutpurse", "Creature — Spirit", ravB2CutpurseOracle, 2, 2)
	for opp.Hand.Size() > 0 {
		opp.Graveyard.PushTop(opp.Hand.Cards[0])
		opp.Hand.Cards = opp.Hand.Cards[1:]
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamage(g, me.ID, purse, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Dimir Guildmage ------------------------------------------------------------

func TestRavB2DimirGuildmageDrawsAndDiscardsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Dimir Guildmage", "Creature — Human Wizard", ravB2DimirMageOracle, 2, 2)
	// The harness starts in the draw step, where a sorcery-speed row is shut.
	if err := effectstest.RavB2Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "C", "C", "C", "U"); err == nil {
		t.Fatal("activated a sorcery-speed ability in the draw step")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.AdvanceToMain(t, g)
	hand := opp.Hand.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "C", "C", "C", "U")
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Errorf("the target drew %d, want 1", got)
	}
	effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	effectstest.RavB2MustActivate(t, g, me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "C", "C", "C", "B")
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target owes %d discards, want 1", n)
	}
}
