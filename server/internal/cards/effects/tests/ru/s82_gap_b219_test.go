package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b219_test.go — Commander Legends: Battle for Baldur's Gate red creatures,
// GAP batch 219 (#5948). Each test plays the real catalog entry.

const (
	gapB219QuasitOracle     = "06cf7b77-ffd7-4192-903e-19bd06887ba1"
	gapB219CavesOracle      = "e4d2dcbb-9910-482e-9e46-dfed2097509b"
	gapB219DeathKissOracle  = "42410e81-f4a6-4501-82e4-6e745934976c"
	gapB219ElturelOracle    = "373ccd23-5c98-4368-995a-e6503cd0dbd8"
	gapB219FangDragonOracle = "c075637a-e94c-4b8d-8a9e-20ee3dfa8fe7"
	gapB219FlutistOracle    = "5bf88cac-5997-4254-8349-2473de854c50"
	gapB219GenasiOracle     = "b3601821-a0f5-4aac-b24e-25498a82d821"
	gapB219GnollOracle      = "ff8fc718-6218-48f3-981d-c4d2bc468493"
	gapB219GutOracle        = "60e7130a-d321-4e50-93a4-999013c76054"
	gapB219BalladeerOracle  = "e6acafdf-e3ad-4ac2-bb26-cec37c7ca115"
)

func TestGapB219EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB219QuasitOracle:            "Bothersome Quasit",
		gapB219CavesOracle:             "Caves of Chaos Adventurer",
		gapB219DeathKissOracle:         "Death Kiss",
		gapB219ElturelOracle:           "Elturel Survivors",
		gapB219FangDragonOracle:        "Fang Dragon",
		gapB219FangDragonOracle + "#1": "Forktail Sweep",
		gapB219FlutistOracle:           "Firbolg Flutist",
		gapB219GenasiOracle:            "Genasi Enforcers",
		gapB219GnollOracle:             "Gnoll War Band",
		gapB219GutOracle:               "Gut, True Soul Zealot",
		gapB219BalladeerOracle:         "Insufferable Balladeer",
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

func gapB219Read(g *game.Game, id uuid.UUID) (c game.Card) {
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
	return c
}

func gapB219PickCard(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
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

// --- Bothersome Quasit ---

func TestGapB219QuasitGoadsOnANoncreatureSpellAndGoadedCreaturesCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Bothersome Quasit", "Creature — Demon", gapB219QuasitOracle, false)
	goaded := effectstest.B12Creature(g, opp.ID, "Goaded Bear", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free Bear", "Creature — Bear", 2, 2)
	spell := effectstest.HandSorcery(me, "Noncreature Spell", "{R}")
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	gapB219PickCard(t, g, me.ID, goaded)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, goaded) != me.ID {
		t.Fatal("the target was not goaded by the Quasit's controller")
	}
	attacker := effectstest.B12Creature(g, me.ID, "My Attacker", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(goaded, attacker); err == nil {
		t.Error("a goaded creature an opponent controls blocked")
	}
	if err := g.DeclareBlocker(free, attacker); err != nil {
		t.Errorf("a creature that is not goaded could not block: %v", err)
	}
}

func TestGapB219QuasitIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Bothersome Quasit", "Creature — Demon", gapB219QuasitOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	cast := game.Card{InstanceID: uuid.New(), Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(cast)
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, cast.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, bear) != uuid.Nil {
		t.Error("a creature spell goaded a creature")
	}
}

// --- Caves of Chaos Adventurer ---

func TestGapB219CavesTakesTheInitiativeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Caves of Chaos Adventurer", TypeLine: "Creature — Human Barbarian", OracleID: gapB219CavesOracle, Power: 5, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Initiative != me.ID {
		t.Error("its controller did not take the initiative")
	}
}

func TestGapB219CavesAttackExilesTheTopCardForAnOrdinaryPlayUnlessADungeonIsCompleted(t *testing.T) {
	for _, completed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		caves := effectstest.B12Push(g, me.ID, "Caves of Chaos Adventurer", "Creature — Human Barbarian", gapB219CavesOracle, 5, 3)
		if completed {
			me.CompletedDungeons = append(me.CompletedDungeons, game.UndercityKey)
		}
		top := uuid.New()
		g.WithWriteLock(func() {
			me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
		})
		effectstest.DeclareAttack(t, g, opp.ID, caves)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(top) {
			t.Fatalf("completed=%v: the top card was not exiled", completed)
		}
		perm := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile)
		if perm == nil {
			t.Fatalf("completed=%v: no permission to play the exiled card", completed)
		}
		if free := perm.Cost == "{0}"; free != completed {
			t.Errorf("completed=%v: the permission is free = %v", completed, free)
		}
	}
}

// --- Death Kiss ---

func TestGapB219DeathKissDoublesAnOpponentsAttackerAimedAtAnotherOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Death Kiss", "Creature — Beholder", gapB219DeathKissOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, other.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, theirs); p != 6 {
		t.Errorf("the attacker has power %d, want 6", p)
	}
}

func TestGapB219DeathKissDoesNotDoubleACreatureAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Death Kiss", "Creature — Beholder", gapB219DeathKissOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, theirs); p != 3 {
		t.Errorf("a creature attacking Death Kiss's controller has power %d, want 3", p)
	}
}

