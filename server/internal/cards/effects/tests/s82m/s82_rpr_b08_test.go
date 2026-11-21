package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprMasterfulReplicationOracle = "cc6e6e16-da2b-4fc0-91ad-b76f0a101306"
	rprMigratoryRouteOracle       = "f6a74582-62b6-48d1-89d6-e26059229840"
	rprOrderOfMidnightOracle      = "2db17cfc-0947-407f-bd53-d64a48eec314"
	rprOrzhovLocketOracle         = "8aaf248e-66cf-447b-b062-9aa2ab1ec864"
	rprPhyrexianGhoulOracle       = "748c2723-dc59-4aa7-9667-d4679448c8b6"
	rprPhyrexianTriniformOracle   = "f8526a36-eb8f-457d-b532-e12b47622cfa"
	rprRiseAndShineOracle         = "59925461-99f5-4217-8343-49c9b1ce24c1"
	rprRootOutOracle              = "bec03edc-261a-4b02-be37-2006f657e905"
	rprSaheelisArtistryOracle     = "0038e74a-b950-47ff-8853-dba6a3b31091"
	rprScytheclawOracle           = "a75ddd9c-2574-46e9-ac85-58082789ff3e"
	rprSeraphOfTheMassesOracle    = "5d6f75f1-dd9c-4d94-aea8-3bdec41eaa29"
)

func TestRprB08EveryCardIsRegisteredFull(t *testing.T) {
	for name, id := range map[string]string{
		"Masterful Replication": rprMasterfulReplicationOracle,
		"Migratory Route":       rprMigratoryRouteOracle,
		"Order of Midnight":     rprOrderOfMidnightOracle,
		"Alter Fate":            rprOrderOfMidnightOracle + "#1",
		"Orzhov Locket":         rprOrzhovLocketOracle,
		"Phyrexian Ghoul":       rprPhyrexianGhoulOracle,
		"Phyrexian Triniform":   rprPhyrexianTriniformOracle,
		"Rise and Shine":        rprRiseAndShineOracle,
		"Root Out":              rprRootOutOracle,
		"Saheeli's Artistry":    rprSaheelisArtistryOracle,
		"Scytheclaw":            rprScytheclawOracle,
		"Seraph of the Masses":  rprSeraphOfTheMassesOracle,
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func rprTokens(g *game.Game, name string) []uuid.UUID { return effectstest.B39TokensNamed(g, name) }

func TestRprB08PhyrexianGhoulSacrificesForPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.B12Push(g, me.ID, "Phyrexian Ghoul", "Creature — Phyrexian Zombie", rprPhyrexianGhoulOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
	if p, tough := effectstest.MkmB1PT(t, g, ghoul); p != 4 || tough != 4 {
		t.Errorf("Ghoul is %d/%d, want 4/4", p, tough)
	}
	// Strict: an opponent's creature is not a legal sacrifice.
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{theirs}, Strict: true,
	}); err == nil {
		t.Error("sacrificed an opponent's creature")
	}
}

func TestRprB08ScytheclawLivingWeaponAndHalvesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Scytheclaw", "Artifact — Equipment", rprScytheclawOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	germs := rprTokens(g, "Phyrexian Germ")
	if len(germs) != 1 {
		t.Fatalf("%d Germs, want 1", len(germs))
	}
	if p, tough := effectstest.MkmB1PT(t, g, germs[0]); p != 1 || tough != 1 {
		t.Errorf("Germ is %d/%d, want 1/1", p, tough)
	}
	// A ready creature with the Scytheclaw attached connects.
	var claw uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Scytheclaw" {
				claw = c.InstanceID
			}
		}
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.P7Activate(t, g, me, claw, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if p, _ := effectstest.MkmB1PT(t, g, bear); p != 3 {
		t.Fatalf("equipped Bear power %d, want 3", p)
	}
	opp.Life = 40
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	for i := 0; i < 6 && opp.Life == 40; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// 40 - 3 = 37; half of 37 rounded up is 19; 37 - 19 = 18.
	if opp.Life != 18 {
		t.Errorf("opponent life %d, want 18", opp.Life)
	}
}

