package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, batch 36 (Modern Horizons 2: suspend, damage, #5501).

const (
	oldB036ResurgentBelief = "0fcfef04-a954-4423-82b0-c70dc64a8fbc"
	oldB036Glimpse         = "12778951-7b11-400b-98ed-06434d7ecc11"
	oldB036ProfaneTutor    = "27a1f42c-0b86-4609-9609-1fa9cab7e7c9"
	oldB036SolTalisman     = "867f0cdf-cf75-40f0-bf1e-ccb92e55e9a3"
	oldB036RiftSower       = "f9deeea1-bf80-4832-aac3-373fc1f4b41d"
	oldB036Feast           = "af384946-e22a-45cd-ba07-1a5fbd4a3f0c"
	oldB036Piru            = "dcd13eab-7ed7-4e95-bdae-d140c1ff84df"
	oldB036Manticore       = "f434590f-3a5d-47dd-8ed0-d2a43a521f86"
	oldB036FodderTosser    = "feb34b5a-d51e-42d7-90c6-c5c84ff00c4e"
	oldB036Pacifism        = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
)

func TestOldB036CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB036ResurgentBelief: "Resurgent Belief",
		oldB036Glimpse:         "Glimpse of Tomorrow",
		oldB036ProfaneTutor:    "Profane Tutor",
		oldB036SolTalisman:     "Sol Talisman",
		oldB036RiftSower:       "Rift Sower",
		oldB036Feast:           "Feast of Sanity",
		oldB036Piru:            "Piru, the Volatile",
		oldB036Manticore:       "Mount Velus Manticore",
		oldB036FodderTosser:    "Fodder Tosser",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB036Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var out game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

// oldB036Suspend suspends the card from the active seat's hand with `pay`
// mana, then walks the table through the countdown until the last time counter
// comes off and the free cast is offered. It returns the card's id and the
// seat that owns it.
func oldB036Suspend(t *testing.T, g *game.Game, name, typeLine, oracle, color string, pay, counters int) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, name, typeLine, oracle)
	effectstest.StampLayout(g, me, id, "normal")
	if strings.Contains(typeLine, "Creature") {
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 1, 3
				}
			}
		})
	}
	effectstest.FillPoolColored(me, color, pay)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend %s: %v", name, err)
	}
	var got int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			got = c.Counters[game.CounterTime]
		}
	})
	if got != counters {
		t.Fatalf("%s has %d time counters after suspending, want %d", name, got, counters)
	}
	for i := 0; i < counters; i++ {
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		effectstest.AdvanceToUpkeepOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatalf("the last time counter of %s coming off offered no cast", name)
	}
	return id, me
}

func oldB036CastFromExile(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, targets ...game.TargetRef) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile", Targets: targets}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB036SuspendCardsDeclareTheirPrintedNumbers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		n      int
		cost   string
	}{
		{"Resurgent Belief", oldB036ResurgentBelief, 2, "{1}{W}"},
		{"Glimpse of Tomorrow", oldB036Glimpse, 3, "{R}{R}"},
		{"Profane Tutor", oldB036ProfaneTutor, 2, "{1}{B}"},
		{"Sol Talisman", oldB036SolTalisman, 3, "{1}"},
		{"Rift Sower", oldB036RiftSower, 2, "{G}"},
	} {
		sa := effectstest.SuspendOf(tc.oracle)
		if sa == nil {
			t.Errorf("%s declares no suspend", tc.name)
			continue
		}
		if sa.Counters != tc.n || sa.Cost != tc.cost {
			t.Errorf("%s: suspend %d—%s, want %d—%s", tc.name, sa.Counters, sa.Cost, tc.n, tc.cost)
		}
	}
}

// --- Resurgent Belief ---------------------------------------------------------

func TestOldB036ResurgentBeliefReturnsEveryEnchantmentFromYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	add := func(p *game.Player, name, typeLine string) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		return id
	}
	e1 := add(me, "Ench One", "Enchantment")
	e2 := add(me, "Ench Two", "Enchantment Creature — Spirit")
	art := add(me, "Rock", "Artifact")
	theirs := add(opp, "Their Ench", "Enchantment")
	id, me := oldB036Suspend(t, g, "Resurgent Belief", "Sorcery", oldB036ResurgentBelief, "W", 2, 2)
	oldB036CastFromExile(t, g, me, id)
	for _, e := range []uuid.UUID{e1, e2} {
		if oldB036Zone(g, e) != game.ZoneBattlefield {
			t.Errorf("an enchantment card is in %q, want the battlefield", oldB036Zone(g, e))
		}
	}
	if oldB036Zone(g, art) != game.ZoneGraveyard {
		t.Error("an artifact card came back")
	}
	if oldB036Zone(g, theirs) != game.ZoneGraveyard {
		t.Error("an opponent's enchantment came back")
	}
}

// --- Glimpse of Tomorrow ----------------------------------------------------------

