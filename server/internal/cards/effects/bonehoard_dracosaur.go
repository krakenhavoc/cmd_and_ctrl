package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonehoard Dracosaur — Creature — Dinosaur Dragon {3}{R}{R}, 5/5
// (EDHREC rank 1375):
//
//	"Flying, first strike
//	 At the beginning of your upkeep, exile the top two cards of your
//	 library. You may play them this turn. If you exiled a land card
//	 this way, create a 3/1 red Dinosaur creature token. If you exiled
//	 a nonland card this way, create a Treasure token."
//
// Two cards a turn plus a body or a Treasure for each kind, on a 5/5
// flying first striker. The exile is the impulse-exile primitive
// with a "play" grant (so the land can be the turn's drop), and the
// riders read the exiled cards where they now sit — the printed
// type line, which for "land card" / "nonland card" is exactly the
// question. Each rider fires at most once ("a land card", not "each
// land card"): two lands is one Dinosaur, one land and one spell is
// one of each, and an empty library exiles nothing and makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ba550a5-81fc-42c2-8df8-ad455d938605",
		Name:            "Bonehoard Dracosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bonehoard Dracosaur — exile the top two cards, play them this turn", func(g *game.Game, item *game.StackItem) error {
				// The riders read what was EXILED, so they wait for
				// the continuation form: a commander among the two
				// cards pauses the batch on its owner's CR 903.9
				// prompt, and the fire-and-forget form's return would
				// be short or empty right when this needs it most
				// (#1587).
				return b12ImpulseExileForTurnThen(g, item, 2, bonehoardExiled)
			}),
		},
	})
}

// bonehoardExiled reads what was exiled and makes the tokens, keyed (#4253
// slice 10a).
var bonehoardExiled = CardTail("then/bonehoard-dracosaur-exiled", func(ctx *Context, o game.TailOutcome) error {
	land, nonland := 0, 0
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if c.IsLand() {
			land = 1
		} else {
			nonland = 1
		}
	}
	// The Treasure waits for a prompt the Dinosaur asks (a CR 616 order
	// of two token replacements; #4162).
	ctx.RecordStep(bonehoardRecord, nil, land, nonland)
	return bonehoardTokens.Run(ctx)
})

// bonehoardRecord is whether a land and whether a nonland card was
// exiled, as 0 or 1.
const bonehoardRecord = "bonehoard-dracosaur/exiled"

// bonehoardTokens is "If you exiled a land card this way, create a 3/1
// red Dinosaur creature token. If you exiled a nonland card this way,
// create a Treasure token."
var bonehoardTokens = RegisterThenSteps("then-steps/bonehoard-dracosaur-tokens",
	Instruction(func(ctx *Context) error {
		if r := ctx.StepRecord(bonehoardRecord); len(r.Ns) < 1 || r.Ns[0] == 0 {
			return nil
		}
		return CreateToken{Template: TokenCard("3/1 red Dinosaur"), N: 1}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		if r := ctx.StepRecord(bonehoardRecord); len(r.Ns) < 2 || r.Ns[1] == 0 {
			return nil
		}
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}),
)
