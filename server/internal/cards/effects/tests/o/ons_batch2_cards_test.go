package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ons_batch2_cards_test.go — Onslaught, batch 2 (#5610, S82): fourteen
// A–C cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	ons2AirborneAidOracle      = "7f715f72-5444-4ab3-a6ca-3d5bfd0d4851"
	ons2AncestorsProphetOracle = "274ce6b0-af29-4afc-873b-b4126a67bd81"
	ons2AnnexOracle            = "5f1f49dd-b7ce-494a-9b4b-ac702d3664c5"
	ons2AphettoGrifterOracle   = "d9d7b7a9-5107-4720-9f92-b49e372372b8"
	ons2AscendingAvenOracle    = "67b4a4bc-0b02-4cef-b5c8-41c839eaf0cd"
	ons2AvaraxOracle           = "2e1deb7e-d40d-437e-b817-b66132d2d1d6"
	ons2AvenFateshaperOracle   = "739d93c4-8d7b-4a54-bb6e-1d283e333392"
	ons2BoneknitterOracle      = "34b2f465-0b15-428a-a1c2-bcd41af07d01"
	ons2CabalExecutionerOracle = "b7817579-df4d-41db-bb90-2aabac58f299"
	ons2CatapultMasterOracle   = "02da2145-0edb-44be-8622-360182a8c25a"
	ons2CatapultSquadOracle    = "f1ff7a67-1add-4de1-bc25-62f63626c43a"
	ons2ContestedCliffsOracle  = "b891a683-2ebc-4e9c-b402-5dd9c1b42b69"
	ons2CrowdFavoritesOracle   = "1ead750f-14a6-4f25-9eb8-9472c2fdac35"
	ons2CruelRevivalOracle     = "836d074a-8888-42a0-8dce-11ffce00d2f5"
)

func TestOns2CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ons2AirborneAidOracle:      "Airborne Aid",
		ons2AncestorsProphetOracle: "Ancestor's Prophet",
		ons2AnnexOracle:            "Annex",
		ons2AphettoGrifterOracle:   "Aphetto Grifter",
		ons2AscendingAvenOracle:    "Ascending Aven",
		ons2AvaraxOracle:           "Avarax",
		ons2AvenFateshaperOracle:   "Aven Fateshaper",
		ons2BoneknitterOracle:      "Boneknitter",
		ons2CabalExecutionerOracle: "Cabal Executioner",
		ons2CatapultMasterOracle:   "Catapult Master",
		ons2CatapultSquadOracle:    "Catapult Squad",
		ons2ContestedCliffsOracle:  "Contested Cliffs",
		ons2CrowdFavoritesOracle:   "Crowd Favorites",
		ons2CruelRevivalOracle:     "Cruel Revival",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %v", name, def.Name, def.Completeness)
		}
	}
}

// Every morph in the batch casts face down for {3} and turns face up
// for the card's own printed cost (CR 702.37a).
func TestOns2MorphCostsAreThePrintedOnes(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Ascending Aven", ons2AscendingAvenOracle, "{2}{U}"},
		{"Boneknitter", ons2BoneknitterOracle, "{2}{B}"},
		{"Cabal Executioner", ons2CabalExecutionerOracle, "{3}{B}{B}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "morph")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no morph cast", tc.name)
			continue
		}
		if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownMorphed {
			t.Errorf("%s morph = %q %q", tc.name, ac.ManaCost, ac.FaceDown.Kind)
		}
		if ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s turns face up for %q, want %q", tc.name, ac.FaceDown.FaceUpCost, tc.faceUp)
		}
	}
}

// ons2FaceDown seeds a face-down morphed copy of a catalog creature on
// the battlefield, known to its controller.
func ons2FaceDown(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, KnownBy: map[uuid.UUID]bool{owner: true},
	}
	c.SetFaceDown(game.FaceDownMorphed)
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestOns2AirborneAidDrawsForEveryBirdAtTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Aven", "Creature — Bird Soldier", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.CastCatalogSpell(t, g, "Airborne Aid", "Sorcery", ons2AirborneAidOracle, nil)
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("hand %d -> %d, want two cards for two Birds", before, got)
	}
}

