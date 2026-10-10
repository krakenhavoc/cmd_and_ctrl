package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// craft_variants_test.go — ADR 0137's 2026-10-10 amendment (#2709):
// "one or more", rules over the chosen set, and graveyard-only
// materials, each pinned on the card that prints it, and each card's
// back face that reads what it was crafted with.

// pushGraveyardMaterial seeds a card into `p`'s graveyard with the
// fields a material test reads.
func pushGraveyardMaterial(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// pushMaterialPermanent puts a permanent with the given fields onto the
// battlefield under `p`.
func pushMaterialPermanent(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	return pushBattlefieldCardWithTimestamp(g, c)
}

// "One or more Dinosaurs": none is too few, a non-Dinosaur is refused,
// and any number mixed from both zones pays. The Raptor's power is the
// materials' total.
func TestSaheelisLatticeCraftsWithOneOrMoreDinosaurs(t *testing.T) {
	g, me, _ := spendTable(t)
	lattice := pushCraftCard(g, me, transformRow(saheelisLatticeOracleID, "Saheeli's Lattice", "Artifact", "{1}{R}",
		"Mastercraft Raptor", "Artifact Creature — Dinosaur", "*", "4", []string{"R"}))
	rex := pushMaterialPermanent(g, me, game.Card{Name: "Rex", TypeLine: "Creature — Dinosaur", Power: 3, Toughness: 3})
	fossil := pushGraveyardMaterial(me, game.Card{Name: "Fossil", TypeLine: "Creature — Dinosaur", Power: 4, Toughness: 4})
	bear := pushGraveyardMaterial(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	floatMana(t, g, me, "{C}{C}{C}{C}{R}")
	if err := activateCraft(g, me, lattice, 0); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("no materials: err = %v, want ErrInvalidParam", err)
	}
	if err := activateCraft(g, me, lattice, 0, rex, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a Bear among the Dinosaurs: err = %v, want ErrIllegalTarget", err)
	}
	if err := activateCraft(g, me, lattice, 0, rex, fossil); err != nil {
		t.Fatalf("two Dinosaurs: %v", err)
	}
	passPriorityAroundTable(t, g)

	raptor := craftedPermanent(t, g, "Mastercraft Raptor")
	if len(raptor.CraftedWith) != 2 {
		t.Fatalf("CraftedWith = %v, want both Dinosaurs", raptor.CraftedWith)
	}
	if p, tough := effectivePower(t, g, raptor.InstanceID), effectiveToughness(t, g, raptor.InstanceID); p != 7 || tough != 4 {
		t.Errorf("Mastercraft Raptor is %d/%d, want 7/4 (3 + 4 power crafted)", p, tough)
	}

	// A material that leaves exile stops counting (CR 702.167c).
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Exile, me.Graveyard, fossil); err != nil {
			t.Fatalf("move the Fossil out of exile: %v", err)
		}
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: fossil, OldZone: game.ZoneExile, NewZone: game.ZoneGraveyard})
	})
	if p := effectivePower(t, g, raptor.InstanceID); p != 3 {
		t.Errorf("power after a material left exile = %d, want 3", p)
	}
}

