package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Halo Forager — Creature — Faerie Rogue {1}{U}{B}, 3/1:
//
//	"Flying
//	 When this creature enters, you may pay {X}. When you do, you may cast
//	 target instant or sorcery card with mana value X from a graveyard
//	 without paying its mana cost. If that spell would be put into a
//	 graveyard, exile it instead."
//
// X is chosen as the enters trigger resolves (MayPayX) and paid through the
// ordinary "you may pay" prompt; paying creates a CR 603.12 reflexive
// trigger that goes on the stack and picks its target then (CR 603.3d) among
// the instant and sorcery cards of every graveyard with mana value exactly
// X. Its resolution offers the free cast (CR 608.2g), exiling the spell
// instead of letting it reach a graveyard.
func init() {
	Register(Spec{
		OracleID:        "210d1077-6e60-4f06-a8a4-10b842979ac5",
		Name:            "Halo Forager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Halo Forager — pay {X}; cast an instant or sorcery card with mana value X from a graveyard free", haloForagerEnters),
		},
	})
}

// haloForagerPaid is the paid half, keyed (#4289): the X paid rides the
// prompt.
var haloForagerPaid = PayUnlessBranches("pay-unless/halo-forager-x-cast",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label:  "Halo Forager — cast an instant or sorcery card with mana value " + strconv.Itoa(r.Number) + " from a graveyard",
			Body:   cmdr2S42HaloForagerCastBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func haloForagerEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Halo Forager",
		Buys:  "cast an instant or sorcery card with mana value X from a graveyard",
		Goal:  XAsHighAsYouCan,
		Then:  haloForagerPaid,
	}.Apply(NewContext(g, item))
}
