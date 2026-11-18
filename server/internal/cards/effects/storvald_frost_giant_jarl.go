package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storvald, Frost Giant Jarl — Legendary Creature — Giant {4}{G}{W}{U}, 7/7
// (Adventures in the Forgotten Realms Commander):
//
//	"Ward {3}
//	 Other creatures you control have ward {3}.
//	 Whenever Storvald enters or attacks, choose one or both —
//	 • Target creature has base power and toughness 7/7 until end of turn.
//	 • Target creature has base power and toughness 1/1 until end of turn."
//
// The printed ward is Ward and the grant is WardGranted over the other
// creatures its controller controls. The trigger is one modal ability on
// enters or attacks, "one or both"; each bullet has its own target (CR 115.3)
// and sets base power and toughness in layer 7b until end of turn, run in
// printed order, so the later 1/1 wins when both name one creature
// (CR 613.7).
//
// No simplification.
func init() {
	setBase := func(label string, p, t int) game.ModeOption {
		return ModeDoing(label, TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				tgt, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return untilEndOfTurn(ctx, tgt.ID, nil, "Storvald, Frost Giant Jarl — "+label, game.SetBasePTMods(p, t)...)
			})
	}
	trig := WhenThisEntersOrAttacks("Storvald, Frost Giant Jarl — choose one or both", func(*game.Game, *game.StackItem) error { return nil })
	trig.Modes = ChooseOneOrMore(
		setBase("Target creature has base power and toughness 7/7 until end of turn.", 7, 7),
		setBase("Target creature has base power and toughness 1/1 until end of turn.", 1, 1),
	)
	Register(Spec{
		OracleID:     "5d7802a7-df6e-47fb-9065-c7461ea24a19",
		Name:         "Storvald, Frost Giant Jarl",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Storvald, Frost Giant Jarl — ward {3}"),
			WardGranted(WardMana("{3}"), "Storvald, Frost Giant Jarl — other creatures you control have ward {3}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.IsCreature()
				}),
			trig,
		},
	})
}
