package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b145_test.go — GAP batch 145 (#5948): Teferi's Puzzle Box, Weapon
// Rack, Elven Bow, Rosethorn Halberd, Icebind Pillar, Mantle of Tides, Shining
// Armor, Assembly-Worker, Clockwork Hydra, Cogwork Assembler and Dutiful
// Replicator. Each test plays the real catalog entry and asserts what its text
// says.

const (
	gap145PuzzleBox   = "37abcc92-9466-47ea-9e0b-5eda2eb62c8e"
	gap145WeaponRack  = "10d366cb-b92f-401c-a88e-60b267bd8ae5"
	gap145ElvenBow    = "43daed96-2216-4f17-82ce-3eccd5d69936"
	gap145Halberd     = "94848b02-757c-456c-a12c-b0be193221d0"
	gap145Pillar      = "18050051-63f6-4091-bf67-7a0701a0791d"
	gap145Mantle      = "56a80bd3-b81b-428d-96e1-f03ab6199cf3"
	gap145Armor       = "090a5f76-206e-4b3b-bc1d-9c9706d9955d"
	gap145Worker      = "751321ac-af07-4b7e-976f-5bf4e4281470"
	gap145Hydra       = "6fd37ff3-e75d-459a-8717-8f6a9865fb32"
	gap145Assembler   = "01cebebb-050c-4b69-8a13-c569183a3cc3"
	gap145Replicator  = "46670c03-d7eb-489f-ba1a-c7f805c486d2"
	gap145EquipLine   = "Artifact — Equipment"
	gap145WorkerLine  = "Artifact Creature — Assembly-Worker"
	gap145TokenGoblin = "Token Creature — Goblin"
)

func TestGap145EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap145PuzzleBox: "Teferi's Puzzle Box", gap145WeaponRack: "Weapon Rack", gap145ElvenBow: "Elven Bow",
		gap145Halberd: "Rosethorn Halberd", gap145Pillar: "Icebind Pillar", gap145Mantle: "Mantle of Tides",
		gap145Armor: "Shining Armor", gap145Worker: "Assembly-Worker", gap145Hydra: "Clockwork Hydra",
		gap145Assembler: "Cogwork Assembler", gap145Replicator: "Dutiful Replicator",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if spec, _ := effects.Lookup(gap145Armor); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Shining Armor printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Teferi's Puzzle Box -------------------------------------------------------

func TestGap145PuzzleBoxCyclesTheHandOfEachPlayerInTheirDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Teferi's Puzzle Box", "Artifact", gap145PuzzleBox, false)
	gap145FillHand(opp, 2)
	var old []uuid.UUID
	for _, c := range opp.Hand.Cards {
		old = append(old, c.InstanceID)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	for i := 0; i < 8 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatal("the player whose draw step it is was not asked to order the cards")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the Box's controller was asked; the player whose step it is chooses")
	}
	// The turn-based draw (CR 504.1) happens before the step's triggers go on
	// the stack, so the drawn card is in the hand being put away too.
	if len(c.ScryCards) != len(old)+1 {
		t.Fatalf("asked to order %d cards, want the two seeded plus the drawn one", len(c.ScryCards))
	}
	old = append([]uuid.UUID(nil), c.ScryCards...)
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, old, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range old {
		if opp.Hand.Contains(id) {
			t.Errorf("card %s is still in hand", id)
		}
		if !opp.Library.Contains(id) {
			t.Errorf("card %s is not in the library", id)
		}
	}
	if got := opp.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (as many drawn as were put away)", got)
	}
}

func TestGap145PuzzleBoxWithOnlyTheDrawnCardPutsItBackAndDrawsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Teferi's Puzzle Box", "Artifact", gap145PuzzleBox, false)
	gap145FillHand(opp, 0)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	drawn := effects.AllHandCardIDs(g, opp.ID)
	if len(drawn) != 1 {
		t.Fatalf("hand %d at the start of the draw step, want the turn's draw", len(drawn))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID); c != nil {
		if err := g.ResolvePutInLibrary(c.ID, opp.ID, drawn, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Hand.Contains(drawn[0]) || !opp.Library.Contains(drawn[0]) {
		t.Error("the drawn card was not put on the bottom of the library")
	}
	if got := opp.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1 (one put away, one drawn)", got)
	}
}

// --- Weapon Rack ---------------------------------------------------------------

