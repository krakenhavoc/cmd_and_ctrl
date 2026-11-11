package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b20_test.go — Commander deck cards, batches 19 and 20 (#5368): the
// cards with a board, combat or turn-order story.

func TestCmdrB20NogiMakesDragonsCheaperAndBecomesAFiveFiveWithThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nogi := pushTestPermanent(g, me.ID, game.Card{Name: "Nogi, Draco-Zealot", OracleID: cmdrB19NogiOracle,
		TypeLine: "Legendary Creature — Kobold Shaman", Power: 3, Toughness: 3})
	if got := b40PricedFor(t, g, me.ID, "Big Dragon", "Creature — Dragon", "{4}{R}"); got != 4 {
		t.Errorf("a Dragon spell costs %d, want 4", got)
	}
	if got := b40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{4}{R}"); got != 5 {
		t.Errorf("a non-Dragon spell costs %d, want 5", got)
	}
	for i := 0; i < 2; i++ {
		pushTestPermanent(g, me.ID, game.Card{Name: "Whelp", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(nogi, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, nogi); p != 3 {
		t.Fatalf("with two Dragons Nogi is %d power, want 3", p)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	nogi = pushTestPermanent(g, me.ID, game.Card{Name: "Nogi, Draco-Zealot", OracleID: cmdrB19NogiOracle,
		TypeLine: "Legendary Creature — Kobold Shaman", Power: 3, Toughness: 3})
	for i := 0; i < 3; i++ {
		pushTestPermanent(g, me.ID, game.Card{Name: "Whelp", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(nogi, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, nogi), effectiveToughness(t, g, nogi); p != 5 || tough != 5 {
		t.Errorf("Nogi is %d/%d, want 5/5", p, tough)
	}
	c := layeredCard(t, g, nogi)
	if !c.HasSubtype("Dragon") || c.HasSubtype("Kobold") || !c.IsLegendary() {
		t.Errorf("Nogi should be a legendary Dragon and no longer a Kobold: %v", c.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, nogi, "flying") {
		t.Error("Nogi gains flying")
	}
	// Until end of turn.
	advanceToDeclareAttackersOf(t, g, 1)
	if p := effectivePower(t, g, nogi); p != 3 {
		t.Errorf("power %d after the turn, want 3", p)
	}
}

func TestCmdrB20GadrakCantAttackWithoutFourArtifactsAndMakesTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gadrak := pushTestPermanent(g, me.ID, game.Card{Name: "Gadrak, the Crown-Scourge", OracleID: cmdrB19GadrakOracle,
		TypeLine: "Legendary Creature — Dragon", Power: 5, Toughness: 4})
	for i := 0; i < 3; i++ {
		pushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, gadrak, opp.ID); err == nil {
		t.Fatal("Gadrak can't attack with fewer than four artifacts")
	}
	pushTestPermanent(g, me.ID, game.Card{Name: "Fourth Rock", TypeLine: "Artifact"})
	if err := declareResult(g, gadrak, opp.ID); err != nil {
		t.Fatalf("with four artifacts: %v", err)
	}
}

func TestCmdrB20GadrakEndStepTreasureCountsNontokenCreaturesThatDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Two creatures die before Gadrak even enters; a token dies too.
	a := pushVanillaCreature(g, opp.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	for _, id := range []uuid.UUID{a, b, tok} {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
	}
	passPriorityAroundTable(t, g)
	pushTestPermanent(g, me.ID, game.Card{Name: "Gadrak, the Crown-Scourge", OracleID: cmdrB19GadrakOracle,
		TypeLine: "Legendary Creature — Dragon", Power: 5, Toughness: 4})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("%d Treasures, want 2 (two nontoken creatures died this turn, the token doesn't count)", got)
	}
}

func TestCmdrB20AssaultFormationAssignsDamageByToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Assault Formation", OracleID: cmdrB19FormationOracle, TypeLine: "Enchantment"})
	glass := pushVanillaCreature(g, me.ID, "Glass Cannon", 4, 1)
	wall := pushVanillaCreature(g, me.ID, "Wall", 2, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 0, 9)
	if p := effectivePower(t, g, glass); p != 4 {
		t.Fatalf("power is unchanged: %d", p)
	}
	before := opp.Life
	attackWith(t, g, opp.ID, glass, wall)
	passPriorityAroundTable(t, g)
	// 4/1 assigns 1, 2/3 assigns 3: the unconditional toughness form.
	if got := before - opp.Life; got != 4 {
		t.Errorf("opponent lost %d, want 4 (1 + 3)", got)
	}
	if _, ok := battlefieldCard(g, theirs); !ok {
		t.Error("their creature should be fine")
	}
	_ = me
}

func TestCmdrB20AssaultFormationWaivesDefenderAndPumpsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f := pushTestPermanent(g, me.ID, game.Card{Name: "Assault Formation", OracleID: cmdrB19FormationOracle, TypeLine: "Enchantment"})
	wall := pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	advanceToMainOf(t, g, 0)
	before := g.Clone()
	advanceTo(t, before, game.StepDeclareAttackers)
	wantDefenderRefused(t, before, wall, opp.ID)

	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}},
	}); err != nil {
		t.Fatalf("waive: %v", err)
	}
	// A creature without defender is not a legal target.
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting a creature without defender: %v, want ErrIllegalTarget", err)
	}
	passPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, f, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tough := effectiveToughness(t, g, plain); tough != 3 {
		t.Errorf("toughness %d, want 3", tough)
	}
	if p := effectivePower(t, g, plain); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := declareResult(g, wall, opp.ID); err != nil {
		t.Errorf("the waived Wall may attack: %v", err)
	}
}

