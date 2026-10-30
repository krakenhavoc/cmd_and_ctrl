package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tony Stark // The Invincible Iron Man — a double-faced Legendary Creature
// (Marvel Super Heroes, #4455):
//
//	Tony Stark {1}{U}, Legendary Creature — Human Artificer Hero
//	  "{1}, {T}: Look at the top four cards of your library. You may reveal
//	   an artifact card from among them and put it into your hand. Put the
//	   rest on the bottom of your library in a random order.
//	   {4}{U}{R}: Transform Tony Stark. Activate only as a sorcery."
//
//	The Invincible Iron Man, Legendary Artifact Creature — Human Hero
//	  "Flying, haste
//	   At the beginning of combat on your turn, you may put an artifact card
//	   from your hand onto the battlefield. If it's an Equipment, attach it
//	   to The Invincible Iron Man."
//
// The look is Dáin's Company's (TakeFromLibraryToHand over the top four, a
// floor of zero for the "may", the taken card revealed, the rest bottomed in
// a random order). The transform is the permanent turning over in place
// (Bruce Banner's TransformThis). The beginning-of-combat put is
// Court of Bounty's PutFromHandOntoBattlefield with the optional floor; its
// keyed tail attaches the card that arrived when it is an Equipment, to The
// Invincible Iron Man (the tail's source), as Armored Skyhunter attaches.
//
// No simplification.
func init() {
	const oracle = "4a46ea59-6c5d-4bed-9951-b5590f9a1bdb"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Tony Stark",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Look at the top four cards of your library. You may reveal an artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Purpose: game.Purpose{
				Answers: game.AnswerValue,
				Tutors:  1,
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
					Match:    Artifact(),
					Max:      1,
					Optional: true,
					Label:    "Tony Stark — you may reveal an artifact card and put it into your hand",
					Key:      takeRevealRestBottomRandom,
				}.Apply(NewContext(g, item))
			},
		}, {
			Label:        "{4}{U}{R}: Transform Tony Stark. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{U}{R}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "The Invincible Iron Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("The Invincible Iron Man — you may put an artifact card from your hand onto the battlefield; attach an Equipment to him",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    Artifact(),
						Optional: true,
						Label:    "The Invincible Iron Man — you may put an artifact card from your hand onto the battlefield",
						ThenTail: ironManAttachEquipment,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// ironManAttachEquipment is "if it's an Equipment, attach it to The
// Invincible Iron Man", keyed: the card that arrived, attached to the
// tail's source. An on-disk key.
var ironManAttachEquipment = PutFromHandTail("the-invincible-iron-man/attach-equipment",
	func(ctx *Context, res PutFromHandResult) error {
		if res.Entered == uuid.Nil {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(res.Entered)
		if !ok || !c.HasSubtype("Equipment") {
			return nil
		}
		return ctx.Game.AttachForEffect(res.Entered, game.TargetRef{Kind: game.TargetCard, ID: res.Source})
	})
