package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	neoB08Naomi       = "cd6018de-5791-4c13-b134-38c410f04872"
	neoB08Kami        = "d7d8784f-8d2d-4f7b-a6fb-5c841936d9e5"
	neoB08Young       = "1f27aa69-d17e-4f1f-9aba-cee205036f93"
	neoB08Okiba       = "142409ce-1187-4554-bc07-6432c4d22dbc"
	neoB08Slash       = "2800830c-73c6-4097-be7b-d698c6813c38"
	neoB08Transfer    = "1e46c709-5278-44f3-9f0e-f71bd9558339"
	neoB08Ink         = "a6d95d26-48d8-4a97-9a42-f492a61300c9"
	neoB08Oath        = "4a4a061d-bc4e-49da-8d20-90fcb571b025"
	neoB08Uprising    = "fe2fe9f2-908b-49a9-9648-b01e0417a364"
	neoB08Synthesizer = "11db299a-0918-46a3-b39b-16acc3f4e281"
	neoB08SkyBlessed  = "dad767d4-e7d4-4fe0-8a18-bb434c5268eb"
	neoB08Heist       = "569da5d3-6af2-4021-a464-b30ebaaf6fa0"
)

func TestNeoB08CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		neoB08Naomi: "Naomi, Pillar of Order", neoB08Kami: "Kami of Terrible Secrets",
		neoB08Young: "When We Were Young", neoB08Okiba: "Okiba Salvage",
		neoB08Slash: "Banishing Slash", neoB08Transfer: "Soul Transfer",
		neoB08Ink: "Assassin's Ink", neoB08Oath: "Imperial Oath",
		neoB08Uprising: "Eiganjo Uprising", neoB08Synthesizer: "Experimental Synthesizer",
		neoB08SkyBlessed: "Sky-Blessed Samurai", neoB08Heist: "Reality Heist",
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

// neoB08Board seeds the artifact and/or the enchantment for `p`.
func neoB08Board(g *game.Game, p *game.Player, artifact, enchantment bool) (art, ench uuid.UUID) {
	if artifact {
		art = b12Permanent(g, p.ID, "Test Rock", "Artifact")
	}
	if enchantment {
		ench = b12Permanent(g, p.ID, "Test Aura", "Enchantment")
	}
	return art, ench
}

func neoB08Samurai(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range tokensNamed(g, "Samurai") {
		if c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func neoB08Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	advanceToMain(t, g)
	return castFromHandErr(g, p, name, typeLine, cost, oracle, params)
}

func TestNeoB08SamuraiTokenIsPrinted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	neoB08Board(g, me, true, true)
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Imperial Oath", "Sorcery", "", neoB08Oath, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	sams := neoB08Samurai(g, me.ID)
	if len(sams) != 3 {
		t.Fatalf("Samurai tokens = %d, want 3", len(sams))
	}
	for _, c := range sams {
		if c.TypeLine != "Token Creature — Samurai" || c.Power != 2 || c.Toughness != 2 || !hasKeywordOf(g, c.InstanceID, "vigilance") {
			t.Errorf("token = %+v", c)
		}
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("Imperial Oath did not ask to scry 3")
	}
}

func TestNeoB08NaomiEnters(t *testing.T) {
	for _, tc := range []struct {
		name      string
		art, ench bool
		want      int
	}{
		{"neither", false, false, 0},
		{"artifact only", true, false, 0},
		{"enchantment only", false, true, 0},
		{"both", true, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			neoB08Board(g, me, tc.art, tc.ench)
			if _, err := neoB08Cast(t, g, me, "Naomi, Pillar of Order", "Legendary Creature — Human Advisor", "", neoB08Naomi, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if n := len(neoB08Samurai(g, me.ID)); n != tc.want {
				t.Errorf("%d Samurai, want %d", n, tc.want)
			}
		})
	}
}

func TestNeoB08NaomiAttacks(t *testing.T) {
	for _, both := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		neoB08Board(g, me, true, both)
		id := b12Push(g, me.ID, "Naomi, Pillar of Order", "Legendary Creature — Human Advisor", neoB08Naomi, 4, 4)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		want := 0
		if both {
			want = 1
		}
		if n := len(neoB08Samurai(g, me.ID)); n != want {
			t.Errorf("both=%v: %d Samurai, want %d", both, n, want)
		}
	}
}

