package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// --- Distant Melody ----------------------------------------------------

func TestDistantMelodyDrawsPerPermanentOfTheChosenTypeChangelingsCounted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	effectstest.PushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	// A non-creature permanent with the type counts: "permanent".
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kindred Banner", TypeLine: "Kindred Artifact — Elf",
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)
	before := effectstest.HandSize(me)
	if got := effectstest.HandSize(me); got != before {
		t.Fatalf("the draw happened before the answer")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	// Two Elves, the Kindred Elf artifact and the changeling; not the
	// opponent's Elf, not the Goblin.
	if got := effectstest.HandSize(me) - before; got != 4 {
		t.Errorf("drew %d, want 4 (two Elves + Kindred artifact + changeling)", got)
	}
}

func TestDistantMelodyAnswerNoCreatureHasDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)
	before := effectstest.HandSize(me)
	effectstest.AnswerCreatureType(t, g, me.ID, "Sliver")
	if got := effectstest.HandSize(me); got != before {
		t.Errorf("drew %d for a type nobody controls", got-before)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) != nil {
		t.Error("the prompt is still open after the answer")
	}
}

func TestDistantMelodyChangelingAloneStillCountsForAnyType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)
	before := effectstest.HandSize(me)
	effectstest.AnswerCreatureType(t, g, me.ID, "Sliver")
	if got := effectstest.HandSize(me) - before; got != 1 {
		t.Errorf("drew %d, want 1: a changeling is of every creature type", got)
	}
}

// An answer outside the vocabulary is refused and the prompt stays open
// (CR 205.3m); a lowercase spelling is normalised.
func TestResolutionCreatureTypeRefusesJunkAndNormalisesCase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)
	c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Not A Type"); err == nil {
		t.Fatal("a string outside game.AllCreatureTypes was accepted")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, g.Seats[1].ID, "Elf"); err == nil {
		t.Fatal("somebody else answered the prompt")
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("a refused answer closed the prompt")
	}
	before := effectstest.HandSize(me)
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "elf"); err != nil {
		t.Fatalf("lowercase elf: %v", err)
	}
	if got := effectstest.HandSize(me) - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Kindred Dominance / Raise the Palisade -----------------------------