// Wretched Bonemass is as big as its materials' total power, and has
// the listed keywords — protection with its quality — that a material
// in exile has.
func TestAltarOfTheWretchedCraftsABonemassWithItsMaterialsKeywords(t *testing.T) {
	g, me, _ := spendTable(t)
	altar := pushCraftCard(g, me, transformRow(altarOfTheWretchedOracleID, "Altar of the Wretched", "Artifact", "{2}{B}",
		"Wretched Bonemass", "Creature — Skeleton Horror", "*", "*", []string{"B"}))
	flyer := pushGraveyardMaterial(me, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying", "protection from red"}})
	bruiser := pushMaterialPermanent(g, me, game.Card{Name: "Bruiser", TypeLine: "Creature — Ogre", Power: 5, Toughness: 5,
		Keywords: []string{"trample", "defender"}})

	bonemass := craftInto(t, g, me, altar, 0, "{C}{C}{B}{B}", "Wretched Bonemass", flyer, bruiser)
	if p, tough := effectivePower(t, g, bonemass.InstanceID), effectiveToughness(t, g, bonemass.InstanceID); p != 7 || tough != 7 {
		t.Errorf("Wretched Bonemass is %d/%d, want 7/7", p, tough)
	}
	for _, kw := range []string{"flying", "trample", "protection from red"} {
		if !effectiveAbilitiesContain(t, g, bonemass.InstanceID, kw) {
			t.Errorf("Wretched Bonemass lacks %q, which a material has", kw)
		}
	}
	if effectiveAbilitiesContain(t, g, bonemass.InstanceID, "defender") {
		t.Error("Wretched Bonemass has defender, which its list does not name")
	}
}

// The bare "Craft with one or more" takes any other permanent and any
// graveyard card; the Effigy counts their colours and taps for them.
func TestSunbirdStandardCraftsAnEffigyOfItsMaterialsColours(t *testing.T) {
	g, me, _ := spendTable(t)
	standard := pushCraftCard(g, me, transformRow(sunbirdStandardOracleID, "Sunbird Standard", "Artifact", "{3}",
		"Sunbird Effigy", "Artifact Creature — Bird Construct", "*", "*", nil))
	land := pushMaterialPermanent(g, me, game.Card{Name: "Wastes", TypeLine: "Basic Land"})
	spell := pushGraveyardMaterial(me, game.Card{Name: "Gold Spell", TypeLine: "Instant", Colors: []string{"U", "G"}})
	red := pushGraveyardMaterial(me, game.Card{Name: "Red Spell", TypeLine: "Sorcery", Colors: []string{"R"}})

	effigy := craftInto(t, g, me, standard, 0, "{C}{C}{C}{C}{C}", "Sunbird Effigy", land, spell, red)
	if p, tough := effectivePower(t, g, effigy.InstanceID), effectiveToughness(t, g, effigy.InstanceID); p != 3 || tough != 3 {
		t.Errorf("Sunbird Effigy is %d/%d, want 3/3 for blue, red and green", p, tough)
	}
	var produced string
	g.ReadSnapshot(func() { produced = sunbirdEffigyProduced(g, me.ID, effigy.InstanceID) })
	if produced != "{U}{R}{G}" {
		t.Errorf("Sunbird Effigy taps for %q, want {U}{R}{G}", produced)
	}
}

// "Two that share a card type": an artifact and a creature that share
// nothing are refused; two creatures pay. The Observatory names a type
// they share, and its tap makes the next spell of that type free.
func TestEyeOfOjerTaqCraftsWithTwoThatShareACardType(t *testing.T) {
	g, me, _ := spendTable(t)
	eye := pushCraftCard(g, me, transformRow(eyeOfOjerTaqOracleID, "Eye of Ojer Taq", "Artifact", "{3}",
		"Apex Observatory", "Artifact", "", "", nil))
	rock := pushMaterialPermanent(g, me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := pushMaterialPermanent(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	elk := pushGraveyardMaterial(me, game.Card{Name: "Elk", TypeLine: "Creature — Elk", Power: 3, Toughness: 3})

	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := activateCraft(g, me, eye, 0, rock, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an artifact and a creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := activateCraft(g, me, eye, 0, bear, elk); err != nil {
		t.Fatalf("two creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)

	apex := craftedPermanent(t, g, "Apex Observatory")
	if !apex.Tapped {
		t.Error("Apex Observatory entered untapped")
	}
	if apex.ChosenOption != "Creature" {
		t.Fatalf("chosen type = %q, want Creature", apex.ChosenOption)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == apex.InstanceID {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, apex.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("tap the Observatory: %v", err)
	}
	passPriorityAroundTable(t, g)

	sorcery := pushHandCardWithManaCost(me, "Big Sorcery", "Sorcery", "{7}")
	if offerWithKey(grantedCardOffers(g, me.ID, sorcery, game.ZoneHand), game.GrantedAltCostNextFree) != nil {
		t.Error("a sorcery was offered the free cast the Observatory gave creature spells")
	}
	beast := pushHandCardWithManaCost(me, "Big Beast", "Creature — Beast", "{7}{G}")
	if offerWithKey(grantedCardOffers(g, me.ID, beast, game.ZoneHand), game.GrantedAltCostNextFree) == nil {
		t.Fatal("the next creature spell was not offered a free cast")
	}
	if err := g.CastSpell(me.ID, beast, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err != nil {
		t.Fatalf("free creature cast: %v", err)
	}
	other := pushHandCardWithManaCost(me, "Second Beast", "Creature — Beast", "{2}")
	if offerWithKey(grantedCardOffers(g, me.ID, other, game.ZoneHand), game.GrantedAltCostNextFree) != nil {
		t.Error("the promise outlived the next creature spell")
	}
}

// "A Dinosaur, a Merfolk, a Pirate, and a Vampire" is four separate
// objects: a changeling fills one part, never two.
func TestThroneOfTheGrimCaptainNeedsOneOfEachSubtype(t *testing.T) {
	g, me, _ := spendTable(t)
	throne := pushCraftCard(g, me, transformRow(throneOfTheGrimCaptainOracleID, "Throne of the Grim Captain", "Legendary Artifact", "{2}",
		"The Grim Captain", "Legendary Creature — Skeleton Spirit Pirate", "7", "7", []string{"B"}))
	dino := pushGraveyardMaterial(me, game.Card{Name: "Dino", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2})
	merfolk := pushMaterialPermanent(g, me, game.Card{Name: "Merrow", TypeLine: "Creature — Merfolk", Power: 1, Toughness: 1})
	pirate := pushGraveyardMaterial(me, game.Card{Name: "Pirate", TypeLine: "Creature — Human Pirate", Power: 1, Toughness: 1})
	changeling := pushGraveyardMaterial(me, game.Card{Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Keywords: []string{"changeling"}})
	vampire := pushGraveyardMaterial(me, game.Card{Name: "Vampire", TypeLine: "Creature — Vampire", Power: 1, Toughness: 1})

	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := activateCraft(g, me, throne, 1, dino, merfolk, pirate); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("three of four: err = %v, want ErrInvalidParam", err)
	}
	if err := activateCraft(g, me, throne, 1, dino, merfolk, pirate, pirate); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("one card named twice: err = %v, want ErrInvalidParam", err)
	}
	g.ReadSnapshot(func() {
		ec := CraftWithEachOf("Dinosaur", "Merfolk", "Pirate", "Vampire").ExilePermanents
		if g.ExilePermanentsPaymentForEffect(me.ID, ec, []uuid.UUID{dino, merfolk, changeling}) != nil {
			t.Error("three cards paid a four-part clause")
		}
		pick := g.ExilePermanentsPaymentForEffect(me.ID, ec, []uuid.UUID{changeling, dino, merfolk, pirate, vampire})
		if len(pick) != 4 || slicesContain(pick, changeling) {
			t.Errorf("payment = %v, want the four single-type cards and the changeling kept", pick)
		}
	})
	if err := activateCraft(g, me, throne, 1, dino, merfolk, pirate, changeling); err != nil {
		t.Fatalf("the changeling as the Vampire: %v", err)
	}
	passPriorityAroundTable(t, g)
	captain := craftedPermanent(t, g, "The Grim Captain")
	if len(captain.CraftedWith) != 4 {
		t.Errorf("CraftedWith = %v, want all four", captain.CraftedWith)
	}
}

func slicesContain(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// The Grim Captain's attack makes each opponent sacrifice a nonland
// permanent, then puts a creature card it was crafted with onto the
// battlefield tapped and attacking.
func TestTheGrimCaptainPutsAMaterialOntoTheBattlefieldAttacking(t *testing.T) {
	g, me, opp := spendTable(t)
	throne := pushCraftCard(g, me, transformRow(throneOfTheGrimCaptainOracleID, "Throne of the Grim Captain", "Legendary Artifact", "{2}",
		"The Grim Captain", "Legendary Creature — Skeleton Spirit Pirate", "7", "7", []string{"B"}))
	dino := pushGraveyardMaterial(me, game.Card{Name: "Dino", TypeLine: "Creature — Dinosaur", Power: 4, Toughness: 4})
	merfolk := pushGraveyardMaterial(me, game.Card{Name: "Merrow", TypeLine: "Creature — Merfolk", Power: 1, Toughness: 1})
	pirate := pushGraveyardMaterial(me, game.Card{Name: "Pirate", TypeLine: "Creature — Human Pirate", Power: 1, Toughness: 1})
	vampire := pushGraveyardMaterial(me, game.Card{Name: "Vampire", TypeLine: "Creature — Vampire", Power: 1, Toughness: 1})
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	captain := craftInto(t, g, me, throne, 1, "{C}{C}{C}{C}", "The Grim Captain", dino, merfolk, pirate, vampire)
	for _, kw := range []string{"menace", "trample", "lifelink", "hexproof"} {
		if !effectiveAbilitiesContain(t, g, captain.InstanceID, kw) {
			t.Errorf("The Grim Captain lacks %s", kw)
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == captain.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	declareAttack(t, g, opp.ID, captain.InstanceID)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, theirs)
	answerChooseCards(t, g, me.ID, dino)
	// What it attacks: the first player on offer, which is the
	// Captain's opponent here.
	var first uuid.UUID
	g.ReadSnapshot(func() { first = g.AttackTargetsForEffect(me.ID)[0].ID })
	if latestOptionPickFor(g, me.ID) != nil {
		answerOptionPick(t, g, me.ID, 0)
	}
	if first != opp.ID {
		t.Fatalf("first attack target %v, want the opponent", first)
	}

	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent kept their only nonland permanent")
	}
	ids := battlefieldIDsNamed(g, "Dino")
	if len(ids) != 1 {
		t.Fatalf("%d Dinos on the battlefield, want the material put in attacking", len(ids))
	}
	var dinoNow game.Card
	g.ReadSnapshot(func() { dinoNow, _ = g.LookupCardForEffect(ids[0]) })
	if !dinoNow.Tapped || dinoNow.AttackingTarget != opp.ID || dinoNow.Controller != me.ID {
		t.Errorf("Dino tapped=%v attacking=%v controller=%v, want tapped, attacking the opponent, mine",
			dinoNow.Tapped, dinoNow.AttackingTarget, dinoNow.Controller)
	}
}

// "Four or more red instant and/or sorcery cards" is graveyard-only: a
// permanent never pays, nor a card of the wrong colour or type, and
// three are too few.
func TestOreRichStalactiteCraftsFromFourOrMoreRedSpellsInTheGraveyard(t *testing.T) {
	g, me, _ := spendTable(t)
	stalactite := pushCraftCard(g, me, transformRow(oreRichStalactiteOracleID, "Ore-Rich Stalactite", "Artifact", "{1}{R}",
		"Cosmium Catalyst", "Artifact", "", "", []string{"R"}))
	var bolts []uuid.UUID
	for i := 0; i < 5; i++ {
		bolts = append(bolts, pushGraveyardMaterial(me, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}))
	}
	blue := pushGraveyardMaterial(me, game.Card{Name: "Opt", TypeLine: "Instant", Colors: []string{"U"}})
	redCreature := pushGraveyardMaterial(me, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"}})
	redArtifact := pushMaterialPermanent(g, me, game.Card{Name: "Red Relic", TypeLine: "Artifact", Colors: []string{"R"}})

	var options []uuid.UUID
	g.ReadSnapshot(func() {
		options = g.ExilePermanentsOptionsForEffect(me.ID, stalactite, CraftWithCardsOrMore(4, "R", "red", "instant", "sorcery").ExilePermanents)
	})
	if len(options) != 5 || slicesContain(options, blue) || slicesContain(options, redCreature) || slicesContain(options, redArtifact) {
		t.Fatalf("options = %v, want the five red instants only", options)
	}
	floatMana(t, g, me, "{C}{C}{C}{R}{R}")
	if err := activateCraft(g, me, stalactite, 0, bolts[:3]...); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("three: err = %v, want ErrInvalidParam", err)
	}
	if err := activateCraft(g, me, stalactite, 0, bolts[0], bolts[1], bolts[2], redArtifact); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a permanent: err = %v, want ErrIllegalTarget", err)
	}
	if err := activateCraft(g, me, stalactite, 0, bolts...); err != nil {
		t.Fatalf("five red instants: %v", err)
	}
	passPriorityAroundTable(t, g)
	catalyst := craftedPermanent(t, g, "Cosmium Catalyst")
	if len(catalyst.CraftedWith) != 5 {
		t.Fatalf("CraftedWith = %v, want five", catalyst.CraftedWith)
	}

	// Cosmium Catalyst: one material at random may be cast for free.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == catalyst.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	floatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, catalyst.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Catalyst: %v", err)
	}
	passPriorityAroundTable(t, g)
	free := 0
	for _, id := range bolts {
		if offerFree(g, me.ID, id) {
			free++
		}
	}
	if free != 1 {
		t.Errorf("%d materials castable for free, want exactly the one chosen at random", free)
	}
}

// offerFree reports whether `id` in exile may be cast now for nothing.
func offerFree(g *game.Game, seat, id uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() {
		c, found := g.LookupCardForEffect(id)
		if !found {
			return
		}
		if p := g.CastPermissionForLocked(seat, c, game.ZoneExile); p != nil && p.Cost == "{0}" {
			ok = true
		}
	})
	return ok
}

// Master's Manufactory makes a Golem only on a turn an artifact entered
// under its controller — the Manufactory's own entry counts.
func TestMastersManufactoryNeedsAnArtifactToHaveEntered(t *testing.T) {
	g, me, _ := spendTable(t)
	mural := pushCraftCard(g, me, transformRow(mastersGuideMuralOracleID, "Master's Guide-Mural", "Artifact", "{3}{W}{U}",
		"Master's Manufactory", "Artifact", "", "", []string{"W", "U"}))
	rock := pushGraveyardMaterial(me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	manufactory := craftInto(t, g, me, mural, 0, "{C}{C}{C}{C}{W}{W}{U}", "Master's Manufactory", rock)

	if err := g.ActivateCatalogAbility(me.ID, manufactory.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the Manufactory entered this turn, so it may make a Golem: %v", err)
	}
	passPriorityAroundTable(t, g)
	golems := battlefieldIDsNamed(g, "Golem")
	if len(golems) != 1 {
		t.Fatalf("%d Golems, want one", len(golems))
	}
	if p := effectivePower(t, g, golems[0]); p != 4 {
		t.Errorf("Golem power %d, want 4", p)
	}

	var entered bool
	g.ReadSnapshot(func() { entered = anArtifactEnteredUnderYouThisTurn(g, me.ID, uuid.Nil) })
	if !entered {
		t.Error("the tally lost this turn's artifact entries")
	}
	g.WithWriteLock(func() { g.TurnTally.EnteredCardTypes = nil })
	g.ReadSnapshot(func() { entered = anArtifactEnteredUnderYouThisTurn(g, me.ID, uuid.Nil) })
	if entered {
		t.Error("the condition held with no artifact entry recorded")
	}
}

// The trigger a Blunderbuss grants names that Blunderbuss: with two of
// them on one creature, each trigger excludes only its own.
func TestDireBlunderbussTriggerNamesItsGrantor(t *testing.T) {
	g, me, _ := spendTable(t)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	var buss [2]uuid.UUID
	for i := range buss {
		c := game.Card{Name: "Dire Blunderbuss", TypeLine: "Artifact — Equipment", OracleID: direFlailOracleID + "#1",
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}}
		buss[i] = pushMaterialPermanent(g, me, c)
	}
	declareAttack(t, g, g.Seats[1].ID, bear)
	var items []*game.StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID == bear {
				items = append(items, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.SourceCardID == bear && it.Kind == game.StackItemTriggered {
				items = append(items, it)
			}
		}
	})
	if len(items) != 2 {
		t.Fatalf("%d granted attack triggers, want one per Blunderbuss", len(items))
	}
	got := map[uuid.UUID]bool{items[0].GrantedBy: true, items[1].GrantedBy: true}
	if !got[buss[0]] || !got[buss[1]] {
		t.Errorf("grantors = %v, %v; want the two Blunderbusses", items[0].GrantedBy, items[1].GrantedBy)
	}
	g.ReadSnapshot(func() {
		ex := direBlunderbussGrantors(g, items[0])
		if len(ex) != 1 || !ex[items[0].GrantedBy] {
			t.Errorf("excluded %v, want only the trigger's own grantor", ex)
		}
	})
}

// The Headdress's attach choice makes the equipped creature a copy of
// a creature card it was crafted with, until it is unattached.
func TestDinosaurHeaddressCopiesAChosenMaterialWhileAttached(t *testing.T) {
	g, me, _ := spendTable(t)
	axe := pushCraftCard(g, me, transformRow(paleontologistsPickAxeOracleID, "Paleontologist's Pick-Axe", "Artifact — Equipment", "{2}",
		"Dinosaur Headdress", "Artifact — Equipment", "", "", nil))
	rex := pushGraveyardMaterial(me, game.Card{Name: "Fossil Rex", TypeLine: "Creature — Dinosaur", Power: 6, Toughness: 6, ManaCost: "{5}{G}"})
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)

	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := activateCraft(g, me, axe, 1, rex); err != nil {
		t.Fatalf("craft: %v", err)
	}
	passPriorityAroundTable(t, g)
	headdress := craftedPermanent(t, g, "Dinosaur Headdress")
	// The enters trigger targets the bear.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			answerPickTarget(t, g, bear)
			break
		}
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, rex)

	var now game.Card
	g.ReadSnapshot(func() { now, _ = g.LookupCardForEffect(bear) })
	if now.Name != "Fossil Rex" {
		t.Fatalf("equipped creature is %q, want a copy of Fossil Rex", now.Name)
	}
	if p := effectivePower(t, g, bear); p != 6 {
		t.Errorf("copy's power %d, want 6", p)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == headdress.InstanceID {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{}
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventUnattach, CardID: headdress.InstanceID})
		g.RecomputeLayersIfStaleLocked()
	})
	g.ReadSnapshot(func() { now, _ = g.LookupCardForEffect(bear) })
	if now.Name != "Grizzly Bears" {
		t.Errorf("unequipped creature is still %q", now.Name)
	}
}

