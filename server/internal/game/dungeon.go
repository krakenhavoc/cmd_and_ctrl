package game

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// dungeon.go is CR 309 (dungeons) and CR 701.49 (venture into the
// dungeon), ADR 0203 (#4096).
//
//	309.2a If a player ventures into the dungeon while they don't own
//	       a dungeon card in the command zone, they choose a dungeon
//	       card they own from outside the game and put it into the
//	       command zone.
//	309.4  Each dungeon card has a series of rooms connected to one
//	       another with arrows. A player uses a venture marker placed
//	       on the dungeon card they own to indicate which room they
//	       are currently in.
//	309.4c Each room has a triggered ability called a room ability
//	       ... "When you move your venture marker into this room,
//	       [effect.]" ... Each room ability is controlled by the player
//	       who owns the dungeon card that is that ability's source.
//	309.6  If a player's venture marker is on the bottommost room of a
//	       dungeon card, and that dungeon card isn't the source of a
//	       room ability that has triggered but not yet left the stack,
//	       the dungeon card's owner removes it from the game. (This is
//	       a state-based action, CR 704.5t.)
//	309.7  A player completes a dungeon as that dungeon card is
//	       removed from the game.
//	701.49b ... they choose an adjacent room, following the direction
//	       of an arrow pointing away from their current room ...
//	701.49c If a player is instructed to venture into the dungeon while
//	       their venture marker is in the bottommost room of a dungeon
//	       card, they remove that dungeon card from the game ... They
//	       then choose an appropriate dungeon card they own from
//	       outside the game, put it into the command zone, and put
//	       their venture marker on the topmost room of that dungeon.
//	701.49d Venture into [quality] is a variant of venture into the
//	       dungeon ... If they already own a dungeon card in the
//	       command zone, they follow the normal procedure.
//
// (Quoted from the September 25, 2026 edition, the one AGENTS.md §6
// pins.)
//
// Four pieces, each with one home:
//
//   - THE DUNGEON CARDS are data: a DungeonDef per dungeon, its rooms
//     and the arrows between them, registered by the catalog
//     (cards/effects/dungeons.go). Another dungeon is another row, not
//     more code here.
//   - THE ROOM ABILITIES are ordinary catalog triggered abilities filed
//     under the dungeon's catalog key (DungeonKey), each watching
//     EventVentureRoom for its own room. The harvester walks every
//     player's dungeon as one more command-zone object
//     (harvestFromDungeonsLocked), so a room that targets (Forge, Trap!,
//     Arena) uses the structured targeting every trigger uses, and the
//     item names its catalog row, so a table with one waiting is a
//     restore point.
//   - THE MARKER is Player.Dungeon: which dungeon the player owns in
//     the command zone, the object's ID (the room abilities' source),
//     and the room the venture marker is in. Plain data, cloned and
//     snapshotted with the seat.
//   - THE CHOICE of a branch is a keyed option pick ("venture/room"),
//     each option naming its room, so the prompt holds no closure and
//     the bot can price a room by name.
//
// The dungeon card is not a Card in a Zone. Nothing in the rules can
// touch it — it is not a permanent, it can't be cast, and it never
// leaves the command zone except by leaving the game (CR 309.2c) — so
// a container every zone verb could resolve would only be a door to
// close. The harvester sees it as a transient object built from the
// marker (dungeonObject), which is all a triggered ability needs.

// DungeonKeyPrefix marks a catalog key as a dungeon's. A real Scryfall
// oracle ID is a UUID, so the prefix can never collide with one.
const DungeonKeyPrefix = "dungeon:"

// DungeonKey is the catalog key a dungeon's room abilities are filed
// under: "dungeon:" + its key. An on-disk identity (a waiting room
// ability's catalog ref names it): never renamed, never reused.
func DungeonKey(key string) string { return DungeonKeyPrefix + key }

// UndercityKey is the Undercity's key (CR 726.2): the dungeon the
// initiative ventures into.
const UndercityKey = "undercity"

// DungeonRoom is one room of a dungeon card (CR 309.4).
type DungeonRoom struct {
	// Key names the room within its dungeon: "secret-entrance". An
	// on-disk identity, carried by Player.Dungeon.Room.
	Key string
	// Name is the printed room name (CR 309.4b: flavour, but it is how
	// a player refers to the room).
	Name string
	// Text is the room ability's printed effect.
	Text string
	// LeadsTo are the rooms the room's arrows point to, by key. Empty
	// for the bottommost room.
	LeadsTo []string
}

