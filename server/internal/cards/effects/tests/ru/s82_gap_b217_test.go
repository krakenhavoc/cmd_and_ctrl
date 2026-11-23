package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b217_test.go — Commander Legends: Battle for Baldur's Gate green
// creatures, GAP batch 217 (#5948). Each test plays the real catalog entry.

const (
	gapB217GiantAnkhegOracle     = "ca905672-50c2-4e19-9d08-2d3f8e3debfd"
	gapB217GreenSlimeOracle      = "0ec5a688-a884-410c-adee-ab784789d6e8"
	gapB217HalsinOracle          = "01ab3603-6c2f-4d65-9bb2-da95efdbe622"
	gapB217LovestruckBeastOracle = "7e84aff5-2cb6-4214-befd-3d2de31229e5"
	gapB217MyconidOracle         = "adf59918-a021-4b4e-b672-d82be960726e"
	gapB217OwlbearCubOracle      = "0c92432d-ac76-47d2-af7a-fdffd4b546a2"
	gapB217OwlbearShepherdOracle = "e81035b5-d922-48f8-af20-0025da4721c6"
	gapB217SharpshooterOracle    = "a25308ae-d757-4868-883d-62bbb00d6cbe"
	gapB217InvokerOracle         = "4a25fab1-cabb-44b9-8a69-01fc1fc46bd4"
	gapB217SkanosOracle          = "250052a7-1ba5-45d2-aa32-b787fd637fa5"
	gapB217SweetGumOracle        = "8ea10a5a-b698-4f7d-a9c8-520f8d835f34"
	gapB217TlincalliOracle       = "79b69b38-c01d-499a-a865-8c1cecf9fae0"
)

func TestGapB217EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB217GiantAnkhegOracle:            "Giant Ankheg",
		gapB217GreenSlimeOracle:             "Green Slime",
		gapB217HalsinOracle:                 "Halsin, Emerald Archdruid",
		gapB217LovestruckBeastOracle:        "Lovestruck Beast",
		gapB217LovestruckBeastOracle + "#1": "Heart's Desire",
		gapB217MyconidOracle:                "Myconid Spore Tender",
		gapB217OwlbearCubOracle:             "Owlbear Cub",
		gapB217OwlbearShepherdOracle:        "Owlbear Shepherd",
		gapB217SharpshooterOracle:           "Sharpshooter Elf",
		gapB217InvokerOracle:                "Silvanus's Invoker",
		gapB217SkanosOracle:                 "Skanos Dragonheart",
		gapB217SweetGumOracle:               "Sweet-Gum Recluse",
		gapB217TlincalliOracle:              "Tlincalli Hunter",
		gapB217TlincalliOracle + "#1":       "Retrieve Prey",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// gapB217Pick answers the pick_target prompt a targeted trigger queues, when
// there is one, with every id given.
func gapB217Pick(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestPickTarget(g, chooser)
	if pick == nil {
		return
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func gapB217Read(g *game.Game, id uuid.UUID) (c game.Card) {
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
	return c
}

// --- Giant Ankheg ---

func TestGapB217GiantAnkhegGivesOtherCreaturesYouControlTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ankheg := effectstest.B12Push(g, me.ID, "Giant Ankheg", "Creature — Insect", gapB217GiantAnkhegOracle, 8, 8)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("another creature you control has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, ankheg, "trample") {
		t.Error("Giant Ankheg has no printed trample")
	}
}

func TestGapB217GiantAnkhegGivesOtherCreaturesWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Giant Ankheg", "Creature — Insect", gapB217GiantAnkhegOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, mine)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting another creature")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Green Slime ---

func TestGapB217GreenSlimeCountersAnArtifactsAbilityAndDestroysIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, rock, "Their Rock — do a thing", nil, &fired)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Green Slime", TypeLine: "Creature — Ooze", OracleID: gapB217GreenSlimeOracle, Power: 2, Toughness: 2})
	effectstest.PickTriggerTarget(t, g, me.ID, ability)
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact whose ability was countered was not destroyed")
	}
}

