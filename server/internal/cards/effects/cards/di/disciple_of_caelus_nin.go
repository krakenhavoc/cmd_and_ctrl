package cardsdi

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciple of Caelus Nin — Creature — Human Wizard {4}{W}, 3/4 (The
// Brothers' War):
//
//	"When this creature enters, starting with you, each player chooses up
//	 to five permanents they control. All permanents other than this
//	 creature that weren't chosen this way phase out.
//	 Permanents can't phase in."
//
// The enters trigger asks each player in turn order from its controller
// for up to five of their own permanents (Foreboding Steamboat's chained
// picks, a keyed run so a table waiting on a seat's answer is a restore
// point), and only after the last answer phases out, simultaneously and in
// one call (CR 702.26a, with attachments following their hosts, CR
// 702.26g), every permanent on the battlefield that is neither this
// creature nor chosen: opponents' and your own, tokens and lands
// included. This creature may itself be chosen, which spends one of its
// controller's five. A Disciple that left before the trigger resolves
// still lets the choices happen and nothing else is spared. The second
// line is the PermanentsCantPhaseIn world rule (The City on the Edge of
// Forever's token): while this creature is on the battlefield nothing
// phases in, at an untap step or anywhere else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e5440ef-6887-4c79-98f5-49db82406d71",
		Name:         "Disciple of Caelus Nin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Disciple of Caelus Nin — each player chooses up to five permanents they control; all others phase out",
				discipleOfCaelusNinEnters),
		},
		Static: []game.StaticAbility{effects.RestrictSelf(game.PermanentsCantPhaseIn)},
	})
}

func discipleOfCaelusNinEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return discipleOfCaelusNinChoose(ctx, effects.SeatsStartingWith(g, ctx.Controller()), 0, nil)
}

// discipleOfCaelusNinChoose asks players[i] for up to five of their
// permanents, then the next player; after the last it phases out the rest.
func discipleOfCaelusNinChoose(ctx *effects.Context, players []uuid.UUID, i int, chosen []uuid.UUID) error {
	if i >= len(players) {
		var out []uuid.UUID
		for _, c := range ctx.Game.Battlefield.Cards {
			if c.InstanceID == ctx.Source() || slices.Contains(chosen, c.InstanceID) {
				continue
			}
			out = append(out, c.InstanceID)
		}
		return effects.PhaseOut{Targets: out}.Apply(ctx)
	}
	return effects.ChoosePermanents{
		Player:   players[i],
		Of:       []uuid.UUID{players[i]},
		Question: "Disciple of Caelus Nin — choose up to five permanents you control; the rest phase out",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var pool []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == of {
					pool = append(pool, c.InstanceID)
				}
			}
			return pool, 0, min(5, len(pool))
		},
		Keyed: discipleOfCaelusNinChosen,
		Carry: effects.PackIDLists(players, chosen),
		Nums:  []int{i},
	}.Apply(ctx)
}

var discipleOfCaelusNinChosen effects.PermanentsThen

func init() {
	discipleOfCaelusNinChosen = effects.RegisterPermanentsPick("prompt-run/disciple-of-caelus-nin-phase-out", effects.PermanentsDef{
		ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
			players, taken := effects.UnpackIDPair(r.Carry)
			i := 0
			if len(r.Nums) > 0 {
				i = r.Nums[0]
			}
			return discipleOfCaelusNinChoose(ctx, players, i+1, append(taken, picked.Cards()...))
		},
	})
}
