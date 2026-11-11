package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Connive // Concoct — split card, oracle 6e07e7ae-1bb7-4831-adb9-a5e71b7d79e3:
//
//	Connive {2}{U/B}{U/B} Sorcery: "Gain control of target creature with
//	 power 2 or less."
//	Concoct {3}{U}{B} Sorcery: "Surveil 3, then return a creature card from
//	 your graveyard to the battlefield."
//
// Two specs, Concoct under "<oracle_id>#1" as Dusk // Dawn is.
//
// Connive is Entrancing Melody's control change: no stated duration, so it
// lasts until the game ends (CR 611.2a); "power 2 or less" is judged at
// announce and again as the spell resolves (CR 608.2b).
//
// Concoct's pick is not targeted and is made after the surveil has put its
// cards into the graveyard (Surveil.ThenTail), so a creature card just
// surveilled away is on offer. It is mandatory (Min 1, Max 1) and the
// creature enters untapped under the caster's control.
//
// No simplification.
func init() {
	const oracle = "6e07e7ae-1bb7-4831-adb9-a5e71b7d79e3"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Connive",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power 2 or less", PowerLE(2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Connive — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Concoct",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 3, ThenTail: ctx.ItemTail(cmdr2S39ConcoctReturn)}.Apply(ctx)
		},
	})
}

// cmdr2S39ConcoctReturn is "then return a creature card from your graveyard
// to the battlefield". An on-disk key: never renamed, never reused.
var cmdr2S39ConcoctReturn = CardTail("cmdr2-s39/concoct-return", func(ctx *Context, _ game.TailOutcome) error {
	return ReturnChosenFromGraveyard{
		Question: "Concoct — return a creature card from your graveyard to the battlefield",
		Match:    game.Card.IsCreature,
		Min:      1,
		Max:      1,
	}.Apply(ctx)
})
