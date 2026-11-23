package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b194_test.go — S82 GAP batch 194 (#5948): Warhammer 40,000
// Commander creatures. Each test plays the real catalog entry.

const (
	gapB194ApostleOracle   = "8a446307-d9ff-4854-90e6-9cdff5b77aa6"
	gapB194ExocrineOracle  = "cdf34da6-d952-43ca-855d-fcd9d49efae6"
	gapB194HeraldOracle    = "ea912e4f-c734-4c7b-9e80-8961ca57cdd3"
	gapB194KeeperOracle    = "23f86ee2-b1e3-45c2-bfd5-044919033922"
	gapB194MarineOracle    = "1f7e85cf-afee-4041-9436-1fb69b4ec292"
	gapB194ScreamerOracle  = "fb839b8c-a2b3-42fb-a7a2-8552b7656004"
	gapB194TerrorOracle    = "557bd3c9-1c60-480d-9650-30c317506097"
	gapB194GhyrsonOracle   = "b110b8e6-859b-4370-aceb-b21c6cf61532"
	gapB194MagnusOracle    = "86c9b01e-18a2-4616-8e04-5343f97d2efe"
	gapB194MutalithOracle  = "7e6ca948-a66d-46a9-9174-ac7b75d7b348"
	gapB194PinkHorrorOracl = "3aa4c374-b146-4890-a8e1-717f214d9741"
)

func TestGapB194CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB194ApostleOracle:   "Dark Apostle",
		gapB194ExocrineOracle:  "Exocrine",
		gapB194HeraldOracle:    "Herald of Slaanesh",
		gapB194KeeperOracle:    "Keeper of Secrets",
		gapB194MarineOracle:    "Noise Marine",
		gapB194ScreamerOracle:  "Screamer-Killer",
		gapB194TerrorOracle:    "The Red Terror",
		gapB194GhyrsonOracle:   "Ghyrson Starn, Kelermorph",
		gapB194MagnusOracle:    "Magnus the Red",
		gapB194MutalithOracle:  "Mutalith Vortex Beast",
		gapB194PinkHorrorOracl: "Pink Horror",
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

// gapB194Settle passes priority until the stack is empty, ordering triggers
// and answering every target prompt with `keep`.
func gapB194Settle(t *testing.T, g *game.Game, who uuid.UUID, keep game.TargetRef) {
	t.Helper()
	for i := 0; i < 96; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, who); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, who, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			if err := g.ResolvePickTarget(p.ID, who, keep); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.CountMayCastPrompts(g) > 0 {
			effectstest.AnswerAllMayCast(t, g, who, false)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func gapB194Opp(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

// --- Dark Apostle ---

func TestGapB194DarkApostleGivesTheNextNoncreatureSpellCascade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCheapLibrary(g, me, 6)
	apostle := effectstest.B12Push(g, me.ID, "Dark Apostle", "Creature — Astartes Warlock", gapB194ApostleOracle, 3, 3)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, apostle, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// A creature spell neither cascades nor uses the promise up.
	effectstest.CastWithCost(t, g, "A Creature", "Creature — Bear", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Fatalf("a creature spell got %d cascade offers, want 0", got)
	}
	effectstest.CastWithCost(t, g, "First Sorcery", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("the next noncreature spell got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Second Sorcery", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("the second noncreature spell got %d cascade offers, want 0", got)
	}
}

func TestGapB194DarkApostleNeedsItsThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	apostle := effectstest.B12Push(g, me.ID, "Dark Apostle", "Creature — Astartes Warlock", gapB194ApostleOracle, 3, 3)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, apostle, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated Gift of Chaos with two mana")
	}
}

// --- Exocrine ---

func gapB194CastExocrine(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	id, err := effectstest.SpmB4Cast(t, g, "Exocrine", "Creature — Tyranid", gapB194ExocrineOracle, "{X}{2}{R}", 2, 2, x)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestGapB194ExocrineEntersWithXCountersAndHitsEveryoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 9)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	hand := me.Hand.Size()
	id := gapB194CastExocrine(t, g, 5)
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if got := effectstest.PlusOnes(g, id); got != 5 {
		t.Errorf("the Exocrine has %d counters, want 5", got)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Exocrine damaged itself")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("my own other creature took no damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.DamageMarked != 5 {
		t.Errorf("the opponent's creature has %d damage, want 5", c.DamageMarked)
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-5 {
			t.Errorf("%s is at %d, want %d (every player takes X)", p.Name, p.Life, lives[p.ID]-5)
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (X = 5 draws a card)", me.Hand.Size(), hand+1)
	}
}

func TestGapB194ExocrineForXFourDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	id := gapB194CastExocrine(t, g, 4)
	gapB194Settle(t, g, me.ID, gapB194Opp(g.Seats[1].ID))
	if got := effectstest.PlusOnes(g, id); got != 4 {
		t.Errorf("the Exocrine has %d counters, want 4", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (X = 4 draws nothing)", me.Hand.Size(), hand)
	}
}

func TestGapB194ExocrineForXZeroDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	gapB194CastExocrine(t, g, 0)
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life || !g.Battlefield.Contains(bear) {
		t.Errorf("X = 0 dealt damage (life %d -> %d)", life, opp.Life)
	}
}

// --- Herald of Slaanesh ---

func TestGapB194HeraldDiscountsDemonSpellsByTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Herald of Slaanesh", "Creature — Demon", gapB194HeraldOracle, 2, 2)
	effectstest.ToMainForCost(t, g)
	demon := effectstest.HandCardFull(me, "A Demon", "Creature — Demon", "{3}{R}", "", []string{"R"})
	bear := effectstest.HandCardFull(me, "A Bear", "Creature — Bear", "{3}{R}", "", []string{"R"})
	effectstest.B06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a non-Demon spell was discounted")
	}
	if err := g.CastSpell(me.ID, demon, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Demon spell was not {2} cheaper: %v", err)
	}
}

