package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b761_cards_test.go — Tempest, batch 761 (S82 older sets, #5628). Every
// test plays the real catalog entry through the engine.

const (
	oldB761Duplicity      = "01072a45-ab1c-4d19-b493-76cf4aae351e"
	oldB761Starke         = "07bc407e-7ba1-45aa-a562-0597ebe56044"
	oldB761OracleEnVec    = "367faa8c-617a-41a0-901c-f65106e86d56"
	oldB761SereneOffering = "6cef0548-6010-4781-8e83-431d014dd1da"
	oldB761RatsOfRath     = "75458175-eeb1-4a6b-93d9-c744c93b3759"
	oldB761Apocalypse     = "82c8f5dd-563d-4fd0-bd43-7ee2001d3777"
	oldB761MaddeningImp   = "97dea59e-2090-4653-9ed9-86820a20605f"
	oldB761CloudchaserEag = "add18787-2329-4355-a524-66fefda8d06e"
	oldB761KnightOfDusk   = "cd1ca3c8-1f03-4c03-a93a-0419ac205c18"
	oldB761Perish         = "dd84d291-cb7a-4f44-81cd-9f1181bb5ae5"
	oldB761ColdStorage    = "e23bf6bd-bc97-425d-902a-c6b606e82b65"
	oldB761Extinction     = "f3f01fe8-0a2c-4afa-94d2-268d27c0d252"
	oldB761Verdigris      = "fb41ea1d-3491-4193-a390-8eb202fbae12"
)

func TestOldB761EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB761Duplicity, oldB761Starke, oldB761OracleEnVec, oldB761SereneOffering, oldB761RatsOfRath,
		oldB761Apocalypse, oldB761MaddeningImp, oldB761CloudchaserEag, oldB761KnightOfDusk, oldB761Perish, oldB761ColdStorage,
		oldB761Extinction, oldB761Verdigris} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB761Cast puts the card in the active seat's hand and casts it strictly.
func oldB761Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.D5681Spell(g, me, uuid.New(), name, typeLine, "", oracle)
	var id uuid.UUID
	g.ReadSnapshot(func() { id = me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID })
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

// oldB761OnBattlefieldNamed finds a permanent by name: a card that changed zones
// is a new object (CR 400.7), so its old instance id no longer names it.
func oldB761OnBattlefieldNamed(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func oldB761Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func oldB761Perm(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestOldB761VerdigrisDestroysAnArtifactAndRefusesAnythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := oldB761Cast(t, g, "Verdigris", "Instant", oldB761Verdigris, oldB761Card(bear)); err == nil {
		t.Fatal("targeted a nonartifact creature")
	}
	if err := oldB761Cast(t, g, "Verdigris", "Instant", oldB761Verdigris, oldB761Card(rock)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature died")
	}
}

func TestOldB761SereneOfferingDestroysTheEnchantmentAndGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	ench := oldB761Perm(g, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment", ManaCost: "{2}{W}{W}"})
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := oldB761Cast(t, g, "Serene Offering", "Instant", oldB761SereneOffering, oldB761Card(bear)); err == nil {
		t.Fatal("targeted a creature")
	}
	life := me.Life
	if err := oldB761Cast(t, g, "Serene Offering", "Instant", oldB761SereneOffering, oldB761Card(ench)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d→%d, want +4 (mana value of {2}{W}{W})", life, me.Life)
	}
}

func TestOldB761CloudchaserEagleDestroysAnEnchantmentWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.PushCatalogPermanent(g, opp.ID, "Their Aura", "Enchantment", "", false)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	eagle := effectstest.CastHoldCreature(t, g, "Cloudchaser Eagle", "Creature — Bird", oldB761CloudchaserEag, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, rock) {
		t.Fatal("an artifact is a legal target")
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived the enters trigger")
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(eagle) {
		t.Error("the artifact was destroyed or the Eagle did not enter")
	}
}

func TestOldB761PerishDestroysGreenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mkCreature := func(owner uuid.UUID, name string, colors []string) uuid.UUID {
		return oldB761Perm(g, owner, game.Card{Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: colors})
	}
	green, mine, red := mkCreature(opp.ID, "Their Green", []string{"G"}), mkCreature(me.ID, "My Green", []string{"G"}), mkCreature(opp.ID, "Their Red", []string{"R"})
	gold := mkCreature(opp.ID, "Their Gold", []string{"G", "W"})
	if err := oldB761Cast(t, g, "Perish", "Sorcery", oldB761Perish); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{green, mine, gold} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a green creature survived: %v", id)
		}
	}
	if !g.Battlefield.Contains(red) {
		t.Error("the red creature died")
	}
}

func TestOldB761ExtinctionDestroysEveryCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.PushTribalCreature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	myGoblin := effectstest.PushTribalCreature(g, me.ID, "My Goblin", "Creature — Goblin Warrior", 1, 1)
	elf := effectstest.PushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Their Shifter", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	if err := oldB761Cast(t, g, "Extinction", "Sorcery", oldB761Extinction); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("resolving did not ask for a creature type")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Fatal("creatures died before the type was chosen")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	if g.Battlefield.Contains(goblin) || g.Battlefield.Contains(myGoblin) {
		t.Error("a Goblin survived")
	}
	if g.Battlefield.Contains(changeling) {
		t.Error("the changeling survived, but it is a Goblin")
	}
	if !g.Battlefield.Contains(elf) {
		t.Error("the Elf died")
	}
}

func TestOldB761ApocalypseExilesEveryPermanentAndDiscardsYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	oppHand := effectstest.HandCardFull(opp, "Their Card", "Instant", "{U}", "", []string{"U"})
	spare := effectstest.HandCardFull(me, "My Spare", "Instant", "{U}", "", []string{"U"})
	if err := oldB761Cast(t, g, "Apocalypse", "Sorcery", oldB761Apocalypse); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a permanent was not exiled: %v", id)
		}
		if effectstest.ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("permanent %v is in %v, want exile", id, effectstest.ZoneOf(g, id))
		}
	}
	if me.Hand.Size() != 0 || !me.Graveyard.Contains(spare) {
		t.Errorf("your hand was not discarded: size %d", me.Hand.Size())
	}
	if !opp.Hand.Contains(oppHand) {
		t.Error("the opponent's hand was discarded")
	}
}

func TestOldB761KnightOfDuskDestroysACreatureBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.PushCatalogPermanent(g, me.ID, "Knight of Dusk", "Creature — Human Knight", oldB761KnightOfDusk, false)
	blocker := effectstest.CtrlPushCreature(g, opp.ID, "Blocker")
	bystander := effectstest.CtrlPushCreature(g, opp.ID, "Bystander")
	effectstest.CrtBlock(g, blocker, knight, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, knight, 0, effectstest.LtCardTarget(bystander), "B", "B"); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a creature that is not blocking the Knight = %v, want ErrIllegalTarget", err)
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	if err := effectstest.CrtActivateRaw(t, g, me.ID, knight, 0, effectstest.LtCardTarget(blocker), "B", "B"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocker survived")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(knight) {
		t.Error("something other than the blocker was destroyed")
	}
}