func TestOns2AncestorsProphetTapsFiveClericsForTenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	prophet := effectstest.B12Push(g, me.ID, "Ancestor's Prophet", "Creature — Human Cleric", ons2AncestorsProphetOracle, 1, 5)
	clerics := []uuid.UUID{prophet}
	for i := 0; i < 4; i++ {
		clerics = append(clerics, effectstest.B12Creature(g, me.ID, "My Cleric", "Creature — Human Cleric", 1, 1))
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	short := append(append([]uuid.UUID{}, clerics[:4]...), bear)
	if err := g.ActivateCatalogAbility(me.ID, prophet, 0, game.ActivateAbilityParams{TapIDs: short}); err == nil {
		t.Fatal("a non-Cleric paid the tap-five cost")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, prophet, 0, game.ActivateAbilityParams{TapIDs: clerics}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range clerics {
		if !effectstest.TmtB1Card(t, g, id).Tapped {
			t.Error("a Cleric did not tap to pay")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+10 {
		t.Errorf("life %d -> %d, want +10", life, me.Life)
	}
}

func TestOns2AnnexTakesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	land := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Annex", effectstest.AuraTypeLine, ons2AnnexOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Annex enchanted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Annex", effectstest.AuraTypeLine, ons2AnnexOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, land); got != me.ID {
		t.Errorf("the land's controller is %s, want the Aura's controller %s", got, me.ID)
	}
}

func TestOns2AphettoGrifterTapsTwoWizardsToTapAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	grifter := effectstest.B12Push(g, me.ID, "Aphetto Grifter", "Creature — Human Wizard", ons2AphettoGrifterOracle, 1, 1)
	wizard := effectstest.B12Creature(g, me.ID, "My Wizard", "Creature — Human Wizard", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	land := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	if err := g.ActivateCatalogAbility(me.ID, grifter, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{grifter, bear}, Targets: effectstest.CardTarget(land),
	}); err == nil {
		t.Fatal("a non-Wizard paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, grifter, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{grifter, wizard}, Targets: effectstest.CardTarget(land),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TmtB1Card(t, g, land).Tapped {
		t.Error("the target land is still untapped")
	}
	if !effectstest.TmtB1Card(t, g, grifter).Tapped || !effectstest.TmtB1Card(t, g, wizard).Tapped {
		t.Error("the two Wizards did not tap to pay")
	}
}

func TestOns2AscendingAvenBlocksOnlyFliersFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	aven := effectstest.B12Push(g, them.ID, "Ascending Aven", "Creature — Bird Soldier", ons2AscendingAvenOracle, 3, 2)
	hidden := ons2FaceDown(g, them.ID, "Ascending Aven", "Creature — Bird Soldier", ons2AscendingAvenOracle, 3, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, bear, bird)
	if r := effectstest.BlockRefusal(t, g, aven, bear); r != game.BlockReasonCantBlockAttacker {
		t.Errorf("refusal %q blocking a creature without flying, want %q", r, game.BlockReasonCantBlockAttacker)
	}
	if r := effectstest.BlockRefusal(t, g, aven, bird); r != "" {
		t.Errorf("the Aven could not block a flier: %q", r)
	}
	if r := effectstest.BlockRefusal(t, g, hidden, bear); r != "" {
		t.Errorf("a face-down Aven could not block a ground creature: %q (it has no text, CR 708.2)", r)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, aven, "flying") {
		t.Error("the face-up Aven has no flying")
	}
}

func TestOns2AvaraxFetchesAnotherAvarax(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Avarax", TypeLine: "Creature — Beast", OracleID: ons2AvaraxOracle},
	)
	id := effectstest.CastCatalogSpell(t, g, "Avarax", "Creature — Beast", ons2AvaraxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after agreeing")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a card not named Avarax was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Avarax")
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Avarax" {
			found = true
		}
	}
	if !found {
		t.Error("the second Avarax is not in hand")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
		t.Error("Avarax has no haste")
	}
}

func TestOns2AvaraxPumpsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avarax := effectstest.B12Push(g, me.ID, "Avarax", "Creature — Beast", ons2AvaraxOracle, 3, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, avarax, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.MkmMPPT(t, g, avarax); p != 4 || th != 3 {
		t.Errorf("Avarax is %d/%d, want 4/3", p, th)
	}
}

func TestOns2AvenFateshaperLooksAtFourOnEntryAndForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.CastCatalogSpell(t, g, "Aven Fateshaper", "Creature — Bird Wizard", ons2AvenFateshaperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger queued no look-at prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Fatalf("looked at %d cards on entry, want 4", len(c.ScryCards))
	}
	reordered := []uuid.UUID{c.ScryCards[3], c.ScryCards[0], c.ScryCards[1], c.ScryCards[2]}
	if err := g.ResolveLookAtTop(c.ID, me.ID, reordered); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	if got := effectstest.LibraryTopNames(me, 4); got[0] != "D" || got[1] != "A" {
		t.Errorf("library top is %v after the reorder, want D first", got)
	}

	shaper := effectstest.B12Push(g, me.ID, "Aven Fateshaper", "Creature — Bird Wizard", ons2AvenFateshaperOracle, 4, 5)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, shaper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 4 {
		t.Fatalf("the activated look-at prompt = %+v, want four cards", c)
	}
}

func TestOns2BoneknitterRegeneratesOnlyAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	knitter := effectstest.B12Push(g, me.ID, "Boneknitter", "Creature — Zombie Cleric", ons2BoneknitterOracle, 1, 1)
	zombie := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, knitter, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("Boneknitter targeted a Bear")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, knitter, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(zombie), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, zombie); n != 1 {
		t.Fatalf("the Zombie has %d regeneration shields, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if !g.Battlefield.Contains(zombie) {
		t.Error("the regenerated Zombie was destroyed")
	}
}

func TestOns2CabalExecutionerMakesTheDamagedPlayerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	exec := effectstest.B12Push(g, me.ID, "Cabal Executioner", "Creature — Human Cleric", ons2CabalExecutionerOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, exec, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the damaged player was not asked to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(ogre) || !g.Battlefield.Contains(mine) {
		t.Error("a creature other than the chosen one left")
	}
}

func TestOns2CabalExecutionerFaceDownHasNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	hidden := ons2FaceDown(g, me.ID, "Cabal Executioner", "Creature — Human Cleric", ons2CabalExecutionerOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, hidden, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil || !g.Battlefield.Contains(bear) {
		t.Error("a face-down Executioner's damage made its victim sacrifice (it has no text, CR 708.2)")
	}
}

func TestOns2CatapultMasterTapsFiveSoldiersToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	master := effectstest.B12Push(g, me.ID, "Catapult Master", "Creature — Human Soldier", ons2CatapultMasterOracle, 3, 3)
	soldiers := []uuid.UUID{master}
	for i := 0; i < 4; i++ {
		soldiers = append(soldiers, effectstest.B12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 1, 1))
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		TapIDs: soldiers[:4], Targets: effectstest.CardTarget(bear),
	}); err == nil {
		t.Fatal("four Soldiers paid a tap-five cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		TapIDs: soldiers, Targets: effectstest.CardTarget(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the target is still on the battlefield")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the target was not exiled")
	}
}

func TestOns2CatapultSquadHitsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	squad := effectstest.B12Push(g, me.ID, "Catapult Squad", "Creature — Human Soldier", ons2CatapultSquadOracle, 2, 1)
	soldier := effectstest.B12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, them.ID, "Their Bears", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	tap := []uuid.UUID{squad, soldier}
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{TapIDs: tap, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("the Squad targeted a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{TapIDs: tap, Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageOn(t, g, attacker); n != 2 {
		t.Errorf("the attacker has %d damage, want 2", n)
	}
}

func TestOns2ContestedCliffsMakesABeastFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	cliffs := effectstest.B12Push(g, me.ID, "Contested Cliffs", "Land", ons2ContestedCliffsOracle, 0, 0)
	beast := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	slots := func(mine uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}
	}
	if err := g.ActivateCatalogAbility(me.ID, cliffs, 0, game.ActivateAbilityParams{Targets: slots(bear)}); err == nil {
		t.Fatal("a non-Beast fought")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, cliffs, 0, game.ActivateAbilityParams{Targets: slots(beast), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature survived 3 damage")
	}
	if n := effectstest.DamageOn(t, g, beast); n != 2 {
		t.Errorf("the Beast has %d damage, want 2", n)
	}
	if !effectstest.TmtB1Card(t, g, cliffs).Tapped {
		t.Error("the Cliffs did not tap")
	}
}

func TestOns2CrowdFavoritesTapsAndGrowsTough(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	fav := effectstest.B12Push(g, me.ID, "Crowd Favorites", "Creature — Human Soldier", ons2CrowdFavoritesOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}{C}{W}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, fav, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear), Strict: true}); err != nil {
		t.Fatalf("activate the tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TmtB1Card(t, g, bear).Tapped {
		t.Error("the target is still untapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, fav, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.MkmMPPT(t, g, fav); p != 4 || th != 9 {
		t.Errorf("Crowd Favorites is %d/%d, want 4/9", p, th)
	}
	if effectstest.TmtB1Card(t, g, fav).Tapped {
		t.Error("Crowd Favorites tapped for abilities with no {T}")
	}
}

func TestOns2CruelRevivalDestroysThroughRegenerationAndReturnsAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	zombie := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	card := effectstest.PushGraveyardCardTyped(me, "Gravedigger", "Creature — Zombie")
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Cruel Revival", "Instant", ons2CruelRevivalOracle, effectstest.CardRefs(zombie)); err == nil {
		t.Fatal("Cruel Revival targeted a Zombie creature")
	}
	effectstest.CastCatalogSpell(t, g, "Cruel Revival", "Instant", ons2CruelRevivalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
		{Kind: game.TargetCard, ID: card, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature regenerated (it can't be regenerated)")
	}
	if !me.Hand.Contains(card) {
		t.Error("the Zombie card did not return to hand")
	}
}

func TestOns2CruelRevivalNeedsNoZombieCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cruel Revival", "Instant", ons2CruelRevivalOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived with no second target chosen")
	}
}
