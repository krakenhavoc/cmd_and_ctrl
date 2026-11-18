package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallowed Respite — Sorcery {W}{U} (Innistrad: Midnight Hunt, #5497):
//
//	"Exile target nonlegendary creature, then return it to the
//	 battlefield under its owner's control. If it entered under your
//	 control, put a +1/+1 counter on it. Otherwise, tap it.
//	 Flashback {1}{W}{U}"
//
// A blink (CR 400.7: the card comes back as a new object) under its
// OWNER's control, so a creature an opponent stole from them goes home.
// Who controls it on return decides the rider: it is the owner, so a
// creature you own returns with a +1/+1 counter, and one you do not own
// returns tapped (Flicker's PlusOneCounters and Tapped, both applied as
// it enters, not afterwards). A token exiled this way is gone (CR 111.8).
// The flashback is the keyword helper's alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2d72454f-3d87-4acc-9f46-b8c579f44af8",
		Name:             "Hallowed Respite",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target nonlegendary creature", Not(Legendary())),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{W}{U}")},
		OnResolve:        hallowedRespiteResolve,
	})
}

func hallowedRespiteResolve(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	id := item.Targets[0].ID
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if c.Owner == ctx.Controller() {
		return Flicker{Target: id, PlusOneCounters: 1}.Apply(ctx)
	}
	return Flicker{Target: id, Tapped: true}.Apply(ctx)
}
