package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b014_cards_test.go — Secret Lair Drop, "Partner" (S82 OLD batch 14,
// #5462): each card driven through a real cast, activation or attack.

const (
	oldB014SophinaOracle  = "095dabda-4570-4764-8d62-4a29eb19d3e6"
	oldB014EllieBMOracle  = "299ff264-c6db-4b48-88fb-ca006cb84f88"
	oldB014OthelmOracle   = "4ee401bf-dc86-41a9-84f9-57aa0a314d08"
	oldB014ElmarOracle    = "5a35201d-e9b1-433a-bb70-3280ceedbcd6"
	oldB014JoelOracle     = "64611260-c9f7-4375-b67a-f88fee6d3bff"
	oldB014WernogOracle   = "6cd03270-54cc-43e1-9b86-70c76960c841"
	oldB014HargildeOracle = "747009e7-da07-48c7-b838-687f19b884ea"
	oldB014TeveshOracle   = "79a1783f-ea7e-4cdf-95f5-53b383d38c66"
	oldB014KratosOracle   = "c000d5c3-5b0c-4c5c-9919-76640da0a1e4"
	oldB014BjornaOracle   = "c880fbdc-bdd9-4f80-81d4-e3e1124f76ca"
	oldB014EllieVHOracle  = "e619cb66-8c0d-408a-829c-6f02bf25fd0f"
	oldB014AtreusOracle   = "f6a4298c-3d9d-43e1-b32c-b168c02ad503"
)

func TestS82OldB014AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB014SophinaOracle, oldB014EllieBMOracle, oldB014OthelmOracle, oldB014ElmarOracle,
		oldB014JoelOracle, oldB014WernogOracle, oldB014HargildeOracle, oldB014TeveshOracle, oldB014KratosOracle,
		oldB014BjornaOracle, oldB014EllieVHOracle, oldB014AtreusOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Sophina, Spearsage Deserter ---------------------------------------------

func TestS82OldB014SophinaInvestigatesForEachNontokenAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sophina := effectstest.B12Push(g, me.ID, "Sophina, Spearsage Deserter", "Legendary Creature — Human Soldier", oldB014SophinaOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tok := effectstest.B12Creature(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, sophina, bear, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Clue"); got != 2 {
		t.Errorf("Clues = %d, want 2 (Sophina and the Bear, not the token)", got)
	}
}

// --- Ellie, Brick Master -----------------------------------------------------

func TestS82OldB014EllieBrickMasterGivesTheAttackerATappedAttackingToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ellie, Brick Master", "Legendary Creature — Human Survivor", oldB014EllieBMOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear2", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, bear2)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Cordyceps Infected" {
			continue
		}
		n++
		if c.Controller != me.ID || !c.Tapped || c.AttackingTarget != opp.ID {
			t.Errorf("token: controller %v tapped %v attacking %v, want mine, tapped, attacking the opponent", c.Controller, c.Tapped, c.AttackingTarget)
		}
	}
	if n != 1 {
		t.Errorf("tokens = %d, want 1 (two attackers at one opponent is one trigger)", n)
	}
}

// --- Othelm, Sigardian Outcast -----------------------------------------------

func TestS82OldB014OthelmReturnsACreatureThatDiedThisTurnTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	othelm := effectstest.B12Push(g, me.ID, "Othelm, Sigardian Outcast", "Legendary Creature — Human", oldB014OthelmOracle, 2, 2)
	effectstest.UntapForTest(g, othelm)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == othelm {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	died := effectstest.B12Creature(g, me.ID, "Fallen", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, died)
	old := effectstest.GraveyardCardOf(g, me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, othelm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(old)}); err == nil {
		t.Fatal("targeted a creature card that did not die this turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, othelm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(died)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, died)
	if !ok || !c.Tapped {
		t.Errorf("the creature came back %v tapped %v, want tapped on the battlefield", ok, c.Tapped)
	}
}

// --- Elmar, Ulvenwald Informant ----------------------------------------------

func TestS82OldB014ElmarUntapsAndInvestigatesOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Elmar, Ulvenwald Informant", "Legendary Creature — Human", oldB014ElmarOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "First Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Clue"); got != 0 {
		t.Fatalf("the first spell investigated: %d Clues", got)
	}
	effectstest.CastCatalogSpell(t, g, "Second Spell", "Sorcery", "", nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, bear); c.Tapped {
		t.Error("the target creature stayed tapped")
	}
	if got := effectstest.B43TokensNamed(g, me.ID, "Clue"); got != 1 {
		t.Errorf("Clues = %d, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Third Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Clue"); got != 1 {
		t.Errorf("a third spell investigated: %d Clues", got)
	}
}

// --- Joel, Resolute Survivor -------------------------------------------------

