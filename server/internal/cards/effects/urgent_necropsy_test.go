package effects

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// urgent_necropsy_test.go — #5240: Urgent Necropsy's "As an additional
// cost to cast this spell, collect evidence X, where X is the total mana
// value of the permanents this spell targets" (CR 701.59a), a cost whose
// amount is read off the targets chosen at CR 601.2c as the total cost is
// determined at CR 601.2f, played through the real catalog entry.

const urgentNecropsyOracle = "378a6203-0d47-4165-bdc2-43dfdfb4d15d"

func TestUrgentNecropsyIsRegistered(t *testing.T) {
	spec, ok := Lookup(urgentNecropsyOracle)
	if !ok {
		t.Fatal("Urgent Necropsy is not registered")
	}
	if spec.Name != "Urgent Necropsy" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	ac := game.AdditionalCostFor(urgentNecropsyOracle)
	if ac == nil || !ac.ExileCards.IsCollectEvidenceFromTargets() {
		t.Fatalf("additional cost %+v, want collect evidence sized by the targets", ac)
	}
	clauses := game.TargetSpecFor(urgentNecropsyOracle).Clauses()
	if len(clauses) != 4 {
		t.Fatalf("%d target clauses, want four", len(clauses))
	}
	for i, c := range clauses {
		if c.Min != 0 || c.Max != 1 || c.Distinct {
			t.Errorf("clause %d is %d..%d distinct=%v, want an \"up to one\" that may repeat a permanent", i, c.Min, c.Max, c.Distinct)
		}
	}
}

// necropsyTable is a main phase on seat 0's turn with Urgent Necropsy in
// hand, {2}{B}{G} in the pool, and an opponent.
func necropsyTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spell := handCardOf(g, me, "Urgent Necropsy", "Instant", "{2}{B}{G}", urgentNecropsyOracle)
	g.WithWriteLock(func() {
		fillPoolColored(me, "B", 2)
		fillPoolColored(me, "G", 2)
	})
	return g, me, opp, spell
}

// necropsyPermanent puts a permanent with a mana cost onto the battlefield.
func necropsyPermanent(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, owner)
	c.TypeLine, c.ManaCost, c.Controller = typeLine, cost, owner
	if c.IsCreature() {
		c.Power, c.Toughness = 2, 2
	}
	if c.IsPlaneswalker() {
		c.Counters = map[string]int{game.CounterLoyalty: 3}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func necropsyCast(g *game.Game, p *game.Player, spell uuid.UUID, targets []game.TargetRef, evidence ...uuid.UUID) error {
	return g.CastSpell(p.ID, spell, game.CastSpellParams{Targets: targets, TargetsStamped: true, ExileIDs: evidence, Strict: true})
}

// slotRef is a target stamped with the clause it answers (0 artifact, 1
// creature, 2 enchantment, 3 planeswalker).
func slotRef(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
}

// Four targets worth 10: cards worth 9 are refused with nothing paid;
// cards worth 10 are exiled with the spell on the stack, and the spell
// destroys all four.
func TestUrgentNecropsyCollectsTheTargetsTotalAndDestroysThem(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	art := necropsyPermanent(g, opp.ID, "Signet", "Artifact", "{2}")
	bear := necropsyPermanent(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}{G}")
	aura := necropsyPermanent(g, opp.ID, "Charm", "Enchantment", "{W}")
	pw := necropsyPermanent(g, opp.ID, "Walker", "Legendary Planeswalker — Test", "{2}{U}{U}")
	targets := []game.TargetRef{slotRef(art, 0), slotRef(bear, 1), slotRef(aura, 2), slotRef(pw, 3)}
	six := gyCard(me, "Six", "Sorcery", "{6}", "")
	three := gyCard(me, "Three", "Sorcery", "{3}", "")
	one := gyCard(me, "One", "Sorcery", "{B}", "")

	if err := necropsyCast(g, me, spell, targets, six, three); err == nil {
		t.Fatal("cards worth 9 paid collect evidence 10")
	}
	if !me.Hand.Contains(spell) || !me.Graveyard.Contains(six) || !me.Graveyard.Contains(three) {
		t.Fatal("the refused cast paid something")
	}
	if err := necropsyCast(g, me, spell, targets, six, three, one); err != nil {
		t.Fatalf("cards worth 10 for targets worth 10: %v", err)
	}
	if !g.Stack.Contains(spell) || !g.Exile.Contains(six) || !g.Exile.Contains(three) || !g.Exile.Contains(one) {
		t.Fatal("the evidence was not exiled with the spell on the stack")
	}
	wantOneCollection(t, g, me.ID, 10, six, three, one)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"artifact": art, "creature": bear, "enchantment": aura, "planeswalker": pw} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s was not destroyed", name)
		}
	}
}

