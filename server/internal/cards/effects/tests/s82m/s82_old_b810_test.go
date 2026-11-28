package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b810_test.go — Mirage, batch 810 (#5633): other creatures.

const (
	oldB810Scouts    = "18710b5e-5403-48cd-9541-8154ba0a0b30"
	oldB810Hyenas    = "2a073f00-0d03-43d0-a0a8-f5bcee0855fc"
	oldB810Cobra     = "456bbc0e-a648-481f-b37e-9996029556bd"
	oldB810Cyclops   = "6c89e4e6-c4d7-416c-b985-41739c6f5b50"
	oldB810Basilisk  = "6e7bb687-8243-4377-80b5-67ce6b65f9b7"
	oldB810Guardian  = "87d734bc-4e4d-4013-91ed-c6bf1dde824e"
	oldB810Coral     = "8a397a64-542f-43fb-bf54-27ac651e1819"
	oldB810Zombie    = "d04a836b-7e42-424d-9600-42a40d0dfccd"
	oldB810Specter   = "da21c21c-88d8-4a93-b2b4-6d7bb28959c9"
	oldB810Fighter   = "ea8d1c55-e045-4a3f-92f0-4ecd86362b97"
	oldB810Viashino  = "f46e06e2-1658-4bec-b94e-91935cc667fb"
	oldB810BearType  = "Creature — Bear"
	oldB810WallType  = "Creature — Wall"
	oldB810Seat0Name = "seat 0"
)

func TestOldB810EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB810Scouts: "Femeref Scouts", oldB810Hyenas: "Gibbering Hyenas", oldB810Cobra: "Crypt Cobra",
		oldB810Cyclops: "Ekundu Cyclops", oldB810Basilisk: "Rock Basilisk", oldB810Guardian: "Spectral Guardian",
		oldB810Coral: "Coral Fighters", oldB810Zombie: "Gravebane Zombie", oldB810Specter: "Dread Specter",
		oldB810Fighter: "Dream Fighter", oldB810Viashino: "Viashino Warrior",
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

func oldB810Colored(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestOldB810GibberingHyenasCantBlockBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := oldB810Colored(g, me.ID, "Black Bear", oldB810BearType, "", 2, 2, "B")
	green := oldB810Colored(g, me.ID, "Green Bear", oldB810BearType, "", 2, 2, "G")
	hyenas := oldB810Colored(g, opp.ID, "Gibbering Hyenas", "Creature — Hyena", oldB810Hyenas, 3, 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, black, green)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hyenas, black); err == nil {
		t.Error("the Hyenas blocked a black creature")
	}
	if err := g.DeclareBlocker(hyenas, green); err != nil {
		t.Errorf("the Hyenas could not block a green creature: %v", err)
	}
}

func TestOldB810CryptCobraPoisonsTheDefenderOnlyWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cobra := oldB810Colored(g, me.ID, "Crypt Cobra", "Creature — Snake", oldB810Cobra, 3, 3, "B")
	effectstest.Pr7bAttack(t, g, opp.ID, cobra)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoisonOn(opp); got != 1 {
		t.Errorf("poison counters %d, want 1 for an unblocked Cobra", got)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	cobra = oldB810Colored(g, me.ID, "Crypt Cobra", "Creature — Snake", oldB810Cobra, 3, 3, "B")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", oldB810WallType, 0, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, cobra)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: cobra})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoisonOn(opp); got != 0 {
		t.Errorf("poison counters %d, want 0 for a blocked Cobra", got)
	}
}

func TestOldB810EkunduCyclopsMustAttackOnceAnotherCreatureDoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyclops := effectstest.B12Push(g, me.ID, "Ekundu Cyclops", "Creature — Cyclops", oldB810Cyclops, 3, 4)
	other := effectstest.B12Creature(g, me.ID, "Bear", oldB810BearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	// Nothing declared: nothing is owed, so the pass is free of the requirement.
	var owed map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if len(owed) != 0 {
		t.Errorf("owed with an empty declaration = %v, want nothing", owed)
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	var e *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &e) {
		t.Fatalf("pass with the Cyclops home = %v, want *AttackRequirementError", err)
	}
	if err := g.DeclareAttacker(cyclops, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker Cyclops: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both attacking: %v", err)
	}
}

