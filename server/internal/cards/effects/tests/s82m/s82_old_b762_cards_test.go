package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b762_cards_test.go — Tempest, batch 762 (S82 older sets, #5628). Every
// test plays the real catalog entry through the engine.

const (
	oldB762LowlandGiant = "041cc2d0-eb67-484e-8da1-e91ddb2b3f70"
	oldB762Shapeshifter = "17f2342b-4dd8-445a-a7c1-4a15a5688294"
	oldB762Servant      = "4e389b61-7b31-42ba-a19b-2502f63fd34a"
	oldB762MarbleTitan  = "62674cfd-f400-44d9-bfff-15bdfbd243fe"
	oldB762HeartSliver  = "6dd1f375-e8f2-4725-8fbd-9750c4860820"
	oldB762TalonSliver  = "74910702-e385-48f7-a6b3-064ebed08be7"
	oldB762Rootwater    = "8a449915-8bd7-4a9c-abb1-a95852f33695"
	oldB762Dryad        = "976968c5-7ef4-4451-86ff-03a011bbcca7"
	oldB762Warhounds    = "ae997fee-2c1f-43c8-9dba-8b409d8e449e"
	oldB762MoggConscr   = "b015686d-07e1-46c7-8e51-2bec646860d6"
	oldB762ApesOfRath   = "b52f5850-f1aa-412a-bae9-2ce76c287188"
	oldB762Eladamri     = "dcbb6b0e-2219-4ba4-9720-3e58b59b4328"
	oldB762Mindwhip     = "e9e43839-7cc1-40ff-a8ec-a47316de6440"

	oldB762Pacifism   = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	oldB762WildGrowth = "706ae742-1807-44b7-a4fa-f2e26f61519a"
)

func TestOldB762EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB762LowlandGiant, oldB762Shapeshifter, oldB762Servant, oldB762MarbleTitan, oldB762HeartSliver,
		oldB762TalonSliver, oldB762Rootwater, oldB762Dryad, oldB762Warhounds, oldB762MoggConscr, oldB762ApesOfRath, oldB762Eladamri, oldB762Mindwhip} {
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

func oldB762Has(t *testing.T, g *game.Game, id uuid.UUID, keyword string) bool {
	t.Helper()
	return effectstest.HasKeywordOnBattlefield(t, g, id, keyword)
}

func TestOldB762SliversGiveEverySliverTheirKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
	}{{"Heart Sliver", oldB762HeartSliver, "haste"}, {"Talon Sliver", oldB762TalonSliver, "first strike"}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := effectstest.B12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 1, 1)
			mine := effectstest.B12Creature(g, me.ID, "My Sliver", "Creature — Sliver", 1, 1)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			for _, id := range []uuid.UUID{lord, mine, theirs} {
				if !oldB762Has(t, g, id, tc.keyword) {
					t.Errorf("a Sliver lacks %s", tc.keyword)
				}
			}
			if oldB762Has(t, g, bear, tc.keyword) {
				t.Errorf("a Bear has %s", tc.keyword)
			}
		})
	}
}

func TestOldB762MindwhipSliverGrantsEverySliverTheSacrificeDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mindwhip Sliver", "Creature — Sliver", oldB762Mindwhip, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.HandCardFull(opp, "Card A", "Instant", "{U}", "", []string{"U"})
	b := effectstest.HandCardFull(opp, "Card B", "Instant", "{U}", "", []string{"U"})
	hand := opp.Hand.Size()
	activate := func(id uuid.UUID) error {
		return effectstest.CrtActivateRaw(t, g, me.ID, id, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, "C", "C")
	}
	if err := activate(bear); err == nil {
		t.Fatal("a Bear activated the Sliver ability")
	}
	me.ManaPool = nil
	if err := activate(other); err != nil {
		t.Fatalf("activate on another Sliver: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) {
		t.Error("the Sliver was not sacrificed")
	}
	if opp.Hand.Size() != hand-1 || opp.Graveyard.Size() != 1 {
		t.Errorf("the opponent did not discard exactly one card: hand %d→%d, graveyard %d", hand, opp.Hand.Size(), opp.Graveyard.Size())
	}
	_, _ = a, b
	// Only as a sorcery.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	another := effectstest.B12Creature(g, me.ID, "Third Sliver", "Creature — Sliver", 1, 1)
	if err := activate(another); err == nil {
		t.Fatal("activated outside the main phase")
	}
}