// Tetzin's trigger fires for itself and another double-faced artifact,
// not for a single-faced one.
func TestTetzinTriggersOnDoubleFacedArtifacts(t *testing.T) {
	g, me, _ := spendTable(t)
	tetzin := pushCraftCard(g, me, transformRow(tetzinGnomeChampionOracleID, "Tetzin, Gnome Champion", "Legendary Artifact Creature — Gnome", "{U}{R}{W}",
		"The Golden-Gear Colossus", "Legendary Artifact Creature — Gnome", "6", "6", []string{"U", "R", "W"}))
	var source game.Card
	g.ReadSnapshot(func() { source, _ = g.LookupCardForEffect(tetzin) })
	dfc := pushCraftCard(g, me, transformRow(clayFiredBricksOracleID, "Clay-Fired Bricks", "Artifact", "{1}{W}",
		"Cosmium Kiln", "Artifact", "", "", []string{"W"}))
	rock := pushMaterialPermanent(g, me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	g.ReadSnapshot(func() {
		if !tetzinDoubleFacedArtifactEntered(game.Event{Kind: game.EventETB, CardID: tetzin}, &source, game.Characteristic{}, g) {
			t.Error("Tetzin's own entry did not trigger it")
		}
		if !tetzinDoubleFacedArtifactEntered(game.Event{Kind: game.EventETB, CardID: dfc}, &source, game.Characteristic{}, g) {
			t.Error("another double-faced artifact's entry did not trigger it")
		}
		if tetzinDoubleFacedArtifactEntered(game.Event{Kind: game.EventETB, CardID: rock}, &source, game.Characteristic{}, g) {
			t.Error("a single-faced artifact's entry triggered it")
		}
	})
}

// The enumerator offers craft's variants as payments the engine takes:
// the floor and the graveyard's cheap fuel for "one or more", each
// shared type for "two that share a card type".
func TestCraftVariantPaymentsAreEnumeratedAndLegal(t *testing.T) {
	g, me, _ := spendTable(t)
	lattice := pushCraftCard(g, me, transformRow(saheelisLatticeOracleID, "Saheeli's Lattice", "Artifact", "{1}{R}",
		"Mastercraft Raptor", "Artifact Creature — Dinosaur", "*", "4", []string{"R"}))
	pushMaterialPermanent(g, me, game.Card{Name: "Rex", TypeLine: "Creature — Dinosaur", Power: 3, Toughness: 3})
	pushGraveyardMaterial(me, game.Card{Name: "Fossil", TypeLine: "Creature — Dinosaur", Power: 4, Toughness: 4})
	pushGraveyardMaterial(me, game.Card{Name: "Bones", TypeLine: "Creature — Dinosaur", Power: 1, Toughness: 1})
	floatMana(t, g, me, "{C}{C}{C}{C}{R}")

	counts := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != lattice {
			continue
		}
		var p struct {
			ExilePermanentIDs []string `json:"exile_permanent_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		counts[len(p.ExilePermanentIDs)] = true
	}
	if !counts[1] || !counts[2] || !counts[3] {
		t.Errorf("payment sizes offered = %v, want the floor (1), the graveyard (2) and everything (3)", counts)
	}
}
