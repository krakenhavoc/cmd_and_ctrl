package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 799 (Mirage, #5633): Goblin Scouts, Ether Well, Prismatic
// Boon, Stupor, Prismatic Lace, Ashen Powder, Political Trickery and
// Disempower.

const (
	oldB799Scouts   = "87b55d99-badc-45e3-8364-173a400c4d7e"
	oldB799Well     = "a2b6faca-3321-45d2-a027-ff3a8ae61f10"
	oldB799Boon     = "aee5dda1-2181-40d5-9846-85d154e1d5da"
	oldB799Stupor   = "c19e3560-d304-4d91-aa2a-463d7dbb35bb"
	oldB799Lace     = "d4a3c9a5-3c91-4135-a19d-98eb5bcf5e00"
	oldB799Powder   = "d906965a-0f33-4d75-8abe-66f2cf4817b8"
	oldB799Trickery = "e0eaa43f-1c46-4299-9ecd-5f4a32d292cd"
	oldB799Dis      = "ff9f4bfe-1585-4eaa-943c-1e85ed79a820"
)

func TestOldB799EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB799Scouts: "Goblin Scouts", oldB799Well: "Ether Well", oldB799Boon: "Prismatic Boon",
		oldB799Stupor: "Stupor", oldB799Lace: "Prismatic Lace", oldB799Powder: "Ashen Powder",
		oldB799Trickery: "Political Trickery", oldB799Dis: "Disempower",
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

func oldB799Main(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat]
}

func TestOldB799GoblinScoutsMakesThreeMountainwalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	effectstest.CastCatalogSpell(t, g, "Goblin Scouts", "Sorcery", oldB799Scouts, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Goblin Scout"); n != 3 {
		t.Fatalf("%d Goblin Scout tokens, want 3", n)
	}
	tok := effectstest.BattlefieldTokenNamed(t, g, "Goblin Scout")
	if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "mountainwalk") {
		t.Error("the token has no mountainwalk")
	}
	if p, tough := effectstest.EffectivePower(t, g, tok.InstanceID), effectstest.EffectiveToughness(t, g, tok.InstanceID); p != 1 || tough != 1 {
		t.Errorf("token is %d/%d, want 1/1", p, tough)
	}
	if got := effectstest.EffectiveColors(t, g, tok.InstanceID); !slices.Equal(got, []string{"R"}) {
		t.Errorf("token colors %v, want red", got)
	}
}

func TestOldB799EtherWellTucksOnTopAndOffersARedCreatureTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	opp := g.Seats[1]
	green := effectstest.B16Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1, "G")
	red := effectstest.B16Creature(g, opp.ID, "Imp", "Creature — Imp", 1, 1, "R")
	effectstest.CastCatalogSpell(t, g, "Ether Well", "Instant", oldB799Well, effectstest.LtCardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if top := effectstest.LibraryTopIDs(opp, 1); len(top) != 1 || top[0] != green {
		t.Errorf("the green creature is not on top: %v", top)
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("a nonred creature was offered the bottom")
	}
	effectstest.CastCatalogSpell(t, g, "Ether Well", "Instant", oldB799Well, effectstest.LtCardTarget(red))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a red creature was not offered the top or the bottom")
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{red}, nil); err != nil {
		t.Fatalf("put on the bottom: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if bottom := effectstest.LibraryBottomIDs(opp, 1); len(bottom) != 1 || bottom[0] != red {
		t.Error("the red creature is not on the bottom of its owner's library")
	}
	if g.Battlefield.Contains(red) {
		t.Error("the red creature is still on the battlefield")
	}
}

func TestOldB799PrismaticBoonGivesXCreaturesProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	id := effectstest.HandCardOf(g, me, "Prismatic Boon", "Instant", "{X}{W}{U}", oldB799Boon)
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: two[:1], XValue: 2, Strict: true}); err == nil {
		t.Fatal("cast for X=2 with one target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: two, XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	want := game.ProtectionFromColor("R")
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"A": {a, true}, "B": {b, true}, "C": {c, false}} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, want); got != tc.want {
			t.Errorf("%s has %q: %v, want %v", name, want, got, tc.want)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasEffectiveKeyword(t, g, a, want) {
		t.Error("the protection outlasted the turn")
	}
}

func TestOldB799StuporDiscardsAtRandomThenByChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB799Main(t, g)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		ids = append(ids, effectstest.HandCardOf(g, opp, n, "Sorcery", "{R}", ""))
	}
	effectstest.CastCatalogSpell(t, g, "Stupor", "Sorcery", oldB799Stupor, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != 1 {
		t.Fatalf("%d cards in their graveyard after the random discard, want 1", got)
	}
	var left []uuid.UUID
	for _, c := range opp.Hand.Cards {
		left = append(left, c.InstanceID)
	}
	if len(left) != 2 {
		t.Fatalf("%d cards left, want 2", len(left))
	}
	effectstest.AnswerDiscard(t, g, opp.ID, left[1])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 1 || !opp.Hand.Contains(left[0]) {
		t.Errorf("hand after the chosen discard: %d cards, kept the first %v", opp.Hand.Size(), opp.Hand.Contains(left[0]))
	}
	if opp.Graveyard.Size() != 2 || !opp.Graveyard.Contains(left[1]) {
		t.Errorf("the chosen card is not in their graveyard (%d cards)", opp.Graveyard.Size())
	}
	_ = ids
}

