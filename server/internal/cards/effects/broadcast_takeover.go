package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Broadcast Takeover — {2}{R}{R}{R} Sorcery (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Gain control of all artifacts your opponents control until end of
//	 turn. Untap them. They gain haste until end of turn."
//
// A mass Threaten over artifacts. The set is the artifacts the caster's
// opponents control as the spell resolves (CR 611.2c: an artifact that
// arrives later is not taken), and the three sentences run in printed
// order over that set (CR 608.2c): every artifact changes control first,
// then each is untapped, then each gains haste for the turn. Haste on a
// noncreature artifact matters once it becomes a creature this turn (a
// crewed Vehicle). Control returns at end of turn (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0d221da-16cd-4399-830f-eecc9a0912eb",
		Name:         "Broadcast Takeover",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, spell *Context) error {
			ctx := spell.asGroupMember()
			var taken []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsArtifact() && c.Controller != item.Controller {
					taken = append(taken, c.InstanceID)
				}
			}
			for _, id := range taken {
				if err := (GainControl{
					Target:   id,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Broadcast Takeover — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range taken {
				if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range taken {
				if err := (GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"haste"},
					Label:    "Broadcast Takeover — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
