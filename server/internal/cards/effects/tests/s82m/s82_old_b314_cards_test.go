package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b314_cards_test.go — Battle for Zendikar, batch 314 (S82, #5547).
// Each test plays the real catalog entry through the engine.

const (
	oldB314ArrowsOracle   = "cb15bcc7-151f-44a4-90ee-2d39d8fd5299"
	oldB314BladeOracle    = "cf7ca005-410d-40c8-aae4-89f41f81848b"
	oldB314HammerOracle   = "f353948c-b5f9-4840-a89a-19206dde9d47"
	oldB314CohortOracle   = "015b1006-decb-4d94-940e-bca550e0ad5b"
	oldB314CastigatorOrac = "fbb9dafa-b1aa-401a-b26c-c0c7e8b7222f"
	oldB314RoilOracle     = "38e43d06-33f1-4672-b0be-b0ea8bf36291"
	oldB314SlaughterOracl = "5f6fbb1d-63c9-4913-8e99-9e8c3b4e8d8c"
	oldB314NetworkOracle  = "5a399467-abb7-4f33-962d-8bef5aa9ebf3"
	oldB314KioraOracle    = "0b829189-ff9d-4f6c-9a2a-fdb4a9bfeeb6"
	oldB314HerderOracle   = "3159d289-b822-4657-919a-d712623e622e"
	oldB314MantisOracle   = "481ede6a-3ac3-45bc-85aa-311b3dbefc40"
)

func TestOldB314EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB314ArrowsOracle:   "Pathway Arrows",
		oldB314BladeOracle:    "Hedron Blade",
		oldB314HammerOracle:   "Slab Hammer",
		oldB314CohortOracle:   "Reckless Cohort",
		oldB314CastigatorOrac: "Kor Castigator",
		oldB314RoilOracle:     "Roil's Retribution",
		oldB314SlaughterOracl: "Gruesome Slaughter",
		oldB314NetworkOracle:  "Aligned Hedron Network",
		oldB314KioraOracle:    "Kiora, Master of the Depths",
		oldB314HerderOracle:   "Blight Herder",
		oldB314MantisOracle:   "Giant Mantis",
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

func TestOldB314GiantMantisHasReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Giant Mantis", "Creature — Insect", oldB314MantisOracle, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach")
	}
}

// --- Pathway Arrows --------------------------------------------------------

func TestOldB314PathwayArrowsPingsAndTapsOnlyAColorlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	arrows := effectstest.SeedEquipment(g, me.ID, "Pathway Arrows", oldB314ArrowsOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, arrows, holder)
	colorless := effectstest.B12Push(g, opp.ID, "Drone", "Creature — Eldrazi Drone", "", 3, 3)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, holder, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTarget(colorless)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageOn(t, g, colorless) != 1 {
		t.Errorf("the colorless creature has %d damage, want 1", effectstest.DamageOn(t, g, colorless))
	}
	if !effectstest.TappedOnBattlefield(t, g, colorless) {
		t.Error("a colorless creature dealt damage was not tapped")
	}
	if !effectstest.TappedOnBattlefield(t, g, holder) {
		t.Error("the {T} cost did not tap the holder")
	}
}

func TestOldB314PathwayArrowsDoesNotTapAColoredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	arrows := effectstest.SeedEquipment(g, me.ID, "Pathway Arrows", oldB314ArrowsOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, arrows, holder)
	red := effectstest.B12Push(g, opp.ID, "Red Bear", "Creature — Bear", "", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == red {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.FloatForTest(g, me, "CC")
	effectstest.ApaActivate(t, g, me, holder, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(red)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageOn(t, g, red) != 1 {
		t.Errorf("the red creature has %d damage, want 1", effectstest.DamageOn(t, g, red))
	}
	if effectstest.TappedOnBattlefield(t, g, red) {
		t.Error("a colored creature was tapped")
	}
}

func TestOldB314PathwayArrowsAbilityLeavesWithTheEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, holder, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTarget(holder)}); err == nil {
		t.Error("a creature with no Arrows has the ability")
	}
}

// --- Hedron Blade ----------------------------------------------------------

func TestOldB314HedronBladePumpsAndGrantsDeathtouchWhenBlockedByColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	blade := effectstest.SeedEquipment(g, me.ID, "Hedron Blade", oldB314BladeOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, blade, holder)
	if p, tough := effectstest.PtOf(t, g, holder); p != 3 || tough != 3 {
		t.Fatalf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	drone := effectstest.B12Push(g, opp.ID, "Drone", "Creature — Eldrazi Drone", "", 2, 2)
	effectstest.AttackAndTrigger(t, g, holder)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{drone: holder})
	if !effectstest.HasEffectiveKeyword(t, g, holder, "deathtouch") {
		t.Error("blocked by a colorless creature, the equipped creature did not gain deathtouch")
	}
}

func TestOldB314HedronBladeIgnoresAColoredBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	blade := effectstest.SeedEquipment(g, me.ID, "Hedron Blade", oldB314BladeOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, blade, holder)
	blocker := effectstest.B12Push(g, opp.ID, "Red Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == blocker {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.AttackAndTrigger(t, g, holder)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: holder})
	if effectstest.HasEffectiveKeyword(t, g, holder, "deathtouch") {
		t.Error("a colored blocker gave deathtouch")
	}
}