func TestGapB217GreenSlimeCannotTargetACreatureSourcesAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ench := effectstest.PushCatalogPermanent(g, opp.ID, "Their Aura", "Enchantment", "", false)
	fromBear := effectstest.TaAnnounce(t, g, opp.ID, bear, "Their Bear — do a thing", nil, nil)
	fromRock := effectstest.TaAnnounce(t, g, opp.ID, rock, "Their Rock — do a thing", nil, nil)
	fromEnch := effectstest.TaAnnounce(t, g, opp.ID, ench, "Their Aura — do a thing", nil, nil)
	spec, _ := effects.Lookup(gapB217GreenSlimeOracle)
	var clause *game.TargetSpec
	for _, tr := range spec.Triggered {
		if tr.Targets != nil {
			clause = tr.Targets
		}
	}
	legal := effectstest.TaLegalForSpec(g, me.ID, clause)
	if legal[fromBear] {
		t.Error("Green Slime offered an ability from a creature source")
	}
	if !legal[fromRock] || !legal[fromEnch] {
		t.Errorf("Green Slime must offer artifact and enchantment sources: %v", legal)
	}
}

// --- Halsin ---

func TestGapB217HalsinMakesATokenAGreenFourFourBearUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	halsin := effectstest.PushCatalogPermanent(g, me.ID, "Halsin, Emerald Archdruid", "Legendary Creature — Elf Druid", gapB217HalsinOracle, false)
	token := effectstest.B12Push(g, me.ID, "Human", "Token Creature — Human", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == token {
				g.Battlefield.Cards[i].Colors = []string{"W"}
			}
		}
	})
	nonToken := effectstest.B12Creature(g, me.ID, "Nontoken", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, halsin, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nonToken}}}); err == nil {
		t.Fatal("a nontoken creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, halsin, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: token}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, token); p != 4 || tough != 4 {
		t.Errorf("token is %d/%d, want 4/4", p, tough)
	}
	c := gapB217Read(g, token)
	if !c.HasSubtype("Bear") || !c.HasSubtype("Human") || !c.HasColor("G") || !c.HasColor("W") {
		t.Errorf("token types/colours %v / %v, want Human Bear, white and green (in addition to its others)", c.Effective().Subtypes, c.Effective().Colors)
	}
}

// --- Lovestruck Beast // Heart's Desire ---

func TestGapB217LovestruckBeastCantAttackWithoutAOneOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Lovestruck Beast", "Creature — Beast Noble", gapB217LovestruckBeastOracle, 5, 5)
	effectstest.B12Creature(g, me.ID, "Two-Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(beast, opp.ID); err == nil {
		t.Fatal("the Beast attacked with no 1/1 creature under its controller's control")
	}
	effectstest.B12Creature(g, opp.ID, "Their One-One", "Creature — Human", 1, 1)
	if err := g.DeclareAttacker(beast, opp.ID); err == nil {
		t.Fatal("an opponent's 1/1 let the Beast attack")
	}
}

func TestGapB217LovestruckBeastAttacksWithAOneOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Lovestruck Beast", "Creature — Beast Noble", gapB217LovestruckBeastOracle, 5, 5)
	effectstest.B12Creature(g, me.ID, "One-One", "Creature — Human", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(beast, opp.ID); err != nil {
		t.Fatalf("the Beast could not attack with a 1/1 creature: %v", err)
	}
}

func TestGapB217HeartsDesireCreatesAOneOneWhiteHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Heart's Desire", "Sorcery — Adventure", gapB217LovestruckBeastOracle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Human"); got != 1 {
		t.Errorf("%d Human tokens, want 1", got)
	}
}

// --- Myconid Spore Tender ---

func TestGapB217MyconidSporeTenderDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Myconid Spore Tender", TypeLine: "Creature — Fungus", OracleID: gapB217MyconidOracle, Power: 4, Toughness: 1})
	gapB217Pick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestGapB217MyconidSporeTenderWithNoTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Myconid Spore Tender", TypeLine: "Creature — Fungus", OracleID: gapB217MyconidOracle, Power: 4, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

// --- Owlbear Cub ---

func gapB217Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushCatalogPermanent(g, owner, "Forest", "Basic Land — Forest", "", false)
	}
}

