package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b118_test.go — Ikoria: Lair of Behemoths, batch 118 (#5512):
// pump and combat tricks, creatures with keywords. Each test plays the real
// catalog entry and asserts what its text says.

const (
	oldB118FightAsOne       = "16f93781-6740-40c7-a727-7911beac4e74"
	oldB118SanctuaryLockdwn = "66793ddf-aca3-44cb-8742-4eee178f31c6"
	oldB118PerimeterSgt     = "6e1b21f9-c890-4d27-81d0-d38bf7493d33"
	oldB118LabyrinthRaptor  = "a29b2e93-d872-4be8-8b15-57f57282ed45"
	oldB118ZagothMamba      = "aad05e4e-28cf-4dfe-bf5b-277a50d183ba"
	oldB118Zilortha         = "2dba8d8d-891e-482e-8014-879f701d0619"
	oldB118Howlbonder       = "89645420-1a14-46bf-b27d-966553a95463"
	oldB118Felidar          = "bb504fb1-1d19-4d0b-958f-e8f2433e0575"
	oldB118ProudWildbonder  = "e72f995e-288f-4d32-b067-ae73eea87e67"
	oldB118GlowstoneRecluse = "b2b76070-f103-4f83-ad21-119e6352f2cd"

	oldB118HumanType = "Creature — Human Soldier"
	oldB118BearType  = "Creature — Bear"
)

// oldB118Live is a battlefield permanent with the layers caught up.
func oldB118Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return *c
}

func TestOldB118EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB118FightAsOne:       "Fight as One",
		oldB118SanctuaryLockdwn: "Sanctuary Lockdown",
		oldB118PerimeterSgt:     "Perimeter Sergeant",
		oldB118LabyrinthRaptor:  "Labyrinth Raptor",
		oldB118ZagothMamba:      "Zagoth Mamba",
		oldB118Zilortha:         "Zilortha, Strength Incarnate",
		oldB118Howlbonder:       "Sonorous Howlbonder",
		oldB118Felidar:          "Frondland Felidar",
		oldB118ProudWildbonder:  "Proud Wildbonder",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Fight as One --------------------------------------------------------

func TestOldB118FightAsOneBothModesPumpEachCreatureAndMakeItIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	effectstest.P1g3CastModal(t, g, "Fight as One", "Instant", oldB118FightAsOne, []int{0, 1},
		effectstest.P1g3ModeCard(human, 0), effectstest.P1g3ModeCard(bear, 1))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Human": human, "non-Human": bear} {
		c := oldB118Live(t, g, id)
		if c.PowerForComparison() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("%s is %d/%d, want 3/3", name, c.PowerForComparison(), c.CurrentToughness())
		}
		if !game.HasKeyword(&c, "indestructible") {
			t.Errorf("%s did not gain indestructible", name)
		}
	}
}

func TestOldB118FightAsOneOneModeLeavesTheOtherCreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	effectstest.P1g3CastModal(t, g, "Fight as One", "Instant", oldB118FightAsOne, []int{1}, effectstest.P1g3ModeCard(bear, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if c := oldB118Live(t, g, bear); c.PowerForComparison() != 3 || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("the chosen non-Human is %d power, indestructible %v", c.PowerForComparison(), game.HasKeyword(&c, "indestructible"))
	}
	if c := oldB118Live(t, g, human); c.PowerForComparison() != 2 || game.HasKeyword(&c, "indestructible") {
		t.Errorf("the Human was not chosen but is %d power, indestructible %v", c.PowerForComparison(), game.HasKeyword(&c, "indestructible"))
	}
}

func TestOldB118FightAsOneRefusesTheWrongKindOfCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB118BearType, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fight as One", TypeLine: "Instant", OracleID: oldB118FightAsOne, ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	for name, tc := range map[string]struct {
		mode   int
		target uuid.UUID
	}{
		"a non-Human for the Human mode": {0, bear},
		"a Human for the non-Human mode": {1, human},
		"an opponent's creature":         {1, theirs},
	} {
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{tc.mode}, Targets: []game.TargetRef{effectstest.P1g3ModeCard(tc.target, 0)}})
		if err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