// oldB810BlockAndEndCombat has the opponent's creature block the attacker, plays
// out combat damage and the end of combat step.
func oldB810BlockAndEndCombat(t *testing.T, g *game.Game, opp *game.Player, attacker, blocker uuid.UUID) {
	t.Helper()
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blocker) {
		t.Fatal("the other creature was destroyed before the end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB810RockBasiliskDestroysANonWallItBlocksOrIsBlockedByAtEndOfCombat(t *testing.T) {
	// Becomes blocked by a Bear: the Bear is destroyed at end of combat.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bas := effectstest.B12Push(g, me.ID, "Rock Basilisk", "Creature — Basilisk", oldB810Basilisk, 4, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB810BearType, 2, 9)
	oldB810BlockAndEndCombat(t, g, opp, bas, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the non-Wall blocker survived the end of combat")
	}

	// Blocked by a Wall: nothing.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bas = effectstest.B12Push(g, me.ID, "Rock Basilisk", "Creature — Basilisk", oldB810Basilisk, 4, 5)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", oldB810WallType, 0, 9)
	oldB810BlockAndEndCombat(t, g, opp, bas, wall)
	if !g.Battlefield.Contains(wall) {
		t.Error("a Wall was destroyed")
	}

	// Blocks: the Basilisk (as blocker) destroys the non-Wall attacker.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB810BearType, 2, 9)
	bas = effectstest.B12Push(g, opp.ID, "Rock Basilisk", "Creature — Basilisk", oldB810Basilisk, 4, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{bas: attacker})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the creature the Basilisk blocked survived the end of combat")
	}
}

func TestOldB810DreadSpecterDestroysOnlyNonblackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec := oldB810Colored(g, me.ID, "Dread Specter", "Creature — Specter", oldB810Specter, 2, 2, "B")
	green := oldB810Colored(g, opp.ID, "Green Bear", oldB810BearType, "", 2, 9, "G")
	oldB810BlockAndEndCombat(t, g, opp, spec, green)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack blocker survived the end of combat")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	spec = oldB810Colored(g, me.ID, "Dread Specter", "Creature — Specter", oldB810Specter, 2, 2, "B")
	black := oldB810Colored(g, opp.ID, "Black Bear", oldB810BearType, "", 2, 9, "B")
	oldB810BlockAndEndCombat(t, g, opp, spec, black)
	if !g.Battlefield.Contains(black) {
		t.Error("a black creature was destroyed")
	}
}

func TestOldB810SpectralGuardianGivesNoncreatureArtifactsShroudWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Spectral Guardian", "Creature — Spirit", oldB810Guardian, 2, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"a noncreature artifact": {rock, true}, "an artifact creature": {golem, false}, "the Guardian": {guard, false}} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, game.KeywordShroud); got != tc.want {
			t.Errorf("%s has shroud = %v, want %v", name, got, tc.want)
		}
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(guard) })
	if effectstest.HasEffectiveKeyword(t, g, rock, game.KeywordShroud) {
		t.Error("a noncreature artifact kept shroud while the Guardian was tapped")
	}
}

func TestOldB810CoralFightersLetsYouBottomTheDefendersTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fighters := effectstest.B12Push(g, me.ID, "Coral Fighters", "Creature — Merfolk Soldier", oldB810Coral, 1, 1)
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.Pr7bAttack(t, g, opp.ID, fighters)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no top-or-bottom choice was offered to the attacker's controller")
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("bottom: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The last card of a zone is its top; the first is its bottom.
	if n := len(opp.Library.Cards); n == 0 || opp.Library.Cards[n-1].InstanceID == top {
		t.Error("the looked-at card is still on top of the defender's library")
	}
	if got := opp.Library.Cards[0].InstanceID; got != top {
		t.Error("the looked-at card is not on the bottom of the defender's library")
	}
}

func TestOldB810GravebaneZombieGoesToTheTopOfItsOwnersLibraryInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := g.Seats[0], g.Seats[1]
	zombie := effectstest.B12Push(g, me.ID, "Gravebane Zombie", "Creature — Zombie", oldB810Zombie, 3, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), zombie, 5) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zombie) {
		t.Fatal("the Zombie stayed on the battlefield")
	}
	if me.Graveyard.Contains(zombie) {
		t.Error("the Zombie went to the graveyard")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[n-1].InstanceID != zombie {
		t.Error("the Zombie is not on top of its owner's library")
	}
}

func TestOldB810DreamFighterPhasesOutWithTheCreatureItFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fighter := effectstest.B12Push(g, me.ID, "Dream Fighter", "Creature — Human Soldier", oldB810Fighter, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB810BearType, 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB810BearType, 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, fighter)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{bear: fighter})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPhasedOutForEffect(fighter) || !g.IsPhasedOutForEffect(bear) {
		t.Fatal("the Fighter and the creature it fought did not both phase out")
	}
	if g.IsPhasedOutForEffect(bystander) {
		t.Error("an uninvolved creature phased out")
	}
}

func TestOldB810VanillaCreatures(t *testing.T) {
	for _, tc := range []struct{ id, name string }{{oldB810Scouts, "Femeref Scouts"}, {oldB810Viashino, "Viashino Warrior"}} {
		spec, ok := effects.Lookup(tc.id)
		if !ok {
			t.Fatalf("%s not registered", tc.name)
		}
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Replacements) != 0 {
			t.Errorf("%s has abilities; it is a vanilla creature", tc.name)
		}
	}
}
