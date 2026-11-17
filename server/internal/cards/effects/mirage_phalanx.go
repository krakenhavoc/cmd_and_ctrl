package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirage Phalanx — {4}{R}{R} Creature — Human Soldier 4/4
// (Innistrad: Crimson Vow Commander, #5675):
//
//	"Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control
//	 both of them.)
//	 As long as Mirage Phalanx is paired with another creature, each of
//	 those creatures has "At the beginning of combat on your turn,
//	 create a token that's a copy of this creature, except it has haste
//	 and loses soulbond. Exile it at end of combat.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 grant to both creatures of the pair, so each
// makes a copy of itself at the beginning of its controller's combat.
//
// The token copies the creature's copiable values (CR 707.2) and
// nothing else: not the pair, and not the ability the pair grants,
// which is no part of what a copy copies. The two exceptions are
// copiable (CR 707.9b): haste, and the loss of soulbond
// (game.Card.LoseCopiedKeyword), so the copy never asks to pair. A copy
// of Mirage Phalanx itself keeps the "as long as … is paired" line; it
// does nothing unless another creature's soulbond pairs with the token.
// A creature that has left the battlefield is copied as it last existed
// there (CR 608.2h).
//
// "Exile it at end of combat" is a delayed trigger at the beginning of
// that turn's end of combat step (CR 511.2), naming the token.
//
// No simplification.
const miragePhalanxGrant = "mirage-phalanx/copy"

func init() {
	Register(Spec{
		OracleID:        "d18dfca7-2947-4308-a9ef-501e1b4843f4",
		Name:            "Mirage Phalanx",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Mirage Phalanx"),
		Grants: []AbilityGrant{{
			Key: miragePhalanxGrant,
			Triggered: []game.TriggeredAbility{
				AtBeginningOfYourCombat("Mirage Phalanx — create a token that's a copy of this creature, except it has haste and loses soulbond; exile it at end of combat",
					miragePhalanxCopy),
			},
			Text: "At the beginning of combat on your turn, create a token that's a copy of this creature, except it has haste and loses soulbond. Exile it at end of combat.",
		}},
		Static: []game.StaticAbility{
			PairedCreaturesHave("As long as Mirage Phalanx is paired with another creature, each of those creatures has \"At the beginning of combat on your turn, create a token that's a copy of this creature, except it has haste and loses soulbond. Exile it at end of combat.\"",
				miragePhalanxGrant),
		},
	})
}

// miragePhalanxCopy makes the hasty copy of the ability's source and
// schedules its exile at end of combat.
func miragePhalanxCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	tmpl, ok := TokenCopyTemplate(g, item.SourceCardID)
	if !ok {
		return nil
	}
	TokenCopyGainsHaste(&tmpl)
	tmpl.LoseCopiedKeyword(game.KeywordSoulbond)
	made, err := g.CreateTokensForEffect(item.Controller, copiedToken(tmpl), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Mirage Phalanx — exile the token",
		Cards: made,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