// --- Sanctuary Lockdown --------------------------------------------------

func TestOldB118SanctuaryLockdownPumpsOnlyYourHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", oldB118HumanType, 2, 2)
	effectstest.B12Push(g, me.ID, "Sanctuary Lockdown", "Enchantment", oldB118SanctuaryLockdwn, 0, 0)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"your Human": {mine, 3}, "your Bear": {bear, 2}, "their Human": {theirs, 2}} {
		if c := oldB118Live(t, g, tc.id); c.PowerForComparison() != tc.want || c.CurrentToughness() != tc.want {
			t.Errorf("%s is %d/%d, want %d/%d", name, c.PowerForComparison(), c.CurrentToughness(), tc.want, tc.want)
		}
	}
}

func TestOldB118SanctuaryLockdownTapsAnOpposingCreatureForTwoHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lockdown := effectstest.B12Push(g, me.ID, "Sanctuary Lockdown", "Enchantment", oldB118SanctuaryLockdwn, 0, 0)
	h1 := effectstest.B12Creature(g, me.ID, "Soldier A", oldB118HumanType, 2, 2)
	h2 := effectstest.B12Creature(g, me.ID, "Soldier B", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB118BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	// A non-Human can't pay, one Human is not enough, your own creature
	// is not a legal target, and the cost needs {2}.
	for name, params := range map[string]game.ActivateAbilityParams{
		"a Bear paying":       {Targets: target, TapIDs: []uuid.UUID{h1, bear}, Strict: true},
		"only one Human":      {Targets: target, TapIDs: []uuid.UUID{h1}, Strict: true},
		"a creature of yours": {Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, TapIDs: []uuid.UUID{h1, h2}, Strict: true},
		"an empty pool":       {Targets: target, TapIDs: []uuid.UUID{h1, h2}, Strict: true},
	} {
		if err := g.ActivateCatalogAbility(me.ID, lockdown, 0, params); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, lockdown, 0, game.ActivateAbilityParams{Targets: target, TapIDs: []uuid.UUID{h1, h2}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, victim) {
		t.Error("the opposing creature was not tapped")
	}
	if !effectstest.B13Tapped(t, g, h1) || !effectstest.B13Tapped(t, g, h2) {
		t.Error("the two Humans were not tapped as the cost")
	}
	if effectstest.B13Tapped(t, g, bear) {
		t.Error("the Bear was tapped")
	}
}

// --- Perimeter Sergeant --------------------------------------------------

func TestOldB118PerimeterSergeantPumpsOtherHumansWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sgt := effectstest.B12Push(g, me.ID, "Perimeter Sergeant", "Creature — Human Soldier", oldB118PerimeterSgt, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB118BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", oldB118HumanType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, sgt)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"the Sergeant": {sgt, 3}, "another Human": {other, 3}, "a Bear": {bear, 2}, "their Human": {theirs, 2}} {
		if c := oldB118Live(t, g, tc.id); c.PowerForComparison() != tc.want || c.CurrentToughness() != 2 {
			t.Errorf("%s is %d/%d, want power %d and toughness 2", name, c.PowerForComparison(), c.CurrentToughness(), tc.want)
		}
	}
}

func TestOldB118PerimeterSergeantDoesNothingWhenAnotherHumanAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Perimeter Sergeant", "Creature — Human Soldier", oldB118PerimeterSgt, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Soldier", oldB118HumanType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, other)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c := oldB118Live(t, g, other); c.PowerForComparison() != 2 {
		t.Errorf("the attacking Human is %d power, want 2", c.PowerForComparison())
	}
}

// --- Zagoth Mamba --------------------------------------------------------

