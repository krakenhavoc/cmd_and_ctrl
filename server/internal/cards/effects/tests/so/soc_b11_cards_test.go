package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB11EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB11FieldHistorianOracle: "Quintorius, Field Historian",
		effectstest.SocB11HistoryChaserOracle:  "Quintorius, History Chaser",
		effectstest.SocB11LoremasterOracle:     "Quintorius, Loremaster",
		effectstest.SocB11GuidanceOracle:       "Raffine's Guidance",
		effectstest.SocB11CopseOracle:          "Rain-Slicked Copse",
		effectstest.SocB11ArcOracle:            "Redemption Arc",
		effectstest.SocB11StormsingerOracle:    "Redoubled Stormsinger",
		effectstest.SocB11RetrieverOracle:      "Relic Retriever",
		effectstest.SocB11ClericOracle:         "Remorseful Cleric",
	}
	for oracle, name := range want {
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

// socB11Token makes a token of the table row under `owner`'s control, as
// an effect would, so it counts as having entered this turn.
func socB11Token(t *testing.T, g *game.Game, owner uuid.UUID, row string) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: owner, SourceCardID: uuid.Nil})
		if err := (effects.CreateToken{Template: effects.TokenCard(row), N: 1, Controller: owner}).Apply(ctx); err != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Rain-Slicked Copse", effectstest.SocB11CopseOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Rain-Slicked Copse")
	for _, colour := range []string{"G", "U"} {
		me.ManaPool = nil
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		effectstest.B28TapForMana(t, g, me.ID, id, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cycled, _ := effectstest.CycleFromHand(t, g2, "Rain-Slicked Copse", "Land — Forest Island", effectstest.SocB11CopseOracle, "{C}{C}")
	if !me2.Graveyard.Contains(cycled) {
		t.Error("the cycled land is not in the graveyard")
	}
}

// --- Remorseful Cleric ---

func TestSocB11RemorsefulClericSacrificesToExileAPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cleric := effectstest.B12Push(g, me.ID, "Remorseful Cleric", "Creature — Spirit Cleric", effectstest.SocB11ClericOracle, 2, 1)
	effectstest.AssertKeywords(t, g, cleric, "flying")
	a := effectstest.SocB11Graveyard(opp, "Opposing Dead", "Creature — Bear", "{1}{G}")
	b := effectstest.SocB11Graveyard(opp, "Opposing Spell", "Instant", "{R}")
	mine := effectstest.SocB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")

	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Sacrificing is a cost: the Cleric is gone before the ability resolves.
	if g.Battlefield.Contains(cleric) {
		t.Fatal("the Cleric was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || opp.Graveyard.Size() != 0 {
		t.Errorf("the opponent's graveyard was not exiled (%d cards left)", opp.Graveyard.Size())
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the controller's own graveyard was touched")
	}
}

func TestSocB11RelicRetrieverMakesATreasureAtAnyEndStepAfterACardLeftYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retriever := effectstest.B12Push(g, me.ID, "Relic Retriever", "Creature — Spirit Monkey", effectstest.SocB11RetrieverOracle, 2, 1)
	effectstest.AssertKeywords(t, g, retriever, "first strike")

	// No card left the graveyard: no Treasure at the end step.
	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != 0 {
		t.Fatalf("%d Treasures with no card leaving the graveyard", got)
	}

	// A card leaves the graveyard during the opponent's turn: it is each end
	// step, so the opponent's end step makes one.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	dead := effectstest.SocB11GraveyardCreature(me, "Dead Bear")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneHand)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepEnd)
	if effectstest.TriggerOnStack(g, retriever) == nil {
		t.Fatal("no trigger at the opponent's end step after a card left the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("%d Treasures, want 1", got)
	}
	_ = opp
}

func TestSocB11RelicRetrieverIgnoresACardLeavingAnotherGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	retriever := effectstest.B12Push(g, me.ID, "Relic Retriever", "Creature — Spirit Monkey", effectstest.SocB11RetrieverOracle, 2, 1)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	theirs := effectstest.SocB11GraveyardCreature(opp, "Their Dead")
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneHand)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	if effectstest.TriggerOnStack(g, retriever) != nil {
		t.Error("the trigger fired for a card that left an opponent's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasures, want 0", got)
	}
}

// --- Raffine's Guidance ---

func socB11GuidanceInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Raffine's Guidance", TypeLine: "Enchantment — Aura",
		ManaCost: "{W}", OracleID: effectstest.SocB11GuidanceOracle, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	return id
}

func TestSocB11RaffinesGuidanceIsCastFromTheGraveyardForTwoAndWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	aura := socB11GuidanceInGraveyard(g, me)

	// The printed cost is not an option from the graveyard, and a pool short
	// of {2}{W} cannot pay the other.
	effectstest.B06AddMana(me, "W")
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

	effectstest.B06AddMana(me, "C", "C")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "raffines-guidance-graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast from the graveyard for {2}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not enter")
	}
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tg)
	}
}

func TestSocB11RaffinesGuidanceCastFromHandCostsOnlyWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	aura := effectstest.HandCardFull(me, "Raffine's Guidance", "Enchantment — Aura", "{W}", effectstest.SocB11GuidanceOracle, []string{"W"})
	effectstest.B06AddMana(me, "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast from hand for {W}: %v", err)
	}
}

// --- Redemption Arc ---

func TestSocB11RedemptionArcGivesIndestructibleGoadsAndExilesOnRequest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	arc := effectstest.PushCatalogPermanent(g, me.ID, "Redemption Arc", "Enchantment — Aura", effectstest.SocB11ArcOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == arc {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
			}
		}
	})
	effectstest.AssertKeywords(t, g, bear, "indestructible")
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
	effectstest.B06AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the enchanted creature was not exiled")
	}
}

func TestSocB11RedemptionArcRefusesTheAbilityWithoutTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	arc := effectstest.PushCatalogPermanent(g, me.ID, "Redemption Arc", "Enchantment — Aura", effectstest.SocB11ArcOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == arc {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
			}
		}
	})
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exile ability was paid with one mana")
	}
	if g.Exile.Contains(bear) {
		t.Error("the creature was exiled")
	}
}
