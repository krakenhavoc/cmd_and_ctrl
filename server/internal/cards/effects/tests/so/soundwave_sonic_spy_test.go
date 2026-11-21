package testsso

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const soundwaveOracle = "cc18e782-a965-404f-96f1-348f3ed799fd"

func soundwaveCard(t *testing.T, g *game.Game, me *game.Player, converted bool) uuid.UUID {
	t.Helper()
	row := effectstest.TransformerRow(soundwaveOracle,
		effectstest.DisturbFace{Name: "Soundwave, Sonic Spy", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{1}{W}{U}{B}", Power: "5", Toughness: "4"},
		effectstest.DisturbFace{Name: "Soundwave, Superior Captain", TypeLine: "Legendary Artifact", Cost: ""},
		"More Than Meets the Eye {2}{W}{U}{B}\nWhenever one or more creature tokens you control deal combat damage to a player, exile target instant or sorcery card with mana value equal to the damage dealt from their graveyard. Copy it. You may cast the copy without paying its mana cost. If you do, convert Soundwave.",
		"Whenever you cast a spell with an odd mana value, convert Soundwave. If you do, create Ravage, a legendary 3/3 black Robot artifact creature token with menace and deathtouch.\nWhenever you cast a spell with an even mana value, convert Soundwave. If you do, create Laserbeak, a legendary 2/2 blue Robot artifact creature token with flying and hexproof.",
		"Convert")
	if converted {
		return effectstest.TransformerConverted(t, g, row, me)
	}
	return effectstest.TransformerIn(t, g, row, me)
}

func TestSoundwaveIsFull(t *testing.T) {
	for _, id := range []string{soundwaveOracle, soundwaveOracle + "#1"} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Fatalf("%s: registered=%v spec=%+v", id, ok, spec)
		}
	}
}

func soundwaveToken(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Creature(g, me.ID, "Test Robot", "Token Artifact Creature — Robot", 2, 2)
}

func soundwaveGraveSpell(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}

// Two tokens deal 1 and 2: the damage dealt is 3, so the mana value 3
// instant is the target; casting the copy converts Soundwave.
func TestSoundwaveFrontExilesCopiesAndConverts(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	sw := soundwaveCard(t, g, me, false)
	two := soundwaveGraveSpell(me, "Two Drop", "{1}{U}")
	three := soundwaveGraveSpell(me, "Three Drop", "{1}{U}{U}")
	a, b := soundwaveToken(g, me), soundwaveToken(g, me)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: a, Target: opp.ID, Amount: 1, Combat: true})
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: b, Target: opp.ID, Amount: 2, Combat: true})
			return nil
		})
	})
	g.RunStateChecksForTest()
	if err := effectstest.SgAnswerTargets(t, g, three); err != nil {
		t.Fatalf("targeting the mana value 3 card: %v", err)
	}
	effectstest.TfSettle(t, g)
	if ch := effectstest.ResolutionCastOfferFor(g, me.ID); ch == nil {
		t.Fatalf("no copy offer; exile has three=%v two=%v", g.Exile.Contains(three), g.Exile.Contains(two))
	}
	if !g.Exile.Contains(three) || g.Exile.Contains(two) {
		t.Fatalf("exile: three=%v two=%v, want only the mana value 3 card", g.Exile.Contains(three), g.Exile.Contains(two))
	}
	var copyID uuid.UUID
	for _, id := range effectstest.ResolutionCastOfferFor(g, me.ID).DiscardOptions {
		copyID = id
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, copyID, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the copy: %v", err)
	}
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, sw, 1, "Soundwave, Superior Captain")
}

// Declining the copy is "if you do" failing: Soundwave stays put.
func TestSoundwaveFrontDecliningTheCopyDoesNotConvert(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	sw := soundwaveCard(t, g, me, false)
	two := soundwaveGraveSpell(me, "Two Drop", "{1}{U}")
	soundwaveGraveSpell(me, "Three Drop", "{1}{U}{U}")
	a := soundwaveToken(g, me)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: a, Target: opp.ID, Amount: 2, Combat: true})
			return nil
		})
	})
	g.RunStateChecksForTest()
	if err := effectstest.SgAnswerTargets(t, g, two); err != nil {
		t.Fatalf("targeting the mana value 2 card: %v", err)
	}
	effectstest.TfSettle(t, g)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, sw, 0, "Soundwave, Sonic Spy")
}

// A nontoken creature's damage does not trigger it.
func TestSoundwaveFrontIgnoresNontokenCreatures(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	soundwaveCard(t, g, me, false)
	soundwaveGraveSpell(me, "Two Drop", "{1}{U}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Amount: 2, Combat: true})
			return nil
		})
	})
	g.RunStateChecksForTest()
	effectstest.TfSettle(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil || len(g.Stack.Cards) != 0 {
		t.Fatal("a nontoken creature's damage triggered Soundwave")
	}
}

func soundwaveCast(t *testing.T, g *game.Game, me *game.Player, cost string) {
	t.Helper()
	effectstest.FloatMana(t, g, me, "{W}{W}{W}{W}")
	effectstest.CastFromHandForTest(t, g, me, "Probe", "Instant", cost, "", game.CastSpellParams{})
	effectstest.TfSettle(t, g)
}

func TestSoundwaveBackOddCastConvertsAndMakesRavage(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	sw := soundwaveCard(t, g, me, true)
	soundwaveCast(t, g, me, "{W}")
	effectstest.WantFace(t, g, sw, 0, "Soundwave, Sonic Spy")
	r := effectstest.BattlefieldTokenNamed(t, g, "Ravage")
	if r.Power != 3 || r.Toughness != 3 || !slices.Contains(r.Keywords, "menace") || !slices.Contains(r.Keywords, "deathtouch") {
		t.Fatalf("Ravage = %+v", r)
	}
}

func TestSoundwaveBackEvenCastConvertsAndMakesLaserbeak(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	sw := soundwaveCard(t, g, me, true)
	soundwaveCast(t, g, me, "{W}{W}")
	effectstest.WantFace(t, g, sw, 0, "Soundwave, Sonic Spy")
	l := effectstest.BattlefieldTokenNamed(t, g, "Laserbeak")
	if l.Power != 2 || l.Toughness != 2 || !slices.Contains(l.Keywords, "flying") || !slices.Contains(l.Keywords, "hexproof") {
		t.Fatalf("Laserbeak = %+v", l)
	}
}
