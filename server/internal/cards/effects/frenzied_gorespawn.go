package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frenzied Gorespawn — Creature — Horror {3}{B}{R}, 4/4 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, for each opponent, goad target creature
//	 that player controls. (Until your next turn, those creatures attack
//	 each combat if able and attack a player other than you if able.)
//	 Whenever one or more creatures attack one of your opponents, those
//	 creatures gain menace until end of turn."
//
// The enter trigger is Havoc Eater's goad (goadAndScheduleClear) with the
// required form of the per-opponent clauses (Blatant Thievery's
// perOpponentRequiredClauses: one exact target per live opponent, bound to
// that player). Each pick still legal on resolution is goaded (CR 608.2b).
//
// The second ability is a batch trigger per attacked player (Sabotage
// Strategist's batch read, CR 603.2c): any creature, whoever controls it,
// attacking one of the controller's opponents themselves; an attack on a
// planeswalker or battle is no attack on the player (CR 508.3a, #4839). It
// triggers once for each opponent attacked in a declaration, and those
// creatures are the ones that attacked that opponent, each still on the
// battlefield.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Frenzied Gorespawn — for each opponent, goad target creature that player controls", frenziedGorespawnEnters)
	enters.Targets = perOpponentRequiredClauses("creature", Creature())
	Register(Spec{
		OracleID:     "197809b7-1ddb-4329-8d82-7651696538a8",
		Name:         "Frenzied Gorespawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				defender := attackedPlayer(g, ev)
				if defender == uuid.Nil || defender == source.Controller {
					return false
				}
				p := g.PlayerByIDForEffect(defender)
				return p != nil && !p.Eliminated
			}, "Frenzied Gorespawn — those creatures gain menace until end of turn", frenziedGorespawnMenace)),
		},
	})
}

func frenziedGorespawnEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var goaded []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && onBattlefield(g, t.ID) {
			goaded = append(goaded, t.ID)
		}
	}
	if len(goaded) == 0 {
		return nil
	}
	return goadAndScheduleClear(ctx, goaded)
}

func frenziedGorespawnMenace(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	trig := item.Trigger.Event
	defender := attackedPlayer(g, trig)
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		if ev.Kind != game.EventAttack || ev.Batch != trig.Batch || ev.Target != defender || ev.CardID == uuid.Nil || seen[ev.CardID] {
			continue
		}
		seen[ev.CardID] = true
		if !onBattlefield(g, ev.CardID) {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: ev.CardID, Keywords: []string{"menace"}, Label: "Frenzied Gorespawn — menace"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
