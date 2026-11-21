package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Bonesplitter: the shape every other Equipment varies on ------

func TestEquipAttachesAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.BonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})

	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Fatalf("setup: power %d, want 2", got)
	}
	effectstest.EquipTo(t, g, me.ID, sword, bear)

	if host := effectstest.AttachmentHostOf(t, g, sword); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Bonesplitter is +2/+0, toughness %d, want 2", got)
	}
}

// CR 701.3a — a second activation MOVES the Equipment, and the bonus
// moves with it in the same beat.
func TestReEquipMovesTheBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	second := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ox", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.BonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})

	effectstest.EquipTo(t, g, me.ID, sword, first)
	effectstest.EquipTo(t, g, me.ID, sword, second)

	if got := effectstest.EffectivePower(t, g, first); got != 2 {
		t.Errorf("first creature kept the bonus: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, second); got != 3 {
		t.Errorf("second creature power %d, want 3", got)
	}
}

// CR 702.6a — equip is sorcery-speed.
func TestEquipIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.BonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	// newCatalogGame lands on Upkeep, which is not a sorcery window.
	err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err != game.ErrSorcerySpeedRequired {
		t.Errorf("equip on upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// CR 702.6a — "target creature you control".
func TestEquipRejectsACreatureYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.BonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("equip should reject a creature you do not control")
	}
}

// --- Skullclamp ---------------------------------------------------

// The whole engine: clamp a 1/1, it becomes a 2/0, the lethal-damage
// SBA kills it, the dies trigger draws two.
func TestSkullclampKillsAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	clamp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Skullclamp", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.SkullclampOracle, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()

	effectstest.EquipTo(t, g, me.ID, clamp, token)

	if g.Battlefield.Contains(token) {
		t.Fatal("a clamped 1/1 is a 2/0 and dies to the toughness SBA")
	}
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2", before, got)
	}
	// CR 704.5n: the Clamp itself stays, unattached, ready to move.
	if !g.Battlefield.Contains(clamp) {
		t.Fatal("Skullclamp should stay on the battlefield")
	}
	if host := effectstest.AttachmentHostOf(t, g, clamp); host.Kind != "" {
		t.Errorf("Skullclamp still attached to the dead creature: %+v", host)
	}
}

// --- Lightning Greaves / Swiftfoot Boots --------------------------

func TestLightningGreavesGrantsHasteAndShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	greaves := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lightning Greaves", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.GreavesOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.EquipTo(t, g, me.ID, greaves, bear)

	abilities := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"haste", "shroud"} {
		if !effectstest.ContainsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	// Shroud is the half that must be real: NOBODY may target it,
	// its own controller included (CR 702.18a).
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("shroud should stop even the controller targeting it")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("shroud should stop an opponent targeting it")
			}
		}
	})
}

func TestSwiftfootBootsGrantHexproofAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	boots := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Swiftfoot Boots", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.BootsOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.EquipTo(t, g, me.ID, boots, bear)

	abilities := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"hexproof", "haste"} {
		if !effectstest.ContainsString(abilities, want) {
			t.Errorf("abilities %v missing %q", abilities, want)
		}
	}
	// Hexproof is the asymmetric one — that asymmetry is the whole
	// reason Boots costs a mana more to equip than Greaves.
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != bear {
				continue
			}
			if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("hexproof must not stop the controller targeting it")
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("hexproof should stop an opponent targeting it")
			}
		}
	})
}

// --- Rancor: the first Aura ---------------------------------------

func TestRancorAttachesOnResolutionAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rancor := effectstest.CastCatalogSpell(t, g, "Rancor", effectstest.AuraTypeLine, effectstest.RancorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor should resolve to the battlefield")
	}
	if host := effectstest.AttachmentHostOf(t, g, rancor); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("Rancor AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("enchanted power %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Rancor is +2/+0, toughness %d, want 2", got)
	}
	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "trample") {
		t.Errorf("abilities %v missing trample", ab)
	}
}

// CR 704.5m plus Rancor's own recursion: the host leaves, the Aura
// goes to the graveyard as a state-based action, and the LTB trigger
// hands it straight back.
func TestRancorReturnsToHandWhenItFallsOff(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	rancor := effectstest.CastCatalogSpell(t, g, "Rancor", effectstest.AuraTypeLine, effectstest.RancorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	// MoveCardByID rather than DestroyPermanentForEffect: the public
	// mutator runs the state checks on the way out, and the CR 704.5m
	// unattach IS a state check. Nothing else is on the stack to
	// carry a priority pass here.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		bear,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor should have fallen off the battlefield")
	}
	if !me.Hand.Contains(rancor) {
		t.Error("Rancor's own recursion should have put it back in hand")
	}
}

