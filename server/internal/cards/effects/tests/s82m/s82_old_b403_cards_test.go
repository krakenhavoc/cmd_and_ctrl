package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b403_cards_test.go — Dragon's Maze instants and sorceries, S82 OLD
// batch 403 (#5564): eight cards, each driven through the engine.

const (
	dgmUncoveredClues  = "942d550d-4823-4db8-8197-cbf15045f4f5"
	dgmReapIntellect   = "aa97eb3d-8b8d-418b-8615-839812d60fe0"
	dgmLyevDecree      = "ace64875-ca4b-45c7-8b46-29ef2f2629dd"
	dgmRestoreThePeace = "aeca0661-8b71-4085-81e9-bb0884fc9831"
	dgmDragonshift     = "b02d24fd-62b4-4f78-96af-57f88c98c5c8"
	dgmRenounceGuilds  = "c854f9e8-6a1c-4515-9912-3847978bf5d5"
	dgmAweForTheGuilds = "cb50ec2e-cb27-4ee5-a7de-64ab93619cea"
	dgmWarpedPhysique  = "f45fdef0-15b1-472b-9485-465c49a48a85"
)

func TestS82OldB403EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dgmUncoveredClues: "Uncovered Clues", dgmReapIntellect: "Reap Intellect", dgmLyevDecree: "Lyev Decree",
		dgmRestoreThePeace: "Restore the Peace", dgmDragonshift: "Dragonshift", dgmRenounceGuilds: "Renounce the Guilds",
		dgmAweForTheGuilds: "Awe for the Guilds", dgmWarpedPhysique: "Warped Physique",
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

func TestS82OldB403UncoveredCluesTakesUpToTwoInstantsOrSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	// Top four, top first: instant, creature, sorcery, instant.
	pushTop := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	filler := pushTop("Filler", "Land")
	i2 := pushTop("Third Instant", "Instant")
	s1 := pushTop("Sorcery", "Sorcery")
	cr := pushTop("Creature", "Creature — Bear")
	i1 := pushTop("First Instant", "Instant")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Uncovered Clues", "Sorcery", dgmUncoveredClues, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, i1, s1)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(i1) || !me.Hand.Contains(s1) {
		t.Error("the two chosen cards did not reach the hand")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, me.Hand.Size())
	}
	// The creature and the third instant went to the bottom; the card below the
	// top four stays above them.
	if me.Library.Contains(i1) || me.Library.Contains(s1) {
		t.Error("a taken card stayed in the library")
	}
	if !me.Library.Contains(cr) || !me.Library.Contains(i2) || !me.Library.Contains(filler) {
		t.Error("an untaken card left the library")
	}
}

func TestS82OldB403LyevDecreeDetainsUpToTwoOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Ogre A", "Creature — Ogre", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Ogre B", "Creature — Ogre", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "Ogre C", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Lyev Decree", "Sorcery", dgmLyevDecree, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Lyev Decree", "Sorcery", dgmLyevDecree, effectstest.CardTarget(mine)); err == nil {
		t.Fatal("my own creature was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Lyev Decree", "Sorcery", dgmLyevDecree, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	want := game.CantAttack | game.CantBlock | game.CantActivate | game.CantActivateMana
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.RestrictionsOf(t, g, id); got&want != want {
			t.Errorf("a target's restrictions are %v, want it detained", got.Names())
		}
	}
	if got := effectstest.RestrictionsOf(t, g, c); got&game.CantBlock != 0 {
		t.Error("the untargeted creature was detained")
	}
}

func TestS82OldB403LyevDecreeMayChooseNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lyev Decree", "Sorcery", dgmLyevDecree, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB403AweForTheGuildsStopsOnlyMonocoloredBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	attacker := effectstest.B21Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "R")
	mono := effectstest.B21Push(g, opp.ID, "Mono", "Creature — Bear", "", 2, 2, "G")
	multi := effectstest.B21Push(g, opp.ID, "Multi", "Creature — Bear", "", 2, 2, "G", "W")
	colorless := effectstest.B21Push(g, opp.ID, "Colorless", "Artifact Creature — Golem", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Awe for the Guilds", "Sorcery", dgmAweForTheGuilds, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// A monocoloured creature arriving later this turn can't block either.
	late := effectstest.B21Push(g, opp.ID, "Late", "Creature — Bear", "", 2, 2, "U")
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if effectstest.BlockRefusal(t, g, mono, attacker) == "" {
		t.Error("a monocoloured creature blocked")
	}
	if effectstest.BlockRefusal(t, g, late, attacker) == "" {
		t.Error("a monocoloured creature that arrived later blocked")
	}
	if got := effectstest.BlockRefusal(t, g, multi, attacker); got != "" {
		t.Errorf("a multicoloured creature was refused: %q", got)
	}
	if got := effectstest.BlockRefusal(t, g, colorless, attacker); got != "" {
		t.Errorf("a colourless creature was refused: %q", got)
	}
}

func TestS82OldB403RestoreThePeaceReturnsOnlyCreaturesThatDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dealer := effectstest.B12Creature(g, me.ID, "Dealer", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(dealer, victim, 1) })
	effectstest.CastCatalogSpell(t, g, "Restore the Peace", "Instant", dgmRestoreThePeace, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dealer) || !me.Hand.Contains(dealer) {
		t.Error("the creature that dealt damage was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(victim) || !g.Battlefield.Contains(idle) {
		t.Error("a creature that only took or did nothing was returned")
	}
}

