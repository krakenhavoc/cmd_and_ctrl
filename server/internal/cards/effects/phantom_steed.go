package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phantom Steed — Creature — Horse Illusion {3}{U}, 4/3 (Adventures in the
// Forgotten Realms Commander):
//
//	"Flash
//	 When this creature enters, exile another target creature you control
//	 until this creature leaves the battlefield.
//	 Whenever this creature attacks, create a tapped and attacking token
//	 that's a copy of the exiled card, except it's an Illusion in addition to
//	 its other types. Sacrifice that token at end of combat."
//
// Mardu Siegebreaker's pair on one body: the enters trigger is the
// exile-until-this-leaves with ExileUntil.Linked (CR 610.3, 607.2a), here with a
// required "another target creature", so the Steed leaving before it resolves
// exiles nothing (CR 610.3b). The attack trigger reads the card exiled with that
// ability of this object (Game.ExiledWithRowForEffect) and creates one token
// copy of it as printed (the ruling), an Illusion in addition to its other types
// (CR 707.9b), tapped and attacking (CR 508.4): the controller chooses what it
// attacks, preferring what the Steed attacked (Calamity's shape), and it was
// never declared as an attacker. The token is sacrificed by a delayed trigger
// at the beginning of the end of combat step (CR 603.7).
//
// No simplification.
func init() {
	const exileLabel = "Phantom Steed — exile another target creature you control until this creature leaves the battlefield"
	Register(Spec{
		OracleID:        "22c643b1-62dd-477c-8944-97b3becdfdbe",
		Name:            "Phantom Steed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(exileLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ExileUntil{Target: id, ThisLeaves: true, Linked: true,
					Label: "Phantom Steed — the exiled card returns when Phantom Steed leaves the battlefield"}.Apply(ctx)
			}),
				Another(TargetCreature("another target creature you control", YouControl()))),
			WheneverThisAttacks("Phantom Steed — create a tapped and attacking token that's a copy of the exiled card, except it's an Illusion in addition to its other types",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					exiled := g.ExiledWithRowForEffect(game.PermissionCardRef(ref), game.AbilitySlotTriggered, exileLabel)
					if len(exiled) == 0 {
						return nil
					}
					tmpl := tokenCopyTemplateFromCard(*exiled[len(exiled)-1])
					tmpl.TypeLine = addedSubtypeTypeLine(tmpl.TypeLine, "Illusion")
					return CreateTokensAttackingYourChoice{
						Template: tmpl,
						Copy:     true,
						N:        1,
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "Phantom Steed",
						Tail:     ctx.ItemTail(phantomSteedSacrificeAtEndOfCombat),
					}.Apply(ctx)
				}),
		},
	})
}

// phantomSteedSacrificeAtEndOfCombat is "sacrifice that token at end of
// combat": the token the creation made. An on-disk key: never renamed, never
// reused.
var phantomSteedSacrificeAtEndOfCombat = AttackingTokensThen("attacking-tokens/phantom-steed-sacrifice-at-end-of-combat",
	func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		item := tail.Asked.StackItem()
		if item == nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: "Phantom Steed — sacrifice that token",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(NewContext(ctx.Game, item))
	})
