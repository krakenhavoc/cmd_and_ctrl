package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ivy, Gleeful Spellthief — Legendary Creature — Faerie Rogue {G}{U}, 2/1
// (Dominaria United):
//
//	"Flying
//	 Whenever a player casts a spell that targets only a single creature other
//	 than Ivy, you may copy that spell. The copy targets Ivy. (A copy of an
//	 Aura spell becomes a token.)"
//
// "Targets only a single creature" is CR 115.9c, read as Chef's Kiss reads it:
// every target of the spell is the same creature on the battlefield (a spell
// that names it twice qualifies). Any player's spell counts, including Ivy's
// controller's own, but a spell aimed at Ivy does not. The copy is Feather,
// Radiant Arbiter's: put on the stack under Ivy's controller (CR 707.10), not
// cast, with every target that was the creature now Ivy
// (CopySpellRetargetedForEffect), and only if Ivy is a legal target for the
// copy: a spell that can't target Ivy (hexproof, protection, a restriction)
// is not copied. An Aura copy enters as a token attached to Ivy (CR 707.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "55ad6a6b-1c44-4397-86ae-dd9221892b22",
		Name:            "Ivy, Gleeful Spellthief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast,
				dmuB17CastSpellTargetsOnly(false, func(target game.Card, ivy *game.Card) bool {
					return target.IsCreature() && target.InstanceID != ivy.InstanceID
				}),
				"Ivy, Gleeful Spellthief — you may copy that spell; the copy targets Ivy",
				ivyCopyTheSpell), "Ivy, Gleeful Spellthief — copy that spell? The copy targets Ivy."),
		},
	})
}

func ivyCopyTheSpell(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell := item.Trigger.Event.CardID
	target, ok := g.StackItemSoleTargetForEffect(spell)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	g.CopySpellRetargetedForEffect(spell, item.Controller, target.ID, item.SourceCardID)
	return nil
}