// An artifact creature chosen as both the target artifact and the target
// creature (CR 115.3) is one permanent the spell targets: X counts it once,
// and it is destroyed once.
func TestUrgentNecropsyCountsAPermanentTargetedTwiceOnce(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	golem := necropsyPermanent(g, opp.ID, "Golem", "Artifact Creature — Golem", "{4}")
	four := gyCard(me, "Four", "Sorcery", "{4}", "")
	if err := necropsyCast(g, me, spell, []game.TargetRef{slotRef(golem, 0), slotRef(golem, 1)}, four); err != nil {
		t.Fatalf("collect evidence 4 for a Golem named twice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Fatal("the Golem was not destroyed")
	}
	if n := countIn(opp.Graveyard, golem); n != 1 {
		t.Errorf("the Golem is in its owner's graveyard %d times, want once", n)
	}
}

// No targets: X is 0 and nothing need be exiled, and that is still
// collecting evidence (Evidence Examiner's "whenever you collect evidence"
// would see it). A target worth 3 with an empty graveyard can't be paid.
func TestUrgentNecropsyWithNoTargetsCollectsEvidenceZero(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	bear := necropsyPermanent(g, opp.ID, "Bear", "Creature — Bear", "{2}{G}")
	if err := necropsyCast(g, me, spell, []game.TargetRef{slotRef(bear, 1)}); err == nil {
		t.Fatal("a target worth 3 was paid with an empty graveyard")
	}
	if err := necropsyCast(g, me, spell, nil); err != nil {
		t.Fatalf("Urgent Necropsy with no targets: %v", err)
	}
	wantOneCollection(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a spell with no targets destroyed the Bear")
	}
}

// The view ships the graveyard picker, the target-sized flag and each
// permanent's mana value, the numbers the engine sums; the spell is
// castable with an empty graveyard (X = 0 is always open).
func TestUrgentNecropsyViewAsksAfterTargeting(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	bear := necropsyPermanent(g, opp.ID, "Bear", "Creature — Bear", "{2}{G}")
	face := necropsyPermanent(g, opp.ID, "Hidden", "Creature — Bear", "{5}")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, face).SetFaceDown(game.FaceDownMorphed) })
	gy := gyCard(me, "Evidence", "Sorcery", "{3}", "")

	c := foiHandView(g, me.ID, spell)
	if c == nil || c.AdditionalCost == nil {
		t.Fatal("no additional cost on Urgent Necropsy's hand view")
	}
	ac := c.AdditionalCost
	if !ac.ExileCostFromTargets || ac.ExileCostN != 0 || ac.ExileCostZone != "graveyard" {
		t.Errorf("view %+v, want a target-sized graveyard exile with no count", ac)
	}
	if len(ac.ExileCostOptions) != 1 || ac.ExileCostOptions[0] != gy.String() {
		t.Errorf("exile options %v, want the one graveyard card", ac.ExileCostOptions)
	}
	if ac.ExileCostTargetManaValues[bear.String()] != 3 || ac.ExileCostTargetManaValues[face.String()] != 0 {
		t.Errorf("mana values %v, want the Bear's 3 and the face-down 0", ac.ExileCostTargetManaValues)
	}
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	if !foiCastableHere(t, g, me.ID, spell) {
		t.Error("Urgent Necropsy is not castable with an empty graveyard")
	}
}

