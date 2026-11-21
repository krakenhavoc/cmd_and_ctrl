package testso

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// oaPermission is the play permission `holder` has over the exiled card,
// or nil.
func oaPermission(g *game.Game, holder, id uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				perm = g.CastPermissionForLocked(holder, c, game.ZoneExile)
			}
		}
	})
	return perm
}

// Demonic Tutor: the Agent's controller answers the search, the
// searcher may not, the card is exiled face up instead of reaching the
// hand, the library is still shuffled, and the controller may play the
// card with mana of any colour.
func TestOppositionAgentControlsADemonicTutor(t *testing.T) {
	g, searcher, ctl, third, _ := effectstest.OaTable(t)
	bolt := effectstest.PushLibraryCardForTest(searcher, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle,
	})
	hand := searcher.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)

	// CR 723.5: the searcher makes no decision; nor does a third player.
	if err := effectstest.OaAnswer(g, c, searcher.ID, searcher.ID, bolt); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("the searcher answered their own search: err = %v, want ErrPlayerControlled", err)
	}
	if err := effectstest.OaAnswer(g, c, third.ID, third.ID, bolt); err == nil {
		t.Fatal("a third player answered the search")
	}
	// CR 723.5b: the controller may only answer the search, nothing else
	// as the searcher.
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeDrawCard, Player: searcher.ID, Caller: ctl.ID}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("the controller drew a card as the searcher: err = %v, want ErrPlayerControlled", err)
	}
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, bolt); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}

	if searcher.Hand.Contains(bolt) || searcher.Library.Contains(bolt) {
		t.Fatal("the found card reached the hand or stayed in the library")
	}
	if !effectstest.OaInExile(g, bolt) {
		t.Fatal("the found card is not in exile face up")
	}
	if got := searcher.Hand.Size(); got != hand {
		t.Errorf("searcher's hand went from %d to %d", hand, got)
	}
	shuffled := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == searcher.ID && ev.Label == game.SearchShuffledLabel {
			shuffled = true
		}
	}
	if !shuffled {
		t.Error("the library was not shuffled after the search")
	}
	perm := oaPermission(g, ctl.ID, bolt)
	if perm == nil {
		t.Fatal("the Agent's controller has no permission to play the exiled card")
	}
	if !perm.AnyColor || perm.CastOnly || perm.Duration.Kind != game.WhileInZone {
		t.Errorf("permission = %+v, want play (not cast only), any colour, while it stays exiled", *perm)
	}
	if oaPermission(g, searcher.ID, bolt) != nil || oaPermission(g, third.ID, bolt) != nil {
		t.Error("someone other than the Agent's controller may play the exiled card")
	}
	effectstest.OaAssertControlOver(t, g, searcher)

	// "you may spend mana as though it were mana of any color": the Bolt
	// is cast off a Swamp, by the Agent's controller, once they hold
	// priority.
	effectstest.E2Swamps(g, ctl.ID, 1)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	life := searcher.Life
	if err := g.CastSpell(ctl.ID, bolt, game.CastSpellParams{
		FromZone: "exile", AutoTap: true, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: searcher.ID}},
	}); err != nil {
		t.Fatalf("cast the exiled Bolt off a Swamp: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if searcher.Life != life-3 {
		t.Errorf("searcher's life = %d, want %d", searcher.Life, life-3)
	}
}

// Evolving Wilds: a fetch to the battlefield. The basic is exiled, not
// put onto the battlefield, and is the Agent's controller's to play as a
// land (a play permission, not cast only).
func TestOppositionAgentExilesAFetchedLand(t *testing.T) {
	g, searcher, ctl, _, _ := effectstest.OaTable(t)
	wilds := effectstest.PushCatalogPermanent(g, searcher.ID, "Evolving Wilds", "Land", effectstest.EvolvingWildsOracle, false)
	forest := effectstest.StapleLibraryCard(searcher, "Forest", effectstest.OaForest)
	if err := g.ActivateCatalogAbility(searcher.ID, wilds, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, forest); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	if g.Battlefield.Contains(forest) {
		t.Fatal("the fetched basic entered the battlefield")
	}
	if !effectstest.OaInExile(g, forest) {
		t.Fatal("the fetched basic is not in exile")
	}
	if perm := oaPermission(g, ctl.ID, forest); perm == nil || perm.CastOnly {
		t.Fatalf("permission = %+v, want a play permission for the Agent's controller", perm)
	}
	effectstest.OaAssertControlOver(t, g, searcher)
}

