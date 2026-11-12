package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trove Warden — Creature — Cat Beast {2}{W}{W}, 3/4 (Zendikar Rising
// Commander, #5509):
//
//	"Vigilance
//	 Landfall — Whenever a land you control enters, exile target permanent
//	 card with mana value 3 or less from your graveyard.
//	 When this creature dies, put each permanent card exiled with it onto
//	 the battlefield under the control of that card's owner."
//
// The two triggers are linked (CR 607.2a), Ghost Vacuum's shape: the
// landfall exile is Ghost Vacuum's own body (stdB26GhostVacuumExile, a
// target that left in response exiles nothing), which stamps the link on
// the card (ExileTarget.Linked), naming
// this Warden as the object it is, and the dies trigger reads the cards
// linked to the object it came from, which is the Warden as it last existed
// on the battlefield. A Warden that leaves and comes back is a new object
// (CR 400.7) and has exiled nothing. A card something else moved out of
// exile is no longer "exiled with it" and is not pulled back.
//
// The target is a permanent card you own in your graveyard with mana value
// 3 or less, chosen as the trigger goes on the stack and checked again as it
// resolves (CR 603.3d, CR 608.2b). The cards come back as one entry
// (CR 603.6a), each under its owner's control, so they see each other enter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9056eba4-612b-4e82-8689-fe098241b007",
		Name:            "Trove Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Trove Warden — exile target permanent card with mana value 3 or less from your graveyard (landfall)", stdB26GhostVacuumExile),
				TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard", YouOwn(), Permanent(), ManaValueLE(3))),
			WhenThisDies("Trove Warden — put each permanent card exiled with it onto the battlefield under the control of that card's owner",
				troveWardenReturn),
		},
	})
}

// troveWardenReturn puts every permanent card still exiled with the Warden
// that died onto the battlefield, together, under each card's owner.
func troveWardenReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsPermanent() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromExileTogether{Targets: ids}.Apply(ctx)
}
