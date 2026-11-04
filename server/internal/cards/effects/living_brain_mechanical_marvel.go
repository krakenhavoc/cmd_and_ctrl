package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Brain, Mechanical Marvel — Legendary Artifact Creature — Robot
// Villain {4}, 3/3 (Marvel's Spider-Man, #4460):
//
//	"At the beginning of combat on your turn, target non-Equipment artifact
//	 you control becomes an artifact creature with base power and toughness
//	 3/3 until end of turn. Untap it."
//
// The animation is the crew shape (BecomeCreatureUntilEOT: a layer 4 type
// change and a layer 7b base P/T set at one timestamp, pinned to the object
// so a flickered target is a new object and is not reached, CR 400.7). Base
// power and toughness are set, so counters and other effects still apply on
// top. The target is untapped whether or not it was already a creature; the
// untap is a separate instruction of the same trigger, run after the
// animation. Living Brain may target itself (it is a non-Equipment
// artifact).
//
// No simplification.
func init() {
	const label = "Living Brain, Mechanical Marvel — target non-Equipment artifact you control becomes a 3/3 artifact creature until end of turn; untap it"
	Register(Spec{
		OracleID:     "b255e13d-01b2-4348-80bd-8d9a8f29be91",
		Name:         "Living Brain, Mechanical Marvel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, livingBrainAnimate),
				TargetPermanent("target non-Equipment artifact you control", Artifact(), Not(HasSubtype("Equipment")), YouControl())),
		},
	})
}

// livingBrainAnimate animates and untaps each still-legal target.
func livingBrainAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BecomeCreatureUntilEOT{
			Target:       t.ID,
			SetPower:     3,
			SetToughness: 3,
			Label:        "Living Brain — becomes a 3/3 artifact creature until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
