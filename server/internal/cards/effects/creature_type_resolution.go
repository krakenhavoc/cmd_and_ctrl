package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// creature_type_resolution.go — "choose a creature type" asked while a
// spell or ability RESOLVES (#2382), the sibling of
// ChooseCreatureTypeAsEnters, which asks as a permanent enters and
// stores the answer on it.
//
// Same prompt, same vocabulary (game.AllCreatureTypes), same
// `{creature_type}` answer, same enumerator and client modal. The
// difference is where the answer goes: into a continuation, stored
// nowhere, so the type outlives nothing but the resolution that asked.
// Append-only (the clone gate's rule).

// ChooseCreatureTypeTail queues a resolution-time "Choose a creature
// type." for `chooser` and hands the canonical type to `then`, a
// registered tail (CardTail or creatureTypeChosen, built with ctx.Tail or
// ctx.ItemTail) told the type as its Word. The rest of the effect is data
// (#4253: the closure form, ChooseCreatureTypeThen, is gone), so a table
// waiting on the answer is a restore point; put what the body reads on the
// tail (Carry, N, Words, the asking item).
//
// The tail is told "" when nobody chose: the chooser left the game with
// the prompt open (CR 800.4a). A clause that acts on the type should do
// nothing for "" (creatureTypeChosen does that), and a chain that asks
// several seats should carry on.
//
// Returns false when nothing was queued (the chooser is out of the
// game), in which case the tail is never run and a chain must carry on
// itself.
func ChooseCreatureTypeTail(g *game.Game, chooser, source uuid.UUID, question string, then game.KeyedTail) bool {
	return g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		ThenTail: then,
	}) != uuid.Nil
}

// creatureTypeChosen registers the tail of a creature-type prompt whose
// clause acts on the type: `body` is handed the type (the tail's Word) and
// is not run when nobody chose. It runs against the Context of CardTail.
// The key is an on-disk identity (ledgered "tail <key>"): never renamed,
// never reused.
func creatureTypeChosen(key string, body func(ctx *Context, creatureType string, o game.TailOutcome) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if o.Tail.Word == "" {
			return nil
		}
		return body(ctx, o.Tail.Word, o)
	})
}

// permanentsOfTypeYouControl counts `player`'s permanents of creature
// type `t` — "for each permanent you control of that type". Reads
// through Card.HasSubtype, so a changeling counts for every type
// (CR 702.73a) and a layer-4 grant counts too.
func permanentsOfTypeYouControl(g *game.Game, player uuid.UUID, t string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(t) {
			n++
		}
	}
	return n
}

// creaturesOfTypeYouControl is the creature-only form ("for each
// creature you control of the chosen type"), as instance IDs in
// battlefield order.
func creaturesOfTypeYouControl(g *game.Game, player uuid.UUID, t string) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.HasSubtype(t) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// notOfCreatureType is "creatures that aren't of the chosen type": a
// creature whose types do not include `t`. A changeling is of every
// type, so it is never matched.
func notOfCreatureType(t string) CardPredicate {
	return And(Creature(), Not(OfCreatureType(t)))
}

// seatsFromController lists the living seats in turn order starting
// with `first` — the order "each player chooses" is asked in (CR 101.4:
// active player first, then turn order; the caster stands in for the
// active player, which is exact for a sorcery).
func seatsFromController(g *game.Game, first uuid.UUID) []uuid.UUID {
	start := 0
	for i, p := range g.Seats {
		if p != nil && p.ID == first {
			start = i
		}
	}
	var seats []uuid.UUID
	n := len(g.Seats)
	for k := 0; k < n; k++ {
		p := g.Seats[(start+k)%n]
		if p == nil || p.Eliminated {
			continue
		}
		seats = append(seats, p.ID)
	}
	return seats
}
