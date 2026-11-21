package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sontaran General — Creature — Alien Soldier {4}{R}, 5/5 (Doctor Who):
//
//	"Trample, haste
//	 Battalion — Whenever this creature and at least two other creatures attack, for
//	 each opponent, goad up to one target creature that player controls. Those
//	 creatures can't block this turn. (Until your next turn, they attack each combat
//	 if able and attack a player other than you if able.)"
//
// Battalion is the shared trigger (gtcBattalion). "For each opponent, up to one target
// creature that player controls" is one optional clause per seat (perOpponentClauses):
// the clause of an opponent who has been eliminated or controls no creature is simply
// empty. Each creature chosen is goaded by the General's controller (CR 701.15) and
// can't block this turn; one that has left by resolution is untouched.
//
// No simplification.
func init() {
	battalion := gtcBattalion("Sontaran General — for each opponent, goad up to one target creature that player controls; those creatures can't block this turn",
		sontaranGeneralBattalion)
	battalion.Targets = perOpponentClauses(game.MaxPlayers-1, "creature", Creature())
	Register(Spec{
		OracleID:        "76a416fa-d97b-4c24-af41-b2382f1fd546",
		Name:            "Sontaran General",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered:       []game.TriggeredAbility{battalion},
	})
}

// sontaranGeneralBattalion goads every creature that was chosen and is still on the
// battlefield, then makes each of them unable to block this turn.
func sontaranGeneralBattalion(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var chosen []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && onBattlefield(g, t.ID) {
			chosen = append(chosen, t.ID)
		}
	}
	if len(chosen) == 0 {
		return nil
	}
	if err := goadAndScheduleClear(ctx, chosen); err != nil {
		return err
	}
	for _, id := range chosen {
		if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Sontaran General — can't block this turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
