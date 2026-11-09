package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos, Patron of Chaos — Legendary Creature — Demon {4}{B}{R} (Murders at
// Karlov Manor, #4469):
//
//	"Flying, trample
//	 At the beginning of your end step, target opponent may sacrifice two
//	 nonland, nontoken permanents of their choice. If they don't, you draw
//	 two cards."
//
// The target opponent is the one asked (a yes/no, then the two permanents),
// and the permanents are their own to choose, not targets: protection and
// hexproof do not stop a player sacrificing their own permanents. An opponent
// with fewer than two eligible permanents (or ones that can't be sacrificed)
// cannot, so you draw without a prompt, and one who declines, or whose pick
// no longer holds as it is answered, has not sacrificed either (CR 608.2d).
// Both legs are keyed, so a table waiting on either is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d08ce18a-0fd9-47ac-ad2f-6934b60070f1",
		Name:            "Rakdos, Patron of Chaos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Rakdos — target opponent may sacrifice two nonland, nontoken permanents; if they don't, you draw two cards",
				rakdosAsk),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// rakdosCandidates is the opponent's own nonland, nontoken permanents that
// can be sacrificed, in battlefield order.
func rakdosCandidates(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.IsLand() && !c.IsToken() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func rakdosAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if len(rakdosCandidates(g, t.ID)) < 2 {
		return g.DrawNForEffect(ctx.Controller(), 2)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      t.ID,
		Source:       ctx.Source(),
		Question:     "Rakdos, Patron of Chaos — sacrifice two nonland, nontoken permanents? If you don't, its controller draws two cards",
		AcceptLabel:  "Sacrifice two permanents",
		DeclineLabel: "Don't",
	}, rakdosConfirm, []uuid.UUID{ctx.Controller()})
	return err
}

// rakdosConfirm is the opponent's yes/no: a decline lets Rakdos' controller
// draw, an accept asks for the two permanents.
var rakdosConfirm = game.RegisterConfirmThen("confirm/rakdos-patron-of-chaos-sacrifice", func(g *game.Game, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	candidates := rakdosCandidates(g, r.Chooser)
	if !r.Accepted || len(candidates) < 2 {
		return g.DrawNForEffect(controller, 2)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  r.Chooser,
		Source:   r.Source,
		Question: "Rakdos, Patron of Chaos — choose two nonland, nontoken permanents to sacrifice",
		Cards:    candidates,
		Min:      2,
		Max:      2,
		ThenKey:  rakdosPick,
		Carry:    r.Carry,
		LiveZone: game.ZoneBattlefield,
	})
	return nil
})

// rakdosPick is the two permanents chosen: sacrificed if the pick still holds,
// else the opponent did not sacrifice and Rakdos' controller draws.
var rakdosPick = game.RegisterChooseCardsThen("choose-cards/rakdos-patron-of-chaos-sacrifice", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	live := map[uuid.UUID]bool{}
	for _, id := range rakdosCandidates(g, r.Chooser) {
		live[id] = true
	}
	seen := map[uuid.UUID]bool{}
	ok := len(r.Picked) == 2
	for _, id := range r.Picked {
		if !live[id] || seen[id] {
			ok = false
		}
		seen[id] = true
	}
	if !ok {
		return g.DrawNForEffect(controller, 2)
	}
	for _, id := range r.Picked {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			return err
		}
	}
	return nil
})
