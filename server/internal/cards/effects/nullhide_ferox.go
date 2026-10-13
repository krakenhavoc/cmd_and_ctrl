package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nullhide Ferox — Creature — Beast {2}{G}{G}, 6/6:
//
//	"Hexproof
//	 You can't cast noncreature spells.
//	 {2}: This creature loses all abilities until end of turn. Any
//	 player may activate this ability.
//	 If a spell or ability an opponent controls causes you to discard
//	 this card, put it onto the battlefield instead of putting it into
//	 your graveyard."
//
// The card #1861 was filed for. Every line already existed:
//
//   - hexproof is a printed keyword;
//   - "You can't cast noncreature spells" is a cast restriction on its
//     controller only (YouCantCastUnless with no escape), so an
//     opponent who steals the Ferox is the one restricted;
//   - the {2} ability is an any-player row (ADR 0106, #1793). The
//     activator pays {2} and is "you"; the effect names the Ferox, not
//     its controller. It removes the Ferox's own abilities, hexproof
//     and the cast restriction included, until end of turn (CR
//     613.1f), which is the point of the card;
//   - the discard line is the Baloth replacement (ADR 0153, see
//     obstinate_baloth.go).
//
// The any-player row declares only what it answers: removing hexproof
// is AnswerRestrict (ADR 0142). It buys the activator no amount a bot
// prices, so a bot never pays {2} to do it (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "5b04a337-3152-481e-973f-a11dbd615f93",
		Name:              "Nullhide Ferox",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"hexproof"},
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		CastRestrictions: []game.CastRestriction{
			YouCantCastUnless("You can't cast noncreature spells.", Not(Creature()), nil),
		},
		Activated: []ActivatedAbility{{
			Label:     "{2}: This creature loses all abilities until end of turn. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerRestrict},
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
					return nil
				}
				return untilEndOfTurn(ctx, item.SourceCardID, nil,
					"Nullhide Ferox — loses all abilities", game.LoseAllAbilitiesMod())
			},
		}},
		Replacements: []game.ReplacementEffect{
			OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Nullhide Ferox", nil),
		},
	})
}
