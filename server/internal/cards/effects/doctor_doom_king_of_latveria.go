package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doctor Doom, King of Latveria — Legendary Creature — Human Noble
// Villain {1}{U}{B}{R}, 3/3 (Marvel Super Heroes Commander):
//
//	"Whenever you discard one or more land cards, each opponent loses 2
//	 life.
//	 At the beginning of combat on your turn, target Villain you control
//	 gains menace until end of turn. It connives."
//
// The discard trigger is Captain Howler's batch shape: one trigger per
// batch of discards (ThatManyEvents counts only the land cards), so a
// hand of three lands discarded together drains each opponent once. The
// combat trigger targets a Villain (Doom himself qualifies) and the
// connive runs for the target, not for Doom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c9cd0fd-bfe9-409c-a9a9-146f5f4a2ef4",
		Name:         "Doctor Doom, King of Latveria",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !discardedByYou(ev, source) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand()
			}, "Doctor Doom, King of Latveria — each opponent loses 2 life",
				Do(EachOpponentLosesLife{N: 2})), ThatManyEvents()),
			Targeting(
				AtBeginningOfYourCombat("Doctor Doom, King of Latveria — target Villain you control gains menace until end of turn and connives",
					msc6DoomMenaceAndConnive),
				TargetCreature("target Villain you control", YouControl(), HasSubtype("Villain"))),
		},
	})
}

func msc6DoomMenaceAndConnive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace"},
			Label: "Doctor Doom, King of Latveria — menace"}).Apply(ctx); err != nil {
			return err
		}
		return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{t.ID}}.Apply(ctx)
	}
	return nil
}