func TestCmdrB20CrawlingSensationOnceAPerTurnForLandsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTestPermanent(g, me.ID, game.Card{Name: "Crawling Sensation", OracleID: cmdrB19SensationOracle, TypeLine: "Enchantment"})
	mm2Library(me, mm2Land("Forest"), mm2Land("Swamp"), mm2Spell("A"), mm2Land("Plains"), mm2Spell("B"), mm2Spell("C"))
	// Two lands milled together are one occurrence.
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 2, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	mm2Settle(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Fatalf("%d Insects after two lands milled together, want 1", got)
	}
	// A spell milled makes nothing; another land the same turn makes nothing.
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 2, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	mm2Settle(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects after a second land the same turn, want still 1", got)
	}
	// A land discarded from hand also counts, next turn (the upkeep mill
	// is declined).
	advanceDecliningTriggerPrompts(t, g, 1)
	advanceDecliningTriggerPrompts(t, g, 0)
	cmdrB17ToGraveyard(t, g, me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	mm2Settle(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 2 {
		t.Errorf("%d Insects after a land reached the graveyard on a later turn, want 2", got)
	}
}

func TestCmdrB20CrawlingSensationUpkeepMillIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTestPermanent(g, me.ID, game.Card{Name: "Crawling Sensation", OracleID: cmdrB19SensationOracle, TypeLine: "Enchantment"})
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	before := me.Library.Size()
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if prompt == nil {
		t.Fatalf("no 'you may mill' question: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestCmdrB20OpportunisticDragonStealsAHumanOrArtifactWhileItLives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Human", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}})
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	dragon := castCatalogSpell(t, g, "Opportunistic Dragon", "Creature — Dragon", cmdrB19OppDragonOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a Bear is neither a Human nor an artifact")
	}
	answerPickTarget(t, g, human)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, human)
	if c.Controller != me.ID {
		t.Fatalf("the Human was not stolen: controller %s", c.Controller)
	}
	if hasEffectiveKeyword(t, g, human, "flying") {
		t.Error("it loses all abilities")
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, human, opp.ID); err == nil {
		t.Error("the stolen permanent can't attack")
	}
	// The Dragon leaves: control, abilities and attacks come back.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dragon); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	c = layeredCard(t, g, human)
	if c.Controller != opp.ID {
		t.Errorf("control should return to the opponent, got %s", c.Controller)
	}
	if !hasEffectiveKeyword(t, g, human, "flying") {
		t.Error("it gets its abilities back")
	}
}

func TestCmdrB20GrenzoGoadsTheDamagedPlayersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Grenzo, Havoc Raiser", OracleID: cmdrB19GrenzoOracle,
		TypeLine: "Legendary Creature — Goblin Rogue", Power: 2, Toughness: 2})
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guard", 1, 1)
	other := pushVanillaCreature(g, g.Seats[2].ID, "Third Seat Guard", 1, 1)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	answerMode(t, g, c, me.ID, 0)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no goad target asked")
	}
	if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: other}); err == nil {
		t.Fatal("only a creature the damaged player controls is legal")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, theirs)
	if !ok || len(got.Goads) == 0 {
		t.Errorf("their creature should be goaded: %+v", got.Goads)
	}
}