// DungeonDef is one dungeon card as data.
type DungeonDef struct {
	// Key names the dungeon: "undercity". An on-disk identity.
	Key string
	// Name is the printed name: "Undercity".
	Name string
	// Rooms are the dungeon's rooms, the topmost first (CR 309.4a).
	Rooms []DungeonRoom
	// VentureIntoOnly is the Undercity's "You can't enter this dungeon
	// unless you 'venture into Undercity.'": a plain venture into the
	// dungeon never offers it (CR 701.49d).
	VentureIntoOnly bool
}

// Room is the room with this key, or nil.
func (d *DungeonDef) Room(key string) *DungeonRoom {
	if d == nil {
		return nil
	}
	for i := range d.Rooms {
		if d.Rooms[i].Key == key {
			return &d.Rooms[i]
		}
	}
	return nil
}

// IsBottommost reports whether the room with this key is the
// dungeon's last: no arrow leads away from it.
func (d *DungeonDef) IsBottommost(key string) bool {
	r := d.Room(key)
	return r != nil && len(r.LeadsTo) == 0
}

var (
	dungeonMu    sync.RWMutex
	dungeonDefs  = map[string]*DungeonDef{}
	dungeonOrder []string
)

// RegisterDungeon files a dungeon card's rooms. Called from the
// catalog's init. Panics on a malformed dungeon: no key, no rooms, a
// duplicate key, a room key used twice, an arrow to a room that does
// not exist, or a room nothing leads to other than the first.
func RegisterDungeon(d DungeonDef) {
	if d.Key == "" || d.Name == "" || len(d.Rooms) == 0 {
		panic(fmt.Sprintf("game: RegisterDungeon %q: a dungeon needs a key, a name and rooms", d.Key))
	}
	seen := map[string]bool{}
	for _, r := range d.Rooms {
		if r.Key == "" || r.Name == "" || seen[r.Key] {
			panic(fmt.Sprintf("game: RegisterDungeon %q: room %q has no key, no name, or a duplicate key", d.Key, r.Key))
		}
		seen[r.Key] = true
	}
	reached := map[string]bool{d.Rooms[0].Key: true}
	bottom := 0
	for _, r := range d.Rooms {
		if len(r.LeadsTo) == 0 {
			bottom++
		}
		for _, to := range r.LeadsTo {
			if !seen[to] {
				panic(fmt.Sprintf("game: RegisterDungeon %q: room %q leads to unknown room %q", d.Key, r.Key, to))
			}
			reached[to] = true
		}
	}
	if len(reached) != len(d.Rooms) || bottom != 1 {
		panic(fmt.Sprintf("game: RegisterDungeon %q: every room must be reachable and exactly one must be bottommost", d.Key))
	}
	dungeonMu.Lock()
	defer dungeonMu.Unlock()
	if _, dup := dungeonDefs[d.Key]; dup {
		panic(fmt.Sprintf("game: RegisterDungeon: duplicate dungeon %q", d.Key))
	}
	cp := d
	cp.Rooms = make([]DungeonRoom, len(d.Rooms))
	for i, r := range d.Rooms {
		cp.Rooms[i] = r
		cp.Rooms[i].LeadsTo = append([]string(nil), r.LeadsTo...)
	}
	dungeonDefs[d.Key] = &cp
	dungeonOrder = append(dungeonOrder, d.Key)
}

// DungeonByKey is the registered dungeon with this key, or nil. The
// result is shared: callers must not write through it.
func DungeonByKey(key string) *DungeonDef {
	dungeonMu.RLock()
	defer dungeonMu.RUnlock()
	return dungeonDefs[key]
}

// dungeonsForVenture lists the dungeons a venture may enter, in
// registration order: the one named for "venture into [quality]", or
// every dungeon a plain venture may enter.
func dungeonsForVenture(into string) []*DungeonDef {
	dungeonMu.RLock()
	defer dungeonMu.RUnlock()
	if into != "" {
		if d := dungeonDefs[into]; d != nil {
			return []*DungeonDef{d}
		}
		return nil
	}
	var out []*DungeonDef
	for _, k := range dungeonOrder {
		if d := dungeonDefs[k]; !d.VentureIntoOnly {
			out = append(out, d)
		}
	}
	return out
}

