package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// serras_emissary_test.go — #2742: protection from a card type chosen
// as the permanent enters, for its controller and their creatures.

const serrasEmissaryOracle = "7a56c6e1-0509-4783-9b29-cf3163977166"

// enterEmissaryForTest puts Serra's Emissary onto the battlefield
// under `owner` and runs its as-enters question, leaving the prompt
// open.
func enterEmissaryForTest(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	spec, ok := Lookup(serrasEmissaryOracle)
	if !ok {
		t.Fatal("Serra's Emissary is not registered")
	}
	id := pushPermanentForTest(g, owner, "Serra's Emissary", serrasEmissaryOracle, "Creature — Angel")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				if err := spec.AsEnters(&g.Battlefield.Cards[i], NewContext(g, nil)); err != nil {
					t.Fatalf("AsEnters: %v", err)
				}
			}
		}
	})
	return id
}

func hasProtectionToken(g *game.Game, id uuid.UUID, token string) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				for _, q := range game.ProtectionQualities(&c) {
					if q.Token() == token {
						found = true
					}
				}
			}
		}
	})
	return found
}

// TestSerrasEmissaryOffersEveryCardTypeAndGrantsNothingUnanswered — the
// question offers the nine card types, and while it is open neither
// half of the protection exists.
func TestSerrasEmissaryOffersEveryCardTypeAndGrantsNothingUnanswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	enterEmissaryForTest(t, g, me.ID)

	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatal("no card-type question was asked")
	}
	if got := optionWords(c); !slices.Equal(got, game.ChoosableCardTypes) {
		t.Errorf("offered %v, want %v", got, game.ChoosableCardTypes)
	}
	for _, ct := range game.ChoosableCardTypes {
		if game.ProtectionFromCardType(ct) == "" {
			t.Errorf("%s is offered but makes no protection token", ct)
		} else if _, ok := game.ParseProtectionQuality(game.ProtectionFromCardType(ct)); !ok {
			t.Errorf("%s's token %q does not parse", ct, game.ProtectionFromCardType(ct))
		}
	}
	if got := playerAbilities(g, me); len(got) != 0 {
		t.Errorf("unanswered, the player has %v", got)
	}
	if hasProtectionToken(g, bear, "protection from creatures") {
		t.Error("unanswered, a creature already has protection")
	}
}

// TestSerrasEmissaryProtectsYouAndYourCreaturesFromTheChosenType —
// choose Instant: an opponent's Bolt can target neither the player nor
// their creatures, a creature source still deals the player damage,
// the opponent's own creatures are not protected, and it all ends when
// the Emissary leaves.
func TestSerrasEmissaryProtectsYouAndYourCreaturesFromTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	emissary := enterEmissaryForTest(t, g, opp.ID)
	answerAnchorWord(t, g, opp.ID, "Instant")

	if got := playerAbilities(g, opp); !hasPlayerAbility(got, "protection from instants") {
		t.Fatalf("the Emissary's controller has %v, want protection from instants", got)
	}
	for _, id := range []uuid.UUID{bear, emissary} {
		if !hasProtectionToken(g, id, "protection from instants") {
			t.Errorf("creature %v of the Emissary's controller is not protected from instants", id)
		}
	}
	if hasProtectionToken(g, mine, "protection from instants") {
		t.Error("an opponent's creature is protected too")
	}

	if err := boltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("Bolt at the protected player: got %v, want ErrIllegalTarget", err)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != game.ErrIllegalTarget {
		t.Errorf("Bolt at a protected creature: got %v, want ErrIllegalTarget", err)
	}

	start := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine, opp.ID, 2); err != nil {
			t.Fatalf("creature source: %v", err)
		}
	})
	if opp.Life != start-2 {
		t.Errorf("a creature source dealt %d, want 2 — the protection is from instants only", start-opp.Life)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(emissary); err != nil {
			t.Fatalf("destroy the Emissary: %v", err)
		}
	})
	if got := playerAbilities(g, opp); len(got) != 0 {
		t.Errorf("the player's protection outlived the Emissary: %v", got)
	}
	if hasProtectionToken(g, bear, "protection from instants") {
		t.Error("the creature's protection outlived the Emissary")
	}
}
