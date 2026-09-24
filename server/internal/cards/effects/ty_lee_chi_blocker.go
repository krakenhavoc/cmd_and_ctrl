package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ty Lee, Chi Blocker — Legendary Creature — Human Performer Ally {2}{U}, 2/1:
//
//	"Flash
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When Ty Lee enters, tap up to one target creature. It doesn't untap
//	 during its controller's untap step for as long as you control Ty
//	 Lee."
//
// #1313 (deck tracker #1306): the proof card for untap holds — an
// UntapSkip carrying a CR 611.2b "for as long as you control ~"
// duration (ADR 0058's 2026-09-23 amendment). The lock is read at
// every untap step of the held creature's controller and ends the
// moment Ty Lee leaves or changes control, for good.
//
// #706: prowess is a canonical keyword token the engine turns into a
// trigger (game/prowess.go), so declaring it below is the whole of
// it — and a deck-imported Ty Lee would carry it off Scryfall even
// without the declaration. She is an Ally, so under Sokka, Tenacious
// Tactician she has two instances and gets +2/+2 per spell (CR
// 702.108b).
func init() {
	Register(Spec{
		OracleID:        "081ad4e3-cda3-41cc-890f-412611dc9ea0",
		Name:            "Ty Lee, Chi Blocker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", game.KeywordProwess},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Ty Lee — tap up to one target creature; it doesn't untap while you control Ty Lee",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TapAndHoldWhileYouControlThis(ctx, holdTargetIDs(ctx))
				}),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}
