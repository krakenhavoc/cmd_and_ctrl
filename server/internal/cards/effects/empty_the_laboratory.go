package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Empty the Laboratory — Sorcery {X}{U}{U} (Midnight Hunt Commander):
//
//	"Sacrifice X Zombies, then reveal cards from the top of your library
//	 until you reveal a number of Zombie creature cards equal to the
//	 number of Zombies sacrificed this way. Put those cards onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order."
//
// The sacrifice is on resolution and the caster chooses the Zombies
// (ChoosePermanents, not a target): with fewer than X Zombies they all
// go, and the reveal counts the ones that really left. The reveal is
// revealUntilCount (a token is revealed but never counted, CR 108.2); a
// library with too few Zombie creature cards reveals itself whole, puts
// the Zombies it found onto the battlefield and bottoms the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84f3736f-5b2a-40f9-8c3b-c18f8c637c22",
		Name:         "Empty the Laboratory",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ChoosePermanents{
				Question: "Empty the Laboratory — sacrifice X Zombies",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					match := And(Creature(), HasSubtype("Zombie"))
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && match(g, of, c) {
							out = append(out, c.InstanceID)
						}
					}
					n := min(x, len(out))
					return out, n, n
				},
				Keyed: emptyTheLaboratoryChosen,
			}.Apply(ctx)
		},
	})
}

var emptyTheLaboratoryChosen = RegisterPermanentsPick("prompt-run/empty-the-laboratory-sacrifice", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		if picked.Count() == 0 {
			return nil
		}
		return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), picked.Cards(), ctx.ItemTail(emptyTheLaboratoryReveal))
	},
})

// emptyTheLaboratoryReveal is the reveal, run once the Zombies have gone;
// its count is how many really left. An on-disk key.
var emptyTheLaboratoryReveal = CardTail("then/empty-the-laboratory-reveal", func(ctx *Context, o game.TailOutcome) error {
	n := len(o.IDs)
	if n == 0 {
		return nil
	}
	player := ctx.Controller()
	run, hits := revealUntilCount(ctx, player, tombTyrantZombieCard, n, "Empty the Laboratory — revealed")
	got := make(map[uuid.UUID]bool, len(hits))
	for _, id := range hits {
		got[id] = true
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return got[c.InstanceID] },
		All:      true,
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
})
