package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const evolvingDoorOracle = "30e4849e-c58c-4b65-9fb3-7aa40c944904"

// evolvingDoorTable is a main-phase table with the Door untapped under the
// active player, a library of creature cards with zero to three colours
// (and one green non-creature), and an empty hand.
func evolvingDoorTable(t *testing.T) (g *game.Game, me *game.Player, door uuid.UUID, lib map[string]uuid.UUID) {
	t.Helper()
	g, me, _ = effectstest.ExileCostTable(t)
	door = effectstest.PushCatalogPermanent(g, me.ID, "Evolving Door", "Artifact", evolvingDoorOracle, false)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Colorless Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}", Power: 2, Toughness: 2},
		game.Card{Name: "Mono Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2},
		game.Card{Name: "Mono Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
		game.Card{Name: "Two Hybrid", TypeLine: "Creature — Elf", ManaCost: "{G}{U}", Colors: []string{"G", "U"}, Power: 2, Toughness: 2},
		game.Card{Name: "Three Shard", TypeLine: "Creature — Beast", ManaCost: "{G}{U}{R}", Colors: []string{"G", "U", "R"}, Power: 3, Toughness: 3},
		game.Card{Name: "Green Ritual", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}},
	)
	lib = map[string]uuid.UUID{}
	for i, n := range []string{"Colorless Golem", "Mono Bear", "Mono Elf", "Two Hybrid", "Three Shard", "Green Ritual"} {
		lib[n] = ids[i]
	}
	return g, me, door, lib
}

func evolvingDoorActivate(g *game.Game, me *game.Player, door, sac uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, door, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sac}})
}

// The searchable cards a search prompt offers, by name.
func evolvingDoorOffered(g *game.Game, c *game.PendingChoice, names ...string) map[string]bool {
	out := map[string]bool{}
	for _, n := range names {
		out[n] = effectstest.SearchOptionNamed(g, c, n) != uuid.Nil
	}
	return out
}

// A colourless creature's count is zero, so the Door finds a mono-coloured
// creature card, and only that. The card is exiled, the library shuffled,
// and the cast is paid: with no mana it is refused, with the card's own
// cost it is made during the resolution, sorcery-speed creature or not.
func TestEvolvingDoorColorlessFindsMonoColoredAndTheCastIsPaid(t *testing.T) {
	g, me, door, lib := evolvingDoorTable(t)
	golem := effectstest.B16Creature(g, me.ID, "Field Golem", "Artifact Creature — Golem", 1, 1)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := evolvingDoorActivate(g, me, door, golem); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(golem) || !effectstest.B16Tapped(t, g, door) {
		t.Fatal("the creature is sacrificed and the Door taps")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	got := evolvingDoorOffered(g, c, "Colorless Golem", "Mono Bear", "Mono Elf", "Two Hybrid", "Three Shard", "Green Ritual")
	want := map[string]bool{"Mono Bear": true, "Mono Elf": true}
	for n, offered := range got {
		if offered != want[n] {
			t.Errorf("%s offered = %v, want %v (exactly one colour, creature cards only)", n, offered, want[n])
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Mono Bear")
	bear := lib["Mono Bear"]
	if !g.Exile.Contains(bear) || me.Library.Contains(bear) {
		t.Fatal("the found card is not in exile")
	}
	if me.Library.Size() != 5 {
		t.Fatalf("library holds %d, want the 5 cards not found", me.Library.Size())
	}
	shuffled := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Label == game.SearchShuffledLabel {
			shuffled = true
		}
	}
	if !shuffled {
		t.Error("the library was not shuffled after the search")
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, bear)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the exiled card was cast with no mana: the cast is not free")
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("a refused cast closed the offer")
	}
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("paid cast of the exiled card: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the pool holds %d after paying {1}{G}", me.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Mono Bear") == uuid.Nil {
		t.Error("the cast creature did not resolve onto the battlefield")
	}
}

// A two-colour sacrifice finds a three-colour creature card. The cast is
// optional: declined, the card stays in exile.
func TestEvolvingDoorTwoColorsFindsThreeAndTheCastCanBeDeclined(t *testing.T) {
	g, me, door, lib := evolvingDoorTable(t)
	pair := effectstest.B16Creature(g, me.ID, "Simic Pair", "Creature — Merfolk", 2, 2, "G", "U")
	effectstest.FloatMana(t, g, me, "{C}")
	if err := evolvingDoorActivate(g, me, door, pair); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	got := evolvingDoorOffered(g, c, "Colorless Golem", "Mono Bear", "Two Hybrid", "Three Shard")
	for n, offered := range got {
		if offered != (n == "Three Shard") {
			t.Errorf("%s offered = %v: only the three-colour creature card is", n, offered)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Three Shard")
	shard := lib["Three Shard"]
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, shard)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !g.Exile.Contains(shard) {
		t.Error("a declined card does not stay in exile")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held after the decline")
	}
	if err := g.CastSpell(me.ID, shard, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the offer was declined, but the card can still be cast from exile")
	}
}

// The count is the sacrificed creature as it last existed (CR 608.2h): a
// green token has ceased to exist by the time the ability resolves, and
// still counts one colour.
func TestEvolvingDoorReadsATokensLastKnownColors(t *testing.T) {
	g, me, door, _ := evolvingDoorTable(t)
	token := effectstest.B16Creature(g, me.ID, "Saproling", "Token Creature — Saproling", 1, 1, "G")
	effectstest.FloatMana(t, g, me, "{C}")
	if err := evolvingDoorActivate(g, me, door, token); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Two Hybrid") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Mono Bear") != uuid.Nil {
		t.Error("a one-colour token should find a two-colour creature card, and only that")
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a search that found nothing offers a cast")
	}
}

// Activate only as a sorcery: refused with a spell on the stack, and
// refused without the {1}.
func TestEvolvingDoorIsSorcerySpeedAndCostsOne(t *testing.T) {
	g, me, door, _ := evolvingDoorTable(t)
	bear := effectstest.B16Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2, "G")
	if err := evolvingDoorActivate(g, me, door, bear); err == nil {
		t.Fatal("activated with no mana for {1}")
	}
	var instant uuid.UUID
	g.WithWriteLock(func() { instant = effectstest.PushHandCardWithManaCost(me, "Quick Trick", "Instant", "{0}") })
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := evolvingDoorActivate(g, me, door, bear); err == nil {
		t.Fatal("activated with a spell on the stack: it is sorcery speed")
	}
	if !g.Battlefield.Contains(bear) || effectstest.B16Tapped(t, g, door) {
		t.Fatal("a refused activation paid its costs")
	}
	spec, _ := effects.Lookup(evolvingDoorOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed || spec.Completeness != effects.CompletenessFull {
		t.Error("one sorcery-speed ability, Full")
	}
}