func TestGap145WeaponRackEntersWithThreeAndMovesOneAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rack := effectstest.Msc17Play(t, g, me, game.Card{Name: "Weapon Rack", TypeLine: "Artifact", OracleID: gap145WeaponRack})
	if n := effectstest.B12Counter(t, g, rack, game.CounterPlusOne); n != 3 {
		t.Fatalf("enters with %d counters, want 3", n)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, rack, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("move onto any creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Counter(t, g, rack, game.CounterPlusOne) != 2 || effectstest.B12Counter(t, g, theirs, game.CounterPlusOne) != 1 {
		t.Fatal("exactly one counter should have moved")
	}
	if !effectstest.IsTapped(g, rack) {
		t.Error("the Rack did not tap")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, rack).Tapped = false })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, rack, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Elven Bow -----------------------------------------------------------------

func gap145CastBow(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Elven Bow", TypeLine: gap145EquipLine, OracleID: gap145ElvenBow}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityUntilChoice(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no {2} offer")
	}
	return id
}

func TestGap145ElvenBowPaidMakesAnEquippedElfWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := gap145CastBow(t, g, me)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	var elf uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elf Warrior" && c.Controller == me.ID && c.IsToken() {
			elf = c.InstanceID
		}
	}
	if elf == uuid.Nil {
		t.Fatal("no Elf Warrior token")
	}
	if host := effectstest.AttachmentHostOf(t, g, bow); host.ID != elf {
		t.Fatalf("the Bow is attached to %v, want the Elf Warrior", host.ID)
	}
	if p, tough := effectstest.PtOf(t, g, elf); p != 2 || tough != 3 {
		t.Errorf("the equipped Elf is %d/%d, want 2/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, elf, "reach") {
		t.Error("the equipped Elf has no reach")
	}
}

func TestGap145ElvenBowDeclinedOrUnpaidMakesNothing(t *testing.T) {
	for _, mana := range []string{"", "{C}"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bow := gap145CastBow(t, g, me)
		if mana != "" {
			effectstest.FloatMana(t, g, me, mana)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, mana != "")
		effectstest.PassPriorityAroundTable(t, g)
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Elf Warrior"); n != 0 {
			t.Errorf("pool %q: %d Elf Warriors, want 0", mana, n)
		}
		if host := effectstest.AttachmentHostOf(t, g, bow); host.Kind == game.TargetCard {
			t.Errorf("pool %q: the Bow is attached to %v", mana, host.ID)
		}
	}
}

// --- Rosethorn Halberd ---------------------------------------------------------

func TestGap145HalberdAttachesToANonHumanAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Human Peasant", "Creature — Human Peasant", 1, 1)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 2, 2)
	halberd := effectstest.CastCatalogSpell(t, g, "Rosethorn Halberd", gap145EquipLine, gap145Halberd, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, halberd); host.ID != elf {
		t.Fatalf("attached to %v, want the Elf", host.ID)
	}
	if p, tough := effectstest.PtOf(t, g, elf); p != 4 || tough != 3 {
		t.Errorf("the equipped Elf is %d/%d, want 4/3", p, tough)
	}
}

func TestGap145HalberdCannotTargetAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Human Peasant", "Creature — Human Peasant", 1, 1)
	halberd := effectstest.CastCatalogSpell(t, g, "Rosethorn Halberd", gap145EquipLine, gap145Halberd, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a Human was offered as a target")
	}
	if host := effectstest.AttachmentHostOf(t, g, halberd); host.Kind == game.TargetCard {
		t.Errorf("attached to %v", host.ID)
	}
	if p, _ := effectstest.PtOf(t, g, human); p != 1 {
		t.Errorf("the Human is power %d, want 1", p)
	}
}

// --- Shining Armor -------------------------------------------------------------

func TestGap145ShiningArmorAttachesToAKnightOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	armor := effectstest.CastCatalogSpell(t, g, "Shining Armor", gap145EquipLine, gap145Armor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, knight)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, armor); host.ID != knight {
		t.Fatalf("attached to %v, want the Knight", host.ID)
	}
	if p, tough := effectstest.PtOf(t, g, knight); p != 2 || tough != 4 {
		t.Errorf("the equipped Knight is %d/%d, want 2/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, knight, "vigilance") {
		t.Error("the equipped Knight has no vigilance")
	}
}

func TestGap145ShiningArmorWithNoKnightStaysUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	armor := effectstest.CastCatalogSpell(t, g, "Shining Armor", gap145EquipLine, gap145Armor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a non-Knight was offered")
	}
	if host := effectstest.AttachmentHostOf(t, g, armor); host.Kind == game.TargetCard {
		t.Errorf("attached to %v", host.ID)
	}
}