func TestRprB08SeraphIsAsBigAsYourCreatureCountAndNeedsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seraph := effectstest.B12Push(g, me.ID, "Seraph of the Masses", "Creature — Angel", rprSeraphOfTheMassesOracle, 0, 0)
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	if p, tough := effectstest.MkmB1PT(t, g, seraph); p != 3 || tough != 3 {
		t.Errorf("Seraph is %d/%d, want 3/3", p, tough)
	}
	spec, _ := effects.Lookup(rprSeraphOfTheMassesOracle)
	if spec.TapCost == nil {
		t.Error("Seraph has no convoke")
	}
	effectstest.AssertKeywords(t, g, seraph, "flying")
}

func TestRprB08RootOutDestroysAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Root Out", "Sorcery", rprRootOutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Root Out", "Sorcery", rprRootOutOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if len(rprTokens(g, "Clue")) != 1 {
		t.Errorf("%d Clues, want 1", len(rprTokens(g, "Clue")))
	}
}

func TestRprB08MigratoryRouteMakesFourBirdsAndLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Migratory Route", "Sorcery", rprMigratoryRouteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	birds := rprTokens(g, "Bird")
	if len(birds) != 4 {
		t.Fatalf("%d Birds, want 4", len(birds))
	}
	effectstest.AssertKeywords(t, g, birds[0], "flying")

	id := effectstest.PushCatalogHandCard(me, "Migratory Route", "Sorcery", rprMigratoryRouteOracle)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("landcycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(id) {
		t.Error("landcycling did not fetch the Forest")
	}
}

func TestRprB08TriniformDiesIntoThreeGolemsAndEncores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tri := effectstest.B12Push(g, me.ID, "Phyrexian Triniform", "Artifact Creature — Phyrexian Golem", rprPhyrexianTriniformOracle, 9, 9)
	effectstest.KillCreature(t, g, me.ID, tri)
	effectstest.PassPriorityAroundTable(t, g)
	golems := rprTokens(g, "Phyrexian Golem")
	if len(golems) != 3 {
		t.Fatalf("%d Golems, want 3", len(golems))
	}
	c := effectstest.LayeredCard(t, g, golems[0])
	if !c.IsArtifact() || !c.IsCreature() {
		t.Error("the Golem is not an artifact creature")
	}
	if p, tough := effectstest.MkmB1PT(t, g, golems[0]); p != 3 || tough != 3 {
		t.Errorf("Golem is %d/%d, want 3/3", p, tough)
	}

	g2 := effectstest.NewCatalogGame(t)
	id, _ := effectstest.ActivateFromGraveyard(t, g2, "Phyrexian Triniform", "Artifact Creature — Phyrexian Golem",
		rprPhyrexianTriniformOracle, 9, 9, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}", game.ActivateAbilityParams{})
	if !g2.Exile.Contains(id) {
		t.Error("encore did not exile the card")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	n := 0
	g2.ReadSnapshot(func() {
		for _, c := range g2.Battlefield.Cards {
			if c.Name == "Phyrexian Triniform" {
				n++
			}
		}
	})
	if n != 3 {
		t.Errorf("%d encore tokens, want 3", n)
	}
}

func TestRprB08OrzhovLocketTapsForEitherAndSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	locket := effectstest.B12Push(g, me.ID, "Orzhov Locket", "Artifact", rprOrzhovLocketOracle, 0, 0)
	spec, _ := effects.Lookup(rprOrzhovLocketOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|B}" {
		t.Fatalf("mana abilities: %+v", spec.ManaAbilities)
	}
	// Four hybrid symbols are paid with any mix of {W} and {B}.
	effectstest.AuraSeedLibrary(me, 3)
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{W}{W}{B}{B}")
	effectstest.P7Activate(t, g, me, locket, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(locket) {
		t.Error("the Locket was not sacrificed")
	}
	if me.Hand.Size()-hand != 2 {
		t.Errorf("drew %d, want 2", me.Hand.Size()-hand)
	}
	// Strict: a pool that cannot pay is refused.
	locket2 := effectstest.B12Push(g, me.ID, "Orzhov Locket", "Artifact", rprOrzhovLocketOracle, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, locket2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without paying")
	}
}