// --- Slab Hammer -----------------------------------------------------------

func TestOldB314SlabHammerReturnsALandForAPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	hammer := effectstest.SeedEquipment(g, me.ID, "Slab Hammer", oldB314HammerOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, hammer, holder)
	effectstest.AttackAndTrigger(t, g, holder)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land was not returned to hand")
	}
	if p, tough := effectstest.PtOf(t, g, holder); p != 4 || tough != 4 {
		t.Errorf("the attacker is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB314SlabHammerDecliningGivesNoPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 2)
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	hammer := effectstest.SeedEquipment(g, me.ID, "Slab Hammer", oldB314HammerOracle)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.EquipTo(t, g, me.ID, hammer, holder)
	effectstest.AttackAndTrigger(t, g, holder)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick == nil {
		t.Fatal("no optional pick was offered")
	} else if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Error("the land left though the player declined")
	}
	if p, _ := effectstest.PtOf(t, g, holder); p != 2 {
		t.Errorf("power %d, want 2 (no pump)", p)
	}
}

// --- Reckless Cohort -------------------------------------------------------

func TestOldB314RecklessCohortMustAttackUnlessYouControlAnotherAlly(t *testing.T) {
	for _, tc := range []struct {
		name      string
		other     string
		mustAtack bool
	}{
		{"alone", "", true},
		{"with another Ally", "Creature — Human Ally", false},
		{"with a non-Ally creature", "Creature — Bear", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B12Push(g, me.ID, "Reckless Cohort", "Creature — Human Warrior Ally", oldB314CohortOracle, 2, 2)
			if tc.other != "" {
				effectstest.B12Creature(g, me.ID, "Other", tc.other, 1, 1)
			}
			if got := effectstest.FaMustAttack(t, g); got != tc.mustAtack {
				t.Errorf("must attack = %v, want %v", got, tc.mustAtack)
			}
		})
	}
}

// --- Kor Castigator --------------------------------------------------------

func TestOldB314KorCastigatorCantBeBlockedByEldraziScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	kor := effectstest.B12Push(g, me.ID, "Kor Castigator", "Creature — Kor Wizard Ally", oldB314CastigatorOrac, 3, 1)
	scion := effectstest.B12Push(g, opp.ID, "Eldrazi Scion", "Token Creature — Eldrazi Scion", "", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackAndTrigger(t, g, kor)
	effectstest.OnToBlockers(t, g)
	if err := effectstest.ReqBlock(t, g, kor, scion); err == nil {
		t.Error("an Eldrazi Scion blocked Kor Castigator")
	}
	if err := effectstest.ReqBlock(t, g, kor, bear); err != nil {
		t.Errorf("a Bear could not block Kor Castigator: %v", err)
	}
}

// --- Roil's Retribution ----------------------------------------------------

func TestOldB314RoilsRetributionDividesFiveAmongAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 3)
	effectstest.AttackAndTrigger(t, g, a, b)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Roil's Retribution", TypeLine: "Instant", ManaCost: "{3}{W}{W}", OracleID: oldB314RoilOracle, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "W", "W", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(idle), Distribution: map[uuid.UUID]int{idle: 5}}); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 3, b: 1}}); err == nil {
		t.Fatal("a split of 4 was accepted for 5 damage")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 3, b: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("A took 3 and should be dead")
	}
	if !g.Battlefield.Contains(b) || effectstest.DamageOn(t, g, b) != 2 {
		t.Errorf("B has %d damage, want 2 and alive", effectstest.DamageOn(t, g, b))
	}
}

// --- Gruesome Slaughter ----------------------------------------------------

func TestOldB314GruesomeSlaughterGivesColorlessCreaturesTheTapDamageAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	drone := effectstest.B12Push(g, me.ID, "Drone", "Creature — Eldrazi Drone", "", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gruesome Slaughter", TypeLine: "Sorcery", ManaCost: "{6}", OracleID: oldB314SlaughterOracl, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("a colored creature got the ability")
	}
	effectstest.ApaActivate(t, g, me, drone, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage from the Drone should have killed the 2/3 victim")
	}
}

// --- Aligned Hedron Network ------------------------------------------------

func TestOldB314AlignedHedronNetworkExilesBigCreaturesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Fatty", "Creature — Giant", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Fatty", "Creature — Giant", 7, 7)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 4, 4)
	net := effectstest.CastHoldCreature(t, g, "Aligned Hedron Network", "Artifact", oldB314NetworkOracle, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a creature with power 5 or more stayed on the battlefield")
		}
	}
	if !g.Battlefield.Contains(small) {
		t.Error("a power 4 creature was exiled")
	}
	effectstest.B14Kill(g, net)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	// A returned card is a new object (CR 400.7), so it is found by name.
	if effectstest.FindBattlefieldByName(g, "My Fatty") == uuid.Nil || effectstest.FindBattlefieldByName(g, "Their Fatty") == uuid.Nil {
		t.Error("an exiled creature did not return when the Network left")
	}
	if c, ok := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Their Fatty")); !ok || c.Controller != opp.ID {
		t.Error("their creature did not return under its owner's control")
	}
}

