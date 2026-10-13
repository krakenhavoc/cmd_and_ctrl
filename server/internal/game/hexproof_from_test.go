package game

import (
	"testing"

	"github.com/google/uuid"
)

// hexproof_from_test.go — #3172, ADR 0038's amendment of 2026-10-13:
// CR 702.11e "hexproof from [quality]" at the targeting choke point.
// The headline asymmetry: it asks the SOURCE's quality (protection's
// question) and only refuses an OPPONENT's source (hexproof's).

func TestParseHexproofFromQuality(t *testing.T) {
	good := map[string]ProtectionQualityKind{
		"hexproof from black":         ProtectionQualityColor,
		"Hexproof from Blue":          ProtectionQualityColor,
		"hexproof from planeswalkers": ProtectionQualityCardType,
		"hexproof from monocolored":   ProtectionQualityMonocolored,
		"hexproof from each color":    ProtectionQualityAnyColor,
		"hexproof from all colors":    ProtectionQualityAnyColor,
		"hexproof from Demons":        ProtectionQualitySubtype,
	}
	for tok, kind := range good {
		q, ok := ParseHexproofFromQuality(tok)
		if !ok || q.Kind != kind {
			t.Errorf("ParseHexproofFromQuality(%q) = %+v, %v; want kind %v", tok, q, ok, kind)
		}
	}
	// Refused: the bare keyword (ADR 0038 §6 — never substitute the
	// broad one), a quality outside the grammar, and the protection-only
	// qualities.
	for _, tok := range []string{
		"hexproof", "hexproof from ", "hexproof from nonsense",
		"hexproof from everything", "hexproof from the chosen player",
		"hexproof from mana value 3 or less", "protection from red",
	} {
		if q, ok := ParseHexproofFromQuality(tok); ok {
			t.Errorf("ParseHexproofFromQuality(%q) = %+v, want refused", tok, q)
		}
	}
	// And protection's grammar still does not grow the hexproof-only kinds.
	for _, tok := range []string{"protection from monocolored", "protection from each color"} {
		if q, ok := ParseProtectionQuality(tok); ok {
			t.Errorf("ParseProtectionQuality(%q) = %+v, want refused (ADR 0072 §10)", tok, q)
		}
	}
}

func hexproofFromCreature(owner *Player, tokens ...string) *Card {
	c := NewCard("Hexproof-From Bear", owner.ID)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	c.Keywords = tokens
	return &c
}

// The matrix of qualities against sources, asked straight at the
// choke point. An opponent's source with the quality is refused; the
// same source from the creature's own controller is not.
func TestHexproofFromRefusesAnOpponentsSourceWithTheQuality(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	black := &Characteristic{Colors: []string{"B"}, Types: []string{"Instant"}}
	blue := &Characteristic{Colors: []string{"U"}, Types: []string{"Instant"}}
	gold := &Characteristic{Colors: []string{"U", "B"}, Types: []string{"Instant"}}
	colourless := &Characteristic{Types: []string{"Artifact"}}
	walker := &Characteristic{Colors: []string{"W"}, Types: []string{"Planeswalker"}}

	cases := []struct {
		name    string
		token   string
		src     *Characteristic
		refused bool
	}{
		{"black refuses a black spell", "hexproof from black", black, true},
		{"black lets a blue spell through", "hexproof from black", blue, false},
		{"black refuses a gold black spell", "hexproof from black", gold, true},
		{"black lets a colourless source through", "hexproof from black", colourless, false},
		{"monocolored refuses a mono spell", "hexproof from monocolored", black, true},
		{"monocolored lets a gold spell through", "hexproof from monocolored", gold, false},
		{"monocolored lets a colourless source through", "hexproof from monocolored", colourless, false},
		{"each color refuses any coloured source", "hexproof from each color", blue, true},
		{"each color refuses a gold source", "hexproof from each color", gold, true},
		{"each color lets a colourless source through", "hexproof from each color", colourless, false},
		{"planeswalkers refuses a planeswalker's ability", "hexproof from planeswalkers", walker, true},
		{"planeswalkers lets an instant through", "hexproof from planeswalkers", black, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := hexproofFromCreature(opp, tc.token)
			c.Controller = opp.ID
			g.Battlefield.PushTop(*c)
			defer func() { g.Battlefield.Remove(c.InstanceID) }()
			var got bool
			g.WithWriteLock(func() {
				got = !g.canBeTargetedByLocked(c, ZoneBattlefield, TargetSource{Controller: me.ID, Snapshot: tc.src})
			})
			if got != tc.refused {
				t.Errorf("opponent's source: refused = %v, want %v", got, tc.refused)
			}
			// The creature's own controller is never refused by hexproof.
			g.WithWriteLock(func() {
				got = !g.canBeTargetedByLocked(c, ZoneBattlefield, TargetSource{Controller: opp.ID, Snapshot: tc.src})
			})
			if got {
				t.Error("the creature's own controller must be able to target it")
			}
		})
	}
}

