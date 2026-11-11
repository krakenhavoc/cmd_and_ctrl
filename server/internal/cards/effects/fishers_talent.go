package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fisher's Talent — Enchantment — Class {2}{G}{U} (Bloomburrow Commander):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 At the beginning of your upkeep, look at the top card of your
//	 library. You may reveal it if it's a land card. Create a 1/1 blue
//	 Fish creature token if you revealed it this way. Then draw a card.
//	 {G}{U}: Level 2
//	 If you would create a Fish token, create a 3/3 blue Shark creature
//	 token instead.
//	 {2}{G}{U}: Level 3
//	 If you would create a Shark token, create an 8/8 blue Octopus
//	 creature token instead."
//
// ADR 0071 Class machinery (Hunter's Talent): LevelUp for each level, the
// two replacements gated ActiveWhen Level(2) and Level(3). The upkeep
// look is Sidequest: Catch a Fish's (LookAtTopOfLibraryForEffect); the
// reveal is a "may" asked only when the card is a land (the card stays on
// top either way), the Fish follows only if it was revealed, and the draw
// follows in every case. Both replacements rewrite token kinds on the
// CR 701.7b creation instruction (ReplaceTokenKindsWhere): a Fish becomes
// a Shark, and a Shark — including the one a level-2 replacement just
// produced — becomes an Octopus, so a level-3 Class turns each Fish into
// an 8/8, and any Shark another effect makes too. Counts are kept
// ("that many"), and they apply to tokens created under the Class's
// controller only.
//
// No simplification.
func init() {
	isFish := func(t game.Card) bool { return t.IsToken() && t.HasSubtypeWord("Fish") && t.IsCreature() }
	isShark := func(t game.Card) bool { return t.IsToken() && t.HasSubtypeWord("Shark") && t.IsCreature() }
	rewrite := func(label string, match func(game.Card) bool, kind string, level game.Designation) game.ReplacementEffect {
		return game.ReplacementEffect{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(match)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.ReplaceTokenKindsWhere(match, TokenCard(kind))
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:      label,
			ActiveWhen: level,
		}
	}
	Register(Spec{
		OracleID:     "44ba2aa5-2bf5-4267-ae82-f0daf8f5e3e8",
		Name:         "Fisher's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Fisher's Talent — look at the top card of your library; you may reveal it if it's a land card, then create a 1/1 blue Fish if you did; then draw a card",
				Do(Instruction(fishersTalentLook))),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{G}{U}")),
			LevelUp(3, ManaCost("{2}{G}{U}")),
		},
		Replacements: []game.ReplacementEffect{
			rewrite("Fisher's Talent: a 3/3 blue Shark instead of a Fish", isFish, "3/3 blue Shark", Level(2)),
			rewrite("Fisher's Talent: an 8/8 blue Octopus instead of a Shark", isShark, "8/8 blue Octopus", Level(3)),
		},
	})
}

// fishersTalentLook is the upkeep ability.
func fishersTalentLook(ctx *Context) error {
	looked := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 1)
	if len(looked) > 0 {
		if c, ok := ctx.Game.LookupCardForEffect(looked[0]); ok && c.IsLand() {
			return MayChoice{
				Player:   ctx.Controller(),
				Question: "Fisher's Talent — reveal " + c.Name + " and create a 1/1 blue Fish?",
				YesLabel: "Reveal it",
				NoLabel:  "Don't reveal it",
				Then:     fishersTalentAnswer,
				Carry:    []uuid.UUID{looked[0]},
			}.Apply(ctx)
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// fishersTalentAnswer reveals the card and makes the Fish on a yes, then
// draws either way.
var fishersTalentAnswer = MayYesNo("bloomburrow-commander/fishers-talent-reveal",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if id := carryAt(r, 0); bisonWhistleStillInLibrary(ctx.Game, id) {
			ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Fisher's Talent — reveal the top card")
			if err := (CreateToken{Template: TokenCard("1/1 blue Fish"), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return DrawCards{N: 1}.Apply(ctx)
	},
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return DrawCards{N: 1}.Apply(ctx)
	})
