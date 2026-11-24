package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b039_cards_test.go — S82 older sets, Modern Horizons 2 batch 39
// (#5501). Geyadrone Dihada, Carth the Lion and Dermotaxi are skipped.

const (
	oldB39Dakkon   = "335b5302-2272-405a-b2b8-3943b3df335f"
	oldB39SeaDrake = "44d1029f-604e-44d6-89f6-bd18344f5d79"
	oldB39Skill    = "5c5b308c-07d5-422a-8b45-b8b19cb967ca"
	oldB39Raptor   = "6af53351-ad10-4774-9b5c-7b58bea65bbc"
	oldB39Mirror   = "08aaafc7-5fc0-4c28-a361-2a22536ec080"
	oldB39Respite  = "8c888122-9d05-4641-931c-83f771ba9d09"
	oldB39Redact   = "c2e54e13-dad5-4d01-850f-a1735eea54ae"
)

func TestOldB39CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB39Dakkon: "Dakkon, Shadow Slayer", oldB39SeaDrake: "Sea Drake", oldB39Skill: "Blacksmith's Skill",
		oldB39Raptor: "Sanctuary Raptor", oldB39Mirror: "Void Mirror", oldB39Respite: "Blessed Respite",
		oldB39Redact: "Mystic Redaction",
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

func oldB39Lands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.B12Permanent(g, owner, "Forest", "Basic Land — Forest"))
	}
	return ids
}

func TestOldB39DakkonEntersWithLoyaltyEqualToYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB39Lands(g, me.ID, 4)
	oldB39Lands(g, opp.ID, 2)
	id := effectstest.CastCatalogSpell(t, g, "Dakkon, Shadow Slayer", "Legendary Planeswalker — Dakkon", oldB39Dakkon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Dakkon did not stay on the battlefield")
	}
	if got := effectstest.CountersOn(g, id, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4 (my lands only)", got)
	}
}

func TestOldB39DakkonAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB39Lands(g, me.ID, 6)
	id := effectstest.CastCatalogSpell(t, g, "Dakkon, Shadow Slayer", "Legendary Planeswalker — Dakkon", oldB39Dakkon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	// -3: exile target creature.
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Targets: effectstest.CardTargets(ogre)})
	if !g.Exile.Contains(ogre) {
		t.Error("-3 did not exile the creature")
	}
	if got := effectstest.CountersOn(g, id, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty after -3 is %d, want 3", got)
	}
	// Refuse a second activation in the same turn.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second loyalty ability was activated in one turn")
	}
}

func TestOldB39DakkonPlusOneSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB39Lands(g, me.ID, 3)
	id := effectstest.CastCatalogSpell(t, g, "Dakkon, Shadow Slayer", "Legendary Planeswalker — Dakkon", oldB39Dakkon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if effectstest.PendingOfKind(g, game.PendingChoiceSurveil) == nil {
		t.Error("no surveil prompt")
	}
	if got := effectstest.CountersOn(g, id, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestOldB39DakkonMinusSixPutsAnArtifactFromHandOrGraveyard(t *testing.T) {
	for _, zone := range []string{"hand", "graveyard"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oldB39Lands(g, me.ID, 7)
		id := effectstest.CastCatalogSpell(t, g, "Dakkon, Shadow Slayer", "Legendary Planeswalker — Dakkon", oldB39Dakkon, nil)
		effectstest.PassPriorityAroundTable(t, g)
		var rock uuid.UUID
		if zone == "hand" {
			rock = effectstest.B13HandCard(me, "Rock", "Artifact", "")
		} else {
			rock = effectstest.PushGraveyardCardTyped(me, "Rock", "Artifact")
		}
		effectstest.B16Activate(t, g, me.ID, id, 2, game.ActivateAbilityParams{})
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
		if c == nil {
			t.Fatalf("%s: no pick prompt", zone)
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{rock}); err != nil {
			t.Fatalf("%s: %v", zone, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(rock) {
			t.Errorf("%s: the artifact did not enter", zone)
		}
	}
}

func TestOldB39SeaDrakeReturnsTwoOfYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lands := oldB39Lands(g, me.ID, 3)
	other := effectstest.B12Permanent(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Sea Drake", "Creature — Drake", oldB39SeaDrake, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardTargets(lands[0], lands[1])); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range []game.ZoneKind{game.ZoneHand, game.ZoneHand, game.ZoneBattlefield} {
		if got := effectstest.B12ZoneOf(g, lands[i]); got != want {
			t.Errorf("land %d is in %v, want %v", i, got, want)
		}
	}
	if effectstest.B12ZoneOf(g, other) != game.ZoneBattlefield {
		t.Error("an opponent's land moved")
	}
}

func TestOldB39BlacksmithsSkill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	art := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blacksmith's Skill", "Instant", oldB39Skill, effectstest.CardTarget(art))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Blacksmith's Skill", "Instant", oldB39Skill, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		for _, id := range []uuid.UUID{art, bear} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%v lacks %s", id, kw)
			}
		}
	}
	if p := effectstest.EffectivePower(t, g, art); p != 4 {
		t.Errorf("the artifact creature has power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the plain creature has power %d, want 2", p)
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if effectstest.HasEffectiveKeyword(t, g, art, "hexproof") || effectstest.EffectivePower(t, g, art) != 2 {
		t.Error("the effect outlived the turn")
	}
}

