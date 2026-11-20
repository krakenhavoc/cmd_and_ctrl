package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	exile := ChapterTriggerTargeting(1, "Ajani Fells the Godsire — exile target creature an opponent controls with power 3 or greater",
		TargetCreature("target creature an opponent controls with power 3 or greater", OpponentControls(), PowerGE(3)),
		ExileFirstTarget)
	cat := ChapterTrigger(2, "Ajani Fells the Godsire — create a 2/1 white Cat Warrior, then put a vigilance counter on a creature you control",
		Do(CreateToken{Template: gapB56CatWarriorToken(), N: 1},
			Instruction(func(ctx *Context) error {
				return ChoosePermanents{
					Question:   "Ajani Fells the Godsire — put a vigilance counter on a creature you control",
					Candidates: humiliateCreatures,
					Keyed:      ajaniFellsTheGodsireChosen,
				}.Apply(ctx)
			})))
	double := ChapterTriggerTargeting(3, "Ajani Fells the Godsire — target creature you control gains double strike until end of turn",
		TargetCreature("target creature you control", YouControl()),
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Ajani Fells the Godsire — double strike until end of turn"}.Apply(ctx)
		})
	Register(Spec{
		OracleID:     "0daefecb-7189-44bb-9ca0-96a9572cd240",
		Name:         "Ajani Fells the Godsire",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{exile, cat, double},
	})
}

// ajaniFellsTheGodsireChosen is the vigilance counter on the chosen
// creature, keyed. An on-disk key.
var ajaniFellsTheGodsireChosen = RegisterPermanentsPick("prompt-run/gap-b56-ajani-fells-the-godsire-counter", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (AddCounter{Target: id, Kind: game.CounterVigilance, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
