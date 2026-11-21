package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost Jitte — Legendary Artifact — Equipment {1} (EDHREC rank 4459):
//
//	"Whenever equipped creature deals combat damage, put a charge
//	 counter on Lost Jitte.
//	 Remove a charge counter from Lost Jitte: Choose one —
//	 • Untap target land.
//	 • Target creature can't block this turn.
//	 • Put a +1/+1 counter on equipped creature.
//	 Equip {1}"
//
// Umezawa's Jitte's small cousin: one mana to cast, one to equip, and
// every connection banks a counter to spend on one of three things.
// The counters are the card — a creature that gets through twice can
// unblockably get through a third time and grow while doing it.
//
// A MODAL ACTIVATED ABILITY, which the engine could not express until
// #937 (ADR 0065). The mode and its target are announced together at
// activation (CR 602.2b), and the counter comes off as part of the
// COST — paid at announce (#625), so a response cannot spend the same
// counter twice and a Jitte destroyed in response has already paid.
//
// Three clauses, one per bullet, each with its own target clause
// declared on the OPTION rather than on the ability:
//
//   - "Untap target land" targets a land — ANY land, including an
//     opponent's, which is printed and is occasionally relevant.
//   - "Target creature can't block this turn" is a CR 509.1b
//     restriction, turn-scoped and pinned to the creature with its
//     battlefield-entry stamp, so a creature flickered in response is
//     a new object and blocks freely (CR 400.7, CR 611.2c).
//   - "Put a +1/+1 counter on equipped creature" does not target and
//     is offered whether or not the Jitte is attached; with nothing
//     equipped it simply does nothing, which is what the printed card
//     does.
//
// "Whenever equipped creature deals COMBAT DAMAGE" — to anything. A
// blocked attacker that kills its blocker banks a counter just as a
// connection to a player does, which is the difference between this
// trigger and the Swords' "to a player" one. It is one trigger per
// combat damage step (WheneverEquippedCreatureDealsCombatDamage,
// #5760): a trampler that hits its blocker and the player banks one
// counter, and a double striker banks one in each damage step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9d72f78-2ab5-4e2e-ab7b-ef875e0a0609",
		Name:         "Lost Jitte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverEquippedCreatureDealsCombatDamage("Lost Jitte — put a charge counter on it", effects.B33PutChargeCounterOnSelf),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a charge counter from Lost Jitte: Choose one — untap target land; or target creature can't block this turn; or put a +1/+1 counter on equipped creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
				Cost:    effects.RemoveCountersFromThis("charge", 1),
				Modes: effects.ChooseOne(
					effects.ModeDoing("Untap target land.",
						effects.TargetPermanent("target land", effects.Land()),
						effects.UntapTheModesTarget),
					effects.ModeDoing("Target creature can't block this turn.",
						effects.TargetCreature("target creature"),
						func(_ *game.StackItem, ctx *effects.Context, occ int) error {
							t, ok := effects.ModeTarget(ctx, occ)
							if !ok {
								return nil
							}
							return effects.RestrictUntilEOT{
								Target:       t.ID,
								Restrictions: game.CantBlock,
								Label:        "Lost Jitte — can't block this turn",
							}.Apply(ctx)
						}),
					effects.ModeDoing("Put a +1/+1 counter on equipped creature.",
						nil,
						func(item *game.StackItem, ctx *effects.Context, _ int) error {
							host, ok := effects.B43AttachedCreature(ctx.Game, item.SourceCardID)
							if !ok {
								return nil
							}
							return effects.AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}),
				),
			},
			effects.EquipAbility("{1}"),
		},
	})
}
