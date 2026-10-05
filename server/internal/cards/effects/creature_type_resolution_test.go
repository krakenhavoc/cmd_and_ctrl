package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// creature_type_resolution_test.go — #2382. A creature-type choice made
// as a spell resolves: the prompt is the as-enters one, the rest of the
// card hangs off the answer.

const (
	distantMelodyOracle     = "ac6c5852-71c7-4f19-8c87-ccb345052862"
	kindredDominanceOracle  = "ccaa44f2-96be-44e2-884f-c31baa3908d5"
	raiseThePalisadeOracle  = "f55a3781-fe33-4301-9bb5-6a54b9c13c4f"
	bannerOfKinshipOracle   = "8c220dbd-6572-4715-aae6-dd09a4252d68"
	patriarchsBiddingOracle = "25fc3bc2-d852-4f52-9adb-c7e35c06f3af"
	tribalUnityOracle       = "6b4ee935-4474-4757-9fab-978a2e1e72a0"
	kindredChargeOracle     = "eb0b5b5f-54d3-49c6-9ad8-d44a5551acdb"
)

// answerCreatureType answers the open creature-type prompt owed by
// `chooser`, failing if none is open.
func answerCreatureType(t *testing.T, g *game.Game, chooser uuid.UUID, tribe string) {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceCreatureType)
	if c == nil {
		t.Fatalf("no creature-type prompt is open (answering %s)", tribe)
	}
	if c.Chooser != chooser {
		t.Fatalf("the open creature-type prompt belongs to %s, not %s", c.Chooser, chooser)
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, chooser, tribe); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice(%s): %v", tribe, err)
	}
}

// castAndPause casts the spell, passes priority until it resolves, and
// returns with the creature-type prompt open.
func castAndPause(t *testing.T, g *game.Game, name, typeLine, oracle string) {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatalf("%s: resolving did not pause on a creature-type prompt", name)
	}
}

func hasKw(kws []string, kw string) bool {
	for _, k := range kws {
		if k == kw {
			return true
		}
	}
	return false
}

func battlefieldHas(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.Battlefield.Contains(id) })
	return found
}

func inHandOf(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// --- Distant Melody ----------------------------------------------------

func TestDistantMelodyDrawsPerPermanentOfTheChosenTypeChangelingsCounted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	pushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	// A non-creature permanent with the type counts: "permanent".
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kindred Banner", TypeLine: "Kindred Artifact — Elf",
		Owner: me.ID, Controller: me.ID,
	})

	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)
	before := handSize(me)
	if got := handSize(me); got != before {
		t.Fatalf("the draw happened before the answer")
	}
	answerCreatureType(t, g, me.ID, "Elf")
	// Two Elves, the Kindred Elf artifact and the changeling; not the
	// opponent's Elf, not the Goblin.
	if got := handSize(me) - before; got != 4 {
		t.Errorf("drew %d, want 4 (two Elves + Kindred artifact + changeling)", got)
	}
}

func TestDistantMelodyAnswerNoCreatureHasDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)
	before := handSize(me)
	answerCreatureType(t, g, me.ID, "Sliver")
	if got := handSize(me); got != before {
		t.Errorf("drew %d for a type nobody controls", got-before)
	}
	if pendingOfKind(g, game.PendingChoiceCreatureType) != nil {
		t.Error("the prompt is still open after the answer")
	}
}

func TestDistantMelodyChangelingAloneStillCountsForAnyType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)
	before := handSize(me)
	answerCreatureType(t, g, me.ID, "Sliver")
	if got := handSize(me) - before; got != 1 {
		t.Errorf("drew %d, want 1: a changeling is of every creature type", got)
	}
}

