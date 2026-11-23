package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muddle, the Ever-Changing — Legendary Creature — Elemental Otter
// Shapeshifter {2}{U}{R}, 3/3 (Secrets of Strixhaven Commander, #4834):
//
//	"Whenever you cast an instant or sorcery spell, Muddle becomes a copy
//	 of up to one target nonlegendary creature you control until end of
//	 turn, except it has myriad. (Whenever it attacks, for each opponent
//	 other than defending player, you may create a token copy that's
//	 tapped and attacking that player or a planeswalker they control.
//	 Exile the tokens at end of combat.)"
//
// Hulkling, Young Avenger's shape: a duration copy (BecomeCopy, CR
// 707.2) whose "except" clause (CR 707.9a) grants myriad, the granted
// keyword trigger of myriad.go (CR 702.116a). The exception is part of
// the copiable values, so the myriad tokens have myriad too, though they
// enter attacking without being declared and it never triggers for them
// (CR 508.4). Muddle does not keep its own trigger while it is a copy,
// and becoming a copy is neither entering nor leaving (the rulings). With
// no target chosen, or a target no longer legal, nothing happens and
// Muddle gets no myriad.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a8f6e5e-6cde-49df-861c-a26cb09c9b12",
		Name:         "Muddle, the Ever-Changing",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       muddleMyriadGrant,
			Text:      "Myriad",
			Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Muddle, the Ever-Changing")},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Muddle — becomes a copy of up to one target nonlegendary creature you control until end of turn, with myriad", muddleCopy),
				effects.PermanentYouControl("up to one target nonlegendary creature you control", effects.Creature(), effects.Not(effects.Legendary())).WithCount(0, 1)),
		},
	})
}

const muddleMyriadGrant = "muddle-the-ever-changing/myriad"

func muddleCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return effects.BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      targets[0].ID,
		Except: func(v *game.PrintedValues) {
			v.GrantAbility(muddleMyriadGrant)
		},
		Label: "Muddle — becomes a copy until end of turn, except it has myriad",
	}.Apply(ctx)
}
