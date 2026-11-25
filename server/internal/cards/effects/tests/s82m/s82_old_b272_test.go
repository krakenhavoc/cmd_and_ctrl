package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b272_test.go — Conspiracy: Take the Crown, S82 older-sets batch
// 272 (#5540): Auras, counters, enters-and-dies triggers, creatures with
// keywords. Each test plays the real catalog entry.

const (
	oldB272BondsOracle      = "03ce87f1-cdda-4121-8c0a-988642053055"
	oldB272FootstepsOracle  = "732b1314-68bb-4d3d-9f3f-7e66e763caf5"
	oldB272PossessionOracle = "8104a51d-f07b-45d6-aaa7-c6386d6de8bf"
	oldB272SphinxOracle     = "1c9cf20c-51e4-4a87-b9f8-231b6baf5d6b"
	oldB272VampireOracle    = "616b3ee0-4678-4562-a21a-5f2d31b9964c"
	oldB272WardenOracle     = "b50c1dd3-0495-427f-b5a9-9ac44a4554d1"
	oldB272KeeperOracle     = "1fa0f877-e26c-4a26-8c69-4bf8fcdd30b5"
	oldB272SpiritOracle     = "b34449d1-eab2-4be1-8351-568dc5ced7c3"
	oldB272CourierOracle    = "efaa1262-b7ac-49ab-831a-d10b59680691"
	oldB272SentryOracle     = "263b3472-b437-4889-af5e-c7537abb6cb6"
	oldB272SkyscoutOracle   = "381eb146-879e-473c-b628-65f6c0814a29"
	oldB272SpawnOracle      = "c1b8cfe9-e8fc-4321-94ce-86ab1f1ae2d0"
)

func TestOldB272EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB272BondsOracle: "Bonds of Quicksilver", oldB272FootstepsOracle: "Followed Footsteps",
		oldB272PossessionOracle: "Ghostly Possession", oldB272SphinxOracle: "Sphinx of Magosi",
		oldB272VampireOracle: "Havengul Vampire", oldB272WardenOracle: "Throne Warden", oldB272KeeperOracle: "Keeper of Keys",
		oldB272SpiritOracle: "Hollowhenge Spirit", oldB272CourierOracle: "Canal Courier", oldB272SentryOracle: "Ogre Sentry",
		oldB272SkyscoutOracle: "Merfolk Skyscout", oldB272SpawnOracle: "Vertigo Spawn",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB272Enter puts a catalog creature into owner's hand and moves it onto
// the battlefield through the real zone move, then clears its summoning
// sickness so it can attack this turn.
func oldB272Enter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := effectstest.PlayerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func oldB272Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.E2Card(t, g, id).Counters[game.CounterPlusOne]
}

// oldB272AuraOn casts the Aura named `name` from the active seat's hand onto
// `host` and lets it resolve.
func oldB272AuraOn(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura", oracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Bonds of Quicksilver -------------------------------------------------

func TestOldB272BondsOfQuicksilverKeepsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	spec, _ := effects.Lookup(oldB272BondsOracle)
	if !effectstest.HasString(spec.PrintedKeywords, "flash") {
		t.Error("Bonds of Quicksilver has no flash")
	}
	oldB272AuraOn(t, g, "Bonds of Quicksilver", oldB272BondsOracle, ogre)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if !effectstest.E2Card(t, g, ogre).Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestOldB272BondsOfQuicksilverRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Bonds of Quicksilver", "Enchantment — Aura", oldB272BondsOracle,
		[]game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)}); err == nil {
		t.Error("an Aura that enchants a creature enchanted a player")
	}
}

// --- Followed Footsteps ---------------------------------------------------

