package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thief of Existence — Creature — Eldrazi {1}{C}{G}, 3/4 (Modern Horizons 3,
// S82 GAP2 b05 #5948, #6088):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, exile up to one target noncreature, nonland
//	 permanent an opponent controls with mana value 4 or less. If you do,
//	 Thief of Existence gains 'When this creature leaves the battlefield,
//	 target opponent draws a card.'"
//
// Devoid is read from the printed card (CR 702.114a). The cast trigger
// fires from the stack (CR 601.2i, WhenYouCastThisSpell) and resolves above
// the spell, so it resolves even if the spell is countered. "If you do" is
// the exile's continuation, told whether the permanent reached exile. The
// grant is Rivaz of the Claw's shape: a grantAbilities record on the spell
// that lasts for the permanent it becomes (CR 400.7a,
// ForThePermanentThisSpellBecomesDuration), naming this card's bundle; a
// spell already gone from the stack gains nothing. The granted trigger
// targets an opponent as it is put on the stack (CR 603.3d).
//
// No simplification.
func init() {
	exile := effects.WhenYouCastThisSpell("Thief of Existence — exile up to one target noncreature, nonland permanent an opponent controls with mana value 4 or less",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(thiefOfExistenceExiled)}.Apply(ctx)
				}
			}
			return nil
		})
	exile.Targets = effects.TargetPermanent("up to one target noncreature, nonland permanent an opponent controls with mana value 4 or less",
		effects.Noncreature(), effects.Nonland(), effects.OpponentControls(), effects.ManaValueLE(4)).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:        "7abbd5af-a04b-476d-aee5-2b2f5114b019",
		Name:            "Thief of Existence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Grants: []effects.AbilityGrant{{
			Key: thiefOfExistenceGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.WhenThisLeaves("Thief of Existence — target opponent draws a card",
					effects.TheFirstLegalTargetPlayerDrawsACard),
					effects.TargetPlayer("target opponent", effects.Opponent())),
			},
			Text: "When this creature leaves the battlefield, target opponent draws a card.",
		}},
		Triggered: []game.TriggeredAbility{exile},
	})
}

// thiefOfExistenceGrant is the granted bundle's key. An on-disk key.
const thiefOfExistenceGrant = "thief-of-existence/leaves-opponent-draws"

// thiefOfExistenceExiled is "if you do": the spell gains the bundle for the
// permanent it becomes. An on-disk key.
var thiefOfExistenceExiled = effects.CardTail("then/thief-of-existence-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	spell := ctx.Source()
	if ctx.Game.StackItemForEffect(spell) == nil {
		return nil
	}
	ctx.Game.RegisterScopedEffectForEffect(spell,
		[]game.AffectedObject{{ID: spell}},
		[]game.Mod{game.GrantAbilitiesMod(thiefOfExistenceGrant)},
		game.ForThePermanentThisSpellBecomesDuration(spell),
		"Thief of Existence — it gains \"When this creature leaves the battlefield, target opponent draws a card.\"")
	return nil
})
