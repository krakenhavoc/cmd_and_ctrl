package game

import (
	"cmp"
	"slices"

	"github.com/google/uuid"
)

// search_control.go — controlling a player while they search (CR 723.2,
// #3998, ADR 0166's 2026-10-16 amendment): Opposition Agent.
//
//	You control your opponents while they're searching their
//	libraries.
//	While an opponent is searching their library, they exile each card
//	they find. You may play those cards for as long as they remain
//	exiled, and you may spend mana as though it were mana of any color
//	to cast them.
//
// Both statics live in the one search primitive (effect_api.go,
// SearchLibraryThenForEffect), so every search the catalog makes is
// covered by construction: every destination (hand, battlefield, top of
// the library, graveyard, exile) and every count ("a card", "up to N",
// "any number").
//
// # Control (CR 723.2, 723.4, 723.5)
//
// The span of a search the engine can see is its search prompt: the
// searcher looks through their library and decides while it is open,
// and nothing else is decided in between (the prompt is mid-resolution
// and blocks the table). So the control is a PlayerControl record with
// Until "search" and stage ControlSearching, created as the prompt is
// queued, and live exactly while
//
//   - that prompt is still open, and
//   - the Agent that made it is still on the battlefield, under the
//     record's controller, with the ability (CatalogAbilityKey).
//
// Everything ADR 0166 built then reads it unchanged: the prompt is
// shown to the controller with on_behalf_of (CR 723.5), the actions
// layer re-addresses the controller's answer to the searcher and
// refuses the searcher's own, FilterViewFor makes the controller a
// knower of what the searcher knows (the candidates and the hand, CR
// 723.4) for as long as the record is live and not a moment longer, and
// the enumerator offers the controller the prompt. A search the rules
// decide alone (one forced take) opens no prompt and so no control:
// there is no choice for the controller to make. (A search with no match
// used to be in that list; since #4140 it asks.)
//
// The record is plain data with a prompt ID, never a continuation, so a
// table mid-search is still a restore point (ADR 0041).
//
// # Exile and the play permission
//
// The second static is a replacement of the found cards' destination,
// applied where the cards are taken (executeSearchTakeLocked): the
// search's destination becomes exile (face up), the cards still go
// through the shared exit route (so CR 614 and CR 903.9 see the move),
// and the cards that land are granted to the Agent's controller with a
// WhileInZone play permission and AnyColor spending. The rest of the
// search is unchanged: a reveal still happens, the library is still
// shuffled, and the effect's own continuation still runs, with no found
// cards (they did not arrive where it put them, the reading
// landedInZoneLocked gives every "exile it instead").
//
// It applies to every search an opponent makes of their own library
// while an Agent is on the battlefield, whether or not a prompt was
// opened, which is what "they exile each card they find" says. A search
// of another player's library (Bribery) is not "their library" and is
// left alone.
//
// # More than one Agent
//
// One record per Agent, oldest first, so the newest live one decides
// (CR 723.1a), and the found cards go to the newest Agent's controller:
// CR 616.1 lets the affected player, whose decisions that controller is
// making, pick which replacement applies.

const (
	// ControlSearching is a search-scoped control (#3998): live while
	// its prompt is open and its Agent is still there.
	ControlSearching ControlStage = "searching"
	// ControlUntilSearch is Opposition Agent's: control for as long as
	// the controlled player is searching their library (CR 723.2).
	ControlUntilSearch ControlUntil = "search"
)

// CatalogControlsOpponentsSearches reports whether a battlefield
// permanent with this catalog key controls its controller's opponents
// while they search and exiles what they find. Populated from
// effects.Spec.ControlsOpponentsSearches. A nil hook finds no such
// permanent.
var CatalogControlsOpponentsSearches func(key string) bool

// controlsSearches reports whether c, a permanent, has the ability
// right now (CR 613.1f: one that lost its abilities does not).
func controlsSearches(c *Card) bool {
	if CatalogControlsOpponentsSearches == nil || c == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	return key != "" && CatalogControlsOpponentsSearches(key)
}

// searchAgentsOverLocked is every permanent that controls searcher's
// searches right now: the ability, under a controller who is an
// opponent of searcher and still in the game. Oldest first (CR 613.7's
// timestamp, then battlefield order), so the last is the one that
// applies (CR 723.1a). Caller must hold g.mu.
func (g *Game) searchAgentsOverLocked(searcher uuid.UUID) []*Card {
	if g.Battlefield == nil || CatalogControlsOpponentsSearches == nil {
		return nil
	}
	var out []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == uuid.Nil || c.Controller == searcher || !g.playerInGameLocked(c.Controller) {
			continue
		}
		if controlsSearches(c) {
			out = append(out, c)
		}
	}
	slices.SortStableFunc(out, func(a, b *Card) int {
		return cmp.Compare(a.EnteredBattlefieldAt, b.EnteredBattlefieldAt)
	})
	return out
}

// beginSearchControlsLocked records a search-scoped control over
// searcher for each Agent, now that the search prompt `prompt` is open.
// A fresh slice, never a write into the old one, which an undo entry may
// share. Caller must hold g.mu.
func (g *Game) beginSearchControlsLocked(searcher, prompt uuid.UUID) {
	if prompt == uuid.Nil {
		return
	}
	agents := g.searchAgentsOverLocked(searcher)
	if len(agents) == 0 {
		return
	}
	next := make([]PlayerControl, len(g.PlayerControls), len(g.PlayerControls)+len(agents))
	copy(next, g.PlayerControls)
	for _, a := range agents {
		next = append(next, PlayerControl{
			Controller: a.Controller,
			Controlled: searcher,
			Source:     a.InstanceID,
			SourceName: a.Name,
			Stage:      ControlSearching,
			Until:      ControlUntilSearch,
			Prompt:     prompt,
		})
	}
	g.PlayerControls = next
}

// searchControlLiveLocked is controlLiveLocked's ControlSearching case:
// the prompt is open, and the Agent is on the battlefield under the
// record's controller with its ability. Caller must hold g.mu.
func (g *Game) searchControlLiveLocked(r PlayerControl) bool {
	if r.UntilKind() != ControlUntilSearch || r.Prompt == uuid.Nil || g.choiceIndexLocked(r.Prompt) < 0 {
		return false
	}
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == r.Source {
			return c.Controller == r.Controller && controlsSearches(c)
		}
	}
	return false
}

// searchExileHolderLocked is the player the cards searcher finds go to
// (exiled, with a play permission), or uuid.Nil when no Agent applies:
// the newest Agent's controller. Caller must hold g.mu.
func (g *Game) searchExileHolderLocked(searcher uuid.UUID) uuid.UUID {
	agents := g.searchAgentsOverLocked(searcher)
	if len(agents) == 0 {
		return uuid.Nil
	}
	return agents[len(agents)-1].Controller
}

// SearchExilePermission is the play permission Opposition Agent grants
// over the cards it exiled: "You may play those cards for as long as
// they remain exiled, and you may spend mana as though it were mana of
// any color to cast them." Play, not cast, so a land is played as the
// holder's land drop under the usual timing.
func SearchExilePermission() CastPermission {
	return CastPermission{
		Duration: WhileInZoneDuration(),
		AnyColor: true,
	}
}