// --- Blight Herder ---------------------------------------------------------

func TestOldB314BlightHerderPutsTwoOpponentCardsInTheGraveyardForThreeScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushExiledCard(g, opp.ID, "Their A", "Creature — Bear", false)
	b := effectstest.PushExiledCard(g, opp.ID, "Their B", "Creature — Bear", false)
	mine := effectstest.PushExiledCard(g, me.ID, "My Card", "Creature — Bear", false)
	effectstest.FloatForTest(g, me, "CCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Blight Herder", TypeLine: "Creature — Eldrazi Processor", ManaCost: "{5}", OracleID: oldB314HerderOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsPrompt(t, g)
	if effectstest.HasID(p.ChooseCards, mine) {
		t.Error("my own exiled card is offered")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{a}); err == nil {
		t.Fatal("a single card was accepted: it is two or none")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the cards are not in their owner's graveyard")
	}
	if n := effectstest.CountOnBattlefield(g, "Eldrazi Scion", me.ID); n != 3 {
		t.Errorf("%d Eldrazi Scions, want 3", n)
	}
}

func TestOldB314BlightHerderDecliningMakesNoScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushExiledCard(g, opp.ID, "Their A", "Creature — Bear", false)
	effectstest.PushExiledCard(g, opp.ID, "Their B", "Creature — Bear", false)
	effectstest.FloatForTest(g, me, "CCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Blight Herder", TypeLine: "Creature — Eldrazi Processor", ManaCost: "{5}", OracleID: oldB314HerderOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if opp.Graveyard.Contains(a) {
		t.Error("a card moved though the player declined")
	}
	if n := effectstest.CountOnBattlefield(g, "Eldrazi Scion", me.ID); n != 0 {
		t.Errorf("%d Scions after declining, want 0", n)
	}
}

// --- Kiora, Master of the Depths -------------------------------------------

func TestOldB314KioraPlusOneUntapsACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kiora := effectstest.S82Walker(g, me.ID, "Kiora, Master of the Depths", oldB314KioraOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == bear || id == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.S82Activate(t, g, me.ID, kiora, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: land}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedOnBattlefield(t, g, bear) || effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("the creature and the land were not both untapped")
	}
	if n := effectstest.CounterCount(g, kiora, game.CounterLoyalty); n != 5 {
		t.Errorf("loyalty %d, want 5", n)
	}
}

func TestOldB314KioraMinusTwoTakesACreatureAndALandAndMillsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kiora := effectstest.S82Walker(g, me.ID, "Kiora, Master of the Depths", oldB314KioraOracle, 4)
	effectstest.AdvanceToMain(t, g)
	var ids [4]uuid.UUID
	for i, c := range []game.Card{
		effectstest.CardNamed("Land B", "Basic Land — Forest"),
		effectstest.CardNamed("Spell", "Sorcery"),
		effectstest.CardNamed("Creature B", "Creature — Bear"),
		effectstest.CardNamed("Creature A", "Creature — Bear"),
	} {
		c.InstanceID = uuid.New()
		ids[3-i] = c.InstanceID
		effectstest.LibraryTop(me, c)
	}
	creatureA, creatureB, spell, landB := ids[0], ids[1], ids[2], ids[3]
	effectstest.S82Activate(t, g, me.ID, kiora, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolveChooseCards(effectstest.ChooseCardsPrompt(t, g).ID, me.ID, []uuid.UUID{creatureA, creatureB}); err == nil {
		t.Fatal("two creature cards were accepted: it is a creature and/or a land")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, creatureA, landB)
	if !me.Hand.Contains(creatureA) || !me.Hand.Contains(landB) {
		t.Error("the creature and the land are not in hand")
	}
	if !me.Graveyard.Contains(creatureB) || !me.Graveyard.Contains(spell) {
		t.Error("the rest did not go into the graveyard")
	}
	if n := effectstest.CounterCount(g, kiora, game.CounterLoyalty); n != 2 {
		t.Errorf("loyalty %d, want 2", n)
	}
}

func TestOldB314KioraMinusEightMakesAnEmblemAndThreeOctopuses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kiora := effectstest.S82Walker(g, me.ID, "Kiora, Master of the Depths", oldB314KioraOracle, 8)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kiora, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// The emblem is made first, so each Octopus entering triggers it (CR 603.2).
	for i := 0; i < 3; i++ {
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.EmblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Octopus" && c.Controller == me.ID {
			n++
			e := c.Effective()
			if e.Power != 8 || e.Toughness != 8 || !c.IsToken() {
				t.Errorf("Octopus %d/%d, want an 8/8 token", e.Power, e.Toughness)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Octopus tokens, want 3", n)
	}
}

func TestOldB314KioraEmblemMakesAnEnteringCreatureFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kiora := effectstest.S82Walker(g, me.ID, "Kiora, Master of the Depths", oldB314KioraOracle, 8)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kiora, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Big Bear", "Creature — Bear", "", 4, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the entering creature did not fight and kill the victim")
	}
}