func TestS82OldB403RenounceTheGuildsMakesEachPlayerSacrificeAMulticoloredPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B21Push(g, me.ID, "My Gold", "Creature — Bear", "", 2, 2, "W", "U")
	mineMono := effectstest.B21Push(g, me.ID, "My Mono", "Creature — Bear", "", 2, 2, "W")
	theirs := effectstest.B21Push(g, opp.ID, "Their Gold", "Creature — Bear", "", 2, 2, "B", "R")
	theirsMono := effectstest.B21Push(g, opp.ID, "Their Mono", "Creature — Bear", "", 2, 2, "B")
	effectstest.CastCatalogSpell(t, g, "Renounce the Guilds", "Instant", dgmRenounceGuilds, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a multicoloured permanent was not sacrificed")
	}
	if !g.Battlefield.Contains(mineMono) || !g.Battlefield.Contains(theirsMono) {
		t.Error("a monocoloured permanent was sacrificed")
	}
}

func TestS82OldB403WarpedPhysiqueScalesWithYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 2, 6)
	me.Hand.Cards = nil
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Warped Physique", "Instant", dgmWarpedPhysique, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ogre)
	if p, th := c.CurrentPower(), c.CurrentToughness(); p != 5 || th != 3 {
		t.Errorf("the creature is %d/%d, want 5/3 (three cards in hand)", p, th)
	}
}

func TestS82OldB403DragonshiftBecomesABlueAndRedFlyingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"trample"}})
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Dragonshift", "Instant", dgmDragonshift, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if p, th := c.CurrentPower(), c.CurrentToughness(); p != 4 || th != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, th)
	}
	if st := c.Effective().Subtypes; !dgm403Has(st, "Dragon") || dgm403Has(st, "Bear") {
		t.Errorf("subtypes %v, want only Dragon", st)
	}
	cols := effectstest.EffectiveColors(t, g, bear)
	if !dgm403Has(cols, "U") || !dgm403Has(cols, "R") || len(cols) != 2 {
		t.Errorf("colours %v, want blue and red", cols)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("it should have flying and no other abilities")
	}
	if c := effectstest.LayeredCard(t, g, other); c.CurrentPower() != 2 {
		t.Error("an untargeted creature changed")
	}
}

func TestS82OldB403DragonshiftOverloadedChangesEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.HandCardFull(me, "Dragonshift", "Instant", "{1}{U}{R}", dgmDragonshift, []string{"U", "R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("cast overloaded: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{a, b} {
		if l := effectstest.LayeredCard(t, g, c); l.CurrentPower() != 4 || l.CurrentToughness() != 4 {
			t.Errorf("my creature %s is %d/%d, want 4/4", c, l.CurrentPower(), l.CurrentToughness())
		}
	}
	if l := effectstest.LayeredCard(t, g, theirs); l.CurrentPower() != 2 {
		t.Error("an opponent's creature changed")
	}
}

func dgm403Has(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestS82OldB403ReapIntellectExilesUpToXAndSearchesForTheirNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	push := func(z *game.Zone, name, typeLine string) uuid.UUID {
		id := uuid.New()
		z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: opp.ID, Controller: opp.ID})
		return id
	}
	opp.Hand.Cards = nil
	opp.Library.Cards = nil
	dup1 := push(opp.Hand, "Twin Bear", "Creature — Bear")
	dup2 := push(opp.Hand, "Twin Bear", "Creature — Bear")
	other := push(opp.Hand, "Lone Spell", "Sorcery")
	keep := push(opp.Hand, "Third Spell", "Sorcery")
	land := push(opp.Hand, "Forest", "Basic Land — Forest")
	gy := push(opp.Graveyard, "Twin Bear", "Creature — Bear")
	lib := push(opp.Library, "Twin Bear", "Creature — Bear")
	libOther := push(opp.Library, "Unrelated", "Creature — Bear")
	effectstest.CastXSpell(t, g, "Reap Intellect", "Sorcery", dgmReapIntellect, "{X}{2}{U}{B}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.Count != 2 || !pick.PickUpTo {
		t.Fatalf("the pick is count %d up-to %v, want 2 and up to", pick.Count, pick.PickUpTo)
	}
	for _, id := range pick.DiscardOptions {
		if id == land {
			t.Fatal("a land card was offered")
		}
	}
	// Three cards is more than X; two is fine, as is fewer.
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{dup1, other, keep}); err == nil {
		t.Fatal("three cards were accepted for X=2")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{dup1, other}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if !g.Exile.Contains(dup1) || !g.Exile.Contains(other) {
		t.Fatal("the chosen cards were not exiled")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// One search over the names Twin Bear and Lone Spell.
	effectstest.AnswerSearchByID(t, g, me.ID, dup2, gy, lib)
	for _, id := range []uuid.UUID{dup2, gy, lib} {
		if !g.Exile.Contains(id) {
			t.Errorf("a card named like an exiled one was not exiled: %s", id)
		}
	}
	if g.Exile.Contains(keep) || g.Exile.Contains(land) || g.Exile.Contains(libOther) {
		t.Error("a card with another name was exiled")
	}
}

func TestS82OldB403ReapIntellectMayChooseNothingAndTheLibraryStillShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	opp.Hand.Cards = nil
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Lone Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	shuffles := len(effectstest.RandomEvents(g, game.EventLibraryShuffled))
	effectstest.CastXSpell(t, g, "Reap Intellect", "Sorcery", dgmReapIntellect, "{X}{2}{U}{B}", 1,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing nothing was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spell) {
		t.Error("the card left the hand")
	}
	if got := len(effectstest.RandomEvents(g, game.EventLibraryShuffled)); got <= shuffles {
		t.Errorf("their library was not shuffled (%d -> %d)", shuffles, got)
	}
}
