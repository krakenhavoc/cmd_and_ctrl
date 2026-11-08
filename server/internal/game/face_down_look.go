package game

import (
	"slices"

	"github.com/google/uuid"
)

// face_down_look.go — #5251: "you may look at face-down creatures you
// don't control any time".
//
// Two printed forms, one rule:
//
//   - a STATIC ability of a permanent on the battlefield, Found Footage's
//     "You may look at face-down creatures your opponents control any
//     time" (CardDef.LooksAtFaceDownCreatures);
//   - a resolved ability's grant for a duration, Lumbering Laundry's
//     "{2}: Until end of turn, you may look at face-down creatures you
//     don't control any time" (ModLookAtFaceDownCreatures, a
//     ScopedEffect swept at cleanup, CR 514.2).
//
// "Your opponents control" and "you don't control" are the same set of
// creatures here: the engine has no teams, and in a free-for-all every
// other player plays as an individual against you (CR 806.1), so every
// other player is an opponent. A creature you control you may already
// look at (CR 708.5).
//
// NOTHING IS WRITTEN INTO Card.KnownBy, for the reason library_top.go and
// hand_revealed.go give for their standing looks (ADR 0199): the set of
// objects the grant covers changes every time a creature is turned face
// down or up, enters, leaves or changes hands, and the grant itself ends
// the moment its source leaves or the turn's cleanup runs. A stamp on
// each creature would have to be added and taken off again on every one
// of those, and the first one forgotten is a leak nobody reports (ADR
// 0082 A3). So the answer is derived where it is read: the view asks
// FaceDownLookGrantsLocked once per frame, under its read lock, and the
// per-viewer projection makes the grant's holder — and only that seat —
// able to read the card. faceDownViewersLocked, which decides what
// KnownBy is set to, is unchanged; this is the live half beside it.
//
// It is kept apart from KnownBy for a second reason, which is the one
// that matters most: a spectator knows a card only when EVERY seat knows
// it (ADR 0069's 2026-09-30 amendment). In a two-player game a Found
// Footage would make both seats able to look at every face-down
// creature, and if the grant were a knower the spectator would be handed
// the card. A grant is a permission to one player, not a reveal to the
// table, so the projection never counts it toward "every seat".

// ModLookAtFaceDownCreatures is "until end of turn, <Player> may look at
// face-down creatures they don't control any time" (Lumbering Laundry,
// #5251). Reads Player, who must be seated. Scope ScopeGame. A rules
// gate read by no layer: FaceDownLookGrantsLocked reads it, for the
// view. An on-disk identity: never renamed, never reused.
const ModLookAtFaceDownCreatures ModKind = "lookAtFaceDownCreatures"

// LookAtFaceDownCreaturesThisTurnForEffect registers a
// ModLookAtFaceDownCreatures record for `player` until end of turn: it
// lasts the turn whether or not its source stays (CR 611.2a). Registers
// nothing for a player who is not seated or has left the game. Reports
// whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) LookAtFaceDownCreaturesThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModLookAtFaceDownCreatures, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// faceDownLookHoldersLocked is every seated player who may look at
// face-down creatures they don't control right now, without repeats, in
// the order the grants were found: each controller of a battlefield
// permanent with the static, then each live record's player.
//
// The static is read through catalogAbilityKeyOf, so a Found Footage
// that has lost its abilities (CR 613.1f) grants nothing, and a
// phased-out one is not on the battlefield to be read (CR 702.26b). The
// record's duration is tested here as well as in the cleanup sweep, as
// HandRevealedLocked tests its own: the reader has to be right between
// the sweeps.
//
// Caller must hold g.mu (read or write).
func (g *Game) faceDownLookHoldersLocked() []uuid.UUID {
	var out []uuid.UUID
	add := func(id uuid.UUID) {
		if id == uuid.Nil || slices.Contains(out, id) {
			return
		}
		if p := g.playerByIDLocked(id); p == nil || p.Eliminated {
			return
		}
		out = append(out, id)
	}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			if def := catalogDef(catalogAbilityKeyOf(src)); def != nil && def.LooksAtFaceDownCreatures {
				add(src.Controller)
			}
		}
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModLookAtFaceDownCreatures || g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			add(m.Player)
		}
	}
	return out
}

// FaceDownLookGrantsLocked is who may look at each face-down creature on
// the battlefield right now through a look grant, keyed by the
// creature's instance ID: every grant holder but the creature's
// controller, who may look already (CR 708.5) and is answered by
// KnownBy. Nil when nobody holds a grant or no face-down creature is
// there, which is almost every frame of almost every game.
//
// A face-down CREATURE: the CR 708.2 permanent states only, and only
// while the object is a creature, so Yedora's face-down Forest (a land
// its effect listed, #1270) is not covered. A face-down spell on the
// stack is not a creature, and a face-down card in exile is not a
// permanent; neither is covered.
//
// The protocol projection is the reader (protocol.stampFaceDownLooks).
// It must never be written into KnownBy: see the file comment.
//
// Caller must hold g.mu (read or write).
func (g *Game) FaceDownLookGrantsLocked() map[uuid.UUID][]uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	holders := g.faceDownLookHoldersLocked()
	if len(holders) == 0 {
		return nil
	}
	var out map[uuid.UUID][]uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.faceDownPermanent() || !IsCreaturePtr(c) {
			continue
		}
		for _, h := range holders {
			if h == c.Controller {
				continue
			}
			if out == nil {
				out = make(map[uuid.UUID][]uuid.UUID)
			}
			out[c.InstanceID] = append(out[c.InstanceID], h)
		}
	}
	return out
}

// MayLookAtFaceDownCreatureLocked reports whether `viewer` may look at
// the face-down creature `id` through a look grant right now. False for
// a creature the viewer controls (that look is CR 708.5's, answered by
// KnownBy) and for anything that is not a face-down creature on the
// battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) MayLookAtFaceDownCreatureLocked(viewer, id uuid.UUID) bool {
	return slices.Contains(g.FaceDownLookGrantsLocked()[id], viewer)
}
