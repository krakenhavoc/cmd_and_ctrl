package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Flux — Instant {U}:
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control. If it's a Spirit, put a
//	 +1/+1 counter on it."
//
// "Under its OWNER'S control" — Flicker's Controller left zero — is
// the opposite half of the Cloudshift pair: this one hands a stolen
// creature back rather than keeping it.
//
// The Spirit check reads the NEW object's subtypes after it lands
// (CR 400.7 — it is a fresh permanent, but the same printed card, so
// "if it's a Spirit" is answered off Essence Flux's own creature
// type line, not off anything the exile could have changed). A
// non-Spirit blinked for value gets nothing extra, which is the
// printed floor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64824ae5-efab-4b55-9d3c-b9c690bad857",
		Name:         "Essence Flux",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			return ExileTarget{
				Target:   target,
				ThenTail: withCarry(ctx.ItemTail(essenceFluxExiled), target),
			}.Apply(ctx)
		},
	})
}

// essenceFluxCounter is "if it's a Spirit, put a +1/+1 counter on it" as
// a keyed tail (#4253 slice 6). "Spirit" is a word of the spell's text,
// read as the tail runs (CR 612.2, #3984), which the asking item that
// rides the tail gives back. An on-disk key.
var essenceFluxCounter = EnteredTail("essence-flux/counter-if-spirit", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	g := ctx.Game
	c, ok := g.LookupCardForEffect(entered)
	if !ok || !c.HasSubtype(ctx.TextSubtype("Spirit")) {
		return nil
	}
	return g.AddCounterForEffect(entered, "+1/+1", 1)
})

// essenceFluxExiled is the clause after the exile (#4253 slice 9).
var essenceFluxExiled = CardTail("essence-flux/after-exile", func(ctx *Context, o game.TailOutcome) error {
	target := tailCarry0(o)
	exiled := len(o.IDs) == 1

	if !exiled {
		return nil
	}
	return ReturnFromExile{
		Target:   target,
		ThenTail: ctx.ItemTail(essenceFluxCounter),
	}.Apply(ctx)
})
