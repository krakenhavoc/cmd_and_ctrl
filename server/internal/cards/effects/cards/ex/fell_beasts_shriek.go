package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fell Beast's Shriek — Sorcery {U}{R} (Tales of Middle-earth Commander):
//
//	"Each opponent chooses a creature they control. Tap and goad the
//	 chosen creatures.
//	 Splice onto instant or sorcery {2}{U}{R}"
//
// Splice is ADR 0202's seam (Spec.Splice); the engine adds this text to
// the spell it is spliced onto. The opponents choose in turn order, one
// prompt each (Rejoin the Fight's chain: the opponents still to ask and
// the creatures chosen so far ride the prompt's carry), and nothing is
// tapped or goaded until every opponent has chosen, so the picks are
// made as one choice. An opponent with one creature is not asked and
// chooses it; one with none chooses nothing. The chosen creatures are
// tapped, then goaded by the caster until the caster's next turn (CR
// 701.15), the goad Alela's marker and goad.go share.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d660f6a0-3642-4445-b877-fda2a2950680",
		Name:         "Fell Beast's Shriek",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoInstantOrSorcery("{2}{U}{R}"),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return fellBeastsShriekAsk(ctx, effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller()), nil)
		})),
	})
}

// fellBeastsShriekAsk asks the next opponent with a real choice, then
// continues; once the list is empty the chosen creatures are tapped and
// goaded.
func fellBeastsShriekAsk(ctx *effects.Context, opponents, chosen []uuid.UUID) error {
	g := ctx.Game
	for len(opponents) > 0 {
		opp, rest := opponents[0], opponents[1:]
		opponents = rest
		var cands []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == opp && c.IsCreature() {
				cands = append(cands, c.InstanceID)
			}
		}
		switch len(cands) {
		case 0:
			continue
		case 1:
			chosen = append(append([]uuid.UUID(nil), chosen...), cands[0])
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  opp,
			Source:   ctx.Source(),
			Question: "Fell Beast's Shriek — choose a creature you control to tap and be goaded",
			Cards:    cands,
			Min:      1,
			Max:      1,
			Zone:     game.ZoneBattlefield,
			LiveZone: game.ZoneBattlefield,
			ThenKey:  fellBeastsShriekPick,
			Carry:    effects.PackIDLists(rest, chosen),
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	var live []uuid.UUID
	for _, id := range chosen {
		if effects.OnBattlefield(g, id) {
			live = append(live, id)
		}
	}
	if len(live) == 0 {
		return nil
	}
	for _, id := range live {
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return effects.GoadAndScheduleClear(ctx, live)
}

// fellBeastsShriekPick is one opponent's choice answered: the next
// opponent is asked, or the chosen creatures are tapped and goaded. An
// on-disk key: never renamed, never reused.
var fellBeastsShriekPick game.ChooseCardsThen

func init() {
	fellBeastsShriekPick = effects.AskedPick("choose-cards/fell-beasts-shriek-pick", func(ctx *effects.Context, r game.CardsChosen) error {
		rest, chosen := effects.UnpackIDPair(r.Carry)
		return fellBeastsShriekAsk(ctx, rest, append(append([]uuid.UUID(nil), chosen...), r.Picked...))
	})
}