func TestGapB219DeathKissMonstrosityGoadsUpToXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kiss := effectstest.PushCatalogPermanent(g, me.ID, "Death Kiss", "Creature — Beholder", gapB219DeathKissOracle, false)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, kiss, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	gapB219PickCard(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, kiss, game.CounterPlusOne); got != 2 {
		t.Errorf("Death Kiss has %d +1/+1 counters, want 2", got)
	}
	if effectstest.B33Goaded(t, g, a) != me.ID || effectstest.B33Goaded(t, g, b) != me.ID {
		t.Error("the chosen creatures were not goaded")
	}
	if effectstest.B33Goaded(t, g, c) != uuid.Nil {
		t.Error("an unchosen creature was goaded")
	}
}

// --- Elturel Survivors ---

func TestGapB219ElturelSurvivorsGetsPlusXWhileAttackingForTheDefendersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	elturel := effectstest.B12Push(g, me.ID, "Elturel Survivors", "Creature — Tiefling Peasant", gapB219ElturelOracle, 0, 4)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	}
	for i := 0; i < 5; i++ {
		effectstest.PushCatalogPermanent(g, other.ID, "Forest", "Basic Land — Forest", "", false)
	}
	if p, _ := effectstest.PtOf(t, g, elturel); p != 0 {
		t.Fatalf("power %d before attacking, want 0", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, elturel)
	if p, _ := effectstest.PtOf(t, g, elturel); p != 3 {
		t.Errorf("power %d while attacking a player with three lands, want 3", p)
	}
}

// --- Fang Dragon // Forktail Sweep ---

func TestGapB219ForktailSweepDamagesEachCreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Forktail Sweep", "Sorcery — Adventure", gapB219FangDragonOracle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 1-toughness creature I don't control survived")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("a bigger opposing creature or my own creature died")
	}
}

// --- Firbolg Flutist ---

func TestGapB219FirbolgFlutistStealsUntapsAndGivesHasteAndMyriad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	stolen := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Firbolg Flutist", TypeLine: "Creature — Giant Bard", OracleID: gapB219FlutistOracle, Power: 4, Toughness: 4})
	gapB219PickCard(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)
	c := gapB219Read(g, stolen)
	if c.Controller != me.ID || c.Tapped {
		t.Fatalf("stolen creature controlled by %v, tapped %v; want me and untapped", c.Controller, c.Tapped)
	}
	if !effectstest.HasEffectiveKeyword(t, g, stolen, "haste") {
		t.Error("the stolen creature has no haste")
	}
	// It attacks one opponent: myriad makes a token copy attacking the other opponent.
	effectstest.DeclareAttack(t, g, opp.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if tokens := effectstest.B33AttackingTokensNamed(g, me.ID, other.ID, "Their Bear"); len(tokens) != 1 {
		t.Errorf("%d myriad token copies attack the other opponent, want 1", len(tokens))
	}
}

// --- Genasi Enforcers ---

func TestGapB219GenasiEnforcersPumpsOnlyCreaturesNamedGenasiEnforcers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.PushCatalogPermanent(g, me.ID, "Genasi Enforcers", "Creature — Elemental Shaman", gapB219GenasiOracle, false)
	second := effectstest.B12Push(g, me.ID, "Genasi Enforcers", "Creature — Elemental Shaman", gapB219GenasiOracle, 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Genasi Enforcers", "Creature — Elemental Shaman", gapB219GenasiOracle, 1, 3)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{first: 2, second: 2, other: 2, theirs: 1} {
		if p, _ := effectstest.PtOf(t, g, id); p != want {
			t.Errorf("a creature has power %d, want %d", p, want)
		}
	}
}

// --- Gnoll War Band ---

func TestGapB219GnollWarBandCostsOneLessPerOpponentDealtDamage(t *testing.T) {
	for _, damaged := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
		for i := 1; i <= damaged; i++ {
			g.WithWriteLock(func() {
				if err := g.DealDamageToPlayerForEffect(src, g.Seats[i].ID, 1); err != nil {
					t.Fatalf("damage: %v", err)
				}
			})
		}
		effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R")
		card := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: card, Name: "Gnoll War Band", TypeLine: "Creature — Gnoll", OracleID: gapB219GnollOracle,
			ManaCost: "{5}{R}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("damaged=%d: cast with six mana: %v", damaged, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if left := me.ManaPool.Total(); left != damaged {
			t.Errorf("damaged=%d: %d mana left in the pool, want %d", damaged, left, damaged)
		}
	}
}

// --- Gut, True Soul Zealot ---

func TestGapB219GutSacrificesToMakeASkeletonTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gut := effectstest.B12Push(g, me.ID, "Gut, True Soul Zealot", "Legendary Creature — Goblin Shaman", gapB219GutOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, gut)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Skeleton" && c.IsToken() && c.Tapped && c.AttackingTarget == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("no tapped and attacking Skeleton token")
	}
}

func TestGapB219GutDecliningMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gut := effectstest.B12Push(g, me.ID, "Gut, True Soul Zealot", "Legendary Creature — Goblin Shaman", gapB219GutOracle, 2, 2)
	effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, gut)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Skeleton" {
			t.Error("a Skeleton was made without a sacrifice")
		}
	}
}

// --- Insufferable Balladeer ---

func TestGapB219BalladeerMakesATargetUnableToBlockAndGoadsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Insufferable Balladeer", TypeLine: "Creature — Dwarf Bard", OracleID: gapB219BalladeerOracle, Power: 2, Toughness: 1})
	gapB219PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, victim) != me.ID {
		t.Error("the target was not goaded")
	}
	effectstest.BrAttack(t, g, mine)
	if err := g.DeclareBlocker(victim, mine); err == nil {
		t.Error("the target could block this turn")
	}
}
