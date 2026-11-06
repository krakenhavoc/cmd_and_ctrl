package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b24_helpers.go — shared pieces of Avatar: The Last Airbender
// batches 24 and 25 (#4459, #5075). Append-only.

// avB24PutCountersOnEachThen puts n `kind` counters on every listed
// permanent still on the battlefield, one permanent at a time, each as one
// placement put by the resolving item's controller, and runs `then` once
// the last placement has settled (a CR 616 ordering prompt over any of them
// is a restore point: the rest of the list rides the tail, b13's shape with
// the kind and the count as data). Black Sun's Zenith.
func avB24PutCountersOnEachThen(ctx *Context, ids []uuid.UUID, kind string, n int, then game.KeyedTail) error {
	for len(ids) > 0 && !onBattlefield(ctx.Game, ids[0]) {
		ids = ids[1:]
	}
	if len(ids) == 0 || n <= 0 {
		return then.Run(ctx.Game, nil)
	}
	next := ctx.ItemTail(avB24PutCountersOnEachNext)
	next.Carry = append([]uuid.UUID(nil), ids[1:]...)
	next.Word = kind
	next.N = n
	if !then.IsZero() {
		next.Next = &then
	}
	return ctx.Game.AddCounterByKeyedForEffect(ctx.Controller(), ids[0], kind, n, next)
}

// avB24PutCountersOnEachNext is the rest of an avB24PutCountersOnEachThen
// list behind a settled placement: Carry the permanents still owed, Word
// the counter kind, N the count, Next the caller's tail. An on-disk key:
// never renamed, never reused.
var avB24PutCountersOnEachNext game.TailRef

func init() {
	avB24PutCountersOnEachNext = CardTail("avatar-b24/put-n-counters-on-each-next", func(ctx *Context, o game.TailOutcome) error {
		var then game.KeyedTail
		if o.Tail.Next != nil {
			then = *o.Tail.Next
		}
		return avB24PutCountersOnEachThen(ctx, o.Tail.Carry, o.Tail.Word, o.Tail.N, then)
	})
}

// anotherFlierEnteredUnderYourControl is "whenever another creature you
// control with flying enters": another creature entered under the source
// controller's control and has flying as it is on the battlefield, read
// through the layered view. Watcher of the Spheres and Momo, Friendly
// Flier.
func anotherFlierEnteredUnderYourControl(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.HasKeyword(&c, "flying")
}

// avB24FirstSpellWithoutSubtypeOnYourTurn is Momo, Friendly Flier's "the
// first non-<subtype> creature spell with flying you cast during each of
// your turns": the caster is the active player, the spell lacks the
// subtype (a word of the source's text, CR 612.2) and no creature spell
// with flying lacking it has been cast yet this turn (the tally is bumped
// after a spell is priced, so none so far means this is the first). A
// first such spell spends the discount whether or not it had generic mana
// to take it off (the ruling).
func avB24FirstSpellWithoutSubtypeOnYourTurn(subtype string) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil || !IsYourTurn(q.Game, q.Controller) {
			return false
		}
		sub := q.Source.TextSubtype(subtype)
		return !q.Card.HasSubtype(sub) && !q.Game.CastTallyFor(q.Controller).CastFlyingCreatureSpellWithout(sub)
	}
}

// avB24ZeppelinToken is Lita, Mechanical Engineer's Zeppelin: a 5/5
// colorless Vehicle artifact token with flying and crew 3.
func avB24ZeppelinToken() game.Card { return tokenFromCatalog(printedAvB24ZeppelinToken) }

// printedAvB24ZeppelinToken is the Zeppelin as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func printedAvB24ZeppelinToken() tokenTemplate {
	return tokenTemplate{
		Slug: "zeppelin",
		Card: game.Card{
			Name:      "Zeppelin",
			TypeLine:  "Token Artifact — Vehicle",
			Power:     5,
			Toughness: 5,
			Keywords:  []string{"flying"},
		},
		Activated: []game.ActivatedAbilityShape{crewTokenAbility("Zeppelin", 3)},
		Text:      "Flying\nCrew 3",
	}
}