func TestOldB036GlimpseShufflesYourPermanentsInAndPutsThemBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := oldB036Suspend(t, g, "Glimpse of Tomorrow", "Sorcery", oldB036Glimpse, "R", 2, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		// Start from a clean board: only the permanents under test.
		g.Battlefield.Cards = nil
	})
	me.Library.Cards = nil
	opp.Library.Cards = nil
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 green Squirrel"))
	oldB036CastFromExile(t, g, me, id)
	// Four permanents were shuffled in (the token ceased to exist but counts):
	// with only three cards in the library, every card is revealed and returns.
	for _, c := range []uuid.UUID{land, bear, stolen} {
		if oldB036Zone(g, c) != game.ZoneBattlefield {
			t.Errorf("a permanent card is in %q, want the battlefield", oldB036Zone(g, c))
		}
	}
	if oldB036Zone(g, theirs) != game.ZoneBattlefield {
		t.Error("an opponent's own permanent was touched")
	}
	if oldB036Zone(g, token) != "" {
		t.Errorf("the token is in %q, want gone", oldB036Zone(g, token))
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == stolen && c.Controller != me.ID {
				t.Error("the stolen permanent came back under the thief")
			}
		}
	})
}

func TestOldB036GlimpseRevealsOnlyAsManyCardsAsPermanentsShuffled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := oldB036Suspend(t, g, "Glimpse of Tomorrow", "Sorcery", oldB036Glimpse, "R", 2, 3)
	g.WithWriteLock(func() { g.Battlefield.Cards = nil })
	me.Library.Cards = nil
	var junk []uuid.UUID
	for i := 0; i < 6; i++ {
		junk = append(junk, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}))
	}
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oldB036CastFromExile(t, g, me, id)
	// Eight cards are in the library after the shuffle; two are revealed, so at
	// most two permanents are back, and nothing else was moved by the spell.
	back := 0
	for _, c := range []uuid.UUID{land, bear} {
		if oldB036Zone(g, c) == game.ZoneBattlefield {
			back++
		}
	}
	size := 0
	g.ReadSnapshot(func() { size = me.Library.Size() })
	if size != 8-back {
		t.Errorf("library holds %d cards with %d permanents back, want %d", size, back, 8-back)
	}
	for _, j := range junk {
		if z := oldB036Zone(g, j); z != game.ZoneLibrary {
			t.Errorf("a nonpermanent card is in %q, want the library (to the bottom)", z)
		}
	}
}

func TestOldB036GlimpseAurasEnterAfterTheOtherPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := oldB036Suspend(t, g, "Glimpse of Tomorrow", "Sorcery", oldB036Glimpse, "R", 2, 3)
	g.WithWriteLock(func() { g.Battlefield.Cards = nil })
	me.Library.Cards = nil
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura", OracleID: oldB036Pacifism, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttachForTest(t, g, aura, bear)
	oldB036CastFromExile(t, g, me, id)
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if oldB036Zone(g, bear) != game.ZoneBattlefield {
		t.Fatalf("the creature is in %q, want the battlefield", oldB036Zone(g, bear))
	}
	if oldB036Zone(g, aura) != game.ZoneBattlefield {
		t.Errorf("the Aura is in %q: it should enter after the creature, with something to enchant", oldB036Zone(g, aura))
	}
}

// --- Profane Tutor ------------------------------------------------------------------

func TestOldB036ProfaneTutorSearchesForAnyCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	id, me := oldB036Suspend(t, g, "Profane Tutor", "Sorcery", oldB036ProfaneTutor, "B", 2, 2)
	oldB036CastFromExile(t, g, me, id)
	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{needle}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if oldB036Zone(g, needle) != game.ZoneHand {
		t.Errorf("the card is in %q, want the hand", oldB036Zone(g, needle))
	}
}

func TestOldB036ProfaneTutorCannotBeHardCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, "Profane Tutor", "Sorcery", oldB036ProfaneTutor)
	effectstest.StampLayout(g, me, id, "normal")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a card with no mana cost cannot be cast from the hand (CR 118.6)")
	}
}

// --- Sol Talisman ---------------------------------------------------------------------

func TestOldB036SolTalismanTapsForTwoColorlessTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := oldB036Suspend(t, g, "Sol Talisman", "Artifact", oldB036SolTalisman, "R", 1, 3)
	oldB036CastFromExile(t, g, me, id)
	if oldB036Zone(g, id) != game.ZoneBattlefield {
		t.Fatalf("the Talisman is in %q, want the battlefield", oldB036Zone(g, id))
	}
	before := effectstest.PoolSize(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID) - before; got != 2 {
		t.Errorf("the Talisman made %d mana, want 2", got)
	}
}

// --- Rift Sower ----------------------------------------------------------------------------