// An answer outside the vocabulary is refused and the prompt stays open
// (CR 205.3m); a lowercase spelling is normalised.
func TestResolutionCreatureTypeRefusesJunkAndNormalisesCase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)
	c := pendingOfKind(g, game.PendingChoiceCreatureType)
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Not A Type"); err == nil {
		t.Fatal("a string outside game.AllCreatureTypes was accepted")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, g.Seats[1].ID, "Elf"); err == nil {
		t.Fatal("somebody else answered the prompt")
	}
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("a refused answer closed the prompt")
	}
	before := handSize(me)
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "elf"); err != nil {
		t.Fatalf("lowercase elf: %v", err)
	}
	if got := handSize(me) - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Kindred Dominance / Raise the Palisade -----------------------------

func TestKindredDominanceDestroysEveryCreatureNotOfTheType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	changeling := pushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	oppElf := pushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	human := pushTribalCreature(g, opp.ID, "Soldier", "Creature — Human Soldier", 1, 1)

	castAndPause(t, g, "Kindred Dominance", "Sorcery", kindredDominanceOracle)
	if !battlefieldHas(g, goblin) || !battlefieldHas(g, human) {
		t.Fatal("creatures died before the type was chosen")
	}
	answerCreatureType(t, g, me.ID, "Elf")
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{{"Elf", elf, true}, {"opposing Elf", oppElf, true}, {"changeling", changeling, true}, {"Goblin", goblin, false}, {"Human", human, false}} {
		if got := battlefieldHas(g, tc.id); got != tc.want {
			t.Errorf("%s on the battlefield = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestKindredDominanceNamingATypeNobodyHasLeavesOnlyChangelings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	changeling := pushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	castAndPause(t, g, "Kindred Dominance", "Sorcery", kindredDominanceOracle)
	answerCreatureType(t, g, me.ID, "Sliver")
	if battlefieldHas(g, elf) {
		t.Error("the Elf survived a Sliver wipe")
	}
	if !battlefieldHas(g, changeling) {
		t.Error("the changeling died, but it is a Sliver")
	}
}

func TestRaiseThePalisadeBouncesEveryCreatureNotOfTheType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	oppHuman := pushTribalCreature(g, opp.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	changeling := pushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID,
	})

	castAndPause(t, g, "Raise the Palisade", "Sorcery", raiseThePalisadeOracle)
	answerCreatureType(t, g, me.ID, "Elf")
	if !battlefieldHas(g, elf) || !battlefieldHas(g, changeling) || !battlefieldHas(g, land) {
		t.Error("an Elf, the changeling or a land was bounced")
	}
	if battlefieldHas(g, goblin) || !inHandOf(me, goblin) {
		t.Error("the Goblin was not returned to its owner's hand")
	}
	if battlefieldHas(g, oppHuman) || !inHandOf(opp, oppHuman) {
		t.Error("the opposing Human was not returned to its owner's hand")
	}
}

// --- Banner of Kinship -------------------------------------------------

func TestBannerOfKinshipCountersPerCreatureOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elfA := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	oppElf := pushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)

	banner := castCatalogSpell(t, g, "Banner of Kinship", "Artifact", bannerOfKinshipOracle, nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("Banner of Kinship did not ask for a creature type as it entered")
	}
	answerCreatureType(t, g, me.ID, "Elf")

	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == banner {
				if c.NamedTribe != "Elf" {
					t.Errorf("NamedTribe = %q, want Elf", c.NamedTribe)
				}
				if got := c.Counters["fellowship"]; got != 3 {
					t.Errorf("fellowship counters = %d, want 3 (two Elves + the changeling)", got)
				}
			}
		}
	})
	if got := effectivePower(t, g, elfA); got != 4 {
		t.Errorf("my Elf power = %d, want 1 + 3", got)
	}
	if got := effectivePower(t, g, oppElf); got != 1 {
		t.Errorf("the opponent's Elf got the lord (%d); it pumps only creatures you control", got)
	}
	if got := effectivePower(t, g, goblin); got != 2 {
		t.Errorf("the Goblin got the lord (%d)", got)
	}
}