func TestKindredDominanceDestroysEveryCreatureNotOfTheType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	goblin := effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	oppElf := effectstest.PushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	human := effectstest.PushTribalCreature(g, opp.ID, "Soldier", "Creature — Human Soldier", 1, 1)

	effectstest.CastAndPause(t, g, "Kindred Dominance", "Sorcery", effectstest.KindredDominanceOracle)
	if !effectstest.BattlefieldHas(g, goblin) || !effectstest.BattlefieldHas(g, human) {
		t.Fatal("creatures died before the type was chosen")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{{"Elf", elf, true}, {"opposing Elf", oppElf, true}, {"changeling", changeling, true}, {"Goblin", goblin, false}, {"Human", human, false}} {
		if got := effectstest.BattlefieldHas(g, tc.id); got != tc.want {
			t.Errorf("%s on the battlefield = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestKindredDominanceNamingATypeNobodyHasLeavesOnlyChangelings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	effectstest.CastAndPause(t, g, "Kindred Dominance", "Sorcery", effectstest.KindredDominanceOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Sliver")
	if effectstest.BattlefieldHas(g, elf) {
		t.Error("the Elf survived a Sliver wipe")
	}
	if !effectstest.BattlefieldHas(g, changeling) {
		t.Error("the changeling died, but it is a Sliver")
	}
}

func TestRaiseThePalisadeBouncesEveryCreatureNotOfTheType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	goblin := effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	oppHuman := effectstest.PushTribalCreature(g, opp.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastAndPause(t, g, "Raise the Palisade", "Sorcery", effectstest.RaiseThePalisadeOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if !effectstest.BattlefieldHas(g, elf) || !effectstest.BattlefieldHas(g, changeling) || !effectstest.BattlefieldHas(g, land) {
		t.Error("an Elf, the changeling or a land was bounced")
	}
	if effectstest.BattlefieldHas(g, goblin) || !effectstest.InHandOf(me, goblin) {
		t.Error("the Goblin was not returned to its owner's hand")
	}
	if effectstest.BattlefieldHas(g, oppHuman) || !effectstest.InHandOf(opp, oppHuman) {
		t.Error("the opposing Human was not returned to its owner's hand")
	}
}

// --- Banner of Kinship -------------------------------------------------

func TestBannerOfKinshipCountersPerCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elfA := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	oppElf := effectstest.PushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	goblin := effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)

	banner := effectstest.CastCatalogSpell(t, g, "Banner of Kinship", "Artifact", effectstest.BannerOfKinshipOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("Banner of Kinship did not ask for a creature type as it entered")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")

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
	if got := effectstest.EffectivePower(t, g, elfA); got != 4 {
		t.Errorf("my Elf power = %d, want 1 + 3", got)
	}
	if got := effectstest.EffectivePower(t, g, oppElf); got != 1 {
		t.Errorf("the opponent's Elf got the lord (%d); it pumps only creatures you control", got)
	}
	if got := effectstest.EffectivePower(t, g, goblin); got != 2 {
		t.Errorf("the Goblin got the lord (%d)", got)
	}
}

func TestBannerOfKinshipWithNoCreatureOfTheTypeHasNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	banner := effectstest.CastCatalogSpell(t, g, "Banner of Kinship", "Artifact", effectstest.BannerOfKinshipOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Sliver")
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == banner && (c.Counters["fellowship"] != 0 || c.NamedTribe != "Sliver") {
				t.Errorf("banner = tribe %q, counters %v; want Sliver with none", c.NamedTribe, c.Counters)
			}
		}
	})
	if got := effectstest.EffectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2", got)
	}
}

// --- Patriarch's Bidding -----------------------------------------------

