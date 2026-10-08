package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multiform Wonder — Artifact Creature — Construct {5}, 3/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}: This creature gains your choice of flying, vigilance, or
//	 lifelink until end of turn.
//	 Pay {E}: This creature gets +2/-2 or -2/+2 until end of turn."
//
// ADR 0129 §2: both rows pay one energy (PayEnergy). The two choices
// are made as each ability RESOLVES, not when it is activated, so they
// are asked then (PickOption for the keyword, MayChoice for the two
// pumps), the way Lunar Avenger and Endling ask theirs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9849c743-9492-4a1b-83bf-1be54614b85f",
		Name:         "Multiform Wonder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Multiform Wonder", 3),
		},
		Activated: []ActivatedAbility{
			{
				Label: "Pay {E}: This creature gains your choice of flying, vigilance, or lifelink until end of turn.",
				Cost:  PayEnergy(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					keywords := []string{"flying", "vigilance", "lifelink"}
					return PickOption{
						Question: "Multiform Wonder gains which until end of turn?",
						Options: []game.ChoiceOption{
							{Label: "Flying"}, {Label: "Vigilance"}, {Label: "Lifelink"},
						},
						Then: func(ctx *Context, index int) error {
							if index < 0 || index >= len(keywords) {
								return nil
							}
							return thisGainsKeywordUntilEndOfTurn(keywords[index], "Multiform Wonder — "+keywords[index]+" until end of turn")(ctx.Game, ctx.Item)
						},
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "Pay {E}: This creature gets +2/-2 or -2/+2 until end of turn.",
				Cost:  PayEnergy(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Multiform Wonder — +2/-2 or -2/+2 until end of turn?",
						YesLabel: "+2/-2",
						NoLabel:  "-2/+2",
						OnYes: func(ctx *Context) error {
							return thisGetsUntilEndOfTurn(2, -2, "Multiform Wonder — +2/-2 until end of turn")(ctx.Game, ctx.Item)
						},
						OnNo: func(ctx *Context) error {
							return thisGetsUntilEndOfTurn(-2, 2, "Multiform Wonder — -2/+2 until end of turn")(ctx.Game, ctx.Item)
						},
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