func TestOldB118ZagothMambaShrinksAnOpposingCreatureWhenItMutates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mamba := effectstest.B12Push(g, me.ID, "Zagoth Mamba", "Creature — Nightmare Snake", oldB118ZagothMamba, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB118BearType, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB118BearType, 3, 3)
	recluse := effectstest.RprMutatorInHand(t, g, me.ID, game.Card{
		Name: "Glowstone Recluse", TypeLine: "Creature — Spider", OracleID: oldB118GlowstoneRecluse,
		ManaCost: "{2}{G}", Colors: []string{"G"}, Keywords: []string{"reach", "mutate"}, Power: 2, Toughness: 3,
	})
	if err := effectstest.CastMutating(g, me.ID, recluse, mamba, false); err != nil {
		t.Fatalf("mutate onto the mamba: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	for i := 0; i < 8; i++ {
		switch {
		case effectstest.LatestPickTarget(g, me.ID) != nil:
			effectstest.PickTriggerTarget(t, g, me.ID, victim)
		case effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID):
		case !effectstest.StackFullyEmpty(g):
			effectstest.PassPriorityAroundTable(t, g)
		default:
			i = 8
		}
	}
	if c := oldB118Live(t, g, victim); c.PowerForComparison() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the target is %d/%d, want 1/1", c.PowerForComparison(), c.CurrentToughness())
	}
	if c := oldB118Live(t, g, mine); c.PowerForComparison() != 3 {
		t.Errorf("your own creature is %d power, want 3", c.PowerForComparison())
	}
}

// --- Sonorous Howlbonder -------------------------------------------------

func TestOldB118SonorousHowlbonderNeedsThreeBlockersForEachMenaceCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sonorous Howlbonder", "Creature — Human Warrior", oldB118Howlbonder, 2, 2)
	menace := effectstest.BrCreature(g, me.ID, "Menace Bear", oldB118BearType, 2, 2, "menace")
	plain := effectstest.BrCreature(g, me.ID, "Plain Bear", oldB118BearType, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", oldB118BearType, 2, 2)
	effectstest.BrAttack(t, g, menace, plain)

	// The creature without menace is blocked as ever.
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: plain}}); err != nil {
		t.Fatalf("a lone blocker on the creature without menace: %v", err)
	}
}

func TestOldB118SonorousHowlbonderRefusesTwoBlockersOnAMenaceCreatureButAllowsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sonorous Howlbonder", "Creature — Human Warrior", oldB118Howlbonder, 2, 2)
	menace := effectstest.BrCreature(g, me.ID, "Menace Bear", oldB118BearType, 2, 2, "menace")
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", oldB118BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", oldB118BearType, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Blocker C", oldB118BearType, 2, 2)
	effectstest.BrAttack(t, g, menace)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace},
	}), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace}, {Blocker: c, Attacker: menace},
	}); err != nil {
		t.Fatalf("three blockers are legal: %v", err)
	}
}

func TestOldB118SonorousHowlbonderDoesNotBindAnOpponentsMenaceCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Sonorous Howlbonder", "Creature — Human Warrior", oldB118Howlbonder, 2, 2)
	menace := effectstest.BrCreature(g, me.ID, "Menace Bear", oldB118BearType, 2, 2, "menace")
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", oldB118BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", oldB118BearType, 2, 2)
	effectstest.BrAttack(t, g, menace)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace},
	}); err != nil {
		t.Fatalf("the Howlbonder is not yours to bind them: %v", err)
	}
}

// --- Proud Wildbonder ----------------------------------------------------

func TestOldB118ProudWildbonderGivesOnlyYourTramplersTheUnblockedOption(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wild := effectstest.B12Push(g, me.ID, "Proud Wildbonder", "Creature — Human Warrior", oldB118ProudWildbonder, 4, 3)
	trampler := effectstest.BrCreature(g, me.ID, "Trampling Bear", oldB118BearType, 2, 2, "trample")
	plain := effectstest.BrCreature(g, me.ID, "Plain Bear", oldB118BearType, 2, 2)
	theirs := effectstest.BrCreature(g, opp.ID, "Their Trampler", oldB118BearType, 2, 2, "trample")
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"the Wildbonder": {wild, true}, "your trampler": {trampler, true}, "your plain creature": {plain, false}, "their trampler": {theirs, false}} {
		if got := oldB118Live(t, g, tc.id).Effective().AssignsAsThoughUnblocked; got != tc.want {
			t.Errorf("%s: may assign as though unblocked is %v, want %v", name, got, tc.want)
		}
	}
}