func TestS82OldB014JoelGrowsAndDrawsOnceEachTurnForATokenDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	joel := effectstest.B12Push(g, me.ID, "Joel, Resolute Survivor", "Legendary Creature — Human Survivor", oldB014JoelOracle, 4, 4)
	effectstest.AuraSeedLibrary(me, 5)
	hand := effectstest.HandCount(g, me)
	nontoken := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, nontoken)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand {
		t.Fatalf("a nontoken creature dying drew: hand %d, want %d", got, hand)
	}
	tok1 := effectstest.B12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, tok1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand = %d, want %d after a token died", got, hand+1)
	}
	if c, _ := effectstest.BattlefieldCardByID(g, joel); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Joel has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	tok2 := effectstest.B12Creature(g, me.ID, "Goblin2", "Token Creature — Goblin", 1, 1)
	effectstest.KillCreature(t, g, me.ID, tok2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("a second token this turn drew again: hand %d, want %d", got, hand+1)
	}
}

// --- Wernog, Rider's Chaplain ------------------------------------------------

func TestS82OldB014WernogAsksEachOpponentThenInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Wernog, Rider's Chaplain", "Legendary Creature — Human", oldB014WernogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	answers := []bool{true, false, true}
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	for i := 1; i < len(g.Seats); i++ {
		effectstest.AnswerMayChoice(t, g, g.Seats[i].ID, answers[i-1])
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		wantClue, wantLife := 0, lives[i]
		if answers[i-1] {
			wantClue = 1
		} else {
			wantLife--
		}
		if got := effectstest.B43TokensNamed(g, g.Seats[i].ID, "Clue"); got != wantClue {
			t.Errorf("opponent %d has %d Clues, want %d", i, got, wantClue)
		}
		if g.Seats[i].Life != wantLife {
			t.Errorf("opponent %d life %d, want %d", i, g.Seats[i].Life, wantLife)
		}
	}
	if got := effectstest.B43TokensNamed(g, me.ID, "Clue"); got != 3 {
		t.Errorf("my Clues = %d, want 3 (one plus the two opponents who investigated)", got)
	}
}

// --- Hargilde, Kindly Runechanter --------------------------------------------

func TestS82OldB014HargildeAddsTwoColourlessForArtifactsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Hargilde, Kindly Runechanter", "Legendary Creature — Human", oldB014HargildeOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.UntapForTest(g, h)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == h {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, h, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Hargilde: %v", err)
	}
	cost, _ := game.ParseCost("{2}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the two mana refused an artifact spell or an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a nonartifact spell or ability")
	}
}

// --- Tevesh Szat, Doom of Fools ----------------------------------------------

func oldB014Tevesh(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Tevesh Szat, Doom of Fools", "Legendary Planeswalker — Szat", oldB014TeveshOracle, 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 4}
			}
		}
	})
	return id
}

func TestS82OldB014TeveshPlusTwoMakesTwoThrulls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tev := oldB014Tevesh(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tev, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Thrull"); got != 2 {
		t.Fatalf("Thrulls = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thrull" && (c.Power != 0 || c.Toughness != 1 || !c.HasColor("B")) {
			t.Errorf("Thrull is %d/%d %v, want a black 0/1", c.Power, c.Toughness, c.Colors)
		}
	}
	if c, _ := effectstest.BattlefieldCardByID(g, tev); c.Counters[game.CounterLoyalty] != 6 {
		t.Errorf("loyalty %d, want 6", c.Counters[game.CounterLoyalty])
	}
}

func TestS82OldB014TeveshPlusOneSacrificesForCardsAndOneMoreForACommander(t *testing.T) {
	for _, tc := range []struct {
		name      string
		commander bool
		want      int
	}{{"an ordinary creature", false, 2}, {"a commander", true, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			tev := oldB014Tevesh(t, g)
			effectstest.AuraSeedLibrary(me, 6)
			victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
			if tc.commander {
				g.WithWriteLock(func() {
					for i := range g.Battlefield.Cards {
						if g.Battlefield.Cards[i].InstanceID == victim {
							g.Battlefield.Cards[i].IsCommander = true
						}
					}
				})
			}
			hand := effectstest.HandCount(g, me)
			if err := g.ActivateCatalogAbility(me.ID, tev, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("+1: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerOwnPermanents(t, g, me.ID, victim)
			if tc.commander {
				effectstest.AnswerCommanderReturn(t, g, me.ID, false)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.HandCount(g, me); got != hand+tc.want {
				t.Errorf("hand = %d, want %d", got, hand+tc.want)
			}
			if g.Battlefield.Contains(victim) && !tc.commander {
				t.Error("the creature was not sacrificed")
			}
		})
	}
}

func TestS82OldB014TeveshPlusOneMayDeclineToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tev := oldB014Tevesh(t, g)
	effectstest.AuraSeedLibrary(me, 6)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	hand := effectstest.HandCount(g, me)
	if err := g.ActivateCatalogAbility(me.ID, tev, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) || effectstest.HandCount(g, me) != hand {
		t.Error("declining still sacrificed or drew")
	}
}

func TestS82OldB014TeveshMinusTenTakesEveryCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tev := oldB014Tevesh(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tev {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 10}
			}
		}
	})
	onField := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Commander",
		TypeLine: "Legendary Creature — Human", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true})
	notCommander := effectstest.B12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	inZone := uuid.New()
	g.WithWriteLock(func() {
		opp.Command.PushTop(game.Card{InstanceID: inZone, Name: "Sent Home Commander", TypeLine: "Legendary Creature — Human",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, IsCommander: true})
	})
	if err := g.ActivateCatalogAbility(me.ID, tev, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for name, id := range map[string]uuid.UUID{"the commander on the battlefield": onField, "the commander from the command zone": inZone} {
		c, ok := effectstest.BattlefieldCardByID(g, id)
		if !ok || effectstest.ApaLive(g, id).Controller != me.ID {
			t.Errorf("%s is not under my control (on battlefield %v, controller %v)", name, ok, c.Controller)
		}
	}
	if effectstest.ApaLive(g, notCommander).Controller != opp.ID {
		t.Error("a non-commander creature was taken")
	}
}