func TestNeoB08KamiDrawsAndGainsOnlyWithBoth(t *testing.T) {
	for _, both := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		neoB08Board(g, me, true, both)
		hornTopFive(me, hornFive)
		hand, life := me.Hand.Size(), me.Life
		_, err := neoB08Cast(t, g, me, "Kami of Terrible Secrets", "Creature — Spirit", "", neoB08Kami, game.CastSpellParams{Strict: true})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		// The Kami is put in hand by the cast helper and leaves it again.
		wantHand, wantLife := hand, life
		if both {
			wantHand, wantLife = hand+1, life+1
		}
		if me.Hand.Size() != wantHand || me.Life != wantLife {
			t.Errorf("both=%v: hand %d->%d (want %d), life %d->%d (want %d)", both, hand, me.Hand.Size(), wantHand, life, me.Life, wantLife)
		}
	}
}

func TestNeoB08SoulTransferModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	both := append(cardRefs(victim), cardRefs(dead)...)

	// Without an artifact and an enchantment, choosing both is refused.
	if _, err := neoB08Cast(t, g, me, "Soul Transfer", "Sorcery", "", neoB08Transfer, game.CastSpellParams{Modes: []int{0, 1}, Targets: both, Strict: true}); err == nil {
		t.Fatal("both modes were allowed without an artifact and an enchantment")
	}
	// With them, both resolve in printed order.
	neoB08Board(g, me, true, true)
	if _, err := neoB08Cast(t, g, me, "Soul Transfer", "Sorcery", "", neoB08Transfer, game.CastSpellParams{Modes: []int{0, 1}, Targets: both, Strict: true}); err != nil {
		t.Fatalf("cast both: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, victim) != game.ZoneExile {
		t.Errorf("victim in %q, want exile", b12ZoneOf(g, victim))
	}
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to hand")
	}
}

func TestNeoB08SoulTransferOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if _, err := neoB08Cast(t, g, me, "Soul Transfer", "Sorcery", "", neoB08Transfer, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, victim) != game.ZoneExile {
		t.Errorf("victim in %q, want exile", b12ZoneOf(g, victim))
	}
	// A land is no creature or planeswalker.
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if _, err := neoB08Cast(t, g, me, "Soul Transfer", "Sorcery", "", neoB08Transfer, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(land), Strict: true}); err == nil {
		t.Error("a land was a legal target")
	}
}

func TestNeoB08AssassinsInkCostsLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	price := func() int {
		return selfPricedMV(t, g, me, neoB08Ink, "Instant", "{2}{B}{B}")
	}
	if got := price(); got != 4 {
		t.Fatalf("bare price %d, want 4", got)
	}
	b12Permanent(g, me.ID, "Test Rock", "Artifact")
	if got := price(); got != 3 {
		t.Errorf("with an artifact: %d, want 3", got)
	}
	b12Permanent(g, me.ID, "Test Aura", "Enchantment")
	if got := price(); got != 2 {
		t.Errorf("with both: %d, want 2", got)
	}
	// A second artifact does not discount more: it is "if", not "for each".
	b12Permanent(g, me.ID, "Another Rock", "Artifact")
	if got := price(); got != 2 {
		t.Errorf("with two artifacts and an enchantment: %d, want 2", got)
	}
}

func TestNeoB08AssassinsInkOneEnchantmentArtifactGivesBoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Shrine Relic", "Enchantment Artifact")
	if got := selfPricedMV(t, g, me, neoB08Ink, "Instant", "{2}{B}{B}"); got != 2 {
		t.Errorf("one Enchantment Artifact: %d, want 2", got)
	}
}