func TestOldB272FollowedFootstepsCopiesTheEnchantedCreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	oldB272AuraOn(t, g, "Followed Footsteps", oldB272FootstepsOracle, ogre)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.NamedOnBattlefield(g, "Ogre", me.ID); got != 1 {
		t.Fatalf("I control %d Ogre tokens, want 1", got)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ogre" && c.Controller == me.ID {
			token = c
		}
	}
	if !token.IsToken() || token.Power != 3 || token.Toughness != 3 {
		t.Errorf("the copy is token=%v %d/%d, want a 3/3 token", token.IsToken(), token.Power, token.Toughness)
	}
	if got := effectstest.NamedOnBattlefield(g, "Ogre", opp.ID); got != 1 {
		t.Errorf("the original left its controller: %d", got)
	}
}

// --- Ghostly Possession ---------------------------------------------------

func TestOldB272GhostlyPossessionGrantsFlyingAndPreventsCombatDamageBothWays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Keywords: []string{"flying"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	oldB272AuraOn(t, g, "Ghostly Possession", oldB272PossessionOracle, host)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, host), "flying") {
		t.Error("the enchanted creature has no flying")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, host)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, host); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bDamageStep(t, g)
	if got := effectstest.DamageMarkedOn(g, host); got != 0 {
		t.Errorf("the enchanted creature took %d combat damage, want 0", got)
	}
	if got := effectstest.DamageMarkedOn(g, blocker); got != 0 {
		t.Errorf("the blocker took %d combat damage from the enchanted creature, want 0", got)
	}
}

func TestOldB272GhostlyPossessionLeavesNoncombatDamageAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	oldB272AuraOn(t, g, "Ghostly Possession", oldB272PossessionOracle, host)
	g.WithWriteLock(func() {
		if err := (effects.DealDamage{Source: host, Target: host, Amount: 2}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Errorf("damage: %v", err)
		}
	})
	if got := effectstest.DamageMarkedOn(g, host); got != 2 {
		t.Errorf("noncombat damage marked %d, want 2 (only combat damage is prevented)", got)
	}
}

// --- Sphinx of Magosi -----------------------------------------------------

func TestOldB272SphinxOfMagosiDrawsThenGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sphinx := effectstest.PushMonster(g, me.ID, "Sphinx of Magosi", oldB272SphinxOracle, 6, 6)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if got := oldB272Counters(t, g, sphinx); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
}

// --- Havengul Vampire -----------------------------------------------------

func TestOldB272HavengulVampireGrowsOnPlayerDamageAndWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := oldB272Enter(t, g, me.ID, "Havengul Vampire", "Creature — Vampire", oldB272VampireOracle, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, vamp, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB272Counters(t, g, vamp); got != 1 {
		t.Fatalf("counters after combat damage to a player: %d, want 1", got)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB272Counters(t, g, vamp); got != 2 {
		t.Errorf("counters after another creature died: %d, want 2", got)
	}
	effectstest.KillCreature(t, g, me.ID, vamp)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vamp) {
		t.Error("the Vampire is still on the battlefield")
	}
}

func TestOldB272HavengulVampireIgnoresNoncombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := oldB272Enter(t, g, me.ID, "Havengul Vampire", "Creature — Vampire", oldB272VampireOracle, 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: vamp, Target: opp.ID, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB272Counters(t, g, vamp); got != 0 {
		t.Errorf("counters after noncombat damage: %d, want 0", got)
	}
}

// --- Throne Warden --------------------------------------------------------

func TestOldB272ThroneWardenGrowsOnlyForTheMonarch(t *testing.T) {
	for _, monarch := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		warden := oldB272Enter(t, g, me.ID, "Throne Warden", "Creature — Human Soldier", oldB272WardenOracle, 2, 2)
		if monarch {
			effectstest.MonCrown(t, g, me.ID)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) // the crown's own draw trigger beside the Warden's
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if monarch {
			want = 1
		}
		if got := oldB272Counters(t, g, warden); got != want {
			t.Errorf("monarch=%v: counters %d, want %d", monarch, got, want)
		}
	}
}

// --- Keeper of Keys -------------------------------------------------------

