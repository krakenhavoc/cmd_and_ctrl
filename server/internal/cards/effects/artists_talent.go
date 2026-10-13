package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artist's Talent — Enchantment — Class, {1}{R}:
//
//	(Gain the next level as a sorcery to add its ability.)
//	Whenever you cast a noncreature spell, you may discard a card. If
//	you do, draw a card.
//	{2}{R}: Level 2
//	Noncreature spells you cast cost {1} less to cast.
//	{2}{R}: Level 3
//	If a source you control would deal noncombat damage to an opponent
//	or a permanent an opponent controls, it deals that much damage plus
//	2 instead.
//
// The Talent half of Cool but Rude's shape (ADR 0071), one gate per
// line:
//
//   - Level 1 is a cast trigger, always present (CR 716.3). "You may
//     discard a card. If you do, draw a card" is one instruction with a
//     conditional second half: the draw is the run's count, so declining
//     the discard draws nothing.
//   - Level 2 is a cost reduction gated at level 2, mirroring Fortune
//     Teller's Talent. A land is never a spell, so only noncreature
//     SPELLS are priced.
//   - Level 3 is Torbran's replacement without the colour test and with
//     Solphim's noncombat gate: any source you control, noncombat
//     damage only, to an opponent or a permanent an opponent controls
//     (damage to you and your own permanents is untouched), +2 per
//     damage instance. It is gated at level 3 and does not exist below.
//
// No simplification. (#3213, the Kuja deck.)
func init() {
	Register(Spec{
		OracleID:     "ea734863-5793-49d6-b10e-109432aec0a3",
		Name:         "Artist's Talent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{2}{R}")),
			LevelUp(3, ManaCost("{2}{R}")),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), artistsTalentLabel, artistsTalentRummage),
		},
		CostModifiers: []game.CostModifier{
			gatedCostModifier(Level(2), CostsLess(1,
				"Artist's Talent — noncreature spells you cast cost {1} less",
				YourSpell(), NoncreatureSpell())),
		},
		Replacements: []game.ReplacementEffect{{
			ActiveWhen: Level(3),
			Watches:    []game.EventKind{game.EventDealDamage},
			RowLabel:   "Noncombat damage your sources would deal to an opponent or their permanents is increased by 2.",
			AppliesTo:  noncombatDamageFromYourSourceToAnOpponent,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Artist's Talent: +2 noncombat damage",
		}},
	})
}

const artistsTalentLabel = "Artist's Talent — you may discard a card; if you do, draw a card"

func artistsTalentRummage(g *game.Game, item *game.StackItem) error {
	return b39MayDiscardThenDraw(1, false, artistsTalentLabel,
		func(discarded int) int { return discarded })(NewContext(g, item))
}
