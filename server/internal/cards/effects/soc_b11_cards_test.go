package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_cards_test.go — Secrets of Strixhaven Commander, batch 11
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB11FieldHistorianOracle = "28b64167-b7ea-4b4f-8434-061298050366"
	socB11HistoryChaserOracle  = "5c40a8d4-6969-4d71-aba5-5c7e4109b7d5"
	socB11LoremasterOracle     = "c3d38129-b955-4fb5-8486-095f918935b9"
	socB11GuidanceOracle       = "ad3bf726-e0c2-4038-a66e-9a9dff439563"
	socB11CopseOracle          = "633a41f0-3889-49a0-b2fc-9baf59f4ddc2"
	socB11ArcOracle            = "c22e5fc6-9b4e-4197-b472-70d168eb4330"
	socB11StormsingerOracle    = "792c4d38-41e6-49eb-bd4e-de0a5ac8330f"
	socB11RetrieverOracle      = "689c41e9-800a-4998-bb04-74bf90141035"
	socB11ClericOracle         = "c40bc9f6-2b72-4ae1-b912-50075af59628"
)

func TestSocB11EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB11FieldHistorianOracle: "Quintorius, Field Historian",
		socB11HistoryChaserOracle:  "Quintorius, History Chaser",
		socB11LoremasterOracle:     "Quintorius, Loremaster",
		socB11GuidanceOracle:       "Raffine's Guidance",
		socB11CopseOracle:          "Rain-Slicked Copse",
		socB11ArcOracle:            "Redemption Arc",
		socB11StormsingerOracle:    "Redoubled Stormsinger",
		socB11RetrieverOracle:      "Relic Retriever",
		socB11ClericOracle:         "Remorseful Cleric",
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

// socB11Named lists the battlefield permanents with the name.
func socB11Named(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// socB11Graveyard puts a card of the type line in p's graveyard.
func socB11Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// socB11Token makes a token of the table row under `owner`'s control, as
// an effect would, so it counts as having entered this turn.
func socB11Token(t *testing.T, g *game.Game, owner uuid.UUID, row string) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Controller: owner, SourceCardID: uuid.Nil})
		if err := (CreateToken{Template: TokenCard(row), N: 1, Controller: owner}).Apply(ctx); err != nil {
			t.Fatalf("token %q: %v", row, err)
		}
	})
}

// socB11Move moves one card between zones of the same owner.
func socB11Move(t *testing.T, g *game.Game, owner, card uuid.UUID, from, to game.ZoneKind) {
	t.Helper()
	toOwner := owner
	if to == game.ZoneExile {
		toOwner = uuid.Nil
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: from, Owner: owner}, game.ZoneRef{Kind: to, Owner: toOwner}, card); err != nil {
		t.Fatalf("move: %v", err)
	}
}

// --- Rain-Slicked Copse ---

func TestSocB11RainSlickedCopseEntersTappedTapsForEitherColourAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Rain-Slicked Copse", socB11CopseOracle)
	top100AssertEnteredTapped(t, g, id, "Rain-Slicked Copse")
	for _, colour := range []string{"G", "U"} {
		me.ManaPool = nil
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		b28TapForMana(t, g, me.ID, id, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cycled, _ := cycleFromHand(t, g2, "Rain-Slicked Copse", "Land — Forest Island", socB11CopseOracle, "{C}{C}")
	if !me2.Graveyard.Contains(cycled) {
		t.Error("the cycled land is not in the graveyard")
	}
}

// --- Remorseful Cleric ---

func TestSocB11RemorsefulClericSacrificesToExileAPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cleric := b12Push(g, me.ID, "Remorseful Cleric", "Creature — Spirit Cleric", socB11ClericOracle, 2, 1)
	assertKeywords(t, g, cleric, "flying")
	a := socB11Graveyard(opp, "Opposing Dead", "Creature — Bear", "{1}{G}")
	b := socB11Graveyard(opp, "Opposing Spell", "Instant", "{R}")
	mine := socB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")

	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Sacrificing is a cost: the Cleric is gone before the ability resolves.
	if g.Battlefield.Contains(cleric) {
		t.Fatal("the Cleric was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || opp.Graveyard.Size() != 0 {
		t.Errorf("the opponent's graveyard was not exiled (%d cards left)", opp.Graveyard.Size())
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the controller's own graveyard was touched")
	}
}

// --- Relic Retriever ---

func socB11GraveyardCreature(p *game.Player, name string) uuid.UUID {
	return socB11Graveyard(p, name, "Creature — Bear", "{1}{G}")
}

func TestSocB11RelicRetrieverMakesATreasureAtAnyEndStepAfterACardLeftYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retriever := b12Push(g, me.ID, "Relic Retriever", "Creature — Spirit Monkey", socB11RetrieverOracle, 2, 1)
	assertKeywords(t, g, retriever, "first strike")

	// No card left the graveyard: no Treasure at the end step.
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Treasure"); got != 0 {
		t.Fatalf("%d Treasures with no card leaving the graveyard", got)
	}

	// A card leaves the graveyard during the opponent's turn: it is each end
	// step, so the opponent's end step makes one.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	dead := socB11GraveyardCreature(me, "Dead Bear")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneHand)
	advanceToStepOf(t, g, 1, game.StepEnd)
	if triggerOnStack(g, retriever) == nil {
		t.Fatal("no trigger at the opponent's end step after a card left the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("%d Treasures, want 1", got)
	}
	_ = opp
}