func TestOldB761RatsOfRathDestroysOnlyYourOwnPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B12Push(g, me.ID, "Rats of Rath", "Creature — Rat", oldB761RatsOfRath, 2, 1)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Swamp", "Basic Land — Swamp")
	enchant := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	activate := func(target uuid.UUID) error {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
			t.Fatal(err)
		}
		return g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB761Card(target)}})
	}
	for name, id := range map[string]uuid.UUID{"an opponent's artifact": theirs, "an enchantment": enchant} {
		if err := activate(id); err == nil {
			t.Errorf("targeted %s", name)
		}
	}
	me.ManaPool.EmptyPool()
	for _, id := range []uuid.UUID{rock, land, rats} {
		if err := activate(id); err != nil {
			t.Fatalf("activate on %v: %v", id, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("%v survived", id)
		}
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(enchant) {
		t.Error("a permanent that was not a legal target was destroyed")
	}
}

func TestOldB761ColdStorageExilesYourCreatureAndReturnsItUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	storage := effectstest.PushCatalogPermanent(g, me.ID, "Cold Storage", "Artifact", oldB761ColdStorage, false)
	storage2 := effectstest.PushCatalogPermanent(g, me.ID, "Cold Storage", "Artifact", oldB761ColdStorage, false)
	mine := oldB761Perm(g, me.ID, game.Card{Name: "My Bear", OracleID: uuid.NewString(), TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	if err := effectstest.CrtActivateRaw(t, g, me.ID, storage, 0, effectstest.LtCardTarget(theirs), "C", "C", "C"); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("exiled a creature you do not control: %v", err)
	}
	me.ManaPool = nil
	if err := effectstest.CrtActivateRaw(t, g, me.ID, storage, 0, effectstest.LtCardTarget(mine), "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, mine) != game.ZoneExile {
		t.Fatalf("the creature is in %v, want exile", effectstest.ZoneOf(g, mine))
	}
	// A second Cold Storage has exiled nothing: its sacrifice returns nothing.
	if err := g.ActivateCatalogAbility(me.ID, storage2, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice the second: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, mine) != game.ZoneExile {
		t.Fatal("another Cold Storage returned the card")
	}
	if err := g.ActivateCatalogAbility(me.ID, storage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := oldB761OnBattlefieldNamed(g, "My Bear")
	if !ok || c.Controller != me.ID {
		t.Fatalf("the creature did not return under your control: %v %+v", ok, c.Controller)
	}
	if g.Battlefield.Contains(storage) {
		t.Error("the artifact was not sacrificed")
	}
}

func TestOldB761ColdStorageReturnsACreatureYouDoNotOwnUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	storage := effectstest.PushCatalogPermanent(g, me.ID, "Cold Storage", "Artifact", oldB761ColdStorage, false)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear", OracleID: uuid.NewString(), TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	if err := effectstest.CrtActivateRaw(t, g, me.ID, storage, 0, effectstest.LtCardTarget(stolen), "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, storage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := oldB761OnBattlefieldNamed(g, "Stolen Bear")
	if !ok || c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("want it back under my control, owned by the opponent: %v %v/%v", ok, c.Controller, c.Owner)
	}
}

func TestOldB761StarkeDestroysThenItsControllerGainsControlOfStarke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	starke := effectstest.PushCatalogPermanent(g, me.ID, "Starke of Rath", "Legendary Creature — Human Rogue", oldB761Starke, false)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	land := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, starke, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.LtCardTarget(land)}); err == nil {
		t.Fatal("targeted a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, starke, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.LtCardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	c, ok := effectstest.BattlefieldCard(g, starke)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("the creature's controller did not gain control of Starke: %v %+v", ok, c)
	}
	// Indefinite: still theirs a turn later.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, starke); c.Controller != opp.ID {
		t.Error("control of Starke reverted")
	}
}

func TestOldB761StarkeAlsoChangesHandsWhenTheTargetIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	starke := effectstest.PushCatalogPermanent(g, me.ID, "Starke of Rath", "Legendary Creature — Human Rogue", oldB761Starke, false)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"indestructible"}})
	if err := g.ActivateCatalogAbility(me.ID, starke, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.LtCardTarget(golem)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Error("an indestructible permanent was destroyed")
	}
	if c, ok := effectstest.BattlefieldCard(g, starke); !ok || c.Controller != opp.ID {
		t.Error("Starke did not change hands")
	}
}

// --- Maddening Imp ----------------------------------------------------