// DungeonState is a player's dungeon card in the command zone and the
// venture marker on it (CR 309.4).
type DungeonState struct {
	// ID is the dungeon card's object ID: the source of its room
	// abilities, minted as the dungeon enters the command zone. A
	// dungeon completed and entered again is a new object.
	ID uuid.UUID `json:"id"`
	// Key is the dungeon's DungeonDef.Key.
	Key string `json:"key"`
	// Room is the room the venture marker is in, by key.
	Room string `json:"room"`
}

func cloneDungeonState(d *DungeonState) *DungeonState {
	if d == nil {
		return nil
	}
	cp := *d
	return &cp
}

// dungeonObject is the dungeon card as the harvester and a room
// ability's targeting see it: an object in its owner's command zone,
// owned and controlled by them (CR 309.4c), whose catalog key files the
// room abilities. Built fresh on every read; nothing holds it.
func dungeonObject(p *Player) Card {
	name := p.Dungeon.Key
	if d := DungeonByKey(p.Dungeon.Key); d != nil {
		name = d.Name
	}
	return Card{
		InstanceID: p.Dungeon.ID,
		Name:       name,
		OracleID:   DungeonKey(p.Dungeon.Key),
		Owner:      p.ID,
		Controller: p.ID,
	}
}

// harvestFromDungeonsLocked walks every living player's dungeon card
// for room abilities that trigger on this event (CR 309.4c: "As long
// as a dungeon card is in the command zone, its abilities may
// trigger"). One more command-zone object in the harvester's walk, as
// emblems are (harvestFromEmblemsLocked).
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestFromDungeonsLocked(pass *harvestPass) {
	if pass.ev.Kind != EventVentureRoom {
		// Room abilities watch nothing else; a dungeon costs an event
		// that is not a venture one comparison.
		return
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Dungeon == nil {
			continue
		}
		z := newZone(ZoneCommand, p.ID)
		z.PushTop(dungeonObject(p))
		g.harvestFromZone(pass, z)
	}
}

// EventVentureRoom — a player moved their venture marker into a room
// (CR 309.4c's trigger condition): entering a dungeon's topmost room
// or advancing to the next. Actor is the player, CardID the dungeon
// card's object ID, Label "<dungeon key>/<room key>"
// (VentureRoomLabel), Source the object whose
// effect made them venture (uuid.Nil for the initiative's own
// triggers, which have none). The room abilities watch it.
const EventVentureRoom EventKind = "venture_room"

// EventDungeonCompleted — a player completed a dungeon (CR 309.7): the
// dungeon card was removed from the game, by the state-based action
// (CR 704.5t) or by venturing from its bottommost room (CR 701.49c).
// Actor is the player, CardID the dungeon card's object ID, Label the
// dungeon's key. "Whenever you complete a dungeon" watches it, and the
// layer pass bumps on it for "as long as you've completed a dungeon".
const EventDungeonCompleted EventKind = "dungeon_completed"

// ventureRoomQuestion is the branch prompt's reason, and so its name in
// the action dock.
const ventureRoomQuestion = "venture: choose the next room"

// ventureDungeonQuestion is the prompt when more than one dungeon may
// be entered (CR 309.2a).
const ventureDungeonQuestion = "venture: choose a dungeon"

// VentureForEffect is the keyword action (CR 701.49): `player`
// ventures into the dungeon, or into the dungeon named by `into`
// (CR 701.49d: "venture into Undercity" is into = UndercityKey).
// `source` is the object whose effect made them venture.
//
//   - No dungeon: enter one (asked when more than one may be entered)
//     and put the marker on its topmost room.
//   - Marker in any room but the last: advance along an arrow (asked
//     when two arrows lead on).
//   - Marker in the last room: complete the dungeon, then enter one.
//
// Each move emits EventVentureRoom, which triggers that room's
// ability. A player who has left the game ventures nowhere (CR
// 800.4a). It can pause for the room or dungeon choice; the rest runs
// from the answer, through a keyed continuation.
//
// Caller must hold g.mu in write mode.
func (g *Game) VentureForEffect(player, source uuid.UUID, into string) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return nil
	}
	if p.Dungeon != nil {
		d := DungeonByKey(p.Dungeon.Key)
		if d == nil {
			// A restore from a binary with a dungeon this one does not
			// have is refused at load (unknownDungeonKeys), so this is
			// unreachable in play.
			return fmt.Errorf("game: venture: unknown dungeon %q", p.Dungeon.Key)
		}
		room := d.Room(p.Dungeon.Room)
		if room != nil && len(room.LeadsTo) > 0 {
			return g.ventureAdvanceLocked(p, d, room, source)
		}
		// CR 701.49c: the bottommost room. Complete it, then enter one.
		g.completeDungeonLocked(p)
	}
	return g.ventureEnterLocked(p, source, into)
}

