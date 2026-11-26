package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 366 (S82, #5558): Journey into Nyx strive cards — Hour of Need,
// Polymorphous Rush, Silence the Believers, Blinding Flare, Colossal Heroics,
// Kiora's Dismissal, Setessan Tactics, Harness by Force, Nature's Panoply.

const (
	oldB366Hour      = "2002dd67-2347-4869-98de-299074206684"
	oldB366Rush      = "3098626a-b786-4f54-83df-51833729c1ec"
	oldB366Silence   = "3c2de008-5747-49c0-89c1-83331cb55896"
	oldB366Flare     = "3c8b0a8a-8f80-4c3e-8f27-fb1df94aa0d4"
	oldB366Heroics   = "5b01acaf-bcf4-4ac1-ac62-2f25a45c2589"
	oldB366Dismissal = "6f866b14-dcdd-4c7d-b935-6db431d83b86"
	oldB366Tactics   = "755e2f3e-bb32-4902-af8d-d80f9899c69f"
	oldB366Harness   = "8c7e2596-6c21-4ce0-95dc-5b4dccb4090a"
	oldB366Panoply   = "b3126751-27fe-42d3-a001-33ecee72b528"
)

var oldB366Cards = []struct {
	name, oracle, typeLine, cost string
	twoTargetsMV                 int
}{
	{"Hour of Need", oldB366Hour, "Instant", "{2}{U}", 5},
	{"Polymorphous Rush", oldB366Rush, "Instant", "{2}{U}", 5},
	{"Silence the Believers", oldB366Silence, "Instant", "{2}{B}{B}", 7},
	{"Blinding Flare", oldB366Flare, "Sorcery", "{R}", 2},
	{"Colossal Heroics", oldB366Heroics, "Instant", "{2}{G}", 5},
	{"Setessan Tactics", oldB366Tactics, "Instant", "{1}{G}", 3},
	{"Harness by Force", oldB366Harness, "Sorcery", "{1}{R}{R}", 6},
	{"Nature's Panoply", oldB366Panoply, "Instant", "{G}", 4},
}

func TestOldB366CardsAreRegisteredFull(t *testing.T) {
	names := map[string]string{oldB366Dismissal: "Kiora's Dismissal"}
	for _, c := range oldB366Cards {
		names[c.oracle] = c.name
	}
	for oracle, name := range names {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB366MV(t *testing.T, cost string) int {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", cost, err)
	}
	return parsed.ManaValue()
}

func TestOldB366StriveCostsMorePerExtraTarget(t *testing.T) {
	for _, tc := range oldB366Cards {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
			b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
			if got := effectstest.SelfPriced(t, g, me, tc.oracle, tc.typeLine, tc.cost, game.ZoneHand, effectstest.CardRefs(a)).ManaValue(); got != oldB366MV(t, tc.cost) {
				t.Errorf("one target costs mana value %d, want the printed cost", got)
			}
			if got := effectstest.SelfPriced(t, g, me, tc.oracle, tc.typeLine, tc.cost, game.ZoneHand, effectstest.CardRefs(a, b)).ManaValue(); got != tc.twoTargetsMV {
				t.Errorf("two targets cost mana value %d, want %d", got, tc.twoTargetsMV)
			}
		})
	}
	t.Run("Kiora's Dismissal", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		a := effectstest.B12Push(g, me.ID, "Ench A", "Enchantment", "", 0, 0)
		b := effectstest.B12Push(g, me.ID, "Ench B", "Enchantment", "", 0, 0)
		if got := effectstest.SelfPriced(t, g, me, oldB366Dismissal, "Instant", "{U}", game.ZoneHand, effectstest.CardRefs(a, b)).ManaValue(); got != 2 {
			t.Errorf("two targets cost mana value %d, want 2", got)
		}
	})
}

func TestOldB366KiorasDismissalBouncesEachEnchantmentTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := effectstest.B12Push(g, me.ID, "Ench A", "Enchantment", "", 0, 0)
	b := effectstest.B12Push(g, g.Seats[1].ID, "Ench B", "Enchantment", "", 0, 0)
	keep := effectstest.B12Push(g, me.ID, "Ench Keep", "Enchantment", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Kiora's Dismissal", "Instant", oldB366Dismissal, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted enchantment stayed on the battlefield")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted enchantment left")
	}
	if effectstest.B12ZoneOf(g, b) != game.ZoneHand {
		t.Errorf("their enchantment went to %v, want its owner's hand", effectstest.B12ZoneOf(g, b))
	}
	creature := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Kiora's Dismissal", "Instant", oldB366Dismissal, effectstest.CardTargets(creature)); err == nil {
		t.Error("targeted a nonenchantment")
	}
}

func TestOldB366NaturesPanoplyPutsACounterOnEachTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "B", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Nature's Panoply", "Instant", oldB366Panoply, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("a target has %d counters, want 1", n)
		}
	}
	if n := effectstest.B12Counter(t, g, idle, game.CounterPlusOne); n != 0 {
		t.Errorf("an untargeted creature has %d counters", n)
	}
}

func TestOldB366ColossalHeroicsPumpsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	effectstest.ArnTap(t, g, a)
	effectstest.ArnTap(t, g, b)
	effectstest.CastCatalogSpell(t, g, "Colossal Heroics", "Instant", oldB366Heroics, effectstest.CardRefs(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if effectstest.B12Card(t, g, a).Tapped {
		t.Error("the target is still tapped")
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 2 {
		t.Errorf("an untargeted creature has power %d", p)
	}
	if !effectstest.B12Card(t, g, b).Tapped {
		t.Error("an untargeted creature was untapped")
	}
}

func TestOldB366BlindingFlareStopsEachTargetBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 2, 2)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blinding Flare", "Sorcery", oldB366Flare, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.BlockRefusal(t, g, id, attacker) == "" {
			t.Error("a target blocked")
		}
	}
	if got := effectstest.BlockRefusal(t, g, idle, attacker); got != "" {
		t.Errorf("an untargeted creature was refused: %q", got)
	}
}

func TestOldB366HarnessByForceStealsUntapsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 3, 3)
	effectstest.ArnTap(t, g, a)
	effectstest.CastCatalogSpell(t, g, "Harness by Force", "Sorcery", oldB366Harness, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.LayeredCard(t, g, id)
		if c.Controller != me.ID {
			t.Errorf("%s is controlled by someone else", c.Name)
		}
		if c.Tapped {
			t.Errorf("%s is tapped", c.Name)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", c.Name)
		}
	}
}

func TestOldB366SilenceTheBelieversExilesCreaturesAndTheirAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Other", 2, 2)
	auraA := effectstest.B12Push(g, me.ID, "Aura on A", "Enchantment — Aura", "", 0, 0)
	auraOther := effectstest.B12Push(g, me.ID, "Aura on Other", "Enchantment — Aura", "", 0, 0)
	effectstest.NcdAttach(t, g, auraA, a)
	effectstest.NcdAttach(t, g, auraOther, other)
	effectstest.CastCatalogSpell(t, g, "Silence the Believers", "Instant", oldB366Silence, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, auraA} {
		if g.Battlefield.Contains(id) || !effectstest.InExile(g, id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !g.Battlefield.Contains(other) || !g.Battlefield.Contains(auraOther) {
		t.Error("an unrelated creature or Aura was exiled")
	}
}

func TestOldB366HourOfNeedGivesEachControllerASphinx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, opp.ID, "C", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hour of Need", "Instant", oldB366Hour, effectstest.CardRefs(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	count := map[uuid.UUID]int{}
	for _, card := range g.Battlefield.Cards {
		if card.Name != "Sphinx" || !card.IsToken() {
			continue
		}
		count[card.Controller]++
		if p, tough := effectstest.PtOf(t, g, card.InstanceID); p != 4 || tough != 4 {
			t.Errorf("a Sphinx is %d/%d", p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, card.InstanceID, "flying") {
			t.Error("a Sphinx has no flying")
		}
	}
	if count[opp.ID] != 2 || count[me.ID] != 1 {
		t.Errorf("sphinx counts %v, want two for the opponent and one for me", count)
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.InExile(g, id) {
			t.Error("a target was not exiled")
		}
	}
}

func TestOldB366SetessanTacticsPumpsAndGrantsAFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	foe := effectstest.PushVanillaCreature(g, opp.ID, "Foe", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Setessan Tactics", "Instant", oldB366Tactics, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("a target is %d/%d, want 3/3", p, tough)
		}
	}
	// The granted ability: {T}: this creature fights another target creature.
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(a), Strict: true}); err == nil {
		t.Error("the creature fought itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(foe), Strict: true}); err != nil {
		t.Fatalf("activate the granted fight: %v", err)
	}
	if !effectstest.B12Card(t, g, a).Tapped {
		t.Error("the fight did not cost {T}")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the 3/3 foe survived 3 damage")
	}
	if g.Battlefield.Contains(a) {
		// 3 damage dealt to the 3/3 attacker by the 3/3 foe: it dies too.
		t.Error("the fighter survived 3 damage")
	}
}

func TestOldB366PolymorphousRushCopiesTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "Small A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Small B", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Giant Model", 6, 6)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Small", 1, 1)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Polymorphous Rush", "Instant", oldB366Rush, effectstest.CardTargets(theirs)); err == nil {
		t.Error("targeted a creature I do not control")
	}
	effectstest.CastCatalogSpell(t, g, "Polymorphous Rush", "Instant", oldB366Rush, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BlbSTPickOption(t, g, me.ID, "Giant Model")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c := effectstest.LayeredCard(t, g, id); c.Name != "Giant Model" {
			t.Errorf("a target is %q, want a copy of Giant Model", c.Name)
		}
		if p, tough := effectstest.PtOf(t, g, id); p != 6 || tough != 6 {
			t.Errorf("a copy is %d/%d, want 6/6", p, tough)
		}
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.Name != "Their Small" {
		t.Errorf("an untargeted creature became %q", c.Name)
	}
}