// Explosive Vegetation: "up to two" to the battlefield. Both found
// lands are exiled; the one left behind stays in the library.
func TestOppositionAgentExilesAnUpToNSearch(t *testing.T) {
	g, searcher, ctl, _, _ := effectstest.OaTable(t)
	f1 := effectstest.StapleLibraryCard(searcher, "Forest", effectstest.OaForest)
	f2 := effectstest.StapleLibraryCard(searcher, "Forest", effectstest.OaForest)
	f3 := effectstest.StapleLibraryCard(searcher, "Forest", effectstest.OaForest)
	effectstest.CastCatalogSpell(t, g, "Explosive Vegetation", "Sorcery", effectstest.OaExplosiveVegetation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, f1, f2); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	for _, id := range []uuid.UUID{f1, f2} {
		if !effectstest.OaInExile(g, id) || oaPermission(g, ctl.ID, id) == nil {
			t.Errorf("found land %v: not exiled with a permission for the Agent's controller", id)
		}
	}
	if !searcher.Library.Contains(f3) {
		t.Error("the land not found left the library")
	}
	effectstest.OaAssertControlOver(t, g, searcher)
}

// Buried Alive: "up to three" into the graveyard. Exiled instead.
func TestOppositionAgentExilesAGraveyardSearch(t *testing.T) {
	g, searcher, ctl, _, _ := effectstest.OaTable(t)
	b1 := effectstest.StapleLibraryCard(searcher, "Bear A", effectstest.OaBear)
	b2 := effectstest.StapleLibraryCard(searcher, "Bear B", effectstest.OaBear)
	effectstest.CastCatalogSpell(t, g, "Buried Alive", "Sorcery", effectstest.OaBuriedAlive, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, b1, b2); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	for _, id := range []uuid.UUID{b1, b2} {
		if searcher.Graveyard.Contains(id) || !effectstest.OaInExile(g, id) {
			t.Errorf("found creature %v went to the graveyard rather than exile", id)
		}
	}
	effectstest.OaAssertControlOver(t, g, searcher)
}

// Vampiric Tutor: "then shuffle and put that card on top". The card is
// exiled and never reaches the top of the library.
func TestOppositionAgentExilesAToTopTutor(t *testing.T) {
	g, searcher, ctl, _, _ := effectstest.OaTable(t)
	needle := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Vamp Needle"})
	effectstest.CastCatalogSpell(t, g, "Vampiric Tutor", "Instant", effectstest.OaVampiricTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, needle); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	if searcher.Library.Contains(needle) || !effectstest.OaInExile(g, needle) {
		t.Fatal("the tutored card was not exiled instead of going on top")
	}
	effectstest.OaAssertControlOver(t, g, searcher)
}

// The Agent's controller searching their own library is untouched: no
// control, and the card goes where the tutor says.
func TestOppositionAgentLeavesItsControllersOwnSearchAlone(t *testing.T) {
	g, _, ctl, _, _ := effectstest.OaTable(t)
	// Make the Agent's controller the active player casting the tutor.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = ctl.Seat; g.Turn.PriorityHolder = ctl.Seat })
	needle := effectstest.PushLibraryCardForTest(ctl, game.Card{Name: "Own Needle"})
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.ControllerOfPlayer(ctl.ID); got != ctl.ID {
		t.Fatalf("the Agent's controller is controlled by %v during their own search", got)
	}
	effectstest.AnswerSearchByID(t, g, ctl.ID, needle)
	if !ctl.Hand.Contains(needle) {
		t.Fatal("the Agent's controller's own tutor did not reach their hand")
	}
}