func TestBannerOfKinshipWithNoCreatureOfTheTypeHasNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	banner := castCatalogSpell(t, g, "Banner of Kinship", "Artifact", bannerOfKinshipOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Sliver")
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == banner && (c.Counters["fellowship"] != 0 || c.NamedTribe != "Sliver") {
				t.Errorf("banner = tribe %q, counters %v; want Sliver with none", c.NamedTribe, c.Counters)
			}
		}
	})
	if got := effectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2", got)
	}
}

// --- Patriarch's Bidding -----------------------------------------------

func TestPatriarchsBiddingAsksEverySeatInTurnOrderThenReturnsAllChosenTypes(t *testing.T) {
	g := newCatalogGame(t)
	s := g.Seats
	pushGraveyardCardTyped(s[0], "Dead Elf", "Creature — Elf")
	pushGraveyardCardTyped(s[1], "Dead Goblin", "Creature — Goblin")
	pushGraveyardCardTyped(s[2], "Dead Human", "Creature — Human")
	pushGraveyardCardTyped(s[3], "Dead Bear", "Creature — Bear")
	pushGraveyardCardTyped(s[3], "Other Elf", "Creature — Elf")
	pushGraveyardCardTyped(s[0], "Elf Charm", "Kindred Instant — Elf")

	castAndPause(t, g, "Patriarch's Bidding", "Sorcery", patriarchsBiddingOracle)
	for i, tribe := range []string{"Elf", "Goblin", "Human"} {
		c := pendingOfKind(g, game.PendingChoiceCreatureType)
		if c == nil || c.Chooser != s[i].ID {
			t.Fatalf("prompt %d: want one owed by seat %d, got %+v", i, i, c)
		}
		if elfBack := battlefieldHasName(g, "Dead Elf"); elfBack {
			t.Fatal("cards returned before the last seat answered")
		}
		answerCreatureType(t, g, s[i].ID, tribe)
	}
	// Seat 3 names a type with no cards in anyone's yard.
	answerCreatureType(t, g, s[3].ID, "Sliver")

	for _, tc := range []struct {
		name string
		want bool
	}{{"Dead Elf", true}, {"Other Elf", true}, {"Dead Goblin", true}, {"Dead Human", true}, {"Dead Bear", false}, {"Elf Charm", false}} {
		if got := battlefieldHasName(g, tc.name); got != tc.want {
			t.Errorf("%s on the battlefield = %v, want %v", tc.name, got, tc.want)
		}
	}
	// Each returns under its owner's control.
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Other Elf" && c.Controller != s[3].ID {
				t.Errorf("Other Elf came back under %s, want its owner", c.Controller)
			}
		}
	})
}

func battlefieldHasName(g *game.Game, name string) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				found = true
			}
		}
	})
	return found
}

// A seat that leaves while the prompt waits on it chooses nothing, and
// the chain still reaches the seats after it.
func TestPatriarchsBiddingCarriesOnWhenAChooserLeaves(t *testing.T) {
	g := newCatalogGame(t)
	s := g.Seats
	pushGraveyardCardTyped(s[0], "Dead Elf", "Creature — Elf")
	pushGraveyardCardTyped(s[2], "Dead Human", "Creature — Human")
	castAndPause(t, g, "Patriarch's Bidding", "Sorcery", patriarchsBiddingOracle)
	answerCreatureType(t, g, s[0].ID, "Elf")
	if c := pendingOfKind(g, game.PendingChoiceCreatureType); c == nil || c.Chooser != s[1].ID {
		t.Fatalf("want seat 1's prompt next, got %+v", c)
	}
	if err := g.Concede(s[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	answerCreatureType(t, g, s[2].ID, "Human")
	answerCreatureType(t, g, s[3].ID, "Sliver")
	if !battlefieldHasName(g, "Dead Elf") || !battlefieldHasName(g, "Dead Human") {
		t.Error("the chain stopped at the seat that left; the chosen types were not returned")
	}
}

// --- Tribal Unity ------------------------------------------------------

func TestTribalUnityPumpsEveryCreatureOfTheTypeForX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	oppElf := pushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	goblin := pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	changeling := pushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Tribal Unity", TypeLine: "Instant", OracleID: tribalUnityOracle, Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, active.ID, "Elf")

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"my Elf", elf, 4}, {"their Elf", oppElf, 4}, {"changeling", changeling, 4}, {"Goblin", goblin, 2}} {
		if got := effectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s power = %d, want %d", tc.name, got, tc.want)
		}
	}
	// The set is fixed at resolution (CR 611.2c).
	late := pushTribalCreature(g, me.ID, "Late Elf", "Creature — Elf", 1, 1)
	if got := effectivePower(t, g, late); got != 1 {
		t.Errorf("an Elf that arrived afterwards has power %d, want 1", got)
	}
}