func TestGapB194HeraldGivesOtherDemonsHasteButNotOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := effectstest.B12Push(g, me.ID, "Herald of Slaanesh", "Creature — Demon", gapB194HeraldOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Demon", "Creature — Demon", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Demon", "Creature — Demon", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("my other Demon has no haste")
	}
	for name, id := range map[string]uuid.UUID{"the Herald": herald, "a Bear": bear, "an opponent's Demon": theirs} {
		if effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has haste", name)
		}
	}
}

// --- Keeper of Secrets ---

func gapB194CastFromCommand(t *testing.T, g *game.Game, name, typeLine, cost string) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Command.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: active.ID, Controller: active.ID, IsCommander: true})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast: %v", err)
	}
}

func TestGapB194KeeperOfSecretsDamagesForASpellCastFromAnotherZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Keeper of Secrets", "Creature — Demon", gapB194KeeperOracle, 6, 4)
	life := opp.Life
	gapB194CastFromCommand(t, g, "My Commander", "Legendary Creature — Human", "{2}{G}{U}")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 4 (the commander's mana value)", life, opp.Life)
	}
}

func TestGapB194KeeperOfSecretsIgnoresASpellCastFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Keeper of Secrets", "Creature — Demon", gapB194KeeperOracle, 6, 4)
	life := opp.Life
	effectstest.CastWithCost(t, g, "A Sorcery", "Sorcery", "{3}{G}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life {
		t.Errorf("a spell cast from hand dealt damage (%d -> %d)", life, opp.Life)
	}
}

func TestGapB194KeeperOfSecretsOnlyTargetsOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Keeper of Secrets", "Creature — Demon", gapB194KeeperOracle, 6, 4)
	gapB194CastFromCommand(t, g, "My Commander", "Legendary Creature — Human", "{2}{G}")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB194Opp(me.ID)); err == nil {
		t.Fatal("the Keeper's controller was a legal target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB194Opp(opp.ID)); err != nil {
		t.Fatalf("an opponent was refused: %v", err)
	}
}

// --- Noise Marine ---

func TestGapB194NoiseMarineDealsDamageEqualToTheSpellsYouveCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastWithCost(t, g, "First Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := effectstest.SpmB4Cast(t, g, "Noise Marine", "Creature — Astartes Warrior", gapB194MarineOracle, "{4}{R}", 3, 2, 0); err != nil {
		t.Fatal(err)
	}
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 (two spells cast this turn)", life, opp.Life)
	}
}

