package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scuttling Sentinel — Creature — Crab Elf {1}{G/U}{G/U}, 3/2 (Ravnica: Clue
// Edition):
//
//	"Flash
//	 Vigilance
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control. Until end of turn, that creature becomes a blue
//	 Crab in addition to its other types and gains hexproof."
//
// Guru Pathik's "another target creature you control" counter, then
// Unctus's until-end-of-turn "becomes a blue ... in addition" (layers 4
// and 5) plus a hexproof grant. All three halves act on the same object
// and are skipped if it has left the battlefield or the target is no
// longer legal (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Scuttling Sentinel — a +1/+1 counter on another target creature you control; it becomes a blue Crab and gains hexproof until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id := effects.FirstLegalCardTarget(ctx)
			if id == [16]byte{} || !effects.OnBattlefield(g, id) {
				return nil
			}
			if err := g.AddCounterForEffect(id, game.CounterPlusOne, 1); err != nil {
				return err
			}
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddColorsMod(ctx.TextColors("U")...), game.AddSubtypesMod(ctx.TextSubtypes("Crab")...)},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Scuttling Sentinel — becomes a blue Crab until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"hexproof"},
				Label:    "Scuttling Sentinel — hexproof until end of turn",
			}.Apply(ctx)
		})
	enters.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "a9f1802b-f2a0-4902-acc2-e21c14960e4c",
		Name:            "Scuttling Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