func gapB217CubAttack(t *testing.T, lands int) (g *game.Game, me *game.Player, cub, wolf uuid.UUID, libBefore int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB217Lands(g, opp.ID, lands)
	cub = effectstest.PushCatalogPermanent(g, me.ID, "Owlbear Cub", "Creature — Bird Bear", gapB217OwlbearCubOracle, false)
	wolf = uuid.New()
	libBefore = me.Library.Size()
	g.WithWriteLock(func() {
		for i := 0; i < 8; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		}
		me.Library.PushTop(game.Card{InstanceID: wolf, Name: "Library Wolf", TypeLine: "Creature — Wolf", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	effectstest.DeclareAttack(t, g, opp.ID, cub)
	return g, me, cub, wolf, libBefore
}

func TestGapB217OwlbearCubPutsACreatureOntoTheBattlefieldTappedAndAttacking(t *testing.T) {
	g, me, _, wolf, libBefore := gapB217CubAttack(t, 8)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no look-at-the-top prompt for a defender with eight lands")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := gapB217Read(g, wolf)
	if !g.Battlefield.Contains(wolf) || !c.Tapped || c.AttackingTarget != g.Seats[1].ID {
		t.Fatalf("Library Wolf on battlefield %v, tapped %v, attacking %v; want tapped and attacking the defender", g.Battlefield.Contains(wolf), c.Tapped, c.AttackingTarget)
	}
	if me.Library.Size() != libBefore+8 {
		t.Errorf("library has %d cards, want %d (the other eight stayed in it, on the bottom)", me.Library.Size(), libBefore+8)
	}
}

func TestGapB217OwlbearCubDoesNothingAgainstAPlayerWithSevenLands(t *testing.T) {
	g, me, _, wolf, _ := gapB217CubAttack(t, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(wolf) {
		t.Error("the trigger fired against a player with seven lands")
	}
}

// --- Owlbear Shepherd ---

func TestGapB217OwlbearShepherdDrawsOnlyWithTotalPowerEightOrMore(t *testing.T) {
	for _, tc := range []struct {
		power int
		draws bool
	}{{7, false}, {8, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Owlbear Shepherd", "Creature — Goblin Druid", gapB217OwlbearShepherdOracle, false)
		effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", tc.power-1, 3)
		before := me.Hand.Size()
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		got := me.Hand.Size() - before
		if tc.draws && got != 1 || !tc.draws && got != 0 {
			t.Errorf("total power %d: drew %d cards", tc.power, got)
		}
	}
}

// --- Sharpshooter Elf ---

func TestGapB217SharpshooterElfPowerIsYourCreatureCountAndShootsAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Friend A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Friend B", "Creature — Bear", 1, 1)
	flier := effectstest.BrCreature(g, opp.ID, "Their Bird", "Creature — Bird", 2, 3, "flying")
	walker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	elf := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Sharpshooter Elf", TypeLine: "Creature — Elf Ranger", OracleID: gapB217SharpshooterOracle, Power: 0, Toughness: 2, Keywords: []string{"reach"}})
	gapB217Pick(t, g, me.ID, flier)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, elf); p != 3 {
		t.Errorf("Sharpshooter Elf power is %d, want 3 (the creatures you control)", p)
	}
	if g.Battlefield.Contains(flier) {
		t.Error("the flier survived three damage on three toughness")
	}
	if !g.Battlefield.Contains(walker) {
		t.Error("a creature without flying was hit")
	}
}

// --- Silvanus's Invoker ---

func TestGapB217SilvanussInvokerAnimatesAnUntappedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	invoker := effectstest.PushCatalogPermanent(g, me.ID, "Silvanus's Invoker", "Creature — Dragon Druid", gapB217InvokerOracle, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: invoker}}}); err == nil {
		t.Fatal("a creature that is not a land was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := gapB217Read(g, land)
	if c.Tapped {
		t.Error("the land was not untapped")
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 8 || tough != 8 {
		t.Errorf("land is %d/%d, want 8/8", p, tough)
	}
	if !c.HasCardType("land") || !c.HasCardType("creature") || !c.HasSubtype("Elemental") {
		t.Errorf("land types %v / %v, want a land and an Elemental creature", c.Effective().Types, c.Effective().Subtypes)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "trample") || !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land lacks trample or haste")
	}
}

// --- Skanos Dragonheart ---

func TestGapB217SkanosGetsTheGreatestOtherDragonPowerWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skanos := effectstest.B12Push(g, me.ID, "Skanos Dragonheart", "Legendary Creature — Dragon Ranger", gapB217SkanosOracle, 4, 4)
	effectstest.B12Creature(g, me.ID, "Small Dragon", "Creature — Dragon", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 9, 9)
	effectstest.B12Creature(g, me.ID, "Not A Dragon", "Creature — Bear", 7, 7)
	gy := effectstest.PushGraveyardPermanent(me, "Dead Dragon", "Creature — Dragon", "{5}{R}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == gy {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 6, 6
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, skanos)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, skanos); p != 10 || tough != 10 {
		t.Errorf("Skanos is %d/%d, want 10/10 (4 plus the graveyard Dragon's 6)", p, tough)
	}
}

func TestGapB217SkanosWithNoOtherDragonGetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skanos := effectstest.B12Push(g, me.ID, "Skanos Dragonheart", "Legendary Creature — Dragon Ranger", gapB217SkanosOracle, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, skanos)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, skanos); p != 4 || tough != 4 {
		t.Errorf("Skanos is %d/%d, want 4/4", p, tough)
	}
}

// --- Sweet-Gum Recluse ---

func TestGapB217SweetGumRecluseCountersCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.TurnTally.Entered = map[uuid.UUID]int{} })
	fresh := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Fresh Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	recluse := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Sweet-Gum Recluse", TypeLine: "Creature — Spider", OracleID: gapB217SweetGumOracle, Power: 0, Toughness: 3, Keywords: []string{"flash", "reach"}})
	gapB217Pick(t, g, me.ID, fresh, recluse)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, fresh, game.CounterPlusOne); got != 3 {
		t.Errorf("the creature that entered this turn has %d counters, want 3", got)
	}
	if got := effectstest.CounterOn(g, recluse, game.CounterPlusOne); got != 3 {
		t.Errorf("the Recluse has %d counters, want 3", got)
	}
	if got := effectstest.CounterOn(g, old, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature that did not enter this turn has %d counters", got)
	}
}

// --- Tlincalli Hunter // Retrieve Prey ---

func TestGapB217TlincalliHunterOffersAFreeCastOfACreatureFromExileOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainForTest(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tlincalli Hunter", "Creature — Scorpion Scout", gapB217TlincalliOracle, false)
	key := effects.GapB217GrantedAltCostZeroCreatureFromExile
	cast := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{3}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() {
			if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
				t.Fatalf("exile: %v", err)
			}
		})
		return id
	}
	sorcery := cast("Exiled Sorcery", "Sorcery")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, sorcery, game.ZoneExile), key) != nil {
		t.Fatal("the offer reached a noncreature spell")
	}
	first := cast("Exiled Bear", "Creature — Bear")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, first, game.ZoneExile), key) == nil {
		t.Fatal("no free cast offered for a creature spell cast from exile")
	}
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID})
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, inHand, game.ZoneHand), key) != nil {
		t.Error("the offer reached a creature cast from hand")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	second := cast("Second Bear", "Creature — Bear")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, second, game.ZoneExile), key) != nil {
		t.Error("a second free cast was offered in the same turn")
	}
}

func TestGapB217RetrievePreyExilesACreatureCardFromYourGraveyardWithACastPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	rock := effectstest.PushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	legal := effectstest.TaLegalTargets(g, me.ID, gapB217TlincalliOracle+"#1")
	if !legal[mine] || legal[rock] || legal[theirs] {
		t.Fatalf("legal targets %v: want only my creature card", legal)
	}
	effectstest.CastCatalogSpell(t, g, "Retrieve Prey", "Sorcery — Adventure", gapB217TlincalliOracle+"#1", effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) || me.Graveyard.Contains(mine) {
		t.Fatal("the creature card was not exiled from the graveyard")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, mine, game.ZoneExile) == nil {
		t.Error("no permission to cast the exiled card")
	}
}
