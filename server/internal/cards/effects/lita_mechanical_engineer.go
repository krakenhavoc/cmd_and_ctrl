package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lita, Mechanical Engineer — Legendary Artifact Creature — Artificer
// {2}{W}, 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Vigilance
//	 At the beginning of your end step, untap each other artifact
//	 creature you control.
//	 {3}{W}, {T}: Create a 5/5 colorless Vehicle artifact token named
//	 Zeppelin with flying and crew 3."
//
// The end-step trigger untaps every artifact creature its controller
// controls other than Lita as it resolves, read through the layered view,
// so a crewed Vehicle is one and an uncrewed one is not. The Zeppelin is
// a catalog token template (avB24ZeppelinToken): an artifact Vehicle that
// is a creature only while crewed, its crew 3 the same CR 702.122
// ability a printed Vehicle has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "715cb547-726f-4bc8-ad46-297d4dccb881",
		Name:            "Lita, Mechanical Engineer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Lita, Mechanical Engineer — untap each other artifact creature you control", litaUntapArtifactCreatures),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}, {T}: Create a 5/5 colorless Vehicle artifact token named Zeppelin with flying and crew 3.",
			Cost:    Plus(ManaCost("{3}{W}"), TapCost()),
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Effect:  Do(CreateToken{Template: avB24ZeppelinToken(), N: 1}),
		}},
	})
}

// litaUntapArtifactCreatures untaps each other artifact creature the
// trigger's controller controls.
func litaUntapArtifactCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != item.SourceCardID && c.Controller == item.Controller && c.IsCreature() && c.IsArtifact() && c.Tapped {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
