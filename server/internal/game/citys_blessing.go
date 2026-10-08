package game

import "github.com/google/uuid"

// citys_blessing.go is ascend (CR 702.131) and the city's blessing it
// grants.
//
// CR 702.131 (September 25, 2026 edition):
//
//	702.131a Ascend on an instant or sorcery spell represents a spell
//	  ability. It means "If you control ten or more permanents and you
//	  don't have the city's blessing, you get the city's blessing for
//	  the rest of the game."
//	702.131b Ascend on a permanent represents a static ability. It means
//	  "Any time you control ten or more permanents, you get the city's
//	  blessing for the rest of the game."
//	702.131c The city's blessing is a player designation. It stays with
//	  the player for the rest of the game, even if the permanent that
//	  caused them to get it leaves the battlefield or they no longer
//	  control ten or more permanents.
//
// (ADR 0096's 2026-10-08 amendment, #2696.)
//
// A DESIGNATION ON THE PLAYER, NOT A READ OF THE BOARD. Before this
// file the catalog approximated "if you have the city's blessing" with
// "you control ten or more permanents" read live, so a board that
// shrank below ten shut every card that asked. The blessing is earned
// once and kept; the only honest place to keep it is the player, as the
// monarch is kept on the game. It rides the snapshot, undo and the
// wire, and is never cleared by anything.
//
// TWO WAYS TO GET IT, one per half of the rule:
//
//   - A PERMANENT with ascend (702.131b) is a static ability that has
//     been true "any time" the controller has ten permanents. The
//     engine settles it in citysBlessingSweepLocked, run from the same
//     pass as the state-based actions: it is not one (704 does not list
//     it), but that pass is the one place that runs after every action
//     with the board settled and before any player receives priority,
//     which is exactly when an "any time" static is observable. A
//     permanent that enters as the tenth, a token that arrives, a
//     control change, or a keyword granted to a permanent already
//     there all reach it the same way.
//   - An INSTANT or SORCERY with ascend (702.131a) checks as it
//     resolves (ascendSpellLocked), before the rest of the spell, so
//     "if you have the city's blessing, instead …" reads the answer the
//     spell has just earned. A permanent spell is not asked: its ascend
//     is the static above, which begins when it lands.
//
// Neither path ever takes the blessing away.

// KeywordAscend is CR 702.131's canonical token — Scryfall's "Ascend",
// lowercased.
const KeywordAscend = "ascend"

// CitysBlessingThreshold is the ten permanents ascend counts (CR
// 702.131a/b).
const CitysBlessingThreshold = 10

// EventCitysBlessing — a player got the city's blessing (CR 702.131).
// Actor is that player; Source is the ascend permanent or spell whose
// check granted it, uuid.Nil if none is known.
//
// Emitted by grantCitysBlessingLocked only, and only on a real change:
// a player who already has the blessing does not "get" it again, so
// "whenever you get the city's blessing" would fire once per game.
// Added by #2696.
const EventCitysBlessing EventKind = "citys_blessing"

// CitysBlessingForEffect is "if you have the city's blessing". An ID
// that names no seat, and a seat that has left the game, do not have
// it. Caller must hold g.mu (the effect_api.go convention).
func (g *Game) CitysBlessingForEffect(id uuid.UUID) bool {
	p := g.playerByIDLocked(id)
	return p != nil && p.CitysBlessing
}

// grantCitysBlessingLocked is the one write to Player.CitysBlessing.
// It reports whether the player got it just now. source is the object
// whose ascend did it, for the log line.
//
// Caller must hold g.mu in write mode.
func (g *Game) grantCitysBlessingLocked(p *Player, source uuid.UUID) bool {
	if p == nil || p.Eliminated || p.CitysBlessing {
		return false
	}
	p.CitysBlessing = true
	g.EmitEvent(Event{
		Kind:   EventCitysBlessing,
		Actor:  p.ID,
		Source: source,
	})
	return true
}

// permanentsControlledLocked is how many permanents id controls — the
// number ascend counts. Tokens count; a phased-out permanent is not on
// the battlefield slice and so does not (CR 702.26b).
func (g *Game) permanentsControlledLocked(id uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == id {
			n++
		}
	}
	return n
}

// citysBlessingSweepLocked is CR 702.131b: any player who controls an
// ascend permanent and ten or more permanents gets the city's blessing.
// It reports whether anyone did, so the caller can refresh the layer
// pass the new designation feeds (the grant also bumps layerVersion,
// through the event, so the refresh is a no-op for a caller that does
// not care).
//
// One walk over the battlefield, and none at all once every living
// player has the blessing, which is where a long game spends its time.
//
// Caller must hold g.mu in write mode.
func (g *Game) citysBlessingSweepLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	// Per-seat tallies in arrays indexed by seat position, so the common
	// pass (nobody close to ten, or everybody already blessed) allocates
	// nothing: this runs after every action in every game, bots included.
	var (
		counts   [MaxPlayers]int
		ascender [MaxPlayers]uuid.UUID
		waiting  [MaxPlayers]bool
		someone  bool
	)
	for i, p := range g.Seats {
		if i < MaxPlayers && p != nil && !p.Eliminated && !p.CitysBlessing {
			waiting[i], someone = true, true
		}
	}
	if !someone {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		for s, p := range g.Seats {
			if s >= MaxPlayers || !waiting[s] || p.ID != c.Controller {
				continue
			}
			counts[s]++
			if ascender[s] == uuid.Nil && HasKeyword(c, KeywordAscend) {
				ascender[s] = c.InstanceID
			}
			break
		}
	}
	granted := false
	for s, p := range g.Seats {
		if s >= MaxPlayers || !waiting[s] || ascender[s] == uuid.Nil || counts[s] < CitysBlessingThreshold {
			continue
		}
		if g.grantCitysBlessingLocked(p, ascender[s]) {
			granted = true
		}
	}
	return granted
}

// ascendSpellLocked is CR 702.131a for the spell that is resolving: an
// instant or sorcery with ascend checks the controller's permanents
// before any of its other instructions. A copy resolves with the
// ability too (CR 707.10: a copy has the copied spell's text). Quiet
// for a permanent spell, whose ascend is the static in the sweep.
//
// Caller must hold g.mu in write mode.
func (g *Game) ascendSpellLocked(spell *Card, item *StackItem) {
	if spell == nil || item == nil || spell.IsPermanent() || !HasKeyword(spell, KeywordAscend) {
		return
	}
	p := g.playerByIDLocked(item.Controller)
	if p == nil || p.Eliminated || p.CitysBlessing {
		return
	}
	if g.permanentsControlledLocked(p.ID) >= CitysBlessingThreshold {
		g.grantCitysBlessingLocked(p, spell.InstanceID)
	}
}
