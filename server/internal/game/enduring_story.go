package game

import "github.com/google/uuid"

// enduring_story.go is storied (CR 702.195) and the enduring story it
// grants, The Hobbit's version of ascend's city's blessing.
//
// CR 702.195 (September 25, 2026 edition):
//
//	702.195a Storied is a static ability. It means "Any time you control
//	  three or more permanents that are artifacts, Sagas, and/or
//	  legendary and you don't have an enduring story, you have an
//	  enduring story for the rest of the game."
//	702.195b Enduring story is a designation that has no rules meaning
//	  other than to act as a marker that other rules and effects can
//	  identify. Any number of players may have the enduring story
//	  designation at the same time.
//	702.195c After a player gets an enduring story, continuous effects
//	  are reapplied before the game checks to see if the game state or
//	  preceding events have matched any trigger conditions.
//
// (ADR 0096's 2026-10-29 amendment, #4475.)
//
// It is the city's blessing with a different count, so it is a row of the
// same sweep (earned_designations.go) rather than a second mechanism: a
// bool on the player, written once by grantEnduringStoryLocked, never
// cleared, carried by Clone, the snapshot and undo, and public on the
// wire. Storied appears only on permanents, so there is no spell half.

// KeywordStoried is CR 702.195's canonical token — Scryfall's "Storied",
// lowercased.
const KeywordStoried = "storied"

// EnduringStoryThreshold is the three permanents storied counts (CR
// 702.195a).
const EnduringStoryThreshold = 3

// EventEnduringStory — a player got an enduring story (CR 702.195).
// Actor is that player; Source is the storied permanent whose check
// granted it.
//
// Emitted by grantEnduringStoryLocked only, and only on a real change, so
// it fires at most once per player per game. Added by #4475.
const EventEnduringStory EventKind = "enduring_story"

// EnduringStoryForEffect is "if you have an enduring story" / "as long as
// you have an enduring story". An ID that names no seat, and a seat that
// has left the game, do not have one. Caller must hold g.mu (the
// effect_api.go convention).
func (g *Game) EnduringStoryForEffect(id uuid.UUID) bool {
	p := g.playerByIDLocked(id)
	return p != nil && p.EnduringStory
}

// enduringStoryCounts is CR 702.195a's "permanents that are artifacts,
// Sagas, and/or legendary", read off the permanent's characteristics after
// continuous effects. Each permanent counts once.
func enduringStoryCounts(c *Card) bool {
	return c.IsArtifact() || c.IsLegendary() || c.HasSubtype("Saga")
}

func hasEnduringStory(p *Player) bool { return p.EnduringStory }

// grantEnduringStoryLocked is the one write to Player.EnduringStory. It
// reports whether the player got it just now. source is the storied
// permanent, for the log line.
//
// Caller must hold g.mu in write mode.
func (g *Game) grantEnduringStoryLocked(p *Player, source uuid.UUID) bool {
	if p == nil || p.Eliminated || p.EnduringStory {
		return false
	}
	p.EnduringStory = true
	g.EmitEvent(Event{
		Kind:   EventEnduringStory,
		Actor:  p.ID,
		Source: source,
	})
	return true
}
