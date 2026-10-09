package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blaring Captain — Creature — Azra Warrior {3}{B}, 2/2:
//
//	"Partner with Blaring Recruiter (When this creature enters, target
//	 player may put Blaring Recruiter into their hand from their
//	 library, then shuffle.)
//	 Whenever this creature attacks, attacking Warriors get +1/+1 until
//	 end of turn."
//
// Not legendary, so it can't be a commander: of CR 702.124j's two
// abilities only the entry search does anything (PartnerWith, #2142).
//
// "Attacking Warriors" is every attacking Warrior, the Captain
// included, read once as the trigger resolves (CR 611.2c): a Warrior
// put onto the battlefield attacking afterwards does not get the bonus.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1675476c-25b1-46c2-81ce-fef9a2360bcc",
		Name:         "Blaring Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Blaring Captain", "Blaring Recruiter"),
			WheneverThisAttacks("Blaring Captain — attacking Warriors get +1/+1",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() && c.HasSubtype("Warrior") && c.AttackingTarget != uuid.Nil
						},
						Power:     1,
						Toughness: 1,
						Label:     "Blaring Captain — +1/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