func TestGapB194NoiseMarineCascadesWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCheapLibrary(g, me, 3)
	effectstest.CastWithCost(t, g, "Noise Marine", "Creature — Astartes Warrior", "{4}{R}", gapB194MarineOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

// --- Screamer-Killer ---

func TestGapB194ScreamerKillerDealsFiveForABigCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Screamer-Killer", "Creature — Tyranid", gapB194ScreamerOracle, 5, 5)
	life := opp.Life
	effectstest.CastWithCost(t, g, "Four Drop", "Creature — Bear", "{3}{G}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	effectstest.CastWithCost(t, g, "Five Sorcery", "Sorcery", "{4}{G}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life {
		t.Fatalf("a small creature or a noncreature spell dealt damage (%d -> %d)", life, opp.Life)
	}
	effectstest.CastWithCost(t, g, "Five Drop", "Creature — Bear", "{4}{G}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life-5 {
		t.Errorf("opponent life %d -> %d, want 5", life, opp.Life)
	}
}

// --- The Red Terror ---

func gapB194Red(g *game.Game, owner uuid.UUID, name string, power int) uuid.UUID {
	return effectstest.B21Push(g, owner, name, "Creature — Bear", "", power, 5, "R")
}

func TestGapB194RedTerrorGrowsForEachRedSourceYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	terror := effectstest.B21Push(g, me.ID, "The Red Terror", "Legendary Creature — Tyranid", gapB194TerrorOracle, 4, 3, "R")
	a := gapB194Red(g, me.ID, "Red One", 2)
	b := gapB194Red(g, me.ID, "Red Two", 3)
	green := effectstest.B21Push(g, me.ID, "Green", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackWith(t, g, opp.ID, a, b, green)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, terror); got != 2 {
		t.Errorf("The Red Terror has %d counters, want 2 (one per red source, none for the green one)", got)
	}
}

func TestGapB194RedTerrorIgnoresAnOpponentsRedSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	terror := effectstest.B21Push(g, opp.ID, "The Red Terror", "Legendary Creature — Tyranid", gapB194TerrorOracle, 4, 3, "R")
	mine := gapB194Red(g, me.ID, "My Red", 2)
	effectstest.AttackWith(t, g, opp.ID, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, terror); got != 0 {
		t.Errorf("an opponent's Red Terror grew %d off my red source", got)
	}
}

func TestGapB194RedTerrorTriggersOncePerSourceForSeveralRecipients(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	terror := effectstest.B21Push(g, me.ID, "The Red Terror", "Legendary Creature — Tyranid", gapB194TerrorOracle, 4, 3, "R")
	attacker := gapB194Red(g, me.ID, "Red Attacker", 4)
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 1, 5)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 1, 5)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(b1, attacker); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(b2, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment && c.DamageAssignment != nil {
			var entries []game.DamageAssignmentEntry
			for _, b := range c.DamageAssignment.BlockerIDs {
				entries = append(entries, game.DamageAssignmentEntry{BlockerID: b, Amount: c.DamageAssignment.AttackerPower / len(c.DamageAssignment.BlockerIDs)})
			}
			if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 0); err != nil {
				t.Fatalf("ResolveDamageAssignment: %v", err)
			}
			break
		}
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, b := range []uuid.UUID{b1, b2} {
		if c, ok := effectstest.BattlefieldCard(g, b); !ok || c.DamageMarked == 0 {
			t.Fatalf("a blocker took no damage (the test needs the attacker to damage both)")
		}
	}
	if got := effectstest.PlusOnes(g, terror); got != 1 {
		t.Errorf("The Red Terror has %d counters, want 1 for one source dealing damage to several creatures at once", got)
	}
}

// --- Ghyrson Starn, Kelermorph ---

func TestGapB194GhyrsonDealsTwoMoreWhenAnotherSourceDealsExactlyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ghyrson Starn, Kelermorph", "Legendary Creature — Tyranid Human", gapB194GhyrsonOracle, 3, 2)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, pinger)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 1 combat damage and 2 from Ghyrson", life, opp.Life)
	}
}

func TestGapB194GhyrsonIgnoresDamageOtherThanExactlyOneAndItsOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ghyrson := effectstest.B12Push(g, me.ID, "Ghyrson Starn, Kelermorph", "Legendary Creature — Tyranid Human", gapB194GhyrsonOracle, 1, 2)
	two := effectstest.B12Creature(g, me.ID, "Two Power", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ghyrson, two)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want only 1 + 2 combat damage (Ghyrson's own 1 does not trigger him)", life, opp.Life)
	}
}

func TestGapB194GhyrsonHitsAPermanentTheOneDamageWentTo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ghyrson Starn, Kelermorph", "Legendary Creature — Tyranid Human", gapB194GhyrsonOracle, 3, 2)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 9)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AttackIntoBlocks(t, g, opp.ID, pinger)
	if err := g.DeclareBlocker(wall, pinger); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, wall); !ok || c.DamageMarked != 3 {
		t.Errorf("the blocker has %d damage, want 1 + 2", c.DamageMarked)
	}
}

// --- Magnus the Red ---

func TestGapB194MagnusDiscountsInstantsAndSorceriesPerCreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Magnus the Red", "Legendary Creature — Demon Primarch", gapB194MagnusOracle, 4, 5)
	effectstest.B12Creature(g, me.ID, "Token One", "Token Creature — Spirit", 1, 1)
	effectstest.B12Creature(g, me.ID, "Token Two", "Token Creature — Spirit", 1, 1)
	effectstest.B12Creature(g, me.ID, "Not A Token", "Creature — Bear", 1, 1)
	effectstest.ToMainForCost(t, g)
	sorcery := effectstest.HandCardFull(me, "A Sorcery", "Sorcery", "{2}{R}", "", []string{"R"})
	creature := effectstest.HandCardFull(me, "A Creature", "Creature — Bear", "{2}{R}", "", []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was discounted")
	}
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the sorcery was not {2} cheaper for two creature tokens: %v", err)
	}
}

func TestGapB194MagnusNoDiscountWithoutTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Magnus the Red", "Legendary Creature — Demon Primarch", gapB194MagnusOracle, 4, 5)
	effectstest.ToMainForCost(t, g)
	sorcery := effectstest.HandCardFull(me, "A Sorcery", "Sorcery", "{2}{R}", "", []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the sorcery was discounted with no creature tokens")
	}
}

func TestGapB194MagnusMakesASpawnWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	magnus := effectstest.B12Push(g, me.ID, "Magnus the Red", "Legendary Creature — Demon Primarch", gapB194MagnusOracle, 4, 5)
	effectstest.AttackWith(t, g, opp.ID, magnus)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var spawn game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spawn" && c.Controller == me.ID {
			spawn = c
		}
	}
	if spawn.InstanceID == uuid.Nil {
		t.Fatal("no Spawn token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, spawn.InstanceID); p != 3 || tough != 3 || !spawn.HasColor("R") {
		t.Errorf("the Spawn is %d/%d colours %v, want a 3/3 red", p, tough, spawn.Colors)
	}
}

// --- Mutalith Vortex Beast ---

func TestGapB194MutalithFlipsOnePerOpponentAndPaysOutEachResult(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	hand := me.Hand.Size()
	if _, err := effectstest.SpmB4Cast(t, g, "Mutalith Vortex Beast", "Creature — Mutant Beast", gapB194MutalithOracle, "{4}{U}{R}", 6, 6, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if ask.CoinCount != 3 {
		t.Fatalf("the call is for %d coins, want one per opponent (3)", ask.CoinCount)
	}
	if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall) != nil {
		t.Fatal("a second coin call was asked, want one prompt for all three flips")
	}
	wins := me.Hand.Size() - hand
	if wins < 0 || wins > 3 {
		t.Fatalf("drew %d cards", wins)
	}
	lost := 0
	for _, p := range g.Seats[1:] {
		switch d := lives[p.ID] - p.Life; d {
		case 0:
		case 3:
			lost++
		default:
			t.Errorf("%s lost %d life, want 0 or 3", p.Name, d)
		}
	}
	if lost != 3-wins {
		t.Errorf("%d wins and %d opponents hit for 3, want them to add up to 3 flips", wins, lost)
	}
}

// --- Pink Horror ---

func TestGapB194PinkHorrorDealsTwoWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Pink Horror", "Creature — Demon Horror", gapB194PinkHorrorOracl, 4, 4)
	life := opp.Life
	effectstest.CastWithCost(t, g, "A Creature", "Creature — Bear", "{1}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life {
		t.Fatalf("a creature spell triggered it (%d -> %d)", life, opp.Life)
	}
	effectstest.CastWithCost(t, g, "A Sorcery", "Sorcery", "{1}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2", life, opp.Life)
	}
}

func TestGapB194PinkHorrorSplitsIntoTwoBlueHorrorsThatPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	horror := effectstest.B12Push(g, me.ID, "Pink Horror", "Creature — Demon Horror", gapB194PinkHorrorOracl, 4, 4)
	effectstest.KillCreature(t, g, me.ID, horror)
	effectstest.PassPriorityAroundTable(t, g)
	var tokens []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Blue Horror" && c.Controller == me.ID {
			tokens = append(tokens, c)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("%d Blue Horror tokens, want 2", len(tokens))
	}
	for _, c := range tokens {
		if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 2 || tough != 2 || !c.HasColor("U") || !c.HasColor("R") || !c.HasSubtype("Demon") || !c.HasSubtype("Horror") {
			t.Errorf("a Blue Horror is %d/%d, colours %v, types %q", p, tough, c.Colors, c.TypeLine)
		}
	}
	life := opp.Life
	effectstest.CastWithCost(t, g, "A Sorcery", "Sorcery", "{1}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 1 from each of the two Blue Horrors", life, opp.Life)
	}
	life = opp.Life
	effectstest.CastWithCost(t, g, "A Creature", "Creature — Bear", "{1}", "")
	gapB194Settle(t, g, me.ID, gapB194Opp(opp.ID))
	if opp.Life != life {
		t.Errorf("a creature spell pinged (%d -> %d)", life, opp.Life)
	}
}
