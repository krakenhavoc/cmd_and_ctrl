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
				Target: target,
				Then: func(ctx *Context, exiled bool) error {
					if !exiled {
						return nil
					}
					return ReturnFromExile{
						Target: target,
						Then:   essenceFluxCounterIfSpirit,
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}

// essenceFluxCounterIfSpirit is "if it's a Spirit, put a +1/+1
// counter on it" — the ReturnFromExile.Then contract, so it reads
// nothing but the scalar new-object ID it is handed (#1327): an undo
// across the return resolves this against the restored game.
func essenceFluxCounterIfSpirit(g *game.Game, entered uuid.UUID) error {
	if entered == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(entered)
	if !ok || !c.HasSubtype("Spirit") {
		return nil
	}
	return g.AddCounterForEffect(entered, "+1/+1", 1)
}