func TestOldB762EladamriGivesOtherElvesForestwalkAndShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eladamri := effectstest.B12Push(g, me.ID, "Eladamri, Lord of Leaves", "Legendary Creature — Elf Warrior", oldB762Eladamri, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf Druid", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	goblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	for _, id := range []uuid.UUID{mine, theirs} {
		for _, kw := range []string{"forestwalk", "shroud"} {
			if !oldB762Has(t, g, id, kw) {
				t.Errorf("an Elf lacks %s", kw)
			}
		}
	}
	for _, kw := range []string{"forestwalk", "shroud"} {
		if oldB762Has(t, g, eladamri, kw) {
			t.Errorf("Eladamri himself has %s", kw)
		}
		if oldB762Has(t, g, goblin, kw) {
			t.Errorf("a Goblin has %s", kw)
		}
	}
}

func TestOldB762MoggConscriptsAttackOnlyAfterACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gob := effectstest.B12Push(g, me.ID, "Mogg Conscripts", "Creature — Goblin", oldB762MoggConscr, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.P1g2Cast(t, g, "Zap", "Instant", "", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gob, opp.ID); err == nil {
		t.Fatal("the Conscripts attacked after only a noncreature spell")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	gob2 := effectstest.B12Push(g2, me2.ID, "Mogg Conscripts", "Creature — Goblin", oldB762MoggConscr, 2, 2)
	effectstest.AdvanceToMain(t, g2)
	effectstest.P1g2Cast(t, g2, "Bear", "Creature — Bear", "", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(gob2, opp2.ID); err != nil {
		t.Fatalf("the Conscripts could not attack after a creature spell: %v", err)
	}
}

func TestOldB762MarbleTitanKeepsBigCreaturesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.B12Push(g, me.ID, "Marble Titan", "Creature — Giant", oldB762MarbleTitan, 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Big", "Creature — Ogre", 4, 4)
	land := effectstest.B12Permanent(g, me.ID, "My Swamp", "Basic Land — Swamp")
	for _, id := range []uuid.UUID{titan, small, big, land} {
		effectstest.RtRetap(g, id)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	for name, tc := range map[string]struct {
		id         uuid.UUID
		wantTapped bool
	}{"Marble Titan": {titan, true}, "a 2-power creature": {small, false}, "an opponent's 4-power creature": {big, true}, "a land": {land, false}} {
		if tapped, ok := effectstest.BattlefieldCardTapped(g, tc.id); !ok || tapped != tc.wantTapped {
			t.Errorf("%s: tapped=%v, want %v", name, tapped, tc.wantTapped)
		}
	}
}

func TestOldB762ServantOfVolrathMakesYouSacrificeACreatureWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	servant := effectstest.B12Push(g, me.ID, "Servant of Volrath", "Creature — Minion", oldB762Servant, 3, 3)
	keep := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(servant) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(theirs) {
		t.Error("the wrong creature was sacrificed")
	}
}

func TestOldB762WarhoundsPutTheCreatureThatBlocksThemOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hounds := effectstest.B12Push(g, me.ID, "Elven Warhounds", "Creature — Dog", oldB762Warhounds, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, hounds)
	if err := g.DeclareBlocker(blocker, hounds); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Fatal("the blocker is still on the battlefield")
	}
	if n := opp.Library.Size(); n == 0 || opp.Library.Cards[len(opp.Library.Cards)-1].Name != "Blocker" {
		t.Errorf("the blocker is not on top of its owner's library (size %d)", n)
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that did not block was moved")
	}
}

func TestOldB762ApesOfRathDoNotUntapAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apes := effectstest.B12Push(g, me.ID, "Apes of Rath", "Creature — Ape", oldB762ApesOfRath, 5, 4)
	effectstest.BrAttack(t, g, apes)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	_ = opp
	if tapped, _ := effectstest.BattlefieldCardTapped(g, apes); !tapped {
		t.Fatal("the Apes did not tap to attack")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, apes); !tapped {
		t.Error("the Apes untapped during their controller's next untap step")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, apes); tapped {
		t.Error("the Apes stayed tapped through a second untap step")
	}
}