// The bot: with a graveyard worth 3, the enumerator offers casts at the
// Bear (worth 3) paying exactly enough, never at the Dragon (worth 6),
// and every offered payment reaches its targets' X. The engine accepts the
// moves it offers.
func TestUrgentNecropsyIsOfferedWithAPaymentThatMeetsTheTargets(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	bear := necropsyPermanent(g, opp.ID, "Bear", "Creature — Bear", "{2}{G}")
	dragon := necropsyPermanent(g, opp.ID, "Dragon", "Creature — Dragon", "{4}{R}{R}")
	one := gyCard(me, "One", "Sorcery", "{1}", "")
	two := gyCard(me, "Two", "Sorcery", "{2}", "")
	values := map[string]int{one.String(): 1, two.String(): 2}
	mvs := map[string]int{bear.String(): 3, dragon.String(): 6}

	ms := foiCastMoves(g, me.ID, spell)
	if len(ms) == 0 {
		t.Fatal("the enumerator offers no Urgent Necropsy cast")
	}
	atBear := -1
	for i, m := range ms {
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
			ExileIDs []string `json:"exile_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		x, paid := 0, 0
		seen := map[string]bool{}
		for _, tr := range p.Targets {
			if !seen[tr.ID] {
				seen[tr.ID] = true
				x += mvs[tr.ID]
			}
			if tr.ID == dragon.String() {
				t.Errorf("move %q targets the Dragon, worth more than the graveyard", m.Label)
			}
		}
		for _, id := range p.ExileIDs {
			paid += values[id]
		}
		if paid < x {
			t.Errorf("move %q pays %d for targets worth %d", m.Label, paid, x)
		}
		if seen[bear.String()] && atBear < 0 {
			atBear = i
		}
		if len(p.Targets) == 0 && len(p.ExileIDs) != 0 {
			t.Errorf("move %q exiles cards for X = 0", m.Label)
		}
	}
	if atBear < 0 {
		t.Fatal("no move targets the Bear")
	}
	m := ms[atBear]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("the enumerated Urgent Necropsy cast is refused: %v", err)
	}
	if !g.Stack.Contains(spell) || len(me.Graveyard.Cards) != 0 {
		t.Fatal("the enumerated cast did not exile the graveyard for the Bear")
	}
}

// A restore taken with the spell on the stack keeps what was paid and
// what it targets: the exiled evidence stays on the paid record, and the
// restored table resolves the spell.
func TestUrgentNecropsyOnTheStackSurvivesARestore(t *testing.T) {
	g, me, opp, spell := necropsyTable(t)
	bear := necropsyPermanent(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	two := gyCard(me, "Two", "Sorcery", "{2}", "")
	if err := necropsyCast(g, me, spell, []game.TargetRef{slotRef(bear, 1)}, two); err != nil {
		t.Fatalf("cast: %v", err)
	}
	restored := restoreRoundTrip(t, g, true)
	item := restored.StackMeta[spell]
	if item == nil || len(item.Paid.Exiled) != 1 || item.Paid.Exiled[0] != two {
		t.Fatalf("the restored stack item does not keep the exiled evidence: %+v", item)
	}
	passPriorityAroundTable(t, restored)
	if restored.Battlefield.Contains(bear) {
		t.Error("the restored spell did not destroy the Bear")
	}
}

// Register refuses the target-sized clause where the cast path does not
// pay it as printed.
func TestRegisterRefusesAMisplacedTargetSizedCollectEvidence(t *testing.T) {
	sized := func() *game.ExileCost {
		return CollectEvidenceFromTargetsCost().ExileCards
	}
	upTo := func() *game.TargetSpec { return TargetCreature("up to one target creature").WithCount(0, 1) }
	for name, spec := range map[string]Spec{
		"no target clause": {OracleID: "test-5240-none", Name: "No Targets",
			AdditionalCost: CollectEvidenceFromTargetsCost(), OnResolve: func(*game.StackItem, *Context) error { return nil }},
		"a target that must be chosen": {OracleID: "test-5240-must", Name: "Must Target",
			AdditionalCost: CollectEvidenceFromTargetsCost(), Targets: TargetCreature("target creature"),
			OnResolve: func(*game.StackItem, *Context) error { return nil }},
		"beside another component": {OracleID: "test-5240-beside", Name: "Beside",
			AdditionalCost: func() *game.AdditionalCost {
				c := CollectEvidenceFromTargetsCost()
				c.DiscardCards = 1
				return c
			}(), Targets: upTo(), OnResolve: func(*game.StackItem, *Context) error { return nil }},
		"on an optional cost": {OracleID: "test-5240-optional", Name: "Optional",
			OptionalCosts: []game.AdditionalCost{{Optional: true, Key: "evidence", Label: "x", ExileCards: sized()}},
			Targets:       upTo(), OnResolve: func(*game.StackItem, *Context) error { return nil }},
		"on a branch": {OracleID: "test-5240-branch", Name: "Branch",
			AdditionalCost: EitherCost(
				(&game.AdditionalCost{ExileCards: sized(), Label: "x"}).Keyed("evidence"),
				ManaAdditionalCost("{1}").Keyed("mana")),
			Targets: upTo(), OnResolve: func(*game.StackItem, *Context) error { return nil }},
		"on an ability": {OracleID: "test-5240-ability", Name: "Ability",
			Activated: []ActivatedAbility{{Label: "Exile: draw.", Cost: game.AbilityCost{ExileCards: sized()}, Effect: Do()}}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("Register accepted %s", name)
				} else if msg := fmt.Sprint(r); !strings.Contains(msg, "#5240") {
					t.Errorf("Register refused %s for another reason: %s", name, msg)
				}
			}()
			Register(spec)
		})
	}
}

func countIn(z *game.Zone, id uuid.UUID) int {
	n := 0
	for _, c := range z.Cards {
		if c.InstanceID == id {
			n++
		}
	}
	return n
}