// An unknown source (no snapshot) matches no quality: errs weaker.
func TestHexproofFromWithAnUnknownSourceRefusesNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := hexproofFromCreature(opp, "hexproof from each color")
	c.Controller = opp.ID
	var refused bool
	g.WithWriteLock(func() {
		refused = !g.canBeTargetedByLocked(c, ZoneBattlefield, TargetSource{Controller: me.ID})
	})
	if refused {
		t.Error("a source with no characteristics has no quality to refuse")
	}
}

// End to end through CastSpell: the announce gate refuses the black
// spell and accepts the blue one, and the enumerator never offers it.
func TestHexproofFromIsEnforcedAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const oracle = "test-hexproof-from-gate"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return anyCreatureSpec()
		}
		return nil
	})
	theirs := pushProtectedCreature(g, opp, "Their Bear", "hexproof from black")
	ref := TargetRef{Kind: TargetCard, ID: theirs}

	if err := castColouredSpellAt(t, g, me, oracle, []string{"B"}, ref); err != ErrIllegalTarget {
		t.Fatalf("black spell at hexproof-from-black: got %v, want ErrIllegalTarget", err)
	}
	if err := castColouredSpellAt(t, g, me, oracle, []string{"U"}, ref); err != nil {
		t.Fatalf("blue spell at hexproof-from-black must be legal: %v", err)
	}
}

// CR 702.11d/e on a PLAYER (Veil of Summer): refuses an opponent's
// black spell, lets a red one and the player's own through.
func TestHexproofFromOnAPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	const bolt = "test-player-hexproof-from-bolt"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me.ID, "hexproof from black", "Test", uuid.Nil, Duration{})
		g.GrantPlayerStaticForEffect(me.ID, "hexproof from blue", "Test", uuid.Nil, Duration{})
	})
	if err := castAtPlayer(t, g, opp, bolt, []string{"B"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's black spell at the player: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, opp, bolt, []string{"U"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's blue spell at the player: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != nil {
		t.Fatalf("a red spell at hexproof from black and blue: %v", err)
	}
	if err := castAtPlayer(t, g, me, bolt, []string{"B"}, me.ID); err != nil {
		t.Fatalf("your own black spell at yourself: %v", err)
	}
}

// "As though it didn't have hexproof" waives the variant (#1560 A2
// extended): Glaring Spotlight lets its controller's spells through.
func TestHexproofWaiverAlsoWaivesHexproofFrom(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	withHexproofBypasses(t, opponentsCreaturesWaiver(true))
	theirs := pushProtectedCreature(g, opp, "Their Bear", "hexproof from each color")
	src := TargetSource{Controller: me.ID, Snapshot: &Characteristic{Colors: []string{"R"}}}
	var before, after bool
	ask := func() bool {
		var ok bool
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if c := &g.Battlefield.Cards[i]; c.InstanceID == theirs {
					ok = g.canBeTargetedByLocked(c, ZoneBattlefield, src)
				}
			}
		})
		return ok
	}
	before = ask()
	pushWaiver(g, me)
	after = ask()
	if before || !after {
		t.Errorf("before the waiver targetable = %v (want false), after = %v (want true)", before, after)
	}
}

// "Loses hexproof and can't have hexproof" takes the variants too.
func TestCantHaveHexproofStripsHexproofFrom(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{CantHaveKeywordsMod("hexproof")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{AddKeywordsMod("hexproof from black", "trample")}, IndefiniteDuration())
	got := abilitiesOf(t, g, id)
	if hasAbility(got, "hexproof from black") {
		t.Errorf("abilities = %v: a hexproof variant must be stripped by can't-have hexproof", got)
	}
	if !hasAbility(got, "trample") {
		t.Errorf("abilities = %v: only hexproof and its variants are refused", got)
	}
}