// --- Kindred Charge ----------------------------------------------------

func TestKindredChargeCopiesEachCreatureOfTheTypeWithHasteAndExilesThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)

	castAndPause(t, g, "Kindred Charge", "Sorcery", kindredChargeOracle)
	answerCreatureType(t, g, me.ID, "Elf")

	tokens := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if IsToken(c) {
				tokens++
				if !hasKw(c.Keywords, "haste") {
					t.Errorf("token %s has no haste", c.Name)
				}
				if c.Name == "Goblin Guide" {
					t.Error("a Goblin was copied")
				}
			}
		}
	})
	if tokens != 2 {
		t.Fatalf("%d tokens, want 2 (one per Elf)", tokens)
	}
	for i := 0; i < 12 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if IsToken(c) {
				t.Errorf("token %s survived the end step", c.Name)
			}
		}
	})
}

// --- the seam itself ---------------------------------------------------

// Enumerator agreement: every answer legal.EnumerateFor offers for the
// resolution-time prompt is accepted by the dispatcher, each against a
// fresh Clone (which is also the undo path: the clone carries the
// continuation and finishes the card on its own).
func TestResolutionCreatureTypePromptIsAnsweredByTheEnumerator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("the enumerator offers no answer to the resolution-time prompt: the seat would wedge")
	}
	for _, other := range g.Seats[1:] {
		if got := legal.EnumerateFor(g, other.ID); len(got) != 0 {
			for _, m := range got {
				if m.Type == legal.TypeResolveChoice {
					t.Errorf("%s was offered an answer to somebody else's prompt: %q", other.ID, m.Label)
				}
			}
		}
	}
	before := handSize(me)
	for _, m := range moves {
		clone := g.Clone()
		a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}
		if err := actions.Dispatch(clone, a); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
			continue
		}
		if got := handSize(clone.Seats[0]); got <= before {
			t.Errorf("move %q on a clone drew nothing (%d -> %d): the continuation did not survive Clone", m.Label, before, got)
		}
	}
	// The original is untouched by the clones' answers.
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil || handSize(me) != before {
		t.Error("answering a clone changed the original game")
	}
}

// Snapshot mid-pause: a table waiting on the answer holds a closure the
// snapshot cannot carry, so it must say so rather than write a restore
// point that drops the rest of the card; once answered it is clean.
func TestSnapshotMidResolutionCreatureTypePauseIsNotRestorable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	castAndPause(t, g, "Distant Melody", "Sorcery", distantMelodyOracle)

	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames < 1 {
		t.Fatalf("census = %+v, want the creature-type frame counted and the table unrestorable", snap.Continuations)
	}
	if kinds := snap.Continuations.Kinds(); kinds["choiceResumeFrames"] == 0 && len(kinds) == 0 {
		t.Errorf("Kinds() = %v, want the blocking kind named", kinds)
	}
	answerCreatureType(t, g, me.ID, "Elf")
	if after := g.CaptureSnapshot(); after.Continuations.ChoiceResumeFrames != 0 {
		t.Errorf("census after the answer = %+v, want no frames", after.Continuations)
	}
}