func TestSwordOfFeastAndFamineDiscardsAndUntapsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword of Feast and Famine", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.SwordFeastOracle, Owner: me.ID, Controller: me.ID,
	})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Fatalf("equipped power %d, want 4", got)
	}
	oppHand := opp.Hand.Size()

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)

	// #651: the damaged player CHOOSES their discard (CR 701.9a), so
	// the trigger leaves a prompt addressed to them. The untap is not
	// behind it — the lands come back as the trigger resolves.
	if got := effectstest.DiscardOwed(g, opp.ID); got != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != oppHand-1 {
		t.Errorf("opponent hand %d -> %d, want -1", oppHand, got)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == land && c.Tapped {
				t.Error("the land should have untapped")
			}
		}
	})
}

func TestSwordOfFireAndIceDamagesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword of Fire and Ice", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.SwordFireOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	myHand, oppLife := me.Hand.Size(), opp.Life

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	// The trigger targets, so it waits for a pick before it reaches
	// the stack. Point it at the player who just took the hit.
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppLife-2 {
		t.Errorf("life %d -> %d, want -2", oppLife, opp.Life)
	}
	if got := me.Hand.Size(); got != myHand+1 {
		t.Errorf("hand %d -> %d, want +1", myHand, got)
	}
}

// --- Curse of Opulence: the TargetPlayer branch -------------------

func TestCurseOfOpulenceAttachesToAPlayerAndPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	curse := effectstest.CastCatalogSpell(t, g, "Curse of Opulence", effectstest.CurseAuraTypeLine, effectstest.CurseOpulenceOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	host := effectstest.AttachmentHostOf(t, g, curse)
	if host.Kind != game.TargetPlayer || host.ID != victim.ID {
		t.Fatalf("Curse AttachedTo = %+v, want player %s", host, victim.ID)
	}

	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(attacker, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	golds := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gold" && c.Controller == me.ID {
				golds++
			}
		}
	})
	if golds != 1 {
		t.Errorf("Gold tokens for the Curse's controller: %d, want 1", golds)
	}
}

// TestCurseOfOpulenceEachAttackingOpponentAlsoGetsAGold pins the
// closed caveat: "each opponent attacking that player does the same"
// (#1565). The active seat is neither the Curse's controller nor the
// enchanted player, so their attack should mint the Curse's controller
// their usual Gold PLUS a second Gold for the attacking opponent
// themselves — and the enchanted victim gets none.
func TestCurseOfOpulenceEachAttackingOpponentAlsoGetsAGold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	curseCtl := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Curse of Opulence", TypeLine: effectstest.CurseAuraTypeLine,
		OracleID: effectstest.CurseOpulenceOracle, Owner: curseCtl.ID, Controller: curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID},
	})
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(attacker, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	goldsFor := func(id uuid.UUID) int {
		n := 0
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Gold" && c.Controller == id {
					n++
				}
			}
		})
		return n
	}
	if got := goldsFor(curseCtl.ID); got != 1 {
		t.Errorf("Gold tokens for the Curse's controller: %d, want 1", got)
	}
	if got := goldsFor(active.ID); got != 1 {
		t.Errorf("Gold tokens for the attacking opponent: %d, want 1", got)
	}
	if got := goldsFor(victim.ID); got != 0 {
		t.Errorf("Gold tokens for the enchanted victim: %d, want 0", got)
	}
}

// TestCurseOfOpulenceAttackingOpponentGetsOnlyOneGoldPerCombat proves
// the OncePerBatch dedup on the attacking-opponent ability: a
// two-creature swing from the same non-controller opponent still
// mints exactly one Gold for them, matching the "is attacked" batching
// the Curse's own bullet already had.
func TestCurseOfOpulenceAttackingOpponentGetsOnlyOneGoldPerCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	curseCtl := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Curse of Opulence", TypeLine: effectstest.CurseAuraTypeLine,
		OracleID: effectstest.CurseOpulenceOracle, Owner: curseCtl.ID, Controller: curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID},
	})
	a1 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear One", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	a2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear Two", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: a1, Target: victim.ID},
		{Attacker: a2, Target: victim.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gold" && c.Controller == active.ID {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("Gold tokens for the two-creature attacker: %d, want 1", n)
	}
}