// The permission outlives the Agent ("for as long as they remain
// exiled"), and an Agent that has left the battlefield controls nothing.
func TestOppositionAgentPermissionOutlivesTheAgent(t *testing.T) {
	g, searcher, ctl, _, agent := effectstest.OaTable(t)
	needle := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)
	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, needle); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(agent); err != nil {
			t.Fatalf("remove the Agent: %v", err)
		}
	})
	if oaPermission(g, ctl.ID, needle) == nil {
		t.Fatal("the permission ended with the Agent")
	}
	// A second search with no Agent: the searcher decides, and the card
	// goes to the hand.
	second := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Second"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.ControllerOfPlayer(searcher.ID); got != searcher.ID {
		t.Fatalf("with no Agent, the searcher is controlled by %v", got)
	}
	effectstest.AnswerSearchByID(t, g, searcher.ID, second)
	if !searcher.Hand.Contains(second) {
		t.Fatal("with no Agent, the tutor did not reach the hand")
	}
}

// oaSeatView is seat's PlayerView in v.
func oaSeatView(t *testing.T, v protocol.GameView, seat uuid.UUID) protocol.PlayerView {
	t.Helper()
	for _, p := range v.Seats {
		if p.ID == seat.String() {
			return p
		}
	}
	t.Fatalf("no seat %v in the view", seat)
	return protocol.PlayerView{}
}

// oaSearchPrompt is the search prompt in v, or nil.
func oaSearchPrompt(v protocol.GameView) *protocol.PendingChoiceView {
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceSearchLibrary) {
			return &v.PendingChoices[i]
		}
	}
	return nil
}

// oaKnownHand is how many cards of seat's hand the viewer is shown face
// up.
func oaKnownHand(t *testing.T, v protocol.GameView, seat uuid.UUID) int {
	t.Helper()
	n := 0
	for _, c := range oaSeatView(t, v, seat).Hand.Cards {
		if c.KnownByYou && c.Name != "" {
			n++
		}
	}
	return n
}

// CR 723.4 and the view: during the search the controller is shown the
// prompt (with on_behalf_of), its candidates and the searcher's hand; the
// searcher is told who is choosing and sees no candidate list; a third
// player sees neither the candidates nor the hand. Once the search is
// over, the controller sees the hand no longer.
func TestOppositionAgentViewRedaction(t *testing.T) {
	g, searcher, ctl, third, _ := effectstest.OaTable(t)
	needle := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Secret Needle"})
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OaOpenSearch(t, g, searcher, ctl)

	cv := protocol.ViewOfGameFor(g, ctl.ID.String())
	p := oaSearchPrompt(cv)
	if p == nil || p.Chooser != ctl.ID.String() || p.OnBehalfOf != searcher.ID.String() {
		t.Fatalf("controller's prompt = %+v, want it addressed to them on the searcher's behalf", p)
	}
	named := false
	for _, o := range p.Options {
		if o.InstanceID == needle.String() && o.Name == "Secret Needle" {
			named = true
		}
	}
	if !named {
		t.Error("the controller is not shown the candidates (CR 723.4)")
	}
	if searcher.Hand.Size() == 0 {
		t.Fatal("setup: the searcher has no hand to look at")
	}
	if got := oaKnownHand(t, cv, searcher.ID); got != searcher.Hand.Size() {
		t.Errorf("the controller sees %d of the searcher's %d hand cards during the search (CR 723.4)", got, searcher.Hand.Size())
	}
	if len(cv.PlayerControls) != 1 || cv.PlayerControls[0].Until != string(game.ControlUntilSearch) ||
		cv.PlayerControls[0].SourceName != "Opposition Agent" {
		t.Errorf("player_controls = %+v, want the one search control", cv.PlayerControls)
	}

	sv := protocol.ViewOfGameFor(g, searcher.ID.String())
	if p := oaSearchPrompt(sv); p == nil || p.Chooser != ctl.ID.String() || len(p.Options) != 0 {
		t.Errorf("searcher's prompt = %+v, want it shown as the controller's, with no candidate list", p)
	}

	tv := protocol.ViewOfGameFor(g, third.ID.String())
	if p := oaSearchPrompt(tv); p != nil && len(p.Options) != 0 {
		t.Errorf("a third player is shown %d candidates", len(p.Options))
	}
	if got := oaKnownHand(t, tv, searcher.ID); got != 0 {
		t.Errorf("a third player sees %d of the searcher's hand cards", got)
	}

	if err := effectstest.OaAnswer(g, c, ctl.ID, ctl.ID, needle); err != nil {
		t.Fatalf("the controller's answer: %v", err)
	}
	// The found card is face up in exile, so everyone sees it.
	for _, viewer := range []uuid.UUID{ctl.ID, searcher.ID, third.ID} {
		v := protocol.ViewOfGameFor(g, viewer.String())
		seen := false
		for _, e := range v.Exile.Cards {
			if e.InstanceID == needle.String() && e.Name == "Secret Needle" {
				seen = true
			}
		}
		if !seen {
			t.Errorf("viewer %v does not see the exiled card face up", viewer)
		}
	}
	after := protocol.ViewOfGameFor(g, ctl.ID.String())
	if len(after.PlayerControls) != 0 {
		t.Errorf("player_controls after the search = %+v", after.PlayerControls)
	}
	if got := oaKnownHand(t, after, searcher.ID); got != 0 {
		t.Errorf("the controller still sees %d of the searcher's hand cards after the search", got)
	}
}

