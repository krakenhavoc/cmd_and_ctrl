package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// venture.go — the two venture prompts (CR 701.49, ADR 0203 and its
// amendment of 2026-10-19): "venture: choose the next room", an
// option_pick whose options each name the room they enter
// (PickOptionView.Room), and "venture: choose a dungeon", whose options
// name a dungeon and no room (PickOptionView.Dungeon). The enumerator
// offers every arrow and every dungeon; this file says which is worth
// more.
//
// A room is priced by what its ability will do for this seat right now,
// on the board it can see:
//
//   - A room that puts something on TARGET creature is good only with
//     the right creature to take it: Forge and Storeroom (+1/+1
//     counters) want one of the bot's own, Arena (goad), Fungi Cavern
//     (-4/-0) and Twisted Caverns (can't attack) want an opponent's. The
//     target is required, so with only the wrong side's creatures on the
//     board the room helps the enemy.
//   - Trap! (target player loses 5 life) is the strongest room on the
//     Undercity's left path; Dark Pool drains each opponent.
//   - Runestone Caverns (two cards to play) and Muiral's Graveyard (two
//     bodies) are the Mad Mage's good branches; Goblin Bazaar and Goblin
//     Lair are steady value.
//   - Tomb of Annihilation's Oubliette is a shortcut paid for with a
//     card, a creature, an artifact and a land: priced by what the bot
//     would actually lose. Veils of Fear is the slow, cheap path.
//
// A dungeon is priced by how quickly it pays and what it costs: Lost
// Mine of Phandelver is four steady rooms; Tomb of Annihilation is the
// fastest to complete and every room hurts every player alike, which is
// worth it only with life to spare; Dungeon of the Mad Mage is seven
// rooms long and strongest at the bottom.
//
// An option pick with neither a room nor a dungeon on it is not this
// file's.

// ventureRoomValue scores answering a venture prompt `ch` with option
// `index`. The third return is false for an option pick that is not a
// venture choice.
func (st *state) ventureRoomValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || index == nil || *index < 0 || *index >= len(ch.PickOptions) {
		return 0, "", false
	}
	opt := ch.PickOptions[*index]
	if opt.Room == "" {
		if opt.Dungeon == "" {
			return 0, "", false
		}
		return st.ventureDungeonValue(opt.Dungeon), "venture: " + opt.Label, true
	}
	mine, theirs := st.creatureCounts()
	// ownTarget / theirTarget price a room whose required target should
	// be the bot's own creature, or an opponent's.
	ownTarget := func(v float64) float64 {
		if mine > 0 {
			return v
		}
		return -1
	}
	theirTarget := func(v float64) float64 {
		if theirs > 0 {
			return v
		}
		return -1
	}
	v := 1.0
	switch opt.Dungeon + "/" + opt.Room {
	// The Undercity.
	case "undercity/forge":
		v = ownTarget(3.5)
	case "undercity/trap":
		v = 4
	case "undercity/arena":
		v = theirTarget(2.5)
	case "undercity/catacombs":
		v = 3.25
	case "undercity/archives":
		v = 3
	case "undercity/stash":
		v = 2.25
	case "undercity/lost-well":
		v = 2
	// Lost Mine of Phandelver.
	case "lost-mine-of-phandelver/goblin-lair":
		v = 2.5
	case "lost-mine-of-phandelver/mine-tunnels":
		v = 2.25
	case "lost-mine-of-phandelver/storeroom":
		v = ownTarget(2.5)
	case "lost-mine-of-phandelver/dark-pool":
		v = 3
	case "lost-mine-of-phandelver/fungi-cavern":
		v = theirTarget(2.75)
	// Dungeon of the Mad Mage.
	case "dungeon-of-the-mad-mage/goblin-bazaar":
		v = 2.25
	case "dungeon-of-the-mad-mage/twisted-caverns":
		v = theirTarget(2)
	case "dungeon-of-the-mad-mage/runestone-caverns":
		v = 3
	case "dungeon-of-the-mad-mage/muirals-graveyard":
		v = 2.75
	// Tomb of Annihilation.
	case "tomb-of-annihilation/veils-of-fear":
		v = 2.5
	case "tomb-of-annihilation/oubliette":
		v = st.oublietteValue()
	}
	return v, "venture: " + opt.Label, true
}

// ventureDungeonValue prices entering the dungeon with this key.
func (st *state) ventureDungeonValue(key string) float64 {
	switch key {
	case "lost-mine-of-phandelver":
		return 3
	case "tomb-of-annihilation":
		if st.myLife() > 15 {
			return 3.5
		}
		return 1
	case "dungeon-of-the-mad-mage":
		return 2.5
	}
	return 1
}

// oublietteValue prices Oubliette: the fastest way through the Tomb,
// paid with a discard and a creature, an artifact and a land of the
// bot's own — each worth more the more of them it would lose.
func (st *state) oublietteValue() float64 {
	v := 4.0
	if len(st.seatHand()) > 0 {
		v -= 0.75
	}
	var creature, artifact, land bool
	for _, c := range st.bf {
		if c == nil || c.Controller != st.me {
			continue
		}
		creature = creature || isCreature(c)
		artifact = artifact || isArtifact(c)
		land = land || isLand(c)
	}
	if creature {
		v -= 1.5
	}
	if artifact {
		v -= 1
	}
	if land {
		v -= 1
	}
	return v
}

// creatureCounts counts the creatures on the battlefield this seat
// controls and those its opponents control.
func (st *state) creatureCounts() (mine, theirs int) {
	for _, c := range st.bf {
		if c == nil || !isCreature(c) {
			continue
		}
		if c.Controller == st.me {
			mine++
		} else {
			theirs++
		}
	}
	return mine, theirs
}
