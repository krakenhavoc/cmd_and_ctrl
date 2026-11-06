package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chong and Lily, Nomads — Legendary Creature — Human Bard Ally {3}{R},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever one or more Bards you control attack, choose one —
//	 • Put a lore counter on each of any number of target Sagas you
//	   control.
//	 • Creatures you control get +1/+0 until end of turn for each lore
//	   counter among Sagas you control."
//
// One trigger per declaration that includes a Bard you control, read
// through the layered view (OncePerBatch), modal as it goes on the stack
// (CR 603.3c). A lore counter put on a Saga outside its usual timing
// triggers the matching chapter at once (CR 714.2b, the ruling): each
// Saga gets its counter as its own placement, one after another (b24's
// avB24PutCountersOnEachThen). The pump counts the lore counters on the
// Sagas you control as it resolves (CR 608.2h) and reads that once.
//
// No simplification.
func init() {
	trigger := OncePerBatch(On(game.EventAttack, chongAndLilyBardAttacked,
		"Chong and Lily, Nomads — one or more Bards you control attacked: choose one",
		func(*game.Game, *game.StackItem) error { return nil }))
	trigger.Modes = ChooseOne(
		ModeDoing("Put a lore counter on each of any number of target Sagas you control.",
			TargetPermanent("any number of target Sagas you control", OfSubtype("Saga"), YouControl()).WithCount(0, 0),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				var ids []uuid.UUID
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
						ids = append(ids, t.ID)
					}
				}
				return avB24PutCountersOnEachThen(ctx, ids, game.CounterLore, 1, game.KeyedTail{})
			}),
		ModeDoing("Creatures you control get +1/+0 until end of turn for each lore counter among Sagas you control.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.HasSubtype("Saga") {
						n += c.Counters[game.CounterLore]
					}
				}
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: n, Label: "Chong and Lily, Nomads — +1/+0 per lore counter"}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "faffc9fb-32b3-4015-8c3f-e6b45002c050",
		Name:         "Chong and Lily, Nomads",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// chongAndLilyBardAttacked is "one or more Bards you control attack": a
// declaration by the source's controller of a creature that is a Bard as
// it attacks.
func chongAndLilyBardAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Bard"))
}
