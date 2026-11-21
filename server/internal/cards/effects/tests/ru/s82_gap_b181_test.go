package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b181_test.go — GAP batch 181 (#5948, Starter Commander Decks):
// every card driven through the engine.

const (
	gap181SkybonderOracle  = "aab55af7-8084-47b6-910f-75a75a6313a4"
	gap181SkycatOracle     = "84adce5c-39c7-425e-b163-4a1a3977364b"
	gap181LieutenantOracle = "5a040363-dba5-49d2-9756-e14af61cf818"
	gap181FavorOracle      = "fc3e48c9-e303-435c-bdb1-7dc7881a9858"
	gap181InstinctOracle   = "9ba89c25-5302-414f-a2ec-436b8981999d"
	gap181BlessingOracle   = "7f7049e8-49ed-46da-89f7-1e40aefb3b0c"
	gap181InsightOracle    = "09071f5f-c9cb-43a9-b41d-faf6caf3beff"
)

func gap181Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: toughness, Keywords: keywords})
}

func gap181Fund(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func TestS82Gap181CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{gap181SkybonderOracle, gap181SkycatOracle, gap181LieutenantOracle, gap181FavorOracle,
		gap181InstinctOracle, gap181BlessingOracle, gap181InsightOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82Gap181PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Jubilant Skybonder", gap181SkybonderOracle, 2, 2},
		{"Skycat Sovereign", gap181SkycatOracle, 1, 1},
		{"Kangee's Lieutenant", gap181LieutenantOracle, 1, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, "Creature — Test", c.oracle, c.power, c.tough)
		effectstest.AssertKeywords(t, g, id, "flying")
	}
}

// --- Collective Blessing ------------------------------------------------

func TestS82Gap181CollectiveBlessingPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.B12Push(g, me.ID, "Collective Blessing", "Enchantment", gap181BlessingOracle, 0, 0)
	if got := effectstest.EffectivePower(t, g, mine); got != 5 {
		t.Errorf("your Bear has power %d, want 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, mine); got != 5 {
		t.Errorf("your Bear has toughness %d, want 5", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Bear has power %d, want 2", got)
	}
}

// --- Staggering Insight -------------------------------------------------

func TestS82Gap181StaggeringInsightPumpsGivesLifelinkAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Staggering Insight", "Enchantment — Aura", gap181InsightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Error("enchanted creature has no lifelink")
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after combat damage, want +1", hand, me.Hand.Size())
	}
}

func TestS82Gap181StaggeringInsightRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Staggering Insight", "Enchantment — Aura", gap181InsightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("Staggering Insight was cast enchanting an artifact")
	}
}

// --- Karametra's Favor --------------------------------------------------

func TestS82Gap181KarametrasFavorDrawsAndGrantsAnyColorMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedBear(g, me.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Karametra's Favor", "Enchantment — Aura", gap181FavorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after the Aura entered, want +1", hand, me.Hand.Size())
	}
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 1 {
		t.Errorf("enchanted creature has %d granted mana abilities, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, effectstest.SeedBear(g, me.ID)); n != 0 {
		t.Errorf("an unenchanted creature has %d granted mana abilities, want 0", n)
	}
}

// --- Overwhelming Instinct ----------------------------------------------

func TestS82Gap181OverwhelmingInstinctDrawsWhenThreeAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Overwhelming Instinct", "Enchantment", gap181InstinctOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "C", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after three attackers, want +1", hand, me.Hand.Size())
	}
}

func TestS82Gap181OverwhelmingInstinctIgnoresTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Overwhelming Instinct", "Enchantment", gap181InstinctOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after two attackers, want no change", hand, me.Hand.Size())
	}
}

// --- Skycat Sovereign ---------------------------------------------------

func TestS82Gap181SkycatSovereignCountsOtherFliersAndMakesCatBirds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := effectstest.B12Push(g, me.ID, "Skycat Sovereign", "Creature — Elemental Cat", gap181SkycatOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	gap181Creature(g, me.ID, "Flier", 1, 1, "flying")
	gap181Creature(g, me.ID, "Another Flier", 1, 1, "flying")
	gap181Creature(g, me.ID, "Walker", 1, 1)
	gap181Creature(g, opp.ID, "Their Flier", 1, 1, "flying")
	if p, tough := effectstest.EffectivePower(t, g, cat), effectstest.EffectiveToughness(t, g, cat); p != 3 || tough != 3 {
		t.Errorf("Skycat is %d/%d with two other fliers, want 3/3", p, tough)
	}
	gap181Fund(me, "C", "C", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Cat Bird", me.ID); n != 1 {
		t.Fatalf("%d Cat Bird tokens, want 1", n)
	}
	if p := effectstest.EffectivePower(t, g, cat); p != 4 {
		t.Errorf("Skycat has power %d with the token, want 4", p)
	}
}

func TestS82Gap181SkycatSovereignRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.B12Push(g, me.ID, "Skycat Sovereign", "Creature — Elemental Cat", gap181SkycatOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	gap181Fund(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the token ability was activated without {U}")
	}
}

// --- Jubilant Skybonder -------------------------------------------------

func TestS82Gap181JubilantSkybonderTaxesSpellsAtYourFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := effectstest.B12Push(g, me.ID, "Jubilant Skybonder", "Creature — Human Wizard", gap181SkybonderOracle, 2, 2)
	flier := gap181Creature(g, me.ID, "Flier", 1, 1, "flying")
	bear := effectstest.SeedBear(g, me.ID)
	theirs := gap181Creature(g, opp.ID, "Their Flier", 1, 1, "flying")
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }

	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(flier)); got != 3 {
		t.Errorf("opponent's spell at my flier costs %d, want 3", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(sky)); got != 3 {
		t.Errorf("opponent's spell at the Skybonder costs %d, want 3", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(bear)); got != 1 {
		t.Errorf("opponent's spell at my non-flier costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(theirs)); got != 1 {
		t.Errorf("opponent's spell at their own flier costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, me, at(flier)); got != 1 {
		t.Errorf("my own spell at my flier costs %d, want 1", got)
	}
}

// --- Kangee's Lieutenant ------------------------------------------------

func TestS82Gap181KangeesLieutenantPumpsAttackingFliersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lt := effectstest.B12Push(g, me.ID, "Kangee's Lieutenant", "Creature — Bird Soldier", gap181LieutenantOracle, 1, 1)
	flier := gap181Creature(g, me.ID, "Flier", 2, 2, "flying")
	walker := gap181Creature(g, me.ID, "Walker", 2, 2)
	home := gap181Creature(g, me.ID, "Home Flier", 2, 2, "flying")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, lt, flier, walker)
	effectstest.PassPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"Lieutenant": {lt, 2}, "attacking flier": {flier, 3}, "attacking walker": {walker, 2}, "flier that stayed home": {home, 2}} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s has power %d, want %d", name, got, c.want)
		}
	}
}

func TestS82Gap181KangeesLieutenantEncoresFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Kangee's Lieutenant", "Creature — Bird Soldier",
		gap181LieutenantOracle, 1, 1, "{C}{C}{C}{C}{C}{W}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Error("encore did not exile the card")
	}
}
