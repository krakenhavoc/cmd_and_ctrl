package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b787_cards_test.go — Visions, batch 787 (#5631).

const (
	oldB787Tithe      = "318304f5-0929-4cd3-97e6-08ac1ac48aa6"
	oldB787Foreshadow = "4caa97df-1f6c-4806-bf14-e589cc1badb5"
	oldB787Enchantres = "8b983a46-580f-4a94-8ef3-5253e503148e"
	oldB787Drake      = "94806688-329d-429d-8126-2c9187e71f54"
	oldB787Feral      = "f8f3fc11-25b4-4a23-9d79-f380a14f0420"
)

func TestOldB787EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB787Tithe: "Tithe", oldB787Foreshadow: "Foreshadow", oldB787Enchantres: "Femeref Enchantress",
		oldB787Drake: "Kyscu Drake", oldB787Feral: "Feral Instinct",
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

// oldB787DrawsAtTheNextUpkeep checks the delayed draw: nothing before the
// trigger resolves, one card after.
func oldB787DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p.Hand.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

func TestOldB787FeralInstinctPumpsAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Feral Instinct", "Instant", oldB787Feral, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("target %d/%d, want 3/3", p, tt)
	}
	oldB787DrawsAtTheNextUpkeep(t, g, me)
	if p, tt := effectstest.PtOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("next turn %d/%d, want 2/2", p, tt)
	}
}

func TestOldB787FeralInstinctRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Feral Instinct", "Instant", oldB787Feral, effectstest.CardTarget(rock)); err == nil {
		t.Error("targeted a noncreature permanent")
	}
}

func oldB787Plains(me *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"}))
	}
	return out
}

func oldB787InHand(me *game.Player, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if effectstest.InHandOf(me, id) {
			n++
		}
	}
	return n
}

func TestOldB787TitheFetchesOnePlainsWhenTheOpponentIsNotAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plains := oldB787Plains(me, 3)
	effectstest.PushLand(g, me.ID, "Plains")
	effectstest.PushLand(g, opp.ID, "Mountain")
	effectstest.CastCatalogSpell(t, g, "Tithe", "Instant", oldB787Tithe, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB787InHand(me, plains); got != 1 {
		t.Errorf("%d Plains cards in hand, want 1 (level lands is not more)", got)
	}
}

func TestOldB787TitheFetchesTwoPlainsWhenTheOpponentControlsMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plains := oldB787Plains(me, 3)
	// A nonbasic Plains card counts, a non-Plains land does not.
	nonbasic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Kjeldoran Outpost", TypeLine: "Land — Plains"})
	isl := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.PushLand(g, opp.ID, "Mountain")
	effectstest.CastCatalogSpell(t, g, "Tithe", "Instant", oldB787Tithe, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB787InHand(me, append(plains, nonbasic)); got != 2 {
		t.Errorf("%d Plains cards in hand, want 2", got)
	}
	if effectstest.InHandOf(me, isl) {
		t.Error("fetched a non-Plains land")
	}
}

func TestOldB787TitheRefusesYourOwnSelfAsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Tithe", "Instant", oldB787Tithe, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("targeted its own caster")
	}
}

func TestOldB787FemerefEnchantressDrawsForAnyEnchantmentThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	effectstest.B12Push(g, me.ID, "Femeref Enchantress", "Creature — Human Druid", oldB787Enchantres, 1, 2)
	mine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment — Aura")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d after one's own enchantment died, want +1", hand, me.Hand.Size())
	}
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d -> %d after an opponent's enchantment died, want +2", hand, me.Hand.Size())
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("a creature dying drew a card: hand %d", me.Hand.Size())
	}
}

func TestOldB787FemerefEnchantressIgnoresAnEnchantmentExiledNotDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 6)
	effectstest.B12Push(g, me.ID, "Femeref Enchantress", "Creature — Human Druid", oldB787Enchantres, 1, 2)
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(shrine); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an exiled enchantment drew a card: hand %d -> %d", hand, me.Hand.Size())
	}
}

func oldB787Drake787(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.B21Push(g, me.ID, "Kyscu Drake", "Creature — Drake", oldB787Drake, 2, 2)
}

func TestOldB787KyscuDrakePumpsToughnessOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := oldB787Drake787(t, g, me)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	effectstest.B16Activate(t, g, me.ID, drake, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, drake); p != 2 || tt != 3 {
		t.Errorf("Drake %d/%d, want 2/3", p, tt)
	}
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second pump in one turn was allowed")
	}
}

func TestOldB787KyscuDrakeSacrificesWithSpittingDrakeToFetchViashivanDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := oldB787Drake787(t, g, me)
	spitting := effectstest.B21Push(g, me.ID, "Spitting Drake", "Creature — Drake", "", 1, 1)
	other := effectstest.B21Push(g, me.ID, "Other Drake", "Creature — Drake", "", 1, 1)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Viashivan Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	decoy := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Viashivan Whelp", TypeLine: "Creature — Dragon", Power: 1, Toughness: 1})
	if err := g.ActivateCatalogAbility(me.ID, drake, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("paid with a creature not named Spitting Drake")
	}
	effectstest.B16Activate(t, g, me.ID, drake, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{spitting}})
	if g.Battlefield.Contains(drake) || g.Battlefield.Contains(spitting) {
		t.Error("the Kyscu Drake and the Spitting Drake were not both sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("Viashivan Dragon was not put onto the battlefield")
	}
	if g.Battlefield.Contains(decoy) {
		t.Error("a card not named Viashivan Dragon was fetched")
	}
}

func TestOldB787KyscuDrakeCannotSacrificeWithoutSpittingDrake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := oldB787Drake787(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, drake, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated the sacrifice ability with no Spitting Drake")
	}
}

func oldB787Name(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName {
			if c.Chooser != chooser {
				t.Fatalf("the name prompt asks %v, want the caster %v", c.Chooser, chooser)
			}
			if err := g.ResolveCardNameChoice(c.ID, c.Chooser, name); err != nil {
				t.Fatalf("name: %v", err)
			}
			return
		}
	}
	t.Fatal("no card-name prompt")
}

func oldB787CastForeshadow(t *testing.T, g *game.Game, guess string) (me, opp *game.Player, top uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "A", "B", "C")
	ids := effectstest.SeedLibrary(opp, "Counterspell", "Filler", "Filler")
	effectstest.CastCatalogSpell(t, g, "Foreshadow", "Instant", oldB787Foreshadow, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	oldB787Name(t, g, me.ID, guess)
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp, ids[0]
}

func TestOldB787ForeshadowRightGuessDrawsNowAndAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	_, opp, top := oldB787CastForeshadow(t, g, "counterspell")
	if !effectstest.InGraveyardOf(g, opp.ID, top) {
		t.Error("the top card was not milled")
	}
	if me.Hand.Size() != before+1 {
		t.Fatalf("hand %d -> %d after a right guess, want +1", before, me.Hand.Size())
	}
	oldB787DrawsAtTheNextUpkeep(t, g, me)
}

func TestOldB787ForeshadowWrongGuessStillDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	_, opp, top := oldB787CastForeshadow(t, g, "Opt")
	if !effectstest.InGraveyardOf(g, opp.ID, top) {
		t.Error("the top card was not milled")
	}
	if me.Hand.Size() != before {
		t.Fatalf("hand %d -> %d after a wrong guess, want unchanged", before, me.Hand.Size())
	}
	oldB787DrawsAtTheNextUpkeep(t, g, me)
}

func TestOldB787ForeshadowRefusesItsCasterAsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Foreshadow", "Instant", oldB787Foreshadow, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("targeted its own caster")
	}
}
