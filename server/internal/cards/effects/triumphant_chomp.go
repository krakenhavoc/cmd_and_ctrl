package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triumphant Chomp — Sorcery {R} (The Lost Caverns of Ixalan, #4470):
//
//	"Triumphant Chomp deals damage to target creature equal to 2 or the
//	 greatest power among Dinosaurs you control, whichever is greater."
//
// Flame Slash's creature-only target and damage from the spell, with the
// amount read as the spell resolves (CR 608.2h): the greatest power among
// the creatures you control that are Dinosaurs then, after every layer
// (Card.HasSubtype reads the effective types, so a creature made a
// Dinosaur counts, and "Dinosaur" is read through this spell's text, CR
// 612.2), and 2 when that is less or you control none. The target is judged
// at announce (CR 601.2c) and again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d417f574-38a4-405c-83bb-49c5cf542648",
		Name:         "Triumphant Chomp",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: triumphantChompAmount(ctx),
			}.Apply(ctx)
		},
	})
}

// triumphantChompAmount is 2 or the greatest power among Dinosaurs the
// spell's controller controls, whichever is greater.
func triumphantChompAmount(ctx *Context) int {
	dinosaur := ctx.TextSubtype("Dinosaur")
	amount := 2
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.IsCreature() || !c.HasSubtype(dinosaur) {
			continue
		}
		amount = max(amount, c.CurrentPower())
	}
	return amount
}
