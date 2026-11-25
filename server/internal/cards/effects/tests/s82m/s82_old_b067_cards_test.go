package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b067_cards_test.go — Kaldheim Equipment (S82, set OLD batch 067).

const (
	oldB067RavenWings    = "2d668f67-7cde-447a-8eb3-35c0c5c5ab69"
	oldB067TormentorsHlm = "4f5f0988-0b1f-4f4b-bcee-cc76fa484d9e"
	oldB067DraugrsHelm   = "8762e9ec-e026-45aa-9c7b-d894d17c75a2"
	oldB067BeardedAxe    = "a231319d-5d72-4aa6-8868-5987694791cd"
	oldB067RunedCrown    = "b8739fc4-6308-4e49-be2e-d9d92d890ee1"
	oldB067Halvar        = "acfc3e07-757f-4ccf-a65e-bee07ce87f40"
	oldB067Sword         = "acfc3e07-757f-4ccf-a65e-bee07ce87f40#1"
	oldB067Toralf        = "fffcb71f-e802-436e-917d-eaa6607ab74f"
	oldB067Hammer        = "fffcb71f-e802-436e-917d-eaa6607ab74f#1"
)

func TestOldB067EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB067RavenWings:    "Raven Wings",
		oldB067TormentorsHlm: "Tormentor's Helm",
		oldB067DraugrsHelm:   "Draugr's Helm",
		oldB067BeardedAxe:    "Bearded Axe",
		oldB067RunedCrown:    "Runed Crown",
		oldB067Halvar:        "Halvar, God of Battle",
		oldB067Sword:         "Sword of the Realms",
		oldB067Toralf:        "Toralf, God of Fury",
		oldB067Hammer:        "Toralf's Hammer",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func oldB067Attach(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if !effectstest.AttachedTo(t, g, attachment, host) {
		t.Fatal("the attachment did not attach")
	}
}

func TestOldB067RavenWingsPumpsGivesFlyingAndABirdType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	wings := effectstest.SeedEquipment(g, me.ID, "Raven Wings", oldB067RavenWings)
	oldB067Attach(t, g, wings, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 3/2", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("no flying")
	}
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !c.HasSubtype("Bird") || !c.HasSubtype("Bear") {
		t.Errorf("subtypes %v, want Bear and Bird", c.Effective().Subtypes)
	}
}

func TestOldB067TormentorsHelmHurtsTheDefenderWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.SeedBear(g, me.ID)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	helm := effectstest.SeedEquipment(g, me.ID, "Tormentor's Helm", oldB067TormentorsHlm)
	oldB067Attach(t, g, helm, attacker)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 3 {
		t.Fatalf("P/T = %d/%d, want 3/3", p, tough)
	}
	life, other := opp.Life, g.Seats[2].Life
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("defending player's life %d, want %d", opp.Life, life-1)
	}
	if g.Seats[2].Life != other {
		t.Error("a player who was not attacked lost life")
	}
}

func TestOldB067TormentorsHelmDoesNothingUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.SeedBear(g, me.ID)
	helm := effectstest.SeedEquipment(g, me.ID, "Tormentor's Helm", oldB067TormentorsHlm)
	oldB067Attach(t, g, helm, attacker)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("an unblocked attacker cost the defender %d life before damage", life-opp.Life)
	}
}

func TestOldB067DraugrsHelmMayPayForAZombieAndAttaches(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		helm := effectstest.TdCastFromHand(t, g, game.Card{Name: "Draugr's Helm", TypeLine: effectstest.EquipTypeLine, OracleID: oldB067DraugrsHelm}, game.CastSpellParams{})
		if pay {
			effectstest.FloatMana(t, g, me, "{B}{B}{B}")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if ask := effectstest.OptionalOffer(g, me); ask != nil {
			effectstest.AnswerOptional(t, g, me, pay)
		}
		if effectstest.HasPayUnlessFor(g, me.ID) {
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
		}
		effectstest.PassPriorityAroundTable(t, g)
		zombies := effectstest.BattlefieldIDsNamed(g, "Zombie Berserker")
		if !pay {
			if len(zombies) != 0 {
				t.Fatalf("declined, but %d Zombie Berserkers", len(zombies))
			}
			continue
		}
		if len(zombies) != 1 {
			t.Fatalf("paid, but %d Zombie Berserkers", len(zombies))
		}
		if !effectstest.AttachedTo(t, g, helm, zombies[0]) {
			t.Error("the Helm is not attached to the token")
		}
		if p, tough := effectstest.EffectivePTOf(t, g, zombies[0]); p != 4 || tough != 4 {
			t.Errorf("P/T = %d/%d, want 4/4", p, tough)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, zombies[0], "menace") {
			t.Error("no menace")
		}
	}
}

func TestOldB067BeardedAxeCountsDwarvesEquipmentAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	axe := effectstest.SeedEquipment(g, me.ID, "Bearded Axe", oldB067BeardedAxe)
	oldB067Attach(t, g, axe, bear)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Fatalf("with only the Axe, power %d, want 3 (+1)", p)
	}
	effectstest.B12Creature(g, me.ID, "Dwarf", "Creature — Dwarf Warrior", 1, 1)
	effectstest.SeedEquipment(g, me.ID, "Other Equipment", "")
	effectstest.PushVehicleForTest(g, me.ID, "Skiff", "", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 1, 1)
	// Dwarf, Axe, Equipment, Vehicle: four.
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("P/T = %d/%d, want 6/6", p, tough)
	}
}