func oldB761FinishTurnToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 8; i++ {
		c := effectstest.PendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			break
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB761MaddeningImpForcesEveryNonWallCreatureAndDestroysTheOnesThatDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Maddening Imp", "Creature — Imp", oldB761MaddeningImp, false)
	attacker := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Attacker", "", 2, 2))
	laggard := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Laggard", "", 2, 2))
	wall := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	mine := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Imp's controller's", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(opp.ID, imp, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FaMustAttack(t, g) {
		t.Fatal("the active player's creatures were not required to attack")
	}
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.RtRetap(g, laggard) // can't attack
	oldB761FinishTurnToEndStep(t, g)
	if !effects.OnBattlefield(g, attacker) {
		t.Error("the creature that attacked was destroyed")
	}
	if effects.OnBattlefield(g, laggard) {
		t.Error("the creature that did not attack survived")
	}
	if !effects.OnBattlefield(g, wall) || !effects.OnBattlefield(g, mine) {
		t.Error("a Wall or the Imp controller's own creature was destroyed")
	}
}

func TestOldB761MaddeningImpOnlyBeforeCombatAndOnlyOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Maddening Imp", "Creature — Imp", oldB761MaddeningImp, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AcRefused(t, g, opp.ID, imp, 0, game.ActivateAbilityParams{})
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AcRefused(t, g, opp.ID, imp, 0, game.ActivateAbilityParams{})
}

// --- Oracle en-Vec ----------------------------------------------------

func TestOldB761OracleEnVecForcesTheChosenAndBarsTheRestOnTheirNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	oracle := effectstest.PushCatalogPermanent(g, me.ID, "Oracle en-Vec", "Creature — Human Wizard", oldB761OracleEnVec, false)
	chosen := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Chosen", "", 2, 2))
	chosenIdle := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Chosen but tapped", "", 2, 2))
	other := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Other", "", 2, 2))
	if err := g.ActivateCatalogAbility(opp.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("the opponent activated it")
	}
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeted yourself")
	}
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{chosen, chosenIdle}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.AutoAnswerAll(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RtRetap(g, chosenIdle)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, me.ID); err == nil {
		t.Fatal("a creature that was not chosen attacked")
	}
	if err := g.DeclareAttacker(chosen, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	oldB761FinishTurnToEndStep(t, g)
	if !effects.OnBattlefield(g, chosen) {
		t.Error("the chosen creature that attacked was destroyed")
	}
	if effects.OnBattlefield(g, chosenIdle) {
		t.Error("the chosen creature that did not attack survived")
	}
	if !effects.OnBattlefield(g, other) {
		t.Error("a creature that was not chosen was destroyed")
	}
}

func TestOldB761OracleEnVecWithNoCreaturesChosenNoCreatureCanAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	oracle := effectstest.PushCatalogPermanent(g, me.ID, "Oracle en-Vec", "Creature — Human Wizard", oldB761OracleEnVec, false)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked to choose")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, nil); err != nil {
		t.Fatalf("choose none: %v", err)
	}
	effectstest.AutoAnswerAll(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("a creature attacked although none was chosen")
	}
}

// --- Duplicity --------------------------------------------------------

func TestOldB761DuplicityExilesFiveFaceDownDiscardsAtEndStepAndRefundsOnTheUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "L1", "L2", "L3", "L4", "L5", "L6", "L7")
	card := effectstest.HandCardFull(me, "Duplicity", "Enchantment", "", oldB761Duplicity, []string{"U"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range ids {
		if effectstest.ZoneOf(g, id) == game.ZoneExile {
			exiled++
		}
	}
	if exiled != 5 {
		t.Fatalf("%d library cards were exiled, want 5", exiled)
	}
	if got := len(g.ExiledWithForEffect(game.PermissionCardRef{ID: card, Epoch: effectstest.BattlefieldCardPtr(t, g, card).ObjectEpoch})); got != 5 {
		t.Errorf("%d cards are linked to Duplicity, want 5", got)
	}
}

// oldB761LinkTop exiles the top library card face down and linked to `source`,
// as Duplicity's enters trigger does, and names it.
func oldB761LinkTop(t *testing.T, g *game.Game, me *game.Player, source uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = me.Library.Cards[len(me.Library.Cards)-1].InstanceID
		if _, err := g.ExileFaceDownLinkedForEffect(source, me.ID, id); err != nil {
			t.Fatalf("exile linked: %v", err)
		}
	})
	return id
}

