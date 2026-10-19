package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// search_decline_test.go — #4134, CR 701.19: a "you may search" can be
// declined outright. Declining is not "find nothing": it skips the look,
// the shuffle, EventSearchLibrary (so "whenever a player searches"
// triggers stay quiet) and an Opposition Agent's control.

// answerSearchApply answers a search prompt with {apply: v} through the
// actions layer, as `caller` for `player`.
func answerSearchApply(g *game.Game, c *game.PendingChoice, caller, player uuid.UUID, v bool) error {
	raw, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "apply": v})
	return actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: player, Caller: caller, Params: raw})
}

func libraryIDs(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func sameIDList(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// declineVictimTable: the Trophy's victim has a Gaea's Cradle, a library
// of 30 cards with a basic in it, and the caster has an Archivist of
// Oghma (an opponent searching their library gains 1 and draws).
func declineVictimTable(t *testing.T) (g *game.Game, caster, victim *game.Player, target, basic uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	victim = top100Opponent(g)
	caster = top100ActiveSeat(g)
	pushCatalogPermanent(g, caster.ID, "Archivist of Oghma",
		"Creature — Halfling Cleric", archivistOfOghmaOracle, false)
	target = seedPermanentFor(g, victim.ID, "Gaea's Cradle", "Legendary Land")
	for i := 0; i < 30; i++ {
		stapleLibraryCard(victim, "Island Filler", "Creature — Bear")
	}
	basic = stapleLibraryCard(victim, "Forest", "Basic Land — Forest")
	return g, caster, victim, target, basic
}

func castTrophy(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Assassin's Trophy", "Instant", assassinsTrophyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func TestDeclinedTrophySearchDoesNotShuffleOrTriggerArchivist(t *testing.T) {
	g, caster, victim, target, basic := declineVictimTable(t)
	before := libraryIDs(victim)
	life, hand := caster.Life, caster.Hand.Size()

	castTrophy(t, g, target)
	c := searchChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the victim was not offered the search")
	}
	if !c.SearchIsOptional() || !c.SearchIsUndecided() {
		t.Fatalf("optional/undecided = %v/%v, want true/true: an optional search asks before it looks", c.SearchIsOptional(), c.SearchIsUndecided())
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !sameIDList(libraryIDs(victim), before) {
		t.Error("the library changed order: a declined search must not shuffle")
	}
	if g.Battlefield.Contains(basic) {
		t.Error("declining still fetched the basic")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == victim.ID {
			t.Errorf("EventSearchLibrary fired for a search that never happened: %+v", ev)
		}
	}
	if caster.Life != life || caster.Hand.Size() != hand {
		t.Errorf("Archivist of Oghma triggered on a declined search (life %d→%d, hand %d→%d)",
			life, caster.Life, hand, caster.Hand.Size())
	}
}

// The contrast: accepting and finding nothing IS a search.
func TestTrophySearchThatFindsNothingStillShufflesAndTriggersArchivist(t *testing.T) {
	g, caster, victim, target, basic := declineVictimTable(t)
	before := libraryIDs(victim)
	life := caster.Life

	castTrophy(t, g, target)
	answerSearchFailToFind(t, g, victim.ID)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(basic) {
		t.Error("finding nothing fetched the basic")
	}
	if sameIDList(libraryIDs(victim), before) {
		t.Error("the library kept its order: a search that finds nothing still shuffles")
	}
	if caster.Life != life+1 {
		t.Errorf("Archivist of Oghma life %d, want %d: the search happened", caster.Life, life+1)
	}
}

// Path to Exile: the search is the exile's continuation, and declining
// it skips the shuffle exactly the same way.
func TestDeclinedPathToExileSearchDoesNotShuffle(t *testing.T) {
	g := newCatalogGame(t)
	victim := top100Opponent(g)
	creature := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	for i := 0; i < 30; i++ {
		stapleLibraryCard(victim, "Filler", "Creature — Bear")
	}
	stapleLibraryCard(victim, "Forest", "Basic Land — Forest")
	before := libraryIDs(victim)

	castCatalogSpell(t, g, "Path to Exile", "Instant", "d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the creature's controller was not offered the search")
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if !g.Exile.Contains(creature) {
		t.Error("the creature was not exiled")
	}
	if !sameIDList(libraryIDs(victim), before) {
		t.Error("the library changed order after a declined Path to Exile search")
	}
}

// The enumerator offers the decline (it is the prompt's always-legal
// answer) and the engine accepts the move it offered.
func TestEnumeratorOffersDontSearch(t *testing.T) {
	g, _, victim, target, _ := declineVictimTable(t)
	castTrophy(t, g, target)
	c := searchChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	var decline *legal.Move
	for _, m := range legal.EnumerateFor(g, victim.ID) {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Apply    *bool  `json:"apply"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.ChoiceID == c.ID.String() && p.Apply != nil && !*p.Apply {
			m := m
			decline = &m
		}
	}
	if decline == nil {
		t.Fatal("the enumerator did not offer \"don't search\"")
	}
	if !decline.AlwaysLegal {
		t.Error("\"don't search\" should be the prompt's always-legal answer")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: victim.ID, Caller: victim.ID, Params: decline.Params}); err != nil {
		t.Fatalf("the engine refused the move the enumerator offered: %v", err)
	}
}

// A mandatory search carries no decline anywhere: not on the wire, not in
// the move list, and the engine refuses it.
func TestMandatorySearchHasNoDontSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		stapleLibraryCard(me, "Filler", "Creature — Bear")
	}
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, false)
	})
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := answerSearchApply(g, c, me.ID, me.ID, false); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("declining a mandatory search = %v, want ErrInvalidParam", err)
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			Apply *bool `json:"apply"`
		}
		if m.Type == legal.TypeResolveChoice && json.Unmarshal(m.Params, &p) == nil && p.Apply != nil {
			t.Errorf("a mandatory search enumerated an apply move: %s", m.Label)
		}
	}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, pc := range v.PendingChoices {
		if pc.Kind == string(game.PendingChoiceSearchLibrary) && (pc.SearchOptional || pc.SearchUndecided) {
			t.Errorf("a mandatory search is marked optional/undecided on the wire: %+v", pc)
		}
	}
}

// searchPromptView is the search prompt as `viewer` sees it.
func searchPromptView(t *testing.T, g *game.Game, viewer uuid.UUID) protocol.PendingChoiceView {
	t.Helper()
	for _, pc := range protocol.ViewOfGameFor(g, viewer.String()).PendingChoices {
		if pc.Kind == string(game.PendingChoiceSearchLibrary) {
			return pc
		}
	}
	t.Fatal("no search prompt in the view")
	return protocol.PendingChoiceView{}
}

// searchApplyMoves counts the {apply} moves the enumerator offers for the
// open search prompt, by value.
func searchApplyMoves(g *game.Game, seat uuid.UUID) (yes, no int) {
	for _, m := range legal.EnumerateFor(g, seat) {
		var p struct {
			Apply *bool `json:"apply"`
		}
		if m.Type != legal.TypeResolveChoice || json.Unmarshal(m.Params, &p) != nil || p.Apply == nil {
			continue
		}
		if *p.Apply {
			yes++
		} else {
			no++
		}
	}
	return
}

// An optional search asks BEFORE it looks (CR 701.19: looking through the
// library is searching). The wire shows the question with no candidates;
// only "Search" reveals the list, and from then on there is no decline.
func TestOptionalSearchIsAskedBeforeItLooks(t *testing.T) {
	g, _, victim, target, basic := declineVictimTable(t)
	castTrophy(t, g, target)
	c := searchChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("no prompt")
	}

	pc := searchPromptView(t, g, victim.ID)
	if !pc.SearchOptional || !pc.SearchUndecided {
		t.Errorf("search_optional/search_undecided = %v/%v, want true/true", pc.SearchOptional, pc.SearchUndecided)
	}
	if len(pc.Options) != 0 {
		t.Errorf("the candidate list is shown (%d cards) before the searcher decided to search", len(pc.Options))
	}
	if yes, no := searchApplyMoves(g, victim.ID); yes != 1 || no != 1 {
		t.Errorf("enumerator offered %d search / %d don't-search moves, want 1 / 1", yes, no)
	}
	// No card pick is offered or accepted yet.
	for _, m := range legal.EnumerateFor(g, victim.ID) {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if m.Type == legal.TypeResolveChoice && json.Unmarshal(m.Params, &p) == nil && len(p.CardIDs) > 0 {
			t.Errorf("a card pick was enumerated before the search began: %s", m.Label)
		}
	}
	if err := oaAnswer(g, c, victim.ID, victim.ID, basic); err == nil {
		t.Error("a card pick was accepted before the search began")
	}

	if err := answerSearchApply(g, c, victim.ID, victim.ID, true); err != nil {
		t.Fatalf("beginning: %v", err)
	}
	pc = searchPromptView(t, g, victim.ID)
	if pc.SearchOptional || pc.SearchUndecided {
		t.Errorf("a begun search still reports search_optional/search_undecided = %v/%v", pc.SearchOptional, pc.SearchUndecided)
	}
	if len(pc.Options) == 0 {
		t.Error("the searcher should be shown the candidates once the search began")
	}
	if yes, no := searchApplyMoves(g, victim.ID); yes != 0 || no != 0 {
		t.Errorf("a begun search still enumerates %d/%d apply moves", yes, no)
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("declining after looking = %v, want ErrInvalidParam: it would be a free look", err)
	}
	if searchChoiceFor(g, victim.ID) == nil {
		t.Fatal("the refused decline dropped the prompt")
	}
}

// Another seat's view of the same prompt shows nothing of it.
func TestUndecidedSearchIsHiddenFromOtherSeats(t *testing.T) {
	g, caster, victim, target, _ := declineVictimTable(t)
	castTrophy(t, g, target)
	for _, pc := range protocol.ViewOfGameFor(g, caster.ID.String()).PendingChoices {
		if pc.Kind != string(game.PendingChoiceSearchLibrary) {
			continue
		}
		if pc.SearchOptional || pc.SearchUndecided || len(pc.Options) != 0 {
			t.Errorf("another seat sees the search question: %+v", pc)
		}
	}
	_ = victim
}

// --- Opposition Agent ---

// oaDeclineTable: the Trophy's victim is the potential searcher; the
// Agent is on the caster's side of the table.
func oaDeclineTable(t *testing.T) (g *game.Game, caster, victim *game.Player, target, basic uuid.UUID) {
	t.Helper()
	g, caster, victim, target, basic = declineVictimTable(t)
	pushCatalogPermanent(g, caster.ID, "Opposition Agent", "Creature — Human Rogue", oppositionAgentOracle, false)
	return
}

// The question is the searcher's own, asked before any control: the
// candidates are hidden, nobody controls the searcher, and "don't
// search" ends it with no control record, no event and no shuffle.
func TestOppositionAgentNeverControlsADeclinedSearch(t *testing.T) {
	g, caster, victim, target, basic := oaDeclineTable(t)
	before := libraryIDs(victim)

	castTrophy(t, g, target)
	c := searchChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if !c.SearchIsUndecided() {
		t.Fatal("under an Opposition Agent the \"may\" is asked before the search begins")
	}
	if got := g.ControllerOfPlayer(victim.ID); got != victim.ID {
		t.Fatalf("the searcher is controlled by %v before deciding to search", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Fatalf("a control record exists before the search began: %+v", g.PlayerControls)
	}
	view := protocol.ViewOfGameFor(g, victim.ID.String())
	for _, pc := range view.PendingChoices {
		if pc.Kind == string(game.PendingChoiceSearchLibrary) {
			if !pc.SearchUndecided || !pc.SearchOptional {
				t.Errorf("view undecided/optional = %v/%v, want true/true", pc.SearchUndecided, pc.SearchOptional)
			}
			if len(pc.Options) != 0 {
				t.Error("the candidates are shown before the searcher has decided to search")
			}
		}
	}
	// No card pick is accepted yet.
	if err := oaAnswer(g, c, victim.ID, victim.ID, basic); err == nil {
		t.Fatal("a card pick was accepted before the search began")
	}

	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := g.ControllerOfPlayer(victim.ID); got != victim.ID {
		t.Errorf("the searcher is controlled by %v after declining", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("a control record outlived the declined search: %+v", g.PlayerControls)
	}
	if !sameIDList(libraryIDs(victim), before) {
		t.Error("the library changed order after declining")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == victim.ID {
			t.Error("EventSearchLibrary fired for a declined search")
		}
	}
	if oaInExile(g, basic) || g.Battlefield.Contains(basic) {
		t.Error("the basic moved on a declined search")
	}
	_ = caster
}

// Saying yes begins the search: only then does the Agent's controller take
// over, and they cannot undo the searcher's yes by declining.
func TestOppositionAgentControlsASearchOnceItBegins(t *testing.T) {
	g, caster, victim, target, basic := oaDeclineTable(t)
	castTrophy(t, g, target)
	c := searchChoiceFor(g, victim.ID)
	if c == nil || !c.SearchIsUndecided() {
		t.Fatal("expected an undecided prompt")
	}
	// A third seat cannot answer for the searcher.
	if err := answerSearchApply(g, c, caster.ID, caster.ID, true); err == nil {
		t.Fatal("the Agent's controller began the searcher's search")
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, true); err != nil {
		t.Fatalf("beginning the search: %v", err)
	}
	c = oaOpenSearch(t, g, victim, caster)
	if c.SearchIsUndecided() {
		t.Fatal("the prompt is still undecided after the searcher said yes")
	}
	// The controller cannot decline for them.
	if err := answerSearchApply(g, c, caster.ID, caster.ID, false); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("the Agent's controller declined the searcher's search: err = %v, want ErrInvalidParam", err)
	}
	if searchChoiceFor(g, victim.ID) == nil {
		t.Fatal("the refused decline dropped the prompt")
	}
	view := protocol.ViewOfGameFor(g, caster.ID.String())
	for _, pc := range view.PendingChoices {
		if pc.Kind == string(game.PendingChoiceSearchLibrary) && pc.SearchOptional {
			t.Error("the controller is shown a \"don't search\" they may not use")
		}
	}
	if err := oaAnswer(g, c, caster.ID, caster.ID, basic); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	if !oaInExile(g, basic) {
		t.Error("the found basic is not in exile")
	}
	oaAssertControlOver(t, g, victim)
}

// Moon-Blessed Cleric: "you may search ... then shuffle and put that card
// on top". Declining leaves the library exactly as it was; the cleric's
// owner, who may have set up their top cards, keeps them.
func TestMoonBlessedClericDeclinedSearchKeepsTheLibraryOrder(t *testing.T) {
	g, me, opp, ids := rooncClericGame(t)
	before := libraryIDs(me)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if err := answerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != len(ids) || !sameIDList(libraryIDs(me), before) {
		t.Error("the library changed after a declined search")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == me.ID {
			t.Error("a declined search was announced")
		}
	}
	if len(revealsFor(g, opp.ID)) != 0 {
		t.Error("nothing was revealed")
	}
}

// --- #4140: nothing matches ---

// Path to Exile over a library with no basic land: the controller is
// still asked, can decline (no shuffle, no event, no Archivist trigger)
// or find nothing (shuffles, triggers).
func zeroMatchPathTable(t *testing.T) (g *game.Game, caster, victim *game.Player, creature uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	victim = top100Opponent(g)
	caster = top100ActiveSeat(g)
	pushCatalogPermanent(g, caster.ID, "Archivist of Oghma",
		"Creature — Halfling Cleric", archivistOfOghmaOracle, false)
	creature = seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	for i := 0; i < 30; i++ {
		stapleLibraryCard(victim, "Filler", "Creature — Bear")
	}
	return
}

func castZeroMatchPath(t *testing.T, g *game.Game, creature uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Path to Exile", "Instant", "d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)
}

func TestZeroMatchPathToExileDeclineDoesNotShuffleOrTrigger(t *testing.T) {
	g, caster, victim, creature := zeroMatchPathTable(t)
	before := libraryIDs(victim)
	life := caster.Life
	castZeroMatchPath(t, g, creature)
	c := searchChoiceFor(g, victim.ID)
	if c == nil || len(c.SearchCards) != 0 {
		t.Fatalf("a search with no basic land to find should still ask, with no candidates: %+v", c)
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sameIDList(libraryIDs(victim), before) {
		t.Error("the library changed order after a declined search")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == victim.ID {
			t.Error("a declined search was announced")
		}
	}
	if caster.Life != life {
		t.Error("Archivist of Oghma triggered on a declined search")
	}
}

func TestZeroMatchPathToExileFindNothingShufflesAndTriggers(t *testing.T) {
	g, caster, victim, creature := zeroMatchPathTable(t)
	before := libraryIDs(victim)
	life := caster.Life
	castZeroMatchPath(t, g, creature)
	answerSearchFailToFind(t, g, victim.ID)
	passPriorityAroundTable(t, g)
	if sameIDList(libraryIDs(victim), before) {
		t.Error("the library kept its order: a search that finds nothing still shuffles")
	}
	if caster.Life != life+1 {
		t.Errorf("Archivist life %d, want %d: the search happened", caster.Life, life+1)
	}
}

// Under an Opposition Agent the zero-match "may" is asked first too: no
// control until the searcher says yes.
func TestZeroMatchOptionalSearchUnderAnAgentAsksFirst(t *testing.T) {
	g, caster, victim, creature := zeroMatchPathTable(t)
	pushCatalogPermanent(g, caster.ID, "Opposition Agent", "Creature — Human Rogue", oppositionAgentOracle, false)
	castZeroMatchPath(t, g, creature)
	c := searchChoiceFor(g, victim.ID)
	if c == nil || !c.SearchIsUndecided() {
		t.Fatalf("expected an undecided prompt: %+v", c)
	}
	if len(g.PlayerControls) != 0 {
		t.Fatalf("a control record exists before the search began: %+v", g.PlayerControls)
	}
	if err := answerSearchApply(g, c, victim.ID, victim.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("a control record outlived the declined search: %+v", g.PlayerControls)
	}
}
