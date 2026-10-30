package effects

// Genesis Ultimatum — Sorcery {G}{G}{U}{U}{U}{R}{R} (Marvel Super Heroes
// Commander, #4455):
//
//	"Look at the top five cards of your library. Put any number of
//	 permanent cards from among them onto the battlefield and the rest
//	 into your hand. Exile Genesis Ultimatum."
//
// Majestic Genesis's put (any number, a prompt over the permanent cards)
// with the rest going to hand, Risen Reef's rider through the shared
// putRestIntoHand. The exile is its own step: it waits for the put's
// prompt and its rider (CR 608.2c), and the spell then goes to exile, not
// the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06f42f49-32f0-413b-8437-b8efa4d647c2",
		Name:         "Genesis Ultimatum",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				looked := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Item.Controller, 5)
				return PutFromLibraryOntoBattlefield{
					Player:   ctx.Item.Controller,
					Cards:    looked,
					Max:      0,
					Optional: true,
					Label:    "Genesis Ultimatum — put any number of permanent cards onto the battlefield (the rest go into your hand)",
					ThenTail: putRestIntoHandTail,
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// putRestIntoHandTail is "the rest into your hand" as a keyed tail. An
// on-disk key.
var putRestIntoHandTail = PutFromLibraryTail("put-from-library/rest-into-hand", putRestIntoHand)