// --- Frondland Felidar ---------------------------------------------------

func TestOldB118FrondlandFelidarGivesVigilantCreaturesATapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	felidar := effectstest.B12Push(g, me.ID, "Frondland Felidar", "Creature — Cat Beast", oldB118Felidar, 3, 5)
	vigilant := effectstest.BrCreature(g, me.ID, "Vigilant Bear", oldB118BearType, 2, 2, "vigilance")
	plain := effectstest.BrCreature(g, me.ID, "Plain Bear", oldB118BearType, 2, 2)
	theirs := effectstest.BrCreature(g, opp.ID, "Their Vigilant Bear", oldB118BearType, 2, 2, "vigilance")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB118BearType, 2, 2)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, plain); idx >= 0 {
		t.Error("a creature without vigilance has the granted ability")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, theirs); idx >= 0 {
		t.Error("an opponent's vigilant creature has the granted ability")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, felidar); idx < 0 {
		t.Error("the Felidar itself, which has vigilance, lacks the granted ability")
	}
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, vigilant)
	if idx < 0 {
		t.Fatal("your vigilant creature has no granted ability")
	}
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	if err := g.ActivateCatalogAbility(me.ID, vigilant, idx, game.ActivateAbilityParams{Ref: ref, Targets: target, Strict: true}); err == nil {
		t.Fatal("activated {1} with an empty pool")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, vigilant, idx, game.ActivateAbilityParams{Ref: ref, Targets: target, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, victim) {
		t.Error("the target creature was not tapped")
	}
	if !effectstest.B13Tapped(t, g, vigilant) {
		t.Error("the activating creature did not tap for the cost")
	}
}

// --- Zilortha, Strength Incarnate ----------------------------------------

func TestOldB118ZilorthaMakesPowerTheLethalThresholdForYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Zilortha, Strength Incarnate", "Legendary Creature — Dinosaur", oldB118Zilortha, 7, 3)
	small := effectstest.B12Creature(g, me.ID, "Mine 2/6", oldB118BearType, 2, 6)
	big := effectstest.B12Creature(g, me.ID, "Mine 5/2", oldB118BearType, 5, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs 2/6", oldB118BearType, 2, 6)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", oldB118BearType, 1, 1)

	effectstest.Pr6Damage(t, g, src, small, 2)
	if effectstest.FindBattlefieldCardForTest(g, small) != nil {
		t.Error("a 2/6 under Zilortha survived 2 damage; lethal is its power")
	}
	effectstest.Pr6Damage(t, g, src, big, 2)
	if effectstest.FindBattlefieldCardForTest(g, big) == nil {
		t.Error("a 5/2 under Zilortha died to 2 damage; lethal is its power, 5")
	}
	effectstest.Pr6Damage(t, g, src, big, 3)
	if effectstest.FindBattlefieldCardForTest(g, big) != nil {
		t.Error("a 5/2 under Zilortha survived 5 damage")
	}
	effectstest.Pr6Damage(t, g, src, theirs, 2)
	if effectstest.FindBattlefieldCardForTest(g, theirs) == nil {
		t.Error("an opponent's 2/6 died to 2 damage; Zilortha is not theirs")
	}
}

func TestOldB118ZilorthaEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zil := effectstest.B12Push(g, me.ID, "Zilortha, Strength Incarnate", "Legendary Creature — Dinosaur", oldB118Zilortha, 7, 3)
	small := effectstest.B12Creature(g, me.ID, "Mine 2/6", oldB118BearType, 2, 6)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", oldB118BearType, 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(zil); err != nil {
			t.Fatalf("destroy Zilortha: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.Pr6Damage(t, g, src, small, 2)
	if effectstest.FindBattlefieldCardForTest(g, small) == nil {
		t.Error("a 2/6 died to 2 damage with no Zilortha")
	}
}

