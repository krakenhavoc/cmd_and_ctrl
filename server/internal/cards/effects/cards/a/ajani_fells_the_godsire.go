package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani Fells the Godsire — Enchantment — Saga {3}{W}{W} (Modern Horizons
// 3):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Exile target creature an opponent controls with power 3 or
//	 greater.
//	 II — Create a 2/1 white Cat Warrior creature token, then put a
//	 vigilance counter on a creature you control.
//	 III — Target creature you control gains double strike until end of
//	 turn."
//
// I is a targeted chapter (the clause is checked when it goes on the stack
// and again as it resolves, CR 603.3d, 608.2b: a creature whose power fell
// below 3 is no longer a legal target). II makes the Cat Warrior first, so
// the new token is a candidate for the vigilance counter, then asks for the
// creature as it resolves (Choking Miasma's choice, not a target); the
// counter is the engine's keyword counter (game.CounterVigilance), which
// gives the creature vigilance while it is there. III is a targeted
// until-end-of-turn keyword grant.
//
// No simplification.
func init() {
	exile := effects.ChapterTriggerTargeting(1, "Ajani Fells the Godsire — exile target creature an opponent controls with power 3 or greater",
		effects.TargetCreature("target creature an opponent controls with power 3 or greater", effects.OpponentControls(), effects.PowerGE(3)),
		effects.ExileFirstTarget)
	cat := effects.ChapterTrigger(2, "Ajani Fells the Godsire — create a 2/1 white Cat Warrior, then put a vigilance counter on a creature you control",
		effects.Do(effects.CreateToken{Template: effects.GapB56CatWarriorToken(), N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChoosePermanents{
					Question:   "Ajani Fells the Godsire — put a vigilance counter on a creature you control",
					Candidates: effects.HumiliateCreatures,
					Keyed:      ajaniFellsTheGodsireChosen,
				}.Apply(ctx)
			})))
	double := effects.ChapterTriggerTargeting(3, "Ajani Fells the Godsire — target creature you control gains double strike until end of turn",
		effects.TargetCreature("target creature you control", effects.YouControl()),
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Ajani Fells the Godsire — double strike until end of turn"}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "0daefecb-7189-44bb-9ca0-96a9572cd240",
		Name:         "Ajani Fells the Godsire",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{exile, cat, double},
	})
}

// ajaniFellsTheGodsireChosen is the vigilance counter on the chosen
// creature, keyed. An on-disk key.
var ajaniFellsTheGodsireChosen = effects.RegisterPermanentsPick("prompt-run/gap-b56-ajani-fells-the-godsire-counter", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterVigilance, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