// ventureEnterLocked is CR 309.2a / 701.49a: choose a dungeon that may
// be entered and put the marker on its topmost room.
func (g *Game) ventureEnterLocked(p *Player, source uuid.UUID, into string) error {
	cands := dungeonsForVenture(into)
	switch len(cands) {
	case 0:
		return nil
	case 1:
		g.enterDungeonLocked(p, cands[0], source)
		return nil
	}
	opts := make([]ChoiceOption, len(cands))
	for i, d := range cands {
		opts[i] = ChoiceOption{Label: d.Name, Dungeon: d.Key}
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  p.ID,
		Source:   source,
		Question: ventureDungeonQuestion,
		Options:  opts,
		ThenKey:  ventureDungeonThen,
		Carry:    []uuid.UUID{source},
	})
	return nil
}

// ventureAdvanceLocked is CR 701.49b: follow an arrow from `room`.
func (g *Game) ventureAdvanceLocked(p *Player, d *DungeonDef, room *DungeonRoom, source uuid.UUID) error {
	if len(room.LeadsTo) == 1 {
		g.moveVentureMarkerLocked(p, room.LeadsTo[0], source)
		return nil
	}
	opts := make([]ChoiceOption, 0, len(room.LeadsTo))
	for _, k := range room.LeadsTo {
		next := d.Room(k)
		opts = append(opts, ChoiceOption{Label: roomOptionLabel(next), Dungeon: d.Key, Room: next.Key})
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  p.ID,
		Source:   source,
		Question: ventureRoomQuestion,
		Options:  opts,
		ThenKey:  ventureRoomThen,
		Carry:    []uuid.UUID{p.Dungeon.ID, source},
	})
	return nil
}

// roomOptionLabel is how a room reads on the branch prompt: its name
// and what it does.
func roomOptionLabel(r *DungeonRoom) string {
	if r.Text == "" {
		return r.Name
	}
	return r.Name + " — " + r.Text
}

// ventureRoomThen is the branch prompt's keyed continuation (ADR 0203
// §3): the chosen option names its room. Carry is [dungeon object ID,
// venture source]. The marker moves only if the player still owns that
// same dungeon object with the marker where it was asked from — which
// nothing between the question and the answer can change, but a
// restore across a deploy is not trusted to have kept it.
var ventureRoomThen = RegisterOptionPickThen("venture/room", func(g *Game, r OptionPicked) error {
	if r.Option == nil || len(r.Carry) < 2 {
		return nil
	}
	p := g.playerByIDLocked(r.Chooser)
	if p == nil || p.Eliminated || p.Dungeon == nil || p.Dungeon.ID != r.Carry[0] {
		return nil
	}
	d := DungeonByKey(p.Dungeon.Key)
	cur := d.Room(p.Dungeon.Room)
	if cur == nil || !containsString(cur.LeadsTo, r.Option.Room) {
		return nil
	}
	g.moveVentureMarkerLocked(p, r.Option.Room, r.Carry[1])
	return nil
})

// ventureDungeonThen is the dungeon prompt's keyed continuation: the
// chosen option names its dungeon. Carry is [venture source].
var ventureDungeonThen = RegisterOptionPickThen("venture/dungeon", func(g *Game, r OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	p := g.playerByIDLocked(r.Chooser)
	if p == nil || p.Eliminated || p.Dungeon != nil {
		return nil
	}
	d := DungeonByKey(r.Option.Dungeon)
	if d == nil {
		return nil
	}
	source := uuid.Nil
	if len(r.Carry) > 0 {
		source = r.Carry[0]
	}
	g.enterDungeonLocked(p, d, source)
	return nil
})

func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// enterDungeonLocked puts `d` into p's command zone with the marker on
// its topmost room (CR 309.4a), as a new object.
func (g *Game) enterDungeonLocked(p *Player, d *DungeonDef, source uuid.UUID) {
	p.Dungeon = &DungeonState{ID: uuid.New(), Key: d.Key}
	g.moveVentureMarkerLocked(p, d.Rooms[0].Key, source)
}

