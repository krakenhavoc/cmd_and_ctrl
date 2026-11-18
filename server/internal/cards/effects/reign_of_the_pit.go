package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reign of the Pit — Sorcery {4}{B}{B}:
//
//	"Each player sacrifices a creature of their choice. Create an X/X
//	 black Demon creature token with flying, where X is the total power of
//	 the creatures sacrificed this way."
//
// Fleshbag Marauder's edict (every player, the caster included, chooses
// one creature, in APNAP order, and the chosen permanents are sacrificed
// together, CR 101.4 and 608.2e) with a keyed clause behind it that reads
// each sacrificed creature's power as it last existed on the battlefield
// (CR 608.2h and 603.10, b13LastKnownPower), summed over the creatures that
// really left. The token's size is set as it is created, so it is a plain
// template and not a table row; with nothing sacrificed it is a 0/0 that
// dies to the next state-based check, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e1a6337-bb29-450f-bf29-ff71b357a983",
		Name:         "Reign of the Pit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
				Keyed: reignOfThePitEdicted,
			}.Apply(ctx)
		},
	})
}

// reignOfThePitEdicted creates the Demon, sized by the total power of the
// creatures sacrificed this way. An on-disk key.
var reignOfThePitEdicted = RegisterEdictThen("prompt-run/reign-of-the-pit-demon", func(ctx *Context, sacrificed game.PromptedSacrifices) error {
	total := 0
	for _, id := range sacrificed.Cards() {
		total += b13LastKnownPower(ctx.Game, id, game.Characteristic{})
	}
	return CreateToken{
		Controller: ctx.Controller(),
		Template: game.Card{
			Name:           "Demon",
			TypeLine:       "Token Creature — Demon",
			Colors:         []string{"B"},
			Keywords:       []string{"flying"},
			Power:          total,
			Toughness:      total,
			PrintedPTKnown: true,
		},
		N: 1,
	}.Apply(ctx)
})