func TestOldB762RootwaterShamanLetsEnchantCreatureAurasBeCastAtInstantSpeed(t *testing.T) {
	cast := func(t *testing.T, withShaman bool, aura string, target game.TargetRef, g *game.Game) error {
		me := g.Seats[0]
		if withShaman {
			effectstest.B12Push(g, me.ID, "Rootwater Shaman", "Creature — Merfolk Shaman", oldB762Rootwater, 2, 2)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		id := uuid.New()
		oracle := map[string]string{"Pacifism": oldB762Pacifism, "Wild Growth": oldB762WildGrowth}[aura]
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: aura, TypeLine: "Enchantment — Aura", OracleID: oracle, Owner: me.ID, Controller: me.ID})
		})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{target}})
	}
	setup := func(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
		land := effectstest.B12Permanent(g, g.Seats[0].ID, "My Forest", "Basic Land — Forest")
		return g, bear, land
	}
	g, bear, _ := setup(t)
	if err := cast(t, false, "Pacifism", game.TargetRef{Kind: game.TargetCard, ID: bear}, g); err == nil {
		t.Fatal("cast an Aura outside the main phase without the Shaman")
	}
	g, bear, _ = setup(t)
	if err := cast(t, true, "Pacifism", game.TargetRef{Kind: game.TargetCard, ID: bear}, g); err != nil {
		t.Fatalf("an enchant creature Aura could not be cast at instant speed with the Shaman: %v", err)
	}
	g, _, land := setup(t)
	if err := cast(t, true, "Wild Growth", game.TargetRef{Kind: game.TargetCard, ID: land}, g); err == nil {
		t.Fatal("an enchant land Aura was cast at instant speed")
	}
}

func TestOldB762HeartwoodDryadBlocksCreaturesWithShadowAndStillBlocksTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shade := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shadow Dude", TypeLine: "Creature — Spirit",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"shadow"}})
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	dryad := effectstest.B12Push(g, opp.ID, "Heartwood Dryad", "Creature — Dryad", oldB762Dryad, 2, 1)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, shade, walker)
	if reason := effectstest.BlockRefusal(t, g, dryad, shade); reason != "" {
		t.Errorf("the Dryad could not block a creature with shadow: %q", reason)
	}
	if reason := effectstest.BlockRefusal(t, g, dryad, walker); reason != "" {
		t.Errorf("the Dryad could not block a creature without shadow: %q", reason)
	}
	if reason := effectstest.BlockRefusal(t, g, plain, shade); reason != game.BlockReasonShadow {
		t.Errorf("another creature blocking a creature with shadow = %q, want a shadow refusal", reason)
	}
	if oldB762Has(t, g, dryad, "shadow") {
		t.Error("the Dryad has shadow")
	}
}

func TestOldB762EscapedShapeshifterCopiesTheAbilitiesOfOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shifter := effectstest.B12Push(g, me.ID, "Escaped Shapeshifter", "Creature — Shapeshifter", oldB762Shapeshifter, 3, 4)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"}})
	_ = mine
	for _, kw := range []string{"flying", "first strike", "trample", "protection from red"} {
		if oldB762Has(t, g, shifter, kw) {
			t.Fatalf("the Shapeshifter has %s with no opposing creature having it", kw)
		}
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying", "protection from red"}})
	theirTrampler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Trampler", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"trample"}})
	other := effectstest.B12Push(g, opp.ID, "Escaped Shapeshifter", "Creature — Shapeshifter", oldB762Shapeshifter, 3, 4)
	for kw, want := range map[string]bool{"flying": true, "trample": true, "protection from red": true, "first strike": false, "protection from blue": false} {
		if got := oldB762Has(t, g, shifter, kw); got != want {
			t.Errorf("the Shapeshifter has %s = %v, want %v", kw, got, want)
		}
	}
	// A creature of the same name gives nothing, and the copy does not feed back.
	if oldB762Has(t, g, other, "first strike") {
		t.Error("the opposing Shapeshifter has first strike")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirTrampler) })
	if oldB762Has(t, g, shifter, "trample") {
		t.Error("the Shapeshifter kept trample after the creature that had it died")
	}
}