func TestOldB118ZilorthaChangesHowMuchTrampleDamageIsLethal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Zilortha, Strength Incarnate", "Legendary Creature — Dinosaur", oldB118Zilortha, 7, 3)
	trampler := effectstest.BrCreature(g, me.ID, "Trampler", oldB118BearType, 5, 5, "trample")
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker 2/6", oldB118BearType, 2, 6)
	before := opp.Life
	effectstest.BrAttack(t, g, trampler)
	if err := g.DeclareBlocker(blocker, trampler); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	// Lethal to the 2/6 is 2, so 3 of the Trampler's 5 damage goes to the
	// player; the blocker (2 marked, power 2) is destroyed.
	if got := before - opp.Life; got != 3 {
		t.Errorf("the player lost %d life, want 3 (5 power less 2 lethal)", got)
	}
	if effectstest.FindBattlefieldCardForTest(g, blocker) != nil {
		t.Error("the blocker survived its lethal damage")
	}
}

// --- Labyrinth Raptor ----------------------------------------------------

func TestOldB118LabyrinthRaptorMakesTheDefenderSacrificeABlockerOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Labyrinth Raptor", "Creature — Nightmare Dinosaur", oldB118LabyrinthRaptor, 2, 2)
	menace := effectstest.BrCreature(g, me.ID, "Menace Bear", oldB118BearType, 2, 2, "menace")
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", oldB118BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", oldB118BearType, 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB118BearType, 2, 2)
	effectstest.BrAttack(t, g, menace)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace}}); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if p == nil {
		t.Fatal("the defending player was not asked which blocker to sacrifice")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		offered[id] = true
	}
	if !offered[a] || !offered[b] || offered[bystander] || len(offered) != 2 {
		t.Errorf("offered %v, want exactly the two blockers", offered)
	}
	if err := g.ResolveChooseCards(p.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, b) != nil || !opp.Graveyard.Contains(b) {
		t.Error("the chosen blocker was not sacrificed")
	}
	if effectstest.FindBattlefieldCardForTest(g, a) == nil || effectstest.FindBattlefieldCardForTest(g, bystander) == nil {
		t.Error("another creature was sacrificed")
	}
}

func TestOldB118LabyrinthRaptorDoesNothingForACreatureWithoutMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Labyrinth Raptor", "Creature — Nightmare Dinosaur", oldB118LabyrinthRaptor, 2, 2)
	plain := effectstest.BrCreature(g, me.ID, "Plain Bear", oldB118BearType, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", oldB118BearType, 2, 2)
	effectstest.BrAttack(t, g, plain)
	if err := g.DeclareBlocker(a, plain); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID); p != nil {
		t.Error("the defender was asked to sacrifice for a creature without menace")
	}
	if effectstest.FindBattlefieldCardForTest(g, a) == nil {
		t.Error("the blocker was sacrificed")
	}
}

func TestOldB118LabyrinthRaptorPumpsEveryMenaceCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptor := effectstest.B12Push(g, me.ID, "Labyrinth Raptor", "Creature — Nightmare Dinosaur", oldB118LabyrinthRaptor, 2, 2)
	menace := effectstest.BrCreature(g, me.ID, "Menace Bear", oldB118BearType, 2, 2, "menace")
	plain := effectstest.BrCreature(g, me.ID, "Plain Bear", oldB118BearType, 2, 2)
	theirs := effectstest.BrCreature(g, opp.ID, "Their Menace Bear", oldB118BearType, 2, 2, "menace")
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, raptor, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"the Raptor": {raptor, 3}, "your menace creature": {menace, 3}, "your plain creature": {plain, 2}, "their menace creature": {theirs, 2}} {
		if c := oldB118Live(t, g, tc.id); c.PowerForComparison() != tc.want {
			t.Errorf("%s has power %d, want %d", name, c.PowerForComparison(), tc.want)
		}
	}
}