func TestOldB799StuporOnAOneCardHandDiscardsOnlyThatCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB799Main(t, g)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	only := effectstest.HandCardOf(g, opp, "Only", "Sorcery", "{R}", "")
	effectstest.CastCatalogSpell(t, g, "Stupor", "Sorcery", oldB799Stupor, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(only) || opp.Hand.Size() != 0 {
		t.Error("the only card was not discarded")
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceDiscardFromHand) != nil {
		t.Error("a second discard was asked of an empty hand")
	}
}

func TestOldB799PrismaticLaceRecolorsThePermanentForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	rock := effectstest.AcPermanent(g, me.ID, "Rock", "Artifact", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Prismatic Lace", "Instant", oldB799Lace, effectstest.LtCardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceColor); c == nil || !c.ColorMulti {
		t.Fatalf("prompt %+v, want a multi color pick", c)
	}
	effectstest.AnswerColor(t, g, me.ID, "WB")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveColors(t, g, rock); len(got) != 2 || !slices.Contains(got, "W") || !slices.Contains(got, "B") {
		t.Errorf("colors %v, want white and black", got)
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.EffectiveColors(t, g, rock); len(got) != 2 {
		t.Errorf("after the turn colors %v: the change should last indefinitely", got)
	}
}

func TestOldB799PrismaticLaceWithAGoneTargetAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	rock := effectstest.AcPermanent(g, me.ID, "Rock", "Artifact", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Prismatic Lace", "Instant", oldB799Lace, effectstest.LtCardTarget(rock))
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceColor) != nil {
		t.Error("a color was asked for a target that had left")
	}
}

func TestOldB799AshenPowderStealsACreatureFromAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	opp := g.Seats[1]
	theirs := uuid.New()
	mine := uuid.New()
	spell := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Their Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	for _, bad := range []uuid.UUID{mine, spell} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Ashen Powder", "Sorcery", oldB799Powder, effectstest.LtCardTarget(bad)); err == nil {
			t.Fatal("targeted a card that is not a creature card in an opponent's graveyard")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Ashen Powder", "Sorcery", oldB799Powder, effectstest.LtCardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the creature did not enter the battlefield")
	}
	c := effectstest.LayeredCard(t, g, theirs)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %s owner %s, want me controlling the opponent's card", c.Controller, c.Owner)
	}
}

func TestOldB799PoliticalTrickeryExchangesTwoLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB799Main(t, g)
	opp := g.Seats[1]
	mine := effectstest.AcPermanent(g, me.ID, "My Forest", "Basic Land — Forest", 0, 0)
	other := effectstest.AcPermanent(g, me.ID, "My Island", "Basic Land — Island", 0, 0)
	theirs := effectstest.AcPermanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	two := func(a, b uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Political Trickery", "Sorcery", oldB799Trickery, two(mine, other)); err == nil {
		t.Fatal("exchanged two lands I control")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Political Trickery", "Sorcery", oldB799Trickery, two(mine, bear)); err == nil {
		t.Fatal("exchanged a land for a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Political Trickery", "Sorcery", oldB799Trickery, two(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, mine); c.Controller != opp.ID {
		t.Error("my land is not under their control")
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.Controller != me.ID {
		t.Error("their land is not under my control")
	}
	effectstest.RunCleanup(t, g)
	if c := effectstest.LayeredCard(t, g, theirs); c.Controller != me.ID {
		t.Error("the exchange ended with the turn; it has no duration")
	}
}

func TestOldB799DisempowerPutsAnArtifactOrEnchantmentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB799Main(t, g)
	opp := g.Seats[1]
	rock := effectstest.AcPermanent(g, opp.ID, "Rock", "Artifact", 0, 0)
	aura := effectstest.AcPermanent(g, opp.ID, "Aura", "Enchantment", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Disempower", "Instant", oldB799Dis, effectstest.LtCardTarget(bear)); err == nil {
		t.Fatal("targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Disempower", "Instant", oldB799Dis, effectstest.LtCardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Disempower", "Instant", oldB799Dis, effectstest.LtCardTarget(aura))
	effectstest.PassPriorityAroundTable(t, g)
	top := effectstest.LibraryTopIDs(opp, 2)
	if len(top) != 2 || top[0] != aura || top[1] != rock {
		t.Errorf("their library from the top is %v, want the Aura then the Rock", top)
	}
}
