package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b566_test.go — Coldsnap, batch 566 (#5597): ripple and the snow
// taplands. Each test plays the real catalog entry through the real engine.

const (
	b566Flame     = "3c51b991-f6fa-44d2-81a3-0894939f3c22"
	b566Dementia  = "dd7f46c2-98f7-46e6-92ce-51fa95d2f0c8"
	b566Sentinels = "0b98ae2a-56ac-485d-b064-a06e1e8d4055"
	b566Might     = "04647861-0023-432b-9b24-77264f11e470"
	b566Aether    = "060b1a76-22a4-48a5-a353-4d0501367956"
)

func TestS82OldB566CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b566Flame: "Surging Flame", b566Dementia: "Surging Dementia", b566Sentinels: "Surging Sentinels",
		b566Might: "Surging Might", b566Aether: "Surging Aether",
		"75a49bce-ab48-44d6-906d-6f51a30702ba": "Frost Marsh", "66d845a2-be4c-41bb-a357-3650461c03ed": "Highland Weald",
		"20b549e9-26db-40a5-8bcc-e8de6c559d32": "Boreal Shelf", "74a6597c-13ea-47bd-adb3-45671897b11b": "Tresserhorn Sinks",
		"841f0a1d-6f84-45fb-81cc-f2d99fa9e9a2": "Arctic Flats",
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

func TestS82OldB566SnowTaplandsEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct{ name, oracle, a, b string }{
		{"Frost Marsh", "75a49bce-ab48-44d6-906d-6f51a30702ba", "U", "B"},
		{"Highland Weald", "66d845a2-be4c-41bb-a357-3650461c03ed", "R", "G"},
		{"Boreal Shelf", "20b549e9-26db-40a5-8bcc-e8de6c559d32", "W", "U"},
		{"Tresserhorn Sinks", "74a6597c-13ea-47bd-adb3-45671897b11b", "B", "R"},
		{"Arctic Flats", "841f0a1d-6f84-45fb-81cc-f2d99fa9e9a2", "G", "W"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12PlayFromHand(t, g, row.name, "Snow Land", row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatalf("%s: the land enters tapped", row.name)
		}
		if !effectstest.LayeredCard(t, g, land).HasSupertype("snow") {
			t.Errorf("%s: not snow", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.a)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.a {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.a, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.b)
		if got := effectstest.PoolColors(me); len(got) != 2 || got[1] != row.b {
			t.Errorf("%s: also taps for {%s}: pool %v", row.name, row.b, got)
		}
	}
}

// b566Hand puts the real card in `me`'s hand with its printed cost.
func b566Hand(me *game.Player, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	return id
}

// b566Library pushes a real card on top of `me`'s library.
func b566Library(me *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: me.ID, Controller: me.ID})
	return id
}

// b566Bottom answers "put the rest on the bottom of your library in any order"
// with the order offered.
func b566Bottom(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
}

func TestS82OldB566SurgingFlameRipplesIntoACopyOfItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Flame", "Instant", b566Flame, "{1}{R}", 0, 0)
	effectstest.B06AddMana(me, "R", "R")
	// Top four, top first: a Flame, two others, a Flame. The fifth card is out of range.
	deep := b566Library(me, "Surging Flame", "Instant", b566Flame, "{1}{R}")
	var others [2]uuid.UUID
	others[0] = b566Library(me, "Other", "Sorcery", "", "{3}")
	others[1] = b566Library(me, "Other", "Sorcery", "", "{3}")
	top := b566Library(me, "Surging Flame", "Instant", b566Flame, "{1}{R}")
	beyond := uuid.New()
	g.WithWriteLock(func() {
		me.Library.Cards = append([]game.Card{{InstanceID: beyond, Name: "Surging Flame", TypeLine: "Instant", OracleID: b566Flame, Owner: me.ID, Controller: me.ID}}, me.Library.Cards...)
	})
	life := opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	_ = effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top, deep)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("casting the revealed Flame: %v", err)
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, deep)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	b566Bottom(t, g, me)
	// The free Flame was cast, so it ripples too: decline that one.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 4 {
		t.Errorf("the opponent lost %d, want 4 (the cast Flame and the free one)", got)
	}
	for _, id := range append(others[:], deep) {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneLibrary {
			t.Errorf("a revealed card that was not cast is in %v, want the library", z)
		}
	}
	if z := effectstest.B12ZoneOf(g, beyond); z != game.ZoneLibrary {
		t.Errorf("the fifth card from the top is in %v", z)
	}
}

func TestS82OldB566RippleIsAMayAndTheRevealCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Flame", "Instant", b566Flame, "{1}{R}", 0, 0)
	effectstest.B06AddMana(me, "R", "R")
	same := b566Library(me, "Surging Flame", "Instant", b566Flame, "{1}{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a cast was offered although the reveal was declined")
	}
	if effectstest.B12ZoneOf(g, same) != game.ZoneLibrary {
		t.Error("the card left the library")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := 40 - opp.Life; got != 2 {
		t.Errorf("the one Flame dealt %d, want 2", got)
	}
}

func TestS82OldB566SurgingDementiaMakesTheTargetDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	kept := uuid.New()
	pitched := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: kept, Name: "Kept", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: pitched, Name: "Pitched", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Dementia", "Sorcery", b566Dementia, "{1}{B}", 0, 0)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, pitched)
	if !opp.Graveyard.Contains(pitched) || opp.Hand.Size() != 1 {
		t.Errorf("hand %d, chosen card discarded %v; want one card gone, the chosen one", opp.Hand.Size(), opp.Graveyard.Contains(pitched))
	}
}

func TestS82OldB566SurgingAetherBouncesAnyPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.BlbB1Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Aether", "Instant", b566Aether, "{3}{U}", 0, 0)
	effectstest.B06AddMana(me, "U", "U", "U", "U")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(land)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, land); z != game.ZoneHand {
		t.Errorf("the land is in %v, want its owner's hand", z)
	}
}

func TestS82OldB566SurgingMightGivesPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Might", "Enchantment — Aura", b566Might, "{2}{G}", 0, 0)
	effectstest.B06AddMana(me, "G", "G", "G")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4", p, tough)
	}
	if err := g.CastSpell(me.ID, b566Hand(me, "Surging Might", "Enchantment — Aura", b566Might, "{2}{G}", 0, 0), game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an Aura cast with no creature to enchant was accepted")
	}
}

func TestS82OldB566SurgingSentinelsHasFirstStrikeAndRipples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spell := b566Hand(me, "Surging Sentinels", "Creature — Human Soldier", b566Sentinels, "{2}{W}", 2, 1)
	effectstest.B06AddMana(me, "W", "W", "W")
	free := b566Library(me, "Surging Sentinels", "Creature — Human Soldier", b566Sentinels, "{2}{W}")
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == free {
				me.Library.Cards[i].Power, me.Library.Cards[i].Toughness = 2, 1
			}
		}
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	_ = effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, free)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, free, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	b566Bottom(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{spell, free} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("a Sentinels did not reach the battlefield")
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "first strike") {
			t.Error("a Sentinels has no first strike")
		}
	}
}