func TestOldB272KeeperOfKeysMakesYouMonarchAndYourCreaturesUnblockableAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB272Enter(t, g, me.ID, "Keeper of Keys", "Creature — Human Rogue Mutant", oldB272KeeperOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("the monarch is %v, want me", g.Monarch)
	}
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g) // wraps through the other seats' turns first
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, other); err == nil {
		t.Error("a creature I control was blocked after Keeper of Keys's upkeep trigger")
	}
}

// --- Hollowhenge Spirit ---------------------------------------------------

func TestOldB272HollowhengeSpiritRemovesAnAttackerFromCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bCastNow(t, g, "Hollowhenge Spirit", "Creature — Spirit", oldB272SpiritOracle, game.CastSpellParams{})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, attacker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.E2Card(t, g, attacker).AttackingTarget; got != uuid.Nil {
		t.Error("the attacker is still attacking")
	}
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Errorf("the creature removed from combat dealt %d damage", life-opp.Life)
	}
}

// --- Canal Courier --------------------------------------------------------

func oldB272CourierAttack(t *testing.T, g *game.Game, courierTarget, otherTarget uuid.UUID) (courier, blocker uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	courier = oldB272Enter(t, g, me.ID, "Canal Courier", "Creature — Human Rogue", oldB272CourierOracle, 3, 5)
	effectstest.PassPriorityAroundTable(t, g) // the monarch trigger
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	blocker = effectstest.B12Creature(g, g.Seats[1].ID, "Blocker", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(courier, courierTarget); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(other, otherTarget); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return courier, blocker
}

func TestOldB272CanalCourierIsUnblockableWhenAttackingADifferentPlayerThanAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	courier, blocker := oldB272CourierAttack(t, g, g.Seats[1].ID, g.Seats[2].ID)
	if g.Monarch != g.Seats[0].ID {
		t.Error("the Courier did not make me the monarch")
	}
	if err := g.DeclareBlocker(blocker, courier); err == nil {
		t.Error("the Courier was blocked although it attacked a different player than another creature")
	}
}

func TestOldB272CanalCourierCanBeBlockedWhenEveryoneAttacksTheSamePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	courier, blocker := oldB272CourierAttack(t, g, g.Seats[1].ID, g.Seats[1].ID)
	if err := g.DeclareBlocker(blocker, courier); err != nil {
		t.Errorf("the Courier could not be blocked when both attacked the same player: %v", err)
	}
}

// --- Ogre Sentry ----------------------------------------------------------

func TestOldB272OgreSentryCannotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := oldB272Enter(t, g, me.ID, "Ogre Sentry", "Creature — Ogre Warrior", oldB272SentryOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sentry, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

// --- Merfolk Skyscout -----------------------------------------------------

func TestOldB272MerfolkSkyscoutUntapsATargetPermanentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scout := oldB272Enter(t, g, me.ID, "Merfolk Skyscout", "Creature — Merfolk Scout", oldB272SkyscoutOracle, 2, 3)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(scout, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, tapped)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Card(t, g, tapped).Tapped {
		t.Error("the targeted permanent is still tapped")
	}
}

// --- Vertigo Spawn --------------------------------------------------------

func TestOldB272VertigoSpawnTapsAndLocksTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spawn := oldB272Enter(t, g, me.ID, "Vertigo Spawn", "Creature — Illusion", oldB272SpawnOracle, 0, 3)
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Bear",
		Keywords: []string{"vigilance"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(spawn, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.E2Card(t, g, attacker).Tapped {
		t.Fatal("the blocked creature was not tapped")
	}
	// Its controller's next untap step comes after this turn.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain) // this turn's postcombat is over: the NEXT seat-1 turn
	if g.Turn.ActiveSeat != 1 {
		t.Fatal("setup")
	}
	if !effectstest.E2Card(t, g, attacker).Tapped {
		t.Error("the blocked creature untapped in its controller's next untap step")
	}
}
