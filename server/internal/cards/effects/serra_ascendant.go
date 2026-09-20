package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serra Ascendant — Creature — Human Monk {W}, 1/1:
//
//	"Lifelink."
//	"As long as you have 30 or more life, this creature gets +5/+5
//	 and has flying."
//
// A one-mana 6/6 flier with lifelink, which is why it is banned in
// every constructed format that ever allowed it and why it is a
// four-of in a lifegain Commander deck: in a format that starts at
// 40, the clause is ON from the opening draw and stays on until the
// table takes eleven life off its controller.
//
// Both halves are ordinary layer work — a 7c modification and a
// layer-6 keyword grant, each gated on the same condition, rather
// than one static doing two layers' jobs. Splitting them is what
// makes an anthem, a counter and a "loses all abilities" all compose
// correctly with it: the +5/+5 survives a Darksteel Mutation's
// ability wipe (layer 6 is before layer 7) and the flying does not,
// which is exactly right.
//
// # The life condition is an invalidation problem, not a rules one
//
// Writing "as long as you have 30 or more life" is one comparison.
// Making the answer CHANGE when the life total does is the part that
// needed engine work (#1117): the layer engine caches its resolution
// and bumped the version on battlefield moves, tokens, counters,
// control, attachment and tap state — a life total was on none of
// those lists. Without the fix this card is a 6/6 flier that stays a
// 6/6 flier through the Lightning Bolt that should have shrunk it,
// until some unrelated permanent happens to move.
//
// StaticAbility.DependsOnLifeTotal is how these two statics say they
// are reading one. It is Psychosis Crawler's DependsOnHandSize
// exactly — an invalidation hint, opt-in, so a table holding no card
// that reads a life total pays nothing for the many life changes a
// game of Commander makes. See invalidateLayersForLifeChangeLocked
// for why the bump rides the WRITE to the life total rather than an
// event: damage to a player changes a life total without emitting
// EventChangeLife at all, and emits its own event on opposite sides
// of the write on the combat and non-combat routes.
func init() {
	whileAtThirty := SelfWhileYourLifeAtLeast(30)
	Register(Spec{
		OracleID:        "27ad3e00-6ffb-48f7-8469-8868d066d1e2",
		Name:            "Serra Ascendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static: []game.StaticAbility{
			LifeGatedPump(whileAtThirty, 5, 5),
			LifeGatedKeyword(whileAtThirty, "flying"),
		},
	})
}