func TestPatriarchsBiddingAsksEverySeatInTurnOrderThenReturnsAllChosenTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.PushGraveyardCardTyped(s[0], "Dead Elf", "Creature — Elf")
	effectstest.PushGraveyardCardTyped(s[1], "Dead Goblin", "Creature — Goblin")
	effectstest.PushGraveyardCardTyped(s[2], "Dead Human", "Creature — Human")
	effectstest.PushGraveyardCardTyped(s[3], "Dead Bear", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(s[3], "Other Elf", "Creature — Elf")
	effectstest.PushGraveyardCardTyped(s[0], "Elf Charm", "Kindred Instant — Elf")

	effectstest.CastAndPause(t, g, "Patriarch's Bidding", "Sorcery", effectstest.PatriarchsBiddingOracle)
	for i, tribe := range []string{"Elf", "Goblin", "Human"} {
		c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
		if c == nil || c.Chooser != s[i].ID {
			t.Fatalf("prompt %d: want one owed by seat %d, got %+v", i, i, c)
		}
		if elfBack := effectstest.BattlefieldHasName(g, "Dead Elf"); elfBack {
			t.Fatal("cards returned before the last seat answered")
		}
		effectstest.AnswerCreatureType(t, g, s[i].ID, tribe)
	}
	// Seat 3 names a type with no cards in anyone's yard.
	effectstest.AnswerCreatureType(t, g, s[3].ID, "Sliver")

	for _, tc := range []struct {
		name string
		want bool
	}{{"Dead Elf", true}, {"Other Elf", true}, {"Dead Goblin", true}, {"Dead Human", true}, {"Dead Bear", false}, {"Elf Charm", false}} {
		if got := effectstest.BattlefieldHasName(g, tc.name); got != tc.want {
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

// A seat that leaves while the prompt waits on it chooses nothing, and
// the chain still reaches the seats after it.
func TestPatriarchsBiddingCarriesOnWhenAChooserLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.PushGraveyardCardTyped(s[0], "Dead Elf", "Creature — Elf")
	effectstest.PushGraveyardCardTyped(s[2], "Dead Human", "Creature — Human")
	effectstest.CastAndPause(t, g, "Patriarch's Bidding", "Sorcery", effectstest.PatriarchsBiddingOracle)
	effectstest.AnswerCreatureType(t, g, s[0].ID, "Elf")
	if c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType); c == nil || c.Chooser != s[1].ID {
		t.Fatalf("want seat 1's prompt next, got %+v", c)
	}
	if err := g.Concede(s[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	effectstest.AnswerCreatureType(t, g, s[2].ID, "Human")
	effectstest.AnswerCreatureType(t, g, s[3].ID, "Sliver")
	if !effectstest.BattlefieldHasName(g, "Dead Elf") || !effectstest.BattlefieldHasName(g, "Dead Human") {
		t.Error("the chain stopped at the seat that left; the chosen types were not returned")
	}
}

// --- Tribal Unity ------------------------------------------------------

func TestTribalUnityPumpsEveryCreatureOfTheTypeForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	oppElf := effectstest.PushTribalCreature(g, opp.ID, "Elvish Visionary", "Creature — Elf Shaman", 1, 1)
	goblin := effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Mirror Entity", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Tribal Unity", TypeLine: "Instant", OracleID: effectstest.TribalUnityOracle, Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, active.ID, "Elf")

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"my Elf", elf, 4}, {"their Elf", oppElf, 4}, {"changeling", changeling, 4}, {"Goblin", goblin, 2}} {
		if got := effectstest.EffectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s power = %d, want %d", tc.name, got, tc.want)
		}
	}
	// The set is fixed at resolution (CR 611.2c).
	late := effectstest.PushTribalCreature(g, me.ID, "Late Elf", "Creature — Elf", 1, 1)
	if got := effectstest.EffectivePower(t, g, late); got != 1 {
		t.Errorf("an Elf that arrived afterwards has power %d, want 1", got)
	}
}

// --- Kindred Charge ----------------------------------------------------

func TestKindredChargeCopiesEachCreatureOfTheTypeWithHasteAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)

	effectstest.CastAndPause(t, g, "Kindred Charge", "Sorcery", effectstest.KindredChargeOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")

	tokens := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if effects.IsToken(c) {
				tokens++
				if !game.HasKeyword(&c, "haste") {
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
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if effects.IsToken(c) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)

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
	before := effectstest.HandSize(me)
	for _, m := range moves {
		clone := g.Clone()
		a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}
		if err := actions.Dispatch(clone, a); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
			continue
		}
		if got := effectstest.HandSize(clone.Seats[0]); got <= before {
			t.Errorf("move %q on a clone drew nothing (%d -> %d): the continuation did not survive Clone", m.Label, before, got)
		}
	}
	// The original is untouched by the clones' answers.
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil || effectstest.HandSize(me) != before {
		t.Error("answering a clone changed the original game")
	}
}

// Snapshot mid-pause: the rest of the card is a registered tail (#4253's
// second pass, ChooseCreatureTypeTail), so a table waiting on the answer
// is a restore point, and the restored answer still draws.
func TestSnapshotMidResolutionCreatureTypePauseIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushTribalCreature(g, g.Seats[0].ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.CastAndPause(t, g, "Distant Melody", "Sorcery", effectstest.DistantMelodyOracle)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/distant-melody-type-chosen")
	me := g.Seats[0]
	before := effectstest.HandSize(me)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if effectstest.HandSize(me) != before+1 {
		t.Errorf("hand %d after the restored answer, want %d (one Elf)", effectstest.HandSize(me), before+1)
	}
	if after := g.CaptureSnapshot(); after.Continuations.ChoiceResumeFrames != 0 {
		t.Errorf("census after the answer = %+v, want no frames", after.Continuations)
	}
}