func TestOldB39SanctuaryRaptorNeedsThreeTokens(t *testing.T) {
	for _, tokens := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		raptor := effectstest.B12Push(g, me.ID, "Sanctuary Raptor", "Artifact Creature — Bird", oldB39Raptor, 2, 1)
		for i := 0; i < tokens; i++ {
			effectstest.B12Push(g, me.ID, "Token", "Token Creature — Soldier", "", 1, 1)
		}
		effectstest.DeclareAttack(t, g, opp.ID, raptor)
		effectstest.PassPriorityAroundTable(t, g)
		wantPower, wantFS := 2, false
		if tokens >= 3 {
			wantPower, wantFS = 4, true
		}
		if p := effectstest.EffectivePower(t, g, raptor); p != wantPower {
			t.Errorf("%d tokens: power %d, want %d", tokens, p, wantPower)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, raptor, "first strike"); got != wantFS {
			t.Errorf("%d tokens: first strike %v, want %v", tokens, got, wantFS)
		}
	}
}

func TestOldB39VoidMirrorCountersASpellPaidWithoutColoredMana(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pool   string
		cost   string
		colors []string
		want   bool // countered
	}{
		{"colorless only", "CC", "{2}", nil, true},
		{"red mana", "CR", "{1}{R}", []string{"R"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedEquipment(g, g.Seats[1].ID, "Void Mirror", oldB39Mirror)
			g.Battlefield.Cards[len(g.Battlefield.Cards)-1].TypeLine = "Artifact"
			effectstest.FloatForTest(g, me, tc.pool)
			id := effectstest.HandCardFull(me, "Test Rock", "Artifact", tc.cost, "", tc.colors)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := game.ZoneBattlefield
			if tc.want {
				want = game.ZoneGraveyard
			}
			if got := effectstest.B12ZoneOf(g, id); got != want {
				t.Errorf("the spell is in %v, want %v (countered=%v)", got, want, tc.want)
			}
		})
	}
}

func TestOldB39BlessedRespiteShufflesAndPreventsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.PushGraveyardCardTyped(opp, "Buried", "Creature — Bear")
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Blessed Respite", "Instant", oldB39Respite, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) || opp.Library.Size() != lib+1 {
		t.Errorf("graveyard %v, library %d (was %d)", opp.Graveyard.Size(), opp.Library.Size(), lib)
	}
	if effectstest.ScopedReplacementCount(g) != 1 {
		t.Errorf("scoped replacements = %d, want 1 (the combat-damage prevention)", effectstest.ScopedReplacementCount(g))
	}
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	defender := effectstest.B12Creature(g, me.ID, "Defender", "Creature — Bear", 2, 2)
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.DamageOn(t, g, defender); got != 0 {
		t.Errorf("combat damage dealt: %d", got)
	}
}

func TestOldB39MysticRedactionMillsOnDiscardAndScriesAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Mystic Redaction", "Enchantment", oldB39Redact)
	card := effectstest.B13HandCard(me, "Discard Me", "Instant", "")
	before := opp.Library.Size()
	effectstest.LengDiscard(t, g, me, card)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 2 {
		t.Errorf("an opponent milled %d, want 2", got)
	}
	if got := me.Library.Size(); got == 0 {
		t.Error("my library was milled")
	}
	// An opponent's discard does nothing.
	other := effectstest.B13HandCard(opp, "Their Card", "Instant", "")
	before = opp.Library.Size()
	effectstest.LengDiscard(t, g, opp, other)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before {
		t.Error("an opponent's discard milled")
	}
}

func TestOldB39SeaDrakeWithOneLandReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lands := oldB39Lands(g, me.ID, 1)
	effectstest.CastCatalogSpell(t, g, "Sea Drake", "Creature — Drake", oldB39SeaDrake, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("asked for targets that cannot be chosen")
	}
	if effectstest.B12ZoneOf(g, lands[0]) != game.ZoneBattlefield {
		t.Error("the lone land was returned")
	}
}