func TestSocB11RelicRetrieverIgnoresACardLeavingAnotherGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retriever := b12Push(g, me.ID, "Relic Retriever", "Creature — Spirit Monkey", socB11RetrieverOracle, 2, 1)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	theirs := socB11GraveyardCreature(opp, "Their Dead")
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	advanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOnStack(g, retriever) != nil {
		t.Error("the trigger fired for a card that left an opponent's graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures, want 0", got)
	}
}

// --- Raffine's Guidance ---

func socB11GuidanceInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Raffine's Guidance", TypeLine: "Enchantment — Aura",
		ManaCost: "{W}", OracleID: socB11GuidanceOracle, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	return id
}

func TestSocB11RaffinesGuidanceIsCastFromTheGraveyardForTwoAndWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	aura := socB11GuidanceInGraveyard(g, me)

	// The printed cost is not an option from the graveyard, and a pool short
	// of {2}{W} cannot pay the other.
	b06AddMana(me, "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "raffines-guidance-graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("cast from the graveyard with only {W}")
	}
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, FromZone: "graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("cast from the graveyard for its printed cost")
	}

	b06AddMana(me, "C", "C")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "raffines-guidance-graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast from the graveyard for {2}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not enter")
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tg)
	}
}

func TestSocB11RaffinesGuidanceCastFromHandCostsOnlyWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	aura := handCardFull(me, "Raffine's Guidance", "Enchantment — Aura", "{W}", socB11GuidanceOracle, []string{"W"})
	b06AddMana(me, "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast from hand for {W}: %v", err)
	}
}

// --- Redemption Arc ---

func TestSocB11RedemptionArcGivesIndestructibleGoadsAndExilesOnRequest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	arc := pushCatalogPermanent(g, me.ID, "Redemption Arc", "Enchantment — Aura", socB11ArcOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == arc {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
			}
		}
	})
	assertKeywords(t, g, bear, "indestructible")
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(bear)
		if !c.Goaded() {
			t.Error("the enchanted creature is not goaded")
		}
		var goaders []uuid.UUID
		for _, r := range c.Effective().AttackRequirements {
			if r.GoadedBy != uuid.Nil {
				goaders = append(goaders, r.GoadedBy)
			}
		}
		if len(goaders) == 0 {
			t.Error("no goad requirement")
		}
		for _, who := range goaders {
			if who != me.ID {
				t.Errorf("goaded by %v, want the Aura's controller", who)
			}
		}
	})

	// {1}{W}: exile enchanted creature (the Aura's controller pays).
	b06AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the enchanted creature was not exiled")
	}
}

func TestSocB11RedemptionArcRefusesTheAbilityWithoutTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	arc := pushCatalogPermanent(g, me.ID, "Redemption Arc", "Enchantment — Aura", socB11ArcOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == arc {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
			}
		}
	})
	b06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exile ability was paid with one mana")
	}
	if g.Exile.Contains(bear) {
		t.Error("the creature was exiled")
	}
}
