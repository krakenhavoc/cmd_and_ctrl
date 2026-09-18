package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gargos, Vicious Watcher — Legendary Creature — Hydra {3}{G}{G}{G},
// 8/7 (EDHREC rank 4234):
//
//	"Vigilance
//	 Hydra spells you cast cost {4} less to cast.
//	 Whenever a creature you control becomes the target of a spell,
//	 Gargos fights up to one target creature you don't control."
//
// The Hydra tribal commander: a six-mana 8/7 that makes every other
// Hydra four mana cheaper, and punishes anyone who points removal at
// your board by eating a creature in response. It is in this batch
// because the discount needed the S28 cost-modification pipeline
// (#93) — one of the five mechanics the 2026-09-18 re-triage freed.
//
// The discount is a CostModifiers entry with TWO predicates, and
// leaving either off ships a different card: YourSpell, because
// "Hydra spells YOU cast" does not help the opponent's Hydra deck,
// and the subtype test, read off the printed type line of the card
// being cast. It spends against generic mana only and stops at zero
// (CR 601.2f, enforced engine-side), so a {X}{G}{G} Hydra still costs
// {G}{G} and never goes negative.
//
// The fight trigger is the subtle half:
//
//   - "Becomes the target OF A SPELL", not "of a spell or ability".
//     EventBecomesTarget fires for both, so the condition looks the
//     source up on the stack and requires a SPELL item
//     (b40TargetedByASpell). Firing on abilities too would be a
//     strictly better Gargos, which is the one direction this catalog
//     does not ship (#259).
//   - It fires at ANNOUNCE (CR 601.2c), so the fight trigger goes on
//     the stack ABOVE the spell that targeted and resolves first.
//     That is exactly why the card is played: the removal spell often
//     fizzles because Gargos ate its caster's creature first, or the
//     targeted creature is still there and the fight happens anyway.
//   - "A creature you control" includes Gargos himself — he is a
//     creature its controller controls — so pointing a spell at
//     Gargos triggers him.
//   - "UP TO ONE target creature you don't control" is a 0-to-1 slot,
//     so the trigger is legal with no target and simply does nothing
//     when the board is empty of opposing creatures.
//
// Gargos not being on the battlefield when the trigger resolves means
// no fight, since there is nothing to deal the damage; that is CR
// 701.13a and the shared fight body handles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "378fcd0f-1096-4769-9149-55b4a889ff56",
		Name:            "Gargos, Vicious Watcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		CostModifiers: []game.CostModifier{
			CostsLess(4, "Gargos, Vicious Watcher: your Hydra spells cost {4} less", YourSpell(), b40HydraSpell()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b40CreatureYouControlTargetedByASpell(ev, source, g)
				}, "Gargos, Vicious Watcher — fight a creature you don't control", b30SourceFightsFirstLegalTarget),
				TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1),
			),
		},
	})
}