func TestOldB067SwordOfTheRealmsPumpsAndReturnsItsCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Sword of the Realms", oldB067Sword)
	oldB067Attach(t, g, sword, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 4/2", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

func TestOldB067HalvarGivesDoubleStrikeToEnchantedAndEquippedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Halvar, God of Battle", "Legendary Creature — God", oldB067Halvar, false)
	plain := effectstest.SeedBear(g, me.ID)
	equipped := effectstest.SeedBear(g, me.ID)
	enchanted := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	oldB067Attach(t, g, effectstest.SeedEquipment(g, me.ID, "Test Equipment", ""), equipped)
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
	oldB067Attach(t, g, aura, enchanted)
	oldB067Attach(t, g, effectstest.SeedEquipment(g, opp.ID, "Their Equipment", ""), theirs)
	for name, want := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"plain": {plain, false}, "equipped": {equipped, true}, "enchanted": {enchanted, true}, "opponent's": {theirs, false}} {
		if got := effectstest.EffectiveAbilitiesContain(t, g, want.id, "double strike"); got != want.want {
			t.Errorf("%s creature has double strike = %v, want %v", name, got, want.want)
		}
	}
}

func TestOldB067HalvarMovesAnAttachmentAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Halvar, God of Battle", "Legendary Creature — God", oldB067Halvar, false)
	from := effectstest.SeedBear(g, me.ID)
	to := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Test Equipment", "")
	oldB067Attach(t, g, sword, from)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	answers := []uuid.UUID{sword, to}
	for i := 0; i < 6; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil && len(answers) > 0 {
			id := answers[0]
			answers = answers[1:]
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.AttachedTo(t, g, sword, to) {
		for _, c := range g.PendingChoices {
			t.Logf("pending: kind %s chooser %v", c.Kind, c.Chooser)
		}
		t.Logf("stack %d", len(g.StackMeta))
		t.Error("the Equipment did not move to the target creature")
	}
}

func oldB067PushToralf(g *game.Game, me uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, me, "Toralf, God of Fury", "Legendary Creature — God", oldB067Toralf, false)
}

func TestOldB067ToralfRedirectsExcessNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toralf := oldB067PushToralf(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 5, 5)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Ogre", 1, 1)
	life := opp.Life
	effectstest.Cmdr2B08Damage(g, src, victim, 5) // lethal damage is 2: excess 3
	for i := 0; i < 4; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d (3 excess damage)", opp.Life, life-3)
	}
	_ = toralf
	_ = other
}

func TestOldB067ToralfCannotTargetTheDamagedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB067PushToralf(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Ogre", 1, 1)
	effectstest.Cmdr2B08Damage(g, src, victim, 4)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Toralf's trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err == nil {
		t.Error("the damaged permanent was accepted as the target")
	}
}

func TestOldB067ToralfIgnoresDamageThatIsNotExcessOrNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB067PushToralf(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Ogre", 1, 1)
	effectstest.Cmdr2B08Damage(g, src, victim, 3) // exactly lethal: no excess
	effectstest.Cmdr2B08Damage(g, src, mine, 6)   // excess, but it is Toralf's controller's
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Errorf("Toralf triggered: %+v", p)
	}
}

func TestOldB067ToralfsHammerPumpsLegendaryCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.SeedBear(g, me.ID)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	h1 := effectstest.SeedEquipment(g, me.ID, "Toralf's Hammer", oldB067Hammer)
	h2 := effectstest.SeedEquipment(g, me.ID, "Toralf's Hammer", oldB067Hammer)
	oldB067Attach(t, g, h1, plain)
	oldB067Attach(t, g, h2, legend)
	if p, _ := effectstest.EffectivePTOf(t, g, plain); p != 2 {
		t.Errorf("a nonlegendary creature has power %d, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, legend); p != 5 {
		t.Errorf("a legendary creature has power %d, want 5", p)
	}
}

func TestOldB067ToralfsHammerThrowsAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Toralf's Hammer", oldB067Hammer)
	oldB067Attach(t, g, hammer, bear)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.AttachedTo(t, g, hammer, bear) {
		t.Error("the Hammer was not unattached as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if !me.Hand.Contains(hammer) || effectstest.BattlefieldHas(g, hammer) {
		t.Error("the Hammer did not return to its owner's hand")
	}
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the creature was not tapped")
	}
}

func TestOldB067ToralfsHammerRefusesWithoutManaOrWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Toralf's Hammer", oldB067Hammer)
	oldB067Attach(t, g, hammer, bear)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("activated with no mana")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("activated while tapped")
	}
	if !effectstest.AttachedTo(t, g, hammer, bear) {
		t.Error("a refused activation unattached the Hammer")
	}
}

func TestOldB067RunedCrownPutsARuneOntoTheBattlefieldAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.StockLibraryFor(g, me, 6)
	// No Rune is catalogued yet: Song of the Dryads' "enchant permanent"
	// spec under a Rune type line stands in for one.
	rune := effectstest.HandCardFull(me, "Test Rune", "Enchantment — Aura Rune", "", effectstest.SongOfTheDryadsOracle, nil)
	crown := effectstest.TdCastFromHand(t, g, game.Card{Name: "Runed Crown", TypeLine: effectstest.EquipTypeLine, OracleID: oldB067RunedCrown}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 6 && !effectstest.BattlefieldHas(g, rune); i++ {
		if ask := effectstest.OptionalOffer(g, me); ask != nil {
			effectstest.AnswerOptional(t, g, me, true)
		}
		if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		if effectstest.SearchChoiceFor(g, me.ID) != nil {
			effectstest.AnswerSearchByID(t, g, me.ID, rune)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.BattlefieldHas(g, rune) {
		for _, c := range g.PendingChoices {
			t.Logf("pending: kind %s chooser %v", c.Kind, c.Chooser)
		}
		t.Fatal("the Rune did not enter the battlefield")
	}
	if !effectstest.AttachedTo(t, g, rune, crown) {
		t.Error("the Rune is not attached to the Crown")
	}
}