// --- Kratos, Stoic Father ----------------------------------------------------

func oldB014Experience(g *game.Game, p *game.Player) int {
	return effects.ExperienceCounters(g, p.ID)
}

func TestS82OldB014KratosAttackWithGodsIsOneCounterAndEachGodDeathIsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kratos, Stoic Father", "Legendary Creature — God Warrior", oldB014KratosOracle, 4, 4)
	god1 := effectstest.B12Creature(g, me.ID, "God One", "Creature — God", 2, 2)
	god2 := effectstest.B12Creature(g, opp.ID, "God Two", "Creature — God", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, god1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB014Experience(g, me); got != 1 {
		t.Fatalf("attacking with a God gave %d experience counters, want 1", got)
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB014Experience(g, me); got != 1 {
		t.Errorf("a Bear dying gave counters: %d", got)
	}
	effectstest.KillCreature(t, g, opp.ID, god2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB014Experience(g, me); got != 2 {
		t.Errorf("an opponent's God dying gave %d counters total, want 2", got)
	}
}

func TestS82OldB014KratosEndStepPutsCountersEqualToExperience(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kratos, Stoic Father", "Legendary Creature — God Warrior", oldB014KratosOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterExperience, 3); err != nil {
			t.Fatalf("experience: %v", err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, bear); c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("the Bear has %d +1/+1 counters, want 3", c.Counters[game.CounterPlusOne])
	}
}

// --- Bjorna, Nightfall Alchemist ---------------------------------------------

func TestS82OldB014BjornaSacrificesAnArtifactToDamageAndGoad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bj := effectstest.B12Push(g, me.ID, "Bjorna, Nightfall Alchemist", "Legendary Creature — Human", oldB014BjornaOracle, 1, 3)
	effectstest.UntapForTest(g, bj)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bj {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	target := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, bj, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(target), SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	c, _ := effectstest.BattlefieldCardByID(g, target)
	if c.DamageMarked != 1 {
		t.Errorf("damage = %d, want 1", c.DamageMarked)
	}
	if gc, _ := effectstest.BattlefieldCardByID(g, target); !gc.IsGoadedBy(me.ID) {
		t.Error("the creature was not goaded")
	}
}

// --- Ellie, Vengeful Hunter --------------------------------------------------

func TestS82OldB014EllieVengefulHunterPaysLifeAndASacrificeToBurnAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ellie := effectstest.B12Push(g, me.ID, "Ellie, Vengeful Hunter", "Legendary Creature — Human Survivor", oldB014EllieVHOracle, 3, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	life, theirs := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, ellie, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{ellie}}); err == nil {
		t.Fatal("sacrificed Ellie to herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, ellie, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 || opp.Life != theirs-2 {
		t.Errorf("life me %d (want %d), opponent %d (want %d)", me.Life, life-2, opp.Life, theirs-2)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, ellie, "indestructible") {
		t.Error("Ellie did not gain indestructible")
	}
}

// --- Atreus, Impulsive Son ---------------------------------------------------

func TestS82OldB014AtreusDrawsPerExperienceCounterThenDiscardsAndBurnsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	atreus := effectstest.B12Push(g, me.ID, "Atreus, Impulsive Son", "Legendary Creature — God Archer", oldB014AtreusOracle, 2, 4)
	effectstest.UntapForTest(g, atreus)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == atreus {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
		if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterExperience, 2); err != nil {
			t.Fatalf("experience: %v", err)
		}
	})
	effectstest.AuraSeedLibrary(me, 6)
	hand := effectstest.HandCount(g, me)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, atreus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt after the draws")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+2-1 {
		t.Errorf("hand = %d, want %d (two drawn, one discarded)", got, hand+1)
	}
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Life != lives[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, g.Seats[i].Life, lives[i]-2)
		}
	}
	if g.Seats[0].Life != lives[0] {
		t.Errorf("Atreus's controller lost life: %d, want %d", g.Seats[0].Life, lives[0])
	}
}