func TestOldB036RiftSowerArrivesWithHasteAndTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := oldB036Suspend(t, g, "Rift Sower", "Creature — Elf Druid", oldB036RiftSower, "G", 1, 2)
	oldB036CastFromExile(t, g, me, id)
	if oldB036Zone(g, id) != game.ZoneBattlefield {
		t.Fatalf("Rift Sower is in %q, want the battlefield", oldB036Zone(g, id))
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("a suspended creature arrives with haste (CR 702.62a)")
	}
	before := effectstest.PoolSize(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if effectstest.PoolSize(g, me.ID)-before != 1 {
		t.Error("Rift Sower did not make one mana")
	}
}

// --- Feast of Sanity ---------------------------------------------------------------------------

func TestOldB036FeastOfSanityPingsAndGainsPerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Feast of Sanity", "Enchantment", oldB036Feast, false)
	me.Hand.Cards = nil
	effectstest.HandSpell(me, "Discarded", "Sorcery", "")
	life, theirs := me.Life, opp.Life
	effectstest.B30DiscardOne(t, g, me.ID)
	if p := effectstest.PassUntilPickTarget(t, g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, theirs-1)
	}
	if me.Life != life+1 {
		t.Errorf("my life is %d, want %d", me.Life, life+1)
	}
}

// --- Piru, the Volatile -----------------------------------------------------------------------------

func TestOldB036PiruIsSacrificedUnlessYouPayInYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	piru := effectstest.PushCatalogPermanent(g, me.ID, "Piru, the Volatile", "Legendary Creature — Elder Dragon", oldB036Piru, false)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+len(g.Seats)-1+len(g.Seats))%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.AdvanceToUpkeepOf(t, g, indexOf(g, me))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment was offered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if oldB036Zone(g, piru) != game.ZoneGraveyard {
		t.Errorf("Piru is in %q after declining, want the graveyard", oldB036Zone(g, piru))
	}
}

func indexOf(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s == p {
			return i
		}
	}
	return -1
}

func TestOldB036PiruStaysWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	piru := effectstest.PushCatalogPermanent(g, me.ID, "Piru, the Volatile", "Legendary Creature — Elder Dragon", oldB036Piru, false)
	effectstest.AdvanceToUpkeepOf(t, g, indexOf(g, me))
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []string{"R", "W", "B"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if oldB036Zone(g, piru) != game.ZoneBattlefield {
		t.Errorf("Piru is in %q after paying, want the battlefield", oldB036Zone(g, piru))
	}
}

func TestOldB036PiruDyingDealsSevenToEachNonlegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	piru := effectstest.PushCatalogPermanent(g, me.ID, "Piru, the Volatile", "Legendary Creature — Elder Dragon", oldB036Piru, false)
	plain := effectstest.PushVanillaCreature(g, opp.ID, "Plain", 3, 6)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Giant", Power: 9, Toughness: 9, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, piru); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB036Zone(g, plain) != game.ZoneGraveyard || oldB036Zone(g, mine) != game.ZoneGraveyard {
		t.Error("7 damage should kill every nonlegendary creature, yours included")
	}
	if oldB036Zone(g, legend) != game.ZoneBattlefield {
		t.Error("a legendary creature was damaged")
	}
}

// --- Mount Velus Manticore -------------------------------------------------------------------------------

func TestOldB036ManticoreDealsDamageEqualToTheDiscardedCardsTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Mount Velus Manticore", "Enchantment Creature — Manticore", oldB036Manticore, false)
	me.Hand.Cards = nil
	two := effectstest.HandSpell(me, "Gadget", "Artifact Creature — Construct", "")
	life := opp.Life
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard was offered at the beginning of combat")
	}
	effectstest.AnswerDiscard(t, g, me.ID, two)
	if p := effectstest.PassUntilPickTarget(t, g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d (an artifact creature has two card types)", opp.Life, life-2)
	}
}

func TestOldB036ManticoreDoesNothingWhenYouDeclineTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Mount Velus Manticore", "Enchantment Creature — Manticore", oldB036Manticore, false)
	effectstest.HandSpell(me, "Keep", "Sorcery", "")
	life := opp.Life
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.DiscardChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerDiscard(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life)
	}
}

// --- Fodder Tosser -----------------------------------------------------------------------------------------

func TestOldB036FodderTosserDiscardsForTwoDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tosser := effectstest.PushCatalogPermanent(g, me.ID, "Fodder Tosser", "Artifact", oldB036FodderTosser, false)
	card := effectstest.HandSpell(me, "Pitched", "Sorcery", "")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, tosser, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{card}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	if oldB036Zone(g, card) != game.ZoneGraveyard {
		t.Errorf("the discarded card is in %q", oldB036Zone(g, card))
	}
}

func TestOldB036FodderTosserCannotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tosser := effectstest.PushCatalogPermanent(g, me.ID, "Fodder Tosser", "Artifact", oldB036FodderTosser, false)
	card := effectstest.HandSpell(me, "Pitched", "Sorcery", "")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, tosser, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{card}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature is neither a player nor a planeswalker")
	}
}