// Undo restores a clone (ADR 0041): the search-scoped control is plain
// data with a prompt ID, so a clone taken mid-search is controlled the
// same way, and answering it there exiles the card there only.
func TestOppositionAgentControlSurvivesAClone(t *testing.T) {
	g, searcher, ctl, _, _ := effectstest.OaTable(t)
	needle := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Needle"})
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OaOpenSearch(t, g, searcher, ctl)
	clone := g.Clone()
	if got := clone.ControllerOfPlayer(searcher.ID); got != ctl.ID {
		t.Fatalf("the clone's searcher is controlled by %v, want the Agent's controller", got)
	}
	c := effectstest.SearchChoiceFor(clone, searcher.ID)
	if c == nil {
		t.Fatal("the clone has no search prompt")
	}
	if err := effectstest.OaAnswer(clone, c, ctl.ID, ctl.ID, needle); err != nil {
		t.Fatalf("answer on the clone: %v", err)
	}
	if !effectstest.OaInExile(clone, needle) || effectstest.OaInExile(g, needle) {
		t.Fatal("the clone and the original share the take")
	}
	if got := g.ControllerOfPlayer(searcher.ID); got != ctl.ID {
		t.Fatal("answering the clone ended the original's control")
	}
}

// The enumerator: the searcher is offered nothing while controlled, and
// the Agent's controller is offered the search's answers as the searcher
// (Player set to them), which the actions layer accepts.
func TestOppositionAgentEnumeratorOffersTheSearchToTheController(t *testing.T) {
	g, searcher, ctl, third, _ := effectstest.OaTable(t)
	needle := effectstest.PushLibraryCardForTest(searcher, game.Card{Name: "Needle"})
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OaOpenSearch(t, g, searcher, ctl)
	if moves := legal.EnumerateFor(g, searcher.ID); len(moves) != 0 {
		t.Fatalf("the controlled searcher is offered %d moves", len(moves))
	}
	if moves := legal.EnumerateFor(g, third.ID); len(moves) != 0 {
		t.Fatalf("a third player is offered %d moves during the search", len(moves))
	}
	var take *legal.Move
	for _, m := range legal.EnumerateFor(g, ctl.ID) {
		if m.Player != searcher.ID || m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if len(p.CardIDs) == 1 && p.CardIDs[0] == needle.String() {
			m := m
			take = &m
		}
	}
	if take == nil {
		t.Fatal("the controller is not offered the take of the needle as the searcher")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(take.Type), Player: take.Player, Caller: ctl.ID, Params: take.Params}); err != nil {
		t.Fatalf("the enumerated take: %v", err)
	}
	if !effectstest.OaInExile(g, needle) {
		t.Fatal("the enumerated take did not exile the needle")
	}
}