func TestOldB761DuplicityUpkeepExilesTheHandAndTakesBackTheOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToUpkeepOf(t, g, 1) // Duplicity has no trigger before my own turns
	effectstest.SeedLibrary(me, "L1", "L2", "L3", "L4")
	dup := effectstest.PushCatalogPermanent(g, me.ID, "Duplicity", "Enchantment", oldB761Duplicity, false)
	linked := []uuid.UUID{oldB761LinkTop(t, g, me, dup), oldB761LinkTop(t, g, me, dup)}
	// A card exiled with something else stays where it is.
	other := effectstest.B12Permanent(g, me.ID, "Other Exiler", "Artifact")
	stranger := oldB761LinkTop(t, g, me, other)
	held := effectstest.HandCardFull(me, "Held Card", "Instant", "{U}", "", []string{"U"})
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Hand.Contains(held) || effectstest.ZoneOf(g, held) != game.ZoneExile {
		t.Errorf("the hand card is in %v, want exile", effectstest.ZoneOf(g, held))
	}
	for _, id := range linked {
		if !me.Hand.Contains(id) {
			t.Errorf("an exiled card did not come back to hand: %v", effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, stranger) != game.ZoneExile {
		t.Error("a card exiled with another permanent was taken back")
	}
	if c := effectstest.ExiledCard(g, held); c != nil && c.IsKnownTo(me.ID) {
		t.Error("the controller can look at the hand exiled face down")
	}
}

func TestOldB761DuplicityUpkeepDeclinedLeavesEverythingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToUpkeepOf(t, g, 1) // Duplicity has no trigger before my own turns
	effectstest.SeedLibrary(me, "L1", "L2")
	dup := effectstest.PushCatalogPermanent(g, me.ID, "Duplicity", "Enchantment", oldB761Duplicity, false)
	linked := oldB761LinkTop(t, g, me, dup)
	held := effectstest.HandCardFull(me, "Held Card", "Instant", "{U}", "", []string{"U"})
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if !me.Hand.Contains(held) || effectstest.ZoneOf(g, linked) != game.ZoneExile {
		t.Error("declining changed the hand or the exile")
	}
}

func TestOldB761DuplicityEndStepMakesYouDiscardACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Duplicity", "Enchantment", oldB761Duplicity, false)
	a := effectstest.HandCardFull(me, "Card A", "Instant", "{U}", "", []string{"U"})
	b := effectstest.HandCardFull(me, "Card B", "Instant", "{U}", "", []string{"U"})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardPromptFor(g, me) == nil {
		t.Fatal("no discard prompt at the end step")
	}
	effectstest.AnswerDiscard(t, g, me.ID, b)
	if !me.Graveyard.Contains(b) || !me.Hand.Contains(a) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestOldB761DuplicityLeavingPutsEveryExiledCardIntoItsOwnersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "L1", "L2", "L3")
	dup := effectstest.PushCatalogPermanent(g, me.ID, "Duplicity", "Enchantment", oldB761Duplicity, false)
	linked := []uuid.UUID{oldB761LinkTop(t, g, me, dup), oldB761LinkTop(t, g, me, dup)}
	other := effectstest.B12Permanent(g, me.ID, "Other Exiler", "Artifact")
	stranger := oldB761LinkTop(t, g, me, other)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dup) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range linked {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a linked card is in %v, want the graveyard", effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, stranger) != game.ZoneExile {
		t.Error("a card exiled with another permanent was moved")
	}
}

func TestOldB761DuplicityTakenByAnOpponentSendsTheCardsToTheirOwnersGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "L1", "L2")
	dup := effectstest.PushCatalogPermanent(g, me.ID, "Duplicity", "Enchantment", oldB761Duplicity, false)
	linked := oldB761LinkTop(t, g, me, dup)
	effectstest.StealForTest(t, g, dup, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(linked) {
		t.Errorf("the exiled card is in %v after Duplicity changed hands, want its owner's graveyard", effectstest.ZoneOf(g, linked))
	}
}