// --- Mantle of Tides -----------------------------------------------------------

func TestGap145MantleAttachesOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mantle := effectstest.PushCatalogPermanent(g, me.ID, "Mantle of Tides", gap145EquipLine, gap145Mantle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if base := gap145TurnDraws(g, me); base > 1 {
		t.Fatalf("already %d draws this turn", base)
	}
	if gap145TurnDraws(g, me) < 1 {
		gap145Draw(t, g, me, 1)
	}
	if host := effectstest.AttachmentHostOf(t, g, mantle); host.Kind == game.TargetCard {
		t.Fatal("attached on the first draw")
	}
	gap145Draw(t, g, me, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, mantle); host.ID != bear {
		t.Fatalf("attached to %v, want the Bear", host.ID)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 3/4", p, tough)
	}
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	gap145Draw(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the third draw triggered it")
	}
	if host := effectstest.AttachmentHostOf(t, g, mantle); host.ID != bear {
		t.Errorf("moved to %v on the third draw (other %v)", host.ID, other)
	}
}

// --- Icebind Pillar ------------------------------------------------------------

func TestGap145IcebindPillarTapsAnArtifactOrCreatureForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pillar := effectstest.PushCatalogPermanent(g, me.ID, "Icebind Pillar", "Snow Artifact", gap145Pillar, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	gap145PlainLand(g, me.ID, "Island")
	activate := func(target uuid.UUID) error {
		err := g.ActivateCatalogAbility(me.ID, pillar, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, Targets: effectstest.CardRefs(target)})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		return err
	}
	if err := activate(bear); err == nil {
		t.Fatal("paid {S} with a nonsnow land")
	}
	gap145SnowLand(g, me.ID, "Island")
	if err := activate(land); err == nil {
		t.Fatal("a land was a legal target")
	}
	if err := activate(bear); err != nil {
		t.Fatalf("{S}, {T}: %v", err)
	}
	if !effectstest.IsTapped(g, bear) || !effectstest.IsTapped(g, pillar) {
		t.Errorf("bear tapped %v, pillar tapped %v", effectstest.IsTapped(g, bear), effectstest.IsTapped(g, pillar))
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, pillar).Tapped = false })
	gap145SnowLand(g, me.ID, "Island")
	if err := activate(rock); err != nil {
		t.Fatalf("tap an artifact: %v", err)
	}
	if !effectstest.IsTapped(g, rock) {
		t.Error("the artifact did not tap")
	}
}

// --- Assembly-Worker -----------------------------------------------------------

func TestGap145AssemblyWorkerPumpsAnAssemblyWorkerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	worker := effectstest.PushCatalogPermanent(g, me.ID, "Assembly-Worker", gap145WorkerLine, gap145Worker, false)
	pal := effectstest.B12Creature(g, me.ID, "Other Worker", gap145WorkerLine, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.S82b11Activate(t, g, me, worker, 0, bear); err == nil {
		t.Fatal("a Bear was a legal target")
	}
	if err := effectstest.S82b11Activate(t, g, me, worker, 0, pal); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, pal); p != 3 || tough != 3 {
		t.Errorf("the pumped Worker is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the Bear is power %d, want 2", p)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if p, _ := effectstest.PtOf(t, g, pal); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

// --- Clockwork Hydra -----------------------------------------------------------

func TestGap145ClockworkHydraEntersWithFourAndGrowsByTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.MrdEnter(t, g, me, "Clockwork Hydra", "Artifact Creature — Hydra", gap145Hydra)
	if n := gap145PlusOneCounters(t, g, hydra); n != 4 {
		t.Fatalf("entered with %d counters, want 4", n)
	}
	if p, tough := effectstest.PtOf(t, g, hydra); p != 4 || tough != 4 {
		t.Fatalf("a %d/%d Hydra, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gap145PlusOneCounters(t, g, hydra); n != 5 {
		t.Errorf("%d counters after tapping, want 5", n)
	}
}

func TestGap145ClockworkHydraPingsWhenItAttacksAndWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.MrdEnter(t, g, me, "Clockwork Hydra", "Artifact Creature — Hydra", gap145Hydra)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, hydra)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gap145PlusOneCounters(t, g, hydra); n != 3 {
		t.Errorf("%d counters after the attack trigger, want 3", n)
	}
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if want := life - 1 - 3; opp.Life != want {
		t.Errorf("after combat damage life %d, want %d (the Hydra is a 3/3)", opp.Life, want)
	}
}