// moveVentureMarkerLocked is the one write of DungeonState.Room. It
// emits EventVentureRoom, which triggers the room's ability.
func (g *Game) moveVentureMarkerLocked(p *Player, room string, source uuid.UUID) {
	p.Dungeon.Room = room
	g.EmitEvent(Event{
		Kind:   EventVentureRoom,
		Actor:  p.ID,
		CardID: p.Dungeon.ID,
		Source: source,
		Label:  VentureRoomLabel(p.Dungeon.Key, room),
	})
}

// VentureRoomLabel is EventVentureRoom's Label: "<dungeon>/<room>", so
// a room ability and the log both know which dungeon's room it is.
func VentureRoomLabel(dungeon, room string) string { return dungeon + "/" + room }

// SplitVentureRoomLabel undoes VentureRoomLabel.
func SplitVentureRoomLabel(label string) (dungeon, room string) {
	dungeon, room, _ = strings.Cut(label, "/")
	return dungeon, room
}

// completeDungeonLocked removes p's dungeon card from the game (CR
// 309.7): the dungeon is completed. The one write of
// Player.CompletedDungeons.
func (g *Game) completeDungeonLocked(p *Player) {
	if p.Dungeon == nil {
		return
	}
	done := *p.Dungeon
	p.Dungeon = nil
	p.CompletedDungeons = append(p.CompletedDungeons, done.Key)
	g.EmitEvent(Event{
		Kind:   EventDungeonCompleted,
		Actor:  p.ID,
		CardID: done.ID,
		Label:  done.Key,
	})
}

// dungeonCompletionSBALocked is CR 704.5t (309.6): a venture marker on
// a dungeon's bottommost room, with no room ability of that dungeon
// triggered and not yet off the stack, completes the dungeon. Reports
// whether one was.
//
// Caller must hold g.mu in write mode.
func (g *Game) dungeonCompletionSBALocked() bool {
	fired := false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Dungeon == nil {
			continue
		}
		d := DungeonByKey(p.Dungeon.Key)
		if d == nil || !d.IsBottommost(p.Dungeon.Room) {
			continue
		}
		if g.roomAbilityOutstandingLocked(p.Dungeon.ID) {
			continue
		}
		g.completeDungeonLocked(p)
		fired = true
	}
	return fired
}

// roomAbilityOutstandingLocked reports whether a room ability of the
// dungeon object `id` has triggered and not yet left the stack: waiting
// to be put on it, being announced (its targets asked), or on it.
func (g *Game) roomAbilityOutstandingLocked(id uuid.UUID) bool {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == id {
			return true
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == id {
			return true
		}
	}
	for i := range g.PendingChoices {
		if g.PendingChoices[i].Source == id {
			return true
		}
	}
	return false
}

// CompletedADungeon reports whether `player` has completed a dungeon
// this game — "if you've completed a dungeon". Caller must hold g.mu.
func CompletedADungeon(g *Game, player uuid.UUID) bool {
	if g == nil {
		return false
	}
	p := g.playerByIDLocked(player)
	return p != nil && len(p.CompletedDungeons) > 0
}

// CompletedDungeon reports whether `player` has completed the dungeon
// with this key — "if you haven't completed Tomb of Annihilation".
// Caller must hold g.mu.
func CompletedDungeon(g *Game, player uuid.UUID, key string) bool {
	if g == nil {
		return false
	}
	p := g.playerByIDLocked(player)
	return p != nil && containsString(p.CompletedDungeons, key)
}

// DungeonOf is a copy of `player`'s dungeon state, or nil when they own
// no dungeon card in the command zone. Caller must hold g.mu.
func DungeonOf(g *Game, player uuid.UUID) *DungeonState {
	if g == nil {
		return nil
	}
	return cloneDungeonState(g.playerByIDLocked(player).dungeonOrNil())
}

func (p *Player) dungeonOrNil() *DungeonState {
	if p == nil {
		return nil
	}
	return p.Dungeon
}

// unknownDungeonKeys lists the restored seats' dungeons this binary
// cannot run — a dungeon or a room it does not register — as
// checkEffectKeys' "unknown" lines, so the file is refused and kept the
// way an unknown effect key is (ADR 0041 P10): a venture marker nobody
// can advance is a table that cannot move.
func unknownDungeonKeys(seats []playerSnapshot) []string {
	var out []string
	for i := range seats {
		d := seats[i].Dungeon
		if d == nil {
			continue
		}
		def := DungeonByKey(d.Key)
		if def == nil || def.Room(d.Room) == nil {
			out = append(out, "dungeon "+strings.TrimSpace(d.Key+" "+d.Room))
		}
	}
	return out
}
