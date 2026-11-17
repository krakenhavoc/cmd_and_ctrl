package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	broB01RocHunter     = "0ed811b7-6ead-4906-9e40-9096b735e271"
	broB01YotianMedic   = "d218b9ad-3a6f-4bda-90c6-3f869c6c4e6d"
	broB01HonorGuard    = "b68cf4e6-205e-40b6-b28b-f4b7d8af7017"
	broB01WingCommando  = "423c8465-82cd-4f54-b5ac-022b7ebcf3e4"
	broB01ForgeChanter  = "e69a8970-3ead-479d-860a-26ed65f9c846"
	broB01Chaindancer   = "cb690399-0fb8-4bd8-a73d-5e0e594ff2ae"
	broB01RetrievalAgnt = "2a6a1e58-85a9-436e-b1a0-373baa44d6f2"
	broB01Savant        = "6024cb2d-a6e7-4adc-8b34-7f4e0cac16ce"
	broB01Paratrooper   = "cd583365-27cc-4170-b138-4b3dd0d842b6"
	broB01Sprite        = "0dbcc5d8-9533-48fd-a69d-bbc6c72aaac5"
	broB01Tyrant        = "e7c06af6-8d96-4bfa-8baf-9e80f741e8cb"
	broB01ClayRevenant  = "c183186b-d92c-45a2-92d1-42e0df2b441f"
	broB01Bulwark       = "449ed4b4-9e1e-4181-9090-af91802eba49"
)

func TestBroB01CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		broB01RocHunter: "Roc Hunter", broB01YotianMedic: "Yotian Medic",
		broB01HonorGuard: "Tomakul Honor Guard", broB01WingCommando: "Wing Commando",
		broB01ForgeChanter: "Dwarven Forge-Chanter", broB01Chaindancer: "Fallaji Chaindancer",
		broB01RetrievalAgnt: "Retrieval Agent", broB01Savant: "Third Path Savant",
		broB01Paratrooper: "Ambush Paratrooper", broB01Sprite: "Argothian Sprite",
		broB01Tyrant: "Tyrant of Kher Ridges", broB01ClayRevenant: "Clay Revenant",
		broB01Bulwark: "Coastal Bulwark",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBroB01PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Roc Hunter", "Creature — Human Soldier", broB01RocHunter, []string{"reach"}},
		{"Yotian Medic", "Creature — Human Cleric Soldier", broB01YotianMedic, []string{"lifelink"}},
		{"Wing Commando", "Creature — Human Soldier", broB01WingCommando, []string{"flying", "prowess"}},
		{"Dwarven Forge-Chanter", "Creature — Dwarf Wizard", broB01ForgeChanter, []string{"prowess"}},
		{"Ambush Paratrooper", "Creature — Human Soldier", broB01Paratrooper, []string{"flash", "flying"}},
		{"Tyrant of Kher Ridges", "Creature — Dragon", broB01Tyrant, []string{"flying"}},
		{"Coastal Bulwark", "Artifact Creature — Wall", broB01Bulwark, []string{"defender"}},
	} {
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.kws {
			if !hasKeywordOf(g, id, kw) {
				t.Errorf("%s has no %s", tc.name, kw)
			}
		}
	}
}

func TestBroB01HonorGuardWardAsksTheCasterForTwo(t *testing.T) {
	g, me, opp, guard := wardTable(t, "Tomakul Honor Guard", "Creature — Human Soldier", broB01HonorGuard)
	castAtWardedCreature(t, g, opp, guard)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) || hasPayUnlessFor(g, me.ID) {
		t.Fatal("the ward {2} prompt must go to the caster only")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(guard) {
		t.Error("a declined ward must counter the spell")
	}
}

func TestBroB01ForgeChanterWardIsTwoLife(t *testing.T) {
	g, _, opp, chanter := wardTable(t, "Dwarven Forge-Chanter", "Creature — Dwarf Wizard", broB01ForgeChanter)
	castAtWardedCreature(t, g, opp, chanter)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward—pay 2 life prompt for the caster")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("Forge-Chanter ward LifeCost = %d, want 2", prompt.LifeCost)
	}
	life := opp.Life
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
}

// broB01Activate funds `mana` and activates ability idx strictly.
func broB01Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	return g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Strict: true})
}

func TestBroB01ChaindancerGainsDoubleStrikeForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Fallaji Chaindancer", "Creature — Human Soldier", broB01Chaindancer, 2, 4)
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err == nil {
		t.Fatal("activated with one mana")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOf(g, id, "double strike") {
		t.Error("no double strike after the ability resolved")
	}
}

func TestBroB01RetrievalAgentShrinksToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Retrieval Agent", "Creature — Human Soldier", broB01RetrievalAgnt, 2, 5)
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err == nil {
		t.Fatal("activated with one mana")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 3/4", p, tough)
	}
}

func TestBroB01SavantDrawsTwoForSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	id := b12Push(g, me.ID, "Third Path Savant", "Creature — Human Wizard", broB01Savant, 2, 3)
	if err := broB01Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}{C}"); err == nil {
		t.Fatal("activated with six mana")
	}
	hand := me.Hand.Size()
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("cards drawn = %d, want 2", got)
	}
}

func TestBroB01ParatrooperPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Ambush Paratrooper", "Creature — Human Soldier", broB01Paratrooper, 1, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	if err := broB01Activate(t, g, me, id, 0, "{C}{C}{C}{C}"); err == nil {
		t.Fatal("activated with four mana")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id           uuid.UUID
		name         string
		wantP, wantT int
	}{{id, "Paratrooper", 2, 3}, {mine, "my creature", 3, 3}, {foe, "the opponent's creature", 2, 2}} {
		if p, tough := effectivePTOf(t, g, tc.id); p != tc.wantP || tough != tc.wantT {
			t.Errorf("%s P/T = %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
		}
	}
}

func TestBroB01SpriteGrowsAndCantBeBlockedByArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Argothian Sprite", "Creature — Faerie", broB01Sprite, 2, 2)
	if err := broB01Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}{C}"); err == nil {
		t.Fatal("activated with six mana")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, id); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if got := powerOf(t, g, id); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, id)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(golem, id); err == nil {
		t.Error("an artifact creature blocked the Sprite")
	}
	if err := g.DeclareBlocker(bear, id); err != nil {
		t.Errorf("a nonartifact creature could not block: %v", err)
	}
}

func TestBroB01TyrantDealsFourWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = me
	life := opp.Life
	castCatalogSpell(t, g, "Tyrant of Kher Ridges", "Creature — Dragon", broB01Tyrant, nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, g.Seats[g.Turn.ActiveSeat].ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

func TestBroB01TyrantFirebreathing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Tyrant of Kher Ridges", "Creature — Dragon", broB01Tyrant, 4, 5)
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err == nil {
		t.Fatal("activated with colorless mana for {R}")
	}
	if err := broB01Activate(t, g, me, id, 0, "{R}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, id); p != 5 {
		t.Errorf("power = %d, want 5", p)
	}
}

func TestBroB01ClayRevenantEntersTappedAndReturnsFromGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, "Clay Revenant", "Artifact Creature — Golem", broB01ClayRevenant, nil)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok || !c.Tapped {
		t.Fatalf("Clay Revenant on battlefield=%v tapped=%v, want tapped", ok, c.Tapped)
	}

	gy := pushCatalogGraveyardCard(me, "Clay Revenant", "Artifact Creature — Golem", broB01ClayRevenant, 1, 2)
	if err := broB01Activate(t, g, me, gy, 0, "{C}{C}"); err == nil {
		t.Fatal("returned with no black mana")
	}
	if err := broB01Activate(t, g, me, gy, 0, "{B}"); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(gy) {
		t.Error("the card did not return to hand")
	}
}

func TestBroB01BulwarkPumpsWithAnIslandAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Coastal Bulwark", "Artifact Creature — Wall", broB01Bulwark, 1, 3)
	if p, _ := effectivePTOf(t, g, id); p != 1 {
		t.Errorf("power without an Island = %d, want 1", p)
	}
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if p, _ := effectivePTOf(t, g, id); p != 3 {
		t.Errorf("power with an Island = %d, want 3", p)
	}
	seedLibrary(me, "A", "B")
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err == nil {
		t.Fatal("activated with one mana")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := latestChoiceOfKind(g, game.PendingChoiceSurveil)
	if choice == nil {
		t.Fatal("no surveil choice")
	}
	if err := g.ResolveSurveil(choice.ID, me.ID, choice.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("graveyard size = %d, want 1 (the surveilled card)", me.Graveyard.Size())
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}{C}"); err == nil {
		t.Error("activated while tapped")
	}
}