func TestGap145ClockworkHydraBlockingRemovesACounterAndPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Squire", "Creature — Human Soldier", 1, 1)
	hydra := effectstest.MrdEnter(t, g, opp, "Clockwork Hydra", "Artifact Creature — Hydra", gap145Hydra)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hydra, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	pick := effectstest.PassUntilPickTarget(t, g, opp.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, opp.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gap145PlusOneCounters(t, g, hydra); n != 3 {
		t.Errorf("%d counters after blocking, want 3", n)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 1/1 target survived 1 damage")
	}
}

// --- Cogwork Assembler ---------------------------------------------------------

func TestGap145CogworkAssemblerCopiesAnArtifactWithHasteAndExilesItAtEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	asm := effectstest.PushCatalogPermanent(g, me.ID, "Cogwork Assembler", gap145WorkerLine, gap145Assembler, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 6)
	if err := effectstest.S82b11Activate(t, g, me, asm, 0, rock); err == nil {
		t.Fatal("activated for six mana")
	}
	effectstest.FillPool(me, 1)
	if err := effectstest.S82b11Activate(t, g, me, asm, 0, bear); err == nil {
		t.Fatal("a nonartifact creature was a legal target")
	}
	if err := effectstest.S82b11Activate(t, g, me, asm, 0, rock); err != nil {
		t.Fatalf("{7}: %v", err)
	}
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Rock" && c.IsToken() {
			token = c.InstanceID
			if c.Controller != me.ID {
				t.Errorf("the copy is controlled by %v, want the activator", c.Controller)
			}
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of the artifact")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") {
		t.Error("the copy has no haste")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the original left the battlefield")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the copy was not exiled at the end step")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the original was exiled")
	}
}

// --- Dutiful Replicator --------------------------------------------------------

func gap145CastReplicator(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Dutiful Replicator", TypeLine: gap145WorkerLine, OracleID: gap145Replicator, Power: 3, Toughness: 2}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityUntilChoice(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no {1} offer")
	}
}

func TestGap145DutifulReplicatorPaidCopiesAnotherToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", gap145TokenGoblin, 1, 1)
	self := effectstest.B12Creature(g, me.ID, "Dutiful Replicator", "Token "+gap145WorkerLine, 3, 2)
	gap145CastReplicator(t, g, me)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: self}); err == nil {
		t.Fatal("a token named Dutiful Replicator was a legal target")
	}
	effectstest.PickCard(t, g, me.ID, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Goblin")); n != 2 {
		t.Errorf("%d Goblins, want the token and its copy", n)
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Dutiful Replicator")); n != 2 {
		t.Errorf("%d Dutiful Replicators, want the card and the earlier token only", n)
	}
}

func TestGap145DutifulReplicatorDeclinedOrWithNoTokenCopiesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Goblin", gap145TokenGoblin, 1, 1)
	gap145CastReplicator(t, g, me)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Goblin")); n != 1 {
		t.Errorf("declined: %d Goblins, want 1", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B12Creature(g2, me2.ID, "Plain Bear", "Creature — Bear", 2, 2)
	gap145CastReplicator(t, g2, me2)
	effectstest.FloatMana(t, g2, me2, "{C}")
	effectstest.AnswerPayUnless(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.LatestPickTarget(g2, me2.ID) != nil {
		t.Error("asked for a target with no token to copy")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g2, "Plain Bear")); n != 1 {
		t.Errorf("%d Plain Bears, want 1 (a nontoken is not copied)", n)
	}
}

// gap145FillHand sets a seat's hand to exactly n filler cards.
func gap145FillHand(p *game.Player, n int) {
	for p.Hand.Size() > 0 {
		p.Hand.Remove(p.Hand.Cards[0].InstanceID)
	}
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// gap145Draw has p draw n cards (the library gets filler), letting any
// trigger the draws set off resolve after each.
func gap145Draw(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		effectstest.PushLibraryCardForTest(p, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("draw %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func gap145TurnDraws(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TurnTallyFor(p.ID).CardsDrawn })
	return n
}

// gap145SnowLand puts a snow basic of subtype under owner's control.
func gap145SnowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

// gap145PlainLand puts a nonsnow basic of subtype under owner's control.
func gap145PlainLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

func gap145PlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]
}