func TestRprB08OrderOfMidnightCantBlockAndAlterFateReturnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spec, _ := effects.Lookup(rprOrderOfMidnightOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flying" {
		t.Errorf("keywords %v", spec.PrintedKeywords)
	}
	body := effectstest.GraveCreatureMV(me, "Dead Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Alter Fate", "Sorcery — Adventure", rprOrderOfMidnightOracle+"#1",
		[]game.TargetRef{{Kind: game.TargetCard, ID: body}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(body) {
		t.Error("the creature card did not return to hand")
	}

	mid := effectstest.B12Push(g, me.ID, "Order of Midnight", "Creature — Human Knight", rprOrderOfMidnightOracle, 2, 2)
	if effectstest.AuraRestrictions(t, g, mid)&game.CantBlock == 0 {
		t.Error("Order of Midnight can block")
	}
}

func TestRprB08SaheelisArtistryCopiesAnArtifactAndACreatureAsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Odd Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Test Bear", "Creature — Bear", 3, 3)
	effectstest.B02cCastModalSpell(t, g, "Saheeli's Artistry", "Sorcery", rprSaheelisArtistryOracle, []int{0, 1},
		[]game.TargetRef{effectstest.P1g3ModeCard(rock, 0), effectstest.P1g3ModeCard(bear, 1)})
	effectstest.PassPriorityAroundTable(t, g)
	rocks := rprTokens(g, "Odd Rock")
	bears := rprTokens(g, "Test Bear")
	if len(rocks) != 1 || len(bears) != 1 {
		t.Fatalf("copies: %d rocks, %d bears", len(rocks), len(bears))
	}
	c := effectstest.LayeredCard(t, g, bears[0])
	if !c.IsArtifact() || !c.IsCreature() {
		t.Error("the creature copy is not an artifact creature")
	}
	if effectstest.LayeredCard(t, g, rocks[0]).IsCreature() {
		t.Error("the artifact copy became a creature")
	}
	if effectstest.LayeredCard(t, g, bear).IsArtifact() {
		t.Error("the original Bear became an artifact")
	}

	// One bullet only: a creature is not a legal target of the artifact bullet.
	if err := func() error {
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Saheeli's Artistry", TypeLine: "Sorcery",
			OracleID: rprSaheelisArtistryOracle, Owner: active.ID, Controller: active.ID})
		return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
			Targets: []game.TargetRef{effectstest.P1g3ModeCard(bear, 0)}, Strict: true})
	}(); err == nil {
		t.Error("a nonartifact creature was a legal target of the artifact bullet")
	}
}

func TestRprB08MasterfulReplicationBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.B02cCastModalSpell(t, g, "Masterful Replication", "Instant", rprMasterfulReplicationOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	golems := rprTokens(g, "Golem")
	if len(golems) != 2 {
		t.Fatalf("%d Golems, want 2", len(golems))
	}

	g = effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.B12Push(g, me.ID, "Big Engine", "Artifact Creature — Construct", "", 5, 5)
	small := effectstest.B12Push(g, me.ID, "Little Rock", "Artifact", "", 0, 0)
	theirs := effectstest.B12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	effectstest.B02cCastModalSpell(t, g, "Masterful Replication", "Instant", rprMasterfulReplicationOracle, []int{1},
		[]game.TargetRef{effectstest.P1g3ModeCard(big, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, small).Name != "Big Engine" {
		t.Errorf("my other artifact is %q, want a copy of Big Engine", effectstest.LayeredCard(t, g, small).Name)
	}
	if effectstest.LayeredCard(t, g, theirs).Name != "Their Rock" {
		t.Error("an opponent's artifact was copied")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	_, _ = g.AdvanceStep()
	effectstest.PassPriorityAroundTable(t, g)
}

func TestRprB08RiseAndShineAnimatesOneOrEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Rise and Shine", "Sorcery", rprRiseAndShineOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)}}); err == nil {
		t.Error("a creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Rise and Shine", "Sorcery", rprRiseAndShineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, a); !c.IsCreature() || !c.IsArtifact() {
		t.Error("Rock A is not an artifact creature")
	}
	if p, tough := effectstest.MkmB1PT(t, g, a); p != 4 || tough != 4 {
		t.Errorf("Rock A is %d/%d, want 4/4", p, tough)
	}
	if effectstest.LayeredCard(t, g, b).IsCreature() {
		t.Error("Rock B was animated by the targeted mode")
	}

	// Overloaded: each noncreature artifact you control.
	effectstest.CastWithAltCost(t, g, "Rise and Shine", "Sorcery", rprRiseAndShineOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, b); p != 4 || tough != 4 {
		t.Errorf("overloaded: Rock B is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.MkmB1PT(t, g, a); p != 4 {
		t.Errorf("overloaded: Rock A (already a creature) got counters again: power %d", p)
	}
}
