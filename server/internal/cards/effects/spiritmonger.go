package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiritmonger — Creature — Beast {3}{B}{G}, 6/6 (Apocalypse, #5614):
//
//	"Whenever this creature deals damage to a creature, put a +1/+1
//	 counter on this creature.
//	 {B}: Regenerate this creature.
//	 {G}: This creature becomes the color of your choice until end of
//	 turn."
//
// The trigger is one per damage event to a creature, combat or not
// (thisDealtDamageToACreature); two blockers dealt damage at once are two
// triggers and two counters. The counter lands only on the Spiritmonger
// the trigger came from (CounterOnThis: one that left, or came back as a
// new object, gets nothing).
//
// {B} is the shared regenerate row (a shield, CR 701.19a). {G} asks for
// the colour as it resolves (ChooseColorTail), then sets it until end of
// turn with Mondo Gecko's colour-setting mod: the Spiritmonger is that
// colour and no other (CR 105.3), and returns to black and green in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3913c7f2-86b7-4dd4-84a6-9bd4c1c66781",
		Name:         "Spiritmonger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, thisDealtDamageToACreature,
				"Spiritmonger — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{
			regenerateThisRow("{B}"),
			{
				Label:   "{G}: This creature becomes the color of your choice until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    ManaCost("{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
						"Spiritmonger — choose a color", NewContext(g, item).ItemTail(spiritmongerColorChosen))
					return nil
				},
			},
		},
	})
}

// spiritmongerColorChosen makes the Spiritmonger the chosen colour until
// end of turn, if it is still the permanent that asked. An on-disk key.
var spiritmongerColorChosen = CardTail("spiritmonger/color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return ScopedEffectFor{
		Target:   ctx.Item.SourceCardID,
		Mods:     []game.Mod{game.SetColorsMod(color)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Spiritmonger — " + game.ColorName(color),
	}.Apply(ctx)
})