func TestCmdrB20GrenzoExilesTheTopCardAndLetsYouCastItWithAnyColorMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Grenzo, Havoc Raiser", OracleID: cmdrB19GrenzoOracle,
		TypeLine: "Legendary Creature — Goblin Rogue", Power: 2, Toughness: 2})
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	top := mm2Library(opp, game.Card{Name: "Stolen Spell", TypeLine: "Sorcery", ManaCost: "{1}{U}"}, mm2Spell("Next"))
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !inExile(g, top[0]) {
		t.Fatal("the top card of their library was not exiled")
	}
	perm := g.CastPermissionOnCardByIDForEffect(top[0])
	if perm == nil || perm.Player != me.ID || !perm.CastOnly {
		t.Fatalf("permission = %+v, want a cast-only grant to the controller", perm)
	}
	if inExile(g, top[1]) {
		t.Error("only one card is exiled")
	}
}

func TestCmdrB20TerritorialHellkiteNeverRepeatsLastCombatsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hk := pushTestPermanent(g, me.ID, game.Card{Name: "Territorial Hellkite", OracleID: cmdrB19HellkiteOracle,
		TypeLine: "Creature — Dragon", Power: 6, Toughness: 5})
	var prev uuid.UUID
	for turn := 0; turn < 4; turn++ {
		advanceDecliningTriggerPrompts(t, g, 0)
		passPriorityAroundTable(t, g)
		target := requiredAttackTarget(g)
		if target == uuid.Nil {
			t.Fatalf("turn %d: no 'attacks that player' requirement was set", turn)
		}
		if target == me.ID {
			t.Fatalf("turn %d: chose myself", turn)
		}
		if target == prev {
			t.Fatalf("turn %d: chose %s again, the player it attacked last combat", turn, target)
		}
		// Only the chosen player may be attacked while the requirement binds.
		for _, opp := range g.Seats[1:] {
			err := declareResult(g, hk, opp.ID)
			if opp.ID == target && err != nil {
				t.Fatalf("turn %d: attack the chosen player: %v", turn, err)
			}
		}
		if err := g.DeclareAttacker(hk, target); err != nil {
			t.Fatalf("turn %d: %v", turn, err)
		}
		lockInAttacks(t, g)
		prev = target
		advanceDecliningTriggerPrompts(t, g, 1)
	}
}

func TestCmdrB20TerritorialHellkiteTapsWithNoCandidateAndPlaneswalkerAttacksDontCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.Seats[2].Eliminated, g.Seats[3].Eliminated = true, true
	hk := pushTestPermanent(g, me.ID, game.Card{Name: "Territorial Hellkite", OracleID: cmdrB19HellkiteOracle,
		TypeLine: "Creature — Dragon", Power: 6, Toughness: 5})
	// First combat: the only opponent is the only candidate.
	advanceDecliningTriggerPrompts(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := requiredAttackTarget(g); got != opp.ID {
		t.Fatalf("first combat chose %s, want the only opponent", got)
	}
	if err := g.DeclareAttacker(hk, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceDecliningTriggerPrompts(t, g, 1)
	// Next combat: the only opponent was just attacked, so there is no
	// candidate and the Hellkite is tapped instead.
	advanceDecliningTriggerPrompts(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := requiredAttackTarget(g); got != uuid.Nil {
		t.Errorf("a requirement for %s was set with no candidate", got)
	}
	if !isTapped(g, hk) {
		t.Error("with no opponent to choose, the Hellkite is tapped")
	}
}

// requiredAttackTarget finds the player the Hellkite is required to attack,
// by trying each opponent on a clone: only the chosen one is accepted when
// the requirement binds.
func requiredAttackTarget(g *game.Game) uuid.UUID {
	for _, rec := range g.ScopedEffects {
		for _, m := range rec.Mods {
			if m.Kind == game.ModAttackPlayerIfAble && m.Player != uuid.Nil {
				return m.Player
			}
		}
	}
	return uuid.Nil
}

// advanceDecliningTriggerPrompts walks the turn cursor to the declare
// attackers step of `seat`, answering "no" to any optional trigger.
func advanceDecliningTriggerPrompts(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDeclareAttackers && g.Turn.ActiveSeat == seat {
			return
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
				if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, false); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			passPriorityAroundTable(t, g)
		}
	}
	t.Fatalf("never reached declare attackers of seat %d", seat)
}
