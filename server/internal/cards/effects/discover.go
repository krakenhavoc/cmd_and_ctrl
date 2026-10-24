package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discover.go — CR 701.57, ADR 0099: the card-side words for discover.
//
//	return Discover{N: 4}.Apply(ctx)                        // Daring Discovery
//	return Discover{Player: owner, N: mv}.Apply(ctx)        // Zoyowa's Justice
//	OnResolve: ResolveSteps(Discover{N: 10}, Instruction(treasures))   // Hit the Mother Lode
//	Triggered: []game.TriggeredAbility{WheneverYouDiscover(label, effect)} // Curator
//
// The engine half is game/discover.go — the walk it shares with
// cascade, the resolution_cast prompt that casts the card during the
// resolution (CR 608.2g, #4243), the free-cast grant with its timing and
// its mana-value cap, and the decline that puts the card into the hand.
//
// Discover is an INSTRUCTION, not a keyword ability. A card calls it
// from wherever its text says to — a spell's OnResolve, a trigger's or
// an activated ability's Effect — and works out N first. A number read
// off the board ("the greatest power among them") is read then, at
// resolution (CR 608.2h). A number that is a fact about the triggering
// event ("that spell's mana value") is captured on the item by the
// trigger's Build, as cascade captures its mana value.

// KeywordDiscover is the name cards/coverage uses for the mechanic.
const KeywordDiscover = "discover"

// DiscoverStepKey is the step record Discover files its result under
// (game.StepRecord: IDs[0] the discovered card, Ns[0] its mana value),
// for a later step of the same resolution to read with
// ctx.StepRecord(DiscoverStepKey). The discover's question is answered
// after Apply returns, so what a later instruction reads is data the
// engine parks the rest of the effect behind (step_walk.go), not a Go
// closure: Hit the Mother Lode's "if the discovered card's mana value is
// less than 10" is a later step of its ResolveSteps. An empty walk files
// an empty record. On-disk identity: never renamed.
const DiscoverStepKey = "discover"

// Discover is "discover N" (CR 701.57a).
type Discover struct {
	// Player discovers. Zero is the effect's controller; Zoyowa's
	// Justice names the target's owner instead.
	Player uuid.UUID
	// N is the number.
	N int
}

// Apply performs the discover.
func (d Discover) Apply(ctx *Context) error {
	player := d.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	r, err := ctx.Game.DiscoverForEffect(player, ctx.Source(), d.N)
	rec := game.StepRecord{}
	if r.Discovered != uuid.Nil {
		rec = game.StepRecord{IDs: []uuid.UUID{r.Discovered}, Ns: []int{r.ManaValue}}
	}
	ctx.RecordStepRecord(DiscoverStepKey, rec)
	return err
}

// DiscoverN is the whole effect of an ability whose text is "discover
// N" and nothing else — the Caves' sacrifice ability, Quintorius
// Kand's −3, Etali's Favor's ETB.
func DiscoverN(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return Discover{N: n}.Apply(NewContext(g, item))
	}
}

// YouDiscovered — the discover event is yours. For "whenever you
// discover" (Curator of Sun's Creation).
func YouDiscovered(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Actor == source.Controller
}

// WheneverYouDiscover is "whenever you discover" (CR 701.57b). The
// event's Amount is the N, which is what "discover again for the same
// value" reads (item.Trigger.Event.Amount).
func WheneverYouDiscover(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDiscover, YouDiscovered, label, effect)
}
