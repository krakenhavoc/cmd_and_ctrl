package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of Reverence — 1/6 Creature — Spirit Wall for {3}{W} (EDHREC
// rank 4401):
//
//	"Defender, flying
//	 At the beginning of your end step, you may gain life equal to
//	 the power of target creature you control."
//
// A blocker that holds off the air force and turns your biggest
// creature into a lifegain engine every turn without attacking with
// it. In a lifegain deck that end-step trigger is the whole reason
// the card sees play. Roadmap batch 42 (#449), "no new machinery".
//
// The target is chosen when the trigger goes on the stack at the
// beginning of the end step (CR 603.3d), and the power is read when
// the trigger RESOLVES — so pumping the creature in response gains
// you the bigger number, and killing it means the trigger has no
// legal target and does nothing.
//
// LAYERED power (CurrentPower), so counters, anthems and a Giant
// Growth all count, and the answer is whatever the creature's power
// actually is at that moment — including a negative one, which gains
// nothing rather than losing you life (GainLife with a non-positive
// amount is a no-op, and the printed card gains "life equal to the
// power", which for a 0-power creature is zero).
//
// "You may" is a real prompt on the trigger, and "target creature YOU
// CONTROL" means an opponent's fatty is not a legal choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0810983f-818a-43e6-a7b5-ebe0bc8b9f6a",
		Name:            "Wall of Reverence",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					AtYourEndStep("Wall of Reverence — gain life equal to the target's power",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							c, ok := g.LookupCardForEffect(item.Targets[0].ID)
							if !ok {
								return nil
							}
							return GainLife{
								Player: item.Controller,
								Amount: c.CurrentPower(),
							}.Apply(NewContext(g, item))
						}),
					PermanentYouControl("target creature you control", Creature())),
				"Wall of Reverence — gain life equal to the target creature's power?"),
		},
	})
}
