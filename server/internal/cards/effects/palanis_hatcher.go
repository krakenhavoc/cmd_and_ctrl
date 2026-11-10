package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Palani's Hatcher — Creature — Dinosaur {3}{R}{G}, 3/3 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Other Dinosaurs you control have haste.
//	 When this creature enters, create two 0/1 green Dinosaur Egg creature
//	 tokens.
//	 At the beginning of combat on your turn, if you control one or more
//	 Eggs, sacrifice an Egg, then create a 3/3 green Dinosaur creature
//	 token."
//
// The haste line is a layer-6 tribal grant over the Dinosaurs the
// controller owns, the Hatcher excepted. The combat trigger's intervening
// "if" (CR 603.4) is read as it would trigger and again as it resolves, by
// object types (an Egg is a creature subtype, so a changeling counts). The
// controller picks which Egg to sacrifice, and the 3/3 follows the
// sacrifice as a keyed tail: it is made even when the Egg was already gone
// (the sacrifice has nothing to do, the "then" still happens).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce6b6869-e4d7-4156-b66d-57fc224e4b20",
		Name:         "Palani's Hatcher",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Palani's Hatcher — create two 0/1 green Dinosaur Egg creature tokens",
				Do(CreateToken{Template: TokenCard("0/1 green Dinosaur Egg"), N: 2})),
			InterveningIf(palanisHatcherControlsAnEgg,
				AtBeginningOfYourCombat("Palani's Hatcher — sacrifice an Egg, then create a 3/3 green Dinosaur creature token",
					palanisHatcherHatch)),
		},
	})
}

// palanisHatcherControlsAnEgg is "if you control one or more Eggs".
var palanisHatcherControlsAnEgg = CondBoard{
	Name: "you control one or more Eggs",
	Test: func(ctx *Context) bool { return len(palanisHatcherEggs(ctx)) > 0 },
}

func palanisHatcherEggs(ctx *Context) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(OfCreatureType("Egg"), YouControl())) {
		out = append(out, c.InstanceID)
	}
	return out
}

func palanisHatcherHatch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SacrificeChoice{
		Player:     ctx.Controller(),
		Candidates: palanisHatcherEggs(ctx),
		Question:   "Palani's Hatcher — sacrifice an Egg",
		ThenTail:   ctx.Tail(palanisHatcherThen),
	}.Apply(ctx)
}

// palanisHatcherThen creates the 3/3 Dinosaur after the sacrifice. An
// on-disk key: never renamed, never reused.
var palanisHatcherThen = CardTail("then/palanis-hatcher-dinosaur", func(ctx *Context, _ game.TailOutcome) error {
	return CreateToken{Template: TokenCard("3/3 green Dinosaur"), N: 1}.Apply(ctx)
})
