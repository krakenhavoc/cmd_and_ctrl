package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakdos Guildmage — Creature — Zombie Shaman {B/R}{B/R}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"({B/R} can be paid with either {B} or {R}.)
//	 {3}{B}, Discard a card: Target creature gets -2/-2 until end of turn.
//	 {3}{R}: Create a 2/1 red Goblin creature token with haste. Exile it at
//	 the beginning of the next end step."
//
// The first row is Chittering Witch's -2/-2 with a discard cost chosen as it
// is activated (CR 602.2b). The second creates the token and schedules one
// delayed trigger over exactly that token (CR 603.7), as Saheeli Rai's does:
// "the next end step" is any player's, so one made in an end step is exiled
// at the next turn's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "650949ca-609b-46c5-bc6b-6f782cdcb6d2",
		Name:         "Rakdos Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{B}, Discard a card: Target creature gets -2/-2 until end of turn.",
				Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
				Cost:    Plus(ManaCost("{3}{B}"), DiscardACard()),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Rakdos Guildmage — -2/-2", -2, -2),
			},
			{
				Label:   "{3}{R}: Create a 2/1 red Goblin creature token with haste. Exile it at the beginning of the next end step.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{3}{R}"),
				Effect:  rakdosGuildmageGoblin,
			},
		},
	})
}

func rakdosGuildmageGoblin(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	made, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("2/1 red Goblin with haste")), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Rakdos Guildmage — exile the Goblin token",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}