func TestNeoB08AssassinsInkDestroysCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	walker := b12Permanent(g, opp.ID, "Walker", "Planeswalker — Test")
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if _, err := neoB08Cast(t, g, me, "Assassin's Ink", "Instant", "", neoB08Ink, game.CastSpellParams{Targets: cardRefs(land), Strict: true}); err == nil {
		t.Error("a land was a legal target")
	}
	for _, id := range []uuid.UUID{victim, walker} {
		if _, err := neoB08Cast(t, g, me, "Assassin's Ink", "Instant", "", neoB08Ink, game.CastSpellParams{Targets: cardRefs(id), Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(id) {
			t.Errorf("%s was not destroyed", id)
		}
	}
}

func TestNeoB08BanishingSlash(t *testing.T) {
	t.Run("tapped creature with both makes a Samurai", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		neoB08Board(g, me, true, true)
		victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == victim {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		if _, err := neoB08Cast(t, g, me, "Banishing Slash", "Sorcery", "", neoB08Slash, game.CastSpellParams{Targets: cardRefs(victim), Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(victim) || len(neoB08Samurai(g, me.ID)) != 1 {
			t.Errorf("destroyed=%v, Samurai=%d", opp.Graveyard.Contains(victim), len(neoB08Samurai(g, me.ID)))
		}
	})
	t.Run("an untapped creature is not a legal target", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		if _, err := neoB08Cast(t, g, me, "Banishing Slash", "Sorcery", "", neoB08Slash, game.CastSpellParams{Targets: cardRefs(victim), Strict: true}); err == nil {
			t.Error("an untapped creature was targeted")
		}
	})
	t.Run("destroying your only artifact turns the token off", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		art, _ := neoB08Board(g, me, true, true)
		if _, err := neoB08Cast(t, g, me, "Banishing Slash", "Sorcery", "", neoB08Slash, game.CastSpellParams{Targets: cardRefs(art), Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if len(neoB08Samurai(g, me.ID)) != 0 {
			t.Error("a Samurai was made with the artifact gone")
		}
	})
	t.Run("no target, with both, still makes the token", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		neoB08Board(g, me, true, true)
		if _, err := neoB08Cast(t, g, me, "Banishing Slash", "Sorcery", "", neoB08Slash, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if len(neoB08Samurai(g, me.ID)) != 1 {
			t.Error("no Samurai")
		}
	})
}

func TestNeoB08WhenWeWereYoung(t *testing.T) {
	for _, both := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		neoB08Board(g, me, true, both)
		a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
		b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
		if _, err := neoB08Cast(t, g, me, "When We Were Young", "Instant", "", neoB08Young, game.CastSpellParams{Targets: cardRefs(a, b), Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p, tt := effectivePTOf(t, g, a); p != 4 || tt != 4 {
			t.Errorf("A is %d/%d, want 4/4", p, tt)
		}
		if p, tt := effectivePTOf(t, g, b); p != 3 || tt != 3 {
			t.Errorf("B is %d/%d, want 3/3", p, tt)
		}
		for _, id := range []uuid.UUID{a, b} {
			if got := hasKeywordOf(g, id, "lifelink"); got != both {
				t.Errorf("both=%v: lifelink %v", both, got)
			}
		}
	}
}

func TestNeoB08OkibaSalvage(t *testing.T) {
	for _, tc := range []struct {
		name      string
		art, ench bool
		typeLine  string
		want      int
	}{
		{"creature, neither", false, false, "Creature — Bear", 0},
		{"creature, both", true, true, "Creature — Bear", 2},
		{"creature, enchantment only", false, true, "Creature — Bear", 0},
		{"Vehicle is itself the artifact", false, true, "Artifact — Vehicle", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			neoB08Board(g, me, tc.art, tc.ench)
			card := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Returned", TypeLine: tc.typeLine, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			if _, err := neoB08Cast(t, g, me, "Okiba Salvage", "Sorcery", "", neoB08Okiba, game.CastSpellParams{Targets: cardRefs(card), Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if b12ZoneOf(g, card) != game.ZoneBattlefield {
				t.Fatalf("card in %q, want the battlefield", b12ZoneOf(g, card))
			}
			if got := b12Counter(t, g, card, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNeoB08OkibaSalvageRefusesANonCreatureNonVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if _, err := neoB08Cast(t, g, me, "Okiba Salvage", "Sorcery", "", neoB08Okiba, game.CastSpellParams{Targets: cardRefs(card), Strict: true}); err == nil {
		t.Error("a plain artifact card was a legal target")
	}
}

func TestNeoB08ExperimentalSynthesizer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := hornTopFive(me, hornFive)
	id, err := neoB08Cast(t, g, me, "Experimental Synthesizer", "Artifact", "", neoB08Synthesizer, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top[0]) != game.ZoneExile || !b12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Fatalf("entering did not exile the top card playably (zone %q)", b12ZoneOf(g, top[0]))
	}
	// Leaving does it again: sacrifice for the Samurai at sorcery speed.
	b06AddMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(neoB08Samurai(g, me.ID)) != 1 {
		t.Error("no Samurai")
	}
	if b12ZoneOf(g, top[1]) != game.ZoneExile || !b12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Errorf("leaving did not exile the next card (zone %q)", b12ZoneOf(g, top[1]))
	}
}

func TestNeoB08ExperimentalSynthesizerIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hornTopFive(me, hornFive)
	id := b12Push(g, me.ID, "Experimental Synthesizer", "Artifact", neoB08Synthesizer, 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	b06AddMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestNeoB08EiganjoUprising(t *testing.T) {
	for _, tc := range []struct{ x, wantOpp int }{{0, 0}, {1, 0}, {3, 2}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b06AddMana(me, "R", "W")
		for i := 0; i < tc.x; i++ {
			b06AddMana(me, "C")
		}
		if _, err := castFromHandErr(g, me, "Eiganjo Uprising", "Sorcery", "{X}{R}{W}", neoB08Uprising, game.CastSpellParams{XValue: tc.x, Strict: true}); err != nil {
			t.Fatalf("X=%d cast: %v", tc.x, err)
		}
		passPriorityAroundTable(t, g)
		mine, theirs := neoB08Samurai(g, me.ID), neoB08Samurai(g, opp.ID)
		if len(mine) != tc.x || len(theirs) != tc.wantOpp {
			t.Fatalf("X=%d: mine %d theirs %d, want %d and %d", tc.x, len(mine), len(theirs), tc.x, tc.wantOpp)
		}
		for _, c := range mine {
			if !hasKeywordOf(g, c.InstanceID, "menace") || !hasKeywordOf(g, c.InstanceID, "haste") || !hasKeywordOf(g, c.InstanceID, "vigilance") {
				t.Errorf("X=%d: my token lacks menace, haste or vigilance", tc.x)
			}
		}
		for _, c := range theirs {
			if hasKeywordOf(g, c.InstanceID, "menace") || hasKeywordOf(g, c.InstanceID, "haste") || !hasKeywordOf(g, c.InstanceID, "vigilance") {
				t.Errorf("X=%d: an opponent's token has the wrong keywords", tc.x)
			}
		}
	}
}

func TestNeoB08AffinityPrices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, neoB08SkyBlessed, "Enchantment Creature — Human Samurai", "{6}{W}"); got != 7 {
		t.Fatalf("Sky-Blessed Samurai bare = %d, want 7", got)
	}
	if got := selfPricedMV(t, g, me, neoB08Heist, "Instant", "{5}{U}{U}"); got != 7 {
		t.Fatalf("Reality Heist bare = %d, want 7", got)
	}
	b12Permanent(g, me.ID, "Rock A", "Artifact")
	b12Permanent(g, me.ID, "Rock B", "Artifact")
	b12Permanent(g, me.ID, "Aura A", "Enchantment")
	if got := selfPricedMV(t, g, me, neoB08SkyBlessed, "Enchantment Creature — Human Samurai", "{6}{W}"); got != 6 {
		t.Errorf("Sky-Blessed Samurai with one enchantment = %d, want 6", got)
	}
	if got := selfPricedMV(t, g, me, neoB08Heist, "Instant", "{5}{U}{U}"); got != 5 {
		t.Errorf("Reality Heist with two artifacts = %d, want 5", got)
	}
}

func TestNeoB08SkyBlessedSamuraiFlies(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Sky-Blessed Samurai", "Enchantment Creature — Human Samurai", neoB08SkyBlessed, 4, 4)
	if !hasKeywordOf(g, id, "flying") {
		t.Error("no flying")
	}
}

func TestNeoB08RealityHeistTakesUpToTwoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cards := []struct{ name, typeLine string }{
		{"Rock A", "Artifact"}, {"Bear", "Creature — Bear"}, {"Rock B", "Artifact Creature — Myr"},
		{"Rock C", "Artifact"}, {"Island", "Basic Land — Island"}, {"Opt", "Instant"}, {"Bolt", "Instant"},
	}
	top := hornTopFive(me, cards)
	if _, err := neoB08Cast(t, g, me, "Reality Heist", "Instant", "", neoB08Heist, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil || pick.ChooseMax != 2 {
		t.Fatalf("prompt %+v, want a ceiling of 2", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0], top[1]}); err == nil {
		t.Error("a non-artifact was takeable")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0], top[3], top[2]}); err == nil {
		t.Error("three cards were takeable")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0], top[2]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[0]) || !me.Hand.Contains(top[2]) {
		t.Error("the two artifacts are not in hand")
	}
	if !inLibrary(g, me.ID, top[1]) || !inLibrary(g, me.ID, top[3]) {
		t.Error("the rest did not go to the library")
	}
}
