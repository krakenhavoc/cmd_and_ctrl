package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bilbo's Ring — Legendary Artifact — Equipment for {3}:
//
//	"During your turn, equipped creature has hexproof and can't be
//	 blocked.
//	 Whenever equipped creature attacks alone, you draw a card and you
//	 lose 1 life.
//	 Equip Halfling {1} ({1}: Attach to target Halfling you control.
//	 Equip only as a sorcery.)
//	 Equip {4} ({4}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Every clause on this card has a real shape:
//
//   - "During your turn" is a CONDITION on top of the ordinary
//     equipped-creature relation, not a duration of its own — the same
//     g.Turn read IsYourTurn already exposes for "Activate only during
//     your turn". GrantToAttachedWhile and RestrictAttachedWhile (the
//     restriction-bit sibling appended to restrictions.go for this
//     card) both re-read it on every recompute, so the hexproof and
//     the can't-be-blocked bit are only present while it is this
//     Equipment's controller's turn and fall away the instant it
//     isn't — with no bookkeeping and no "clear it at cleanup" step to
//     get wrong.
//   - The two Equip costs are two Spec.Activated entries. "Equip
//     Halfling" narrows the target clause to a Halfling creature the
//     controller controls; the plain Equip {4} keeps the wide "target
//     creature you control" clause EquipAbility already builds.
//   - "Attacks alone" is a real, checkable fact at the instant this
//     creature's EventAttack fires: DeclareAttackers stamps
//     AttackingTarget on every creature in the batch BEFORE
//     commitAttackDeclarationLocked announces any of their events (see
//     the note on attachedCreatureAttackedAlone below), so counting
//     battlefield cards with AttackingTarget set at that moment is
//     exactly "how many creatures were declared as attackers this
//     combat" — CR 508.3's "alone" ruling, not merely "no other
//     attacker is controlled by the same player".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7cc0f6b-6b17-4e76-b2a6-6a6ee2519b61",
		Name:         "Bilbo's Ring",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttachedWhile(duringSourceControllersTurn, "hexproof"),
			RestrictAttachedWhile(duringSourceControllersTurn, game.CantBeBlocked),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, attachedCreatureAttackedAlone,
				"Bilbo's Ring — draw a card, lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Halfling {1}", "{1}",
				TargetCreature("target Halfling you control", YouControl(), Subtype("Halfling"))),
			EquipAbility("{4}"),
		},
	})
}

// duringSourceControllersTurn is "during your turn" as a
// GrantToAttachedWhile / RestrictAttachedWhile condition — read fresh
// on every layer recompute against g.Turn, exactly as the
// activation-side IsYourTurn is.
func duringSourceControllersTurn(_ *game.Card, g *game.Game, source *game.Card) bool {
	return IsYourTurn(g, source.Controller)
}

// attachedCreatureAttackedAlone is Bilbo's Ring's "attacks alone"
// condition: the equipped creature was declared as an attacker, and
// it is the ONLY creature with AttackingTarget set at that instant.
//
// DeclareAttackers stamps AttackingTarget on every creature in the
// batch first and only THEN calls runStateChecksLocked, whose first
// act announces the whole declaration as one event batch — so every
// other attacker in the same declaration already carries its
// AttackingTarget by the time any one of their EventAttack events is
// evaluated here, and a lone declaration is genuinely alone rather
// than "alone so far".
func attachedCreatureAttackedAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attachedCreatureAttacked(ev, source) {
		return false
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n == 1
}
